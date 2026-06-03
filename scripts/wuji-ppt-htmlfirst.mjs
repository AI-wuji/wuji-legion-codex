#!/usr/bin/env node

import fs from "node:fs/promises";
import fsSync from "node:fs";
import path from "node:path";
import { createRequire } from "node:module";
import { fileURLToPath, pathToFileURL } from "node:url";

const DOM_TO_PPTX_VERSION = "1.1.10";
const DEFAULT_VIEWPORT = { width: 1920, height: 1080 };

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

function parseArgs(argv) {
  const args = {};
  for (let index = 0; index < argv.length; index += 1) {
    const key = argv[index];
    if (!key.startsWith("--")) {
      throw new Error(`Unexpected positional argument: ${key}`);
    }
    const value = argv[index + 1];
    if (!value || value.startsWith("--")) {
      args[key.slice(2)] = true;
      continue;
    }
    args[key.slice(2)] = value;
    index += 1;
  }
  return args;
}

function requireArg(args, key) {
  const value = args[key];
  if (typeof value !== "string" || value.length === 0) {
    throw new Error(`Missing required --${key}`);
  }
  return value;
}

function usage() {
  return [
    "Usage:",
    "  node scripts/wuji-ppt-htmlfirst.mjs --workspace <dir> --html <file> --out <pptx> [options]",
    "",
    "Options:",
    "  --title <text>   Deck title when HTML does not expose one.",
    "  --report <file>  Optional JSON report path.",
    "",
    "Exports HTML-first source into an editable PPTX using a browser-rendered dom-to-pptx pipeline.",
    "Prefers <section class=\"slide\"> or <section data-slide> as slide boundaries.",
  ].join("\n");
}

function decodeEntities(text) {
  return String(text)
    .replace(/&nbsp;/g, " ")
    .replace(/&amp;/g, "&")
    .replace(/&lt;/g, "<")
    .replace(/&gt;/g, ">")
    .replace(/&quot;/g, '"')
    .replace(/&#39;/g, "'");
}

function stripTags(html) {
  return decodeEntities(
    String(html)
      .replace(/<script[\s\S]*?<\/script>/gi, " ")
      .replace(/<style[\s\S]*?<\/style>/gi, " ")
      .replace(/<br\s*\/?>/gi, "\n")
      .replace(/<\/p>/gi, "\n")
      .replace(/<\/li>/gi, "\n")
      .replace(/<[^>]+>/g, " ")
      .replace(/[ \t]+\n/g, "\n")
      .replace(/\n[ \t]+/g, "\n")
      .replace(/\n{3,}/g, "\n\n")
      .replace(/[ \t]{2,}/g, " ")
      .trim(),
  );
}

function attrValue(tag, attrName) {
  const match = tag.match(new RegExp(`\\b${attrName}=["']([^"']+)["']`, "i"));
  return match ? decodeEntities(match[1]).trim() : "";
}

function firstMatch(source, pattern) {
  const match = source.match(pattern);
  return match ? decodeEntities(stripTags(match[1])).trim() : "";
}

function extractSections(html) {
  const sections = [...html.matchAll(/<section\b([^>]*)>([\s\S]*?)<\/section>/gi)];
  const slides = [];
  for (const match of sections) {
    const attrs = match[1] || "";
    const body = match[2] || "";
    const classes = attrValue(attrs, "class");
    const hasSlideClass = /\bslide\b/i.test(classes);
    const hasDataSlide = /\bdata-slide(?:=|>|\s|$)/i.test(attrs);
    if (hasSlideClass || hasDataSlide) {
      slides.push({ attrs, body });
    }
  }
  if (slides.length > 0) return slides;
  return [{ attrs: "", body: html }];
}

function extractImage(sectionHtml, htmlDir) {
  const match = sectionHtml.match(/<img\b([^>]*)>/i);
  if (!match) return undefined;
  const src = attrValue(match[1], "src");
  if (!src || /^(https?:|data:)/i.test(src)) return undefined;
  return path.resolve(htmlDir, src);
}

function detectAnimationSignals(source) {
  const checks = [
    { label: "css-animation", pattern: /\banimation\s*:/i },
    { label: "css-transition", pattern: /\btransition\s*:/i },
    { label: "css-keyframes", pattern: /@keyframes/i },
    { label: "animate-class", pattern: /\banimate[-_:a-z0-9]*\b/i },
    { label: "framer-motion", pattern: /\bframer-motion\b/i },
    { label: "gsap", pattern: /\bgsap\b/i },
    { label: "aos", pattern: /\bdata-aos\b|\baos[-_:a-z0-9]*\b/i },
  ];
  return checks.filter((item) => item.pattern.test(source)).map((item) => item.label);
}

function extractSlide(section, htmlDir, index) {
  const title =
    attrValue(section.attrs, "data-title") ||
    firstMatch(section.body, /<h1\b[^>]*>([\s\S]*?)<\/h1>/i) ||
    firstMatch(section.body, /<h2\b[^>]*>([\s\S]*?)<\/h2>/i) ||
    firstMatch(section.body, /<h3\b[^>]*>([\s\S]*?)<\/h3>/i) ||
    `Slide ${index + 1}`;

  const bullets = [...section.body.matchAll(/<li\b[^>]*>([\s\S]*?)<\/li>/gi)]
    .map((match) => stripTags(match[1]))
    .filter(Boolean);
  const paragraphs = [...section.body.matchAll(/<p\b[^>]*>([\s\S]*?)<\/p>/gi)]
    .map((match) => stripTags(match[1]))
    .filter(Boolean);
  const bodyText = stripTags(section.body);

  const summaryLines = bullets.length > 0 ? bullets : paragraphs;
  const fallback = summaryLines.length > 0 ? summaryLines : bodyText.split(/\n+/).map((line) => line.trim()).filter(Boolean);
  const body = fallback.slice(0, 6).join("\n");
  const imagePath = extractImage(section.body, htmlDir);

  return {
    index: index + 1,
    title,
    body,
    imagePath,
    rawTextLength: bodyText.length,
  };
}

function getNodeRoots() {
  return (process.env.NODE_PATH || "")
    .split(path.delimiter)
    .map((value) => value.trim())
    .filter(Boolean);
}

function packageNameToPnpmPrefix(packageName) {
  return `${packageName.replace(/\//g, "+")}@`;
}

function uniqueValues(values) {
  return [...new Set(values.filter(Boolean))];
}

function loadNodePackage(packageName) {
  const require = createRequire(import.meta.url);
  const roots = getNodeRoots();
  const candidates = [packageName];

  for (const root of roots) {
    candidates.push(path.join(root, packageName));
    candidates.push(path.join(root, ".pnpm", "node_modules", packageName));

    const pnpmRoot = path.join(root, ".pnpm");
    if (fsSync.existsSync(pnpmRoot)) {
      const prefix = packageNameToPnpmPrefix(packageName);
      for (const entry of fsSync.readdirSync(pnpmRoot, { withFileTypes: true })) {
        if (!entry.isDirectory() || !entry.name.startsWith(prefix)) {
          continue;
        }
        candidates.push(path.join(pnpmRoot, entry.name, "node_modules", packageName));
      }
    }
  }

  const errors = [];
  for (const candidate of uniqueValues(candidates)) {
    try {
      const mod = require(candidate);
      return mod.default || mod;
    } catch (error) {
      errors.push(`${candidate}: ${error.message}`);
    }
  }

  throw new Error(`Could not load ${packageName}.\n${errors.join("\n")}`);
}

function resolveRepoRoot() {
  return path.resolve(__dirname, "..");
}

async function ensureDomToPptxBundle(repoRoot) {
  const envBundle = process.env.WUJI_DOM_TO_PPTX_BUNDLE;
  if (envBundle && fsSync.existsSync(envBundle)) {
    return path.resolve(envBundle);
  }

  const cachedBundle = path.join(repoRoot, ".wuji-tools", "dom-to-pptx-cache", "package", "dist", "dom-to-pptx.bundle.js");
  if (fsSync.existsSync(cachedBundle)) {
    return cachedBundle;
  }

  const target = path.join(repoRoot, ".wuji-tools", "dom-to-pptx", DOM_TO_PPTX_VERSION, "dist", "dom-to-pptx.bundle.js");
  if (fsSync.existsSync(target)) {
    return target;
  }

  const url = `https://cdn.jsdelivr.net/npm/dom-to-pptx@${DOM_TO_PPTX_VERSION}/dist/dom-to-pptx.bundle.js`;
  const response = await fetch(url);
  if (!response.ok) {
    throw new Error(`Failed to download dom-to-pptx bundle from ${url}: ${response.status} ${response.statusText}`);
  }
  await fs.mkdir(path.dirname(target), { recursive: true });
  await fs.writeFile(target, Buffer.from(await response.arrayBuffer()));
  return target;
}

function resolveBrowserExecutable() {
  const explicit = process.env.WUJI_CHROMIUM_EXECUTABLE;
  if (explicit && fsSync.existsSync(explicit)) {
    return path.resolve(explicit);
  }

  const roots = uniqueValues([
    process.env.LOCALAPPDATA ? path.join(process.env.LOCALAPPDATA, "ms-playwright") : "",
    process.env.USERPROFILE ? path.join(process.env.USERPROFILE, "AppData", "Local", "ms-playwright") : "",
  ]);

  for (const root of roots) {
    if (!root || !fsSync.existsSync(root)) {
      continue;
    }

    const entries = fsSync.readdirSync(root, { withFileTypes: true }).filter((entry) => entry.isDirectory());
    const byNameDesc = [...entries].sort((left, right) => right.name.localeCompare(left.name));

    for (const entry of byNameDesc) {
      if (!entry.name.startsWith("chromium_headless_shell-")) {
        continue;
      }
      const candidate = path.join(root, entry.name, "chrome-headless-shell-win64", "chrome-headless-shell.exe");
      if (fsSync.existsSync(candidate)) {
        return candidate;
      }
    }

    for (const entry of byNameDesc) {
      if (!entry.name.startsWith("chromium-")) {
        continue;
      }
      const candidate = path.join(root, entry.name, "chrome-win64", "chrome.exe");
      if (fsSync.existsSync(candidate)) {
        return candidate;
      }
    }
  }

  const desktopBrowsers = [
    "C:/Program Files/Google/Chrome/Application/chrome.exe",
    "C:/Program Files (x86)/Google/Chrome/Application/chrome.exe",
    "C:/Program Files/Microsoft/Edge/Application/msedge.exe",
    "C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe",
  ];
  for (const candidate of desktopBrowsers) {
    if (fsSync.existsSync(candidate)) {
      return candidate;
    }
  }

  throw new Error(
    "No Chromium executable found for html-first export. Set WUJI_CHROMIUM_EXECUTABLE or install a local Playwright/Chrome browser.",
  );
}

function buildHtmlDocument(html, htmlPath) {
  const baseHref = pathToFileURL(`${path.dirname(htmlPath)}${path.sep}`).href;
  if (/<head\b[^>]*>/i.test(html)) {
    return html.replace(/<head\b[^>]*>/i, (match) => `${match}<base href="${baseHref}">`);
  }
  return `<!doctype html><html><head><base href="${baseHref}"></head><body>${html}</body></html>`;
}

async function waitForDocumentAssets(page) {
  await page.waitForLoadState("load");
  try {
    await page.waitForFunction(
      () => Array.from(document.images).every((img) => img.complete),
      { timeout: 15000 },
    );
  } catch {
    // Best-effort wait; missing or slow images should not block export forever.
  }
  await page.evaluate(async () => {
    if (document.fonts?.ready) {
      await document.fonts.ready;
    }
  });
}

async function freezeAnimations(page) {
  await page.addStyleTag({
    content: [
      "*, *::before, *::after {",
      "  transition: none !important;",
      "  animation-play-state: paused !important;",
      "}",
    ].join("\n"),
  });
  await page.evaluate(() => {
    if (!document.getAnimations) {
      return;
    }
    for (const animation of document.getAnimations()) {
      try {
        animation.pause();
        animation.currentTime = 0;
      } catch {
        // Ignore animations that do not support manual control.
      }
    }
  });
}

async function capturePilotPreview(page, workspace) {
  const previewPath = path.join(workspace, "htmlfirst-browser-preview.png");
  const slideHandle = await page.$("section.slide, section[data-slide]");
  if (slideHandle) {
    await slideHandle.screenshot({ path: previewPath });
    return previewPath;
  }
  await page.screenshot({ path: previewPath, fullPage: false });
  return previewPath;
}

async function exportDeckViaDomToPptx({ htmlDocument, outPath, bundlePath, title, animationSignals }) {
  const playwright = loadNodePackage("playwright");
  const { chromium } = playwright;
  const browserExecutable = resolveBrowserExecutable();
  const browser = await chromium.launch({
    headless: true,
    executablePath: browserExecutable,
  });

  try {
    const page = await browser.newPage({ viewport: DEFAULT_VIEWPORT });
    await page.setContent(htmlDocument, { waitUntil: "load" });
    await waitForDocumentAssets(page);
    if (animationSignals.length > 0) {
      await freezeAnimations(page);
    }
    const pilotPreviewPath = await capturePilotPreview(page, path.dirname(outPath));
    await page.addScriptTag({ path: bundlePath });
    await page.waitForTimeout(200);

    const exportResult = await page.evaluate(async ({ deckTitle }) => {
      const slides = Array.from(document.querySelectorAll("section.slide, section[data-slide]"));
      const targets = slides.length > 0 ? slides : [document.body];
      const fileName = deckTitle ? `${deckTitle}.pptx` : "html-first.pptx";
      const blob = await window.domToPptx.exportToPptx(targets, {
        fileName,
        skipDownload: true,
        svgAsVector: true,
      });
      const buffer = await blob.arrayBuffer();
      const bytes = new Uint8Array(buffer);
      let binary = "";
      const chunkSize = 0x8000;
      for (let offset = 0; offset < bytes.length; offset += chunkSize) {
        binary += String.fromCharCode(...bytes.subarray(offset, offset + chunkSize));
      }
      return {
        base64: btoa(binary),
        size: bytes.length,
        exported_slide_count: targets.length,
      };
    }, { deckTitle: title });

    await fs.mkdir(path.dirname(outPath), { recursive: true });
    await fs.writeFile(outPath, Buffer.from(exportResult.base64, "base64"));
    return {
      browserExecutable,
      exportedSlideCount: exportResult.exported_slide_count,
      outputBytes: exportResult.size,
      pilotPreviewPath,
    };
  } finally {
    await browser.close();
  }
}

async function main() {
  const args = parseArgs(process.argv.slice(2));
  if (args.help) {
    console.log(usage());
    return;
  }

  const workspace = path.resolve(requireArg(args, "workspace"));
  const htmlPath = path.resolve(requireArg(args, "html"));
  const outPath = path.resolve(requireArg(args, "out"));
  const reportPath = args.report ? path.resolve(args.report) : `${outPath}.json`;
  const explicitTitle = typeof args.title === "string" ? args.title.trim() : "";

  await fs.mkdir(workspace, { recursive: true });
  const html = await fs.readFile(htmlPath, "utf8");
  const htmlDir = path.dirname(htmlPath);
  const animationSignals = detectAnimationSignals(html);
  const sections = extractSections(html);
  const slides = sections.map((section, index) => extractSlide(section, htmlDir, index));
  const title =
    explicitTitle ||
    firstMatch(html, /<title\b[^>]*>([\s\S]*?)<\/title>/i) ||
    slides[0]?.title ||
    "HTML-first Deck";

  const repoRoot = resolveRepoRoot();
  const bundlePath = await ensureDomToPptxBundle(repoRoot);
  const htmlDocument = buildHtmlDocument(html, htmlPath);
  const exportMeta = await exportDeckViaDomToPptx({
    htmlDocument,
    outPath,
    bundlePath,
    title,
    animationSignals,
  });

  const report = {
    workspace,
    html: htmlPath,
    output: outPath,
    title,
    slide_count: Math.max(slides.length, exportMeta.exportedSlideCount || 0),
    slides,
    route: "html-first-editable-pptx",
    engine: "dom-to-pptx",
    engine_version: DOM_TO_PPTX_VERSION,
    renderer_mode: "browser-computed-style",
    editable_output: true,
    css_fidelity: "high",
    animation_signals: animationSignals,
    animation_transcoded: false,
    animation_snapshot_strategy: animationSignals.length > 0 ? "frozen-first-frame" : "none",
    browser_executable: exportMeta.browserExecutable,
    bundle_path: bundlePath,
    output_bytes: exportMeta.outputBytes,
    preview_path: exportMeta.pilotPreviewPath,
    warnings:
      animationSignals.length > 0
        ? [
            "Current HTML-first renderer preserves a static browser snapshot, but does not transcode HTML/CSS animation into PowerPoint animation.",
          ]
        : [],
  };

  await fs.mkdir(path.dirname(reportPath), { recursive: true });
  await fs.writeFile(reportPath, `${JSON.stringify(report, null, 2)}\n`, "utf8");
  console.log(reportPath);
}

main().catch((error) => {
  console.error(error.stack || error.message || String(error));
  console.error(usage());
  process.exit(1);
});
