#!/usr/bin/env node

import fs from "node:fs/promises";
import path from "node:path";

import { loadArtifactToolUtils, loadJSZip, relativeFromWorkspace } from "./wuji-ppt-node-utils.mjs";

function usage() {
  return [
    "Usage:",
    "  node scripts/wuji-ppt-template-inspect.mjs --workspace <dir> --pptx <source.pptx> [options]",
    "",
    "Options:",
    "  --out-dir <dir>   Output directory under workspace. Defaults to <workspace>/template-inspect.",
    "  --scale <n>       Render scale. Defaults to 1.",
    "  --slides <csv>    Only render preview/layout for selected 1-based slide numbers.",
    "  --no-preview      Skip preview PNG export and only produce inspect artifacts.",
    "  --no-layout       Skip layout JSON export and only produce inspect artifacts.",
    "",
    "Imports a source PPTX with artifact-tool, renders source slide PNGs/layouts,",
    "extracts package media, scans font names, writes template-inspect.ndjson,",
    "and writes template-manifest.json without spawning external unzip helpers.",
  ].join("\n");
}

function parseSlideSelection(value) {
  if (typeof value !== "string" || value.trim().length === 0) return null;
  const selected = new Set();
  for (const chunk of value.split(",")) {
    const parsed = Number.parseInt(chunk.trim(), 10);
    if (!Number.isInteger(parsed) || parsed < 1) {
      throw new Error(`Invalid --slides entry: ${chunk}`);
    }
    selected.add(parsed);
  }
  return selected.size > 0 ? selected : null;
}

function isWithin(child, parent) {
  const relative = path.relative(parent, child);
  return relative === "" || (!relative.startsWith("..") && !path.isAbsolute(relative));
}

async function writeJson(filePath, value) {
  await fs.mkdir(path.dirname(filePath), { recursive: true });
  await fs.writeFile(filePath, `${JSON.stringify(value, null, 2)}\n`, "utf8");
}

async function loadZipArchive(pptxPath) {
  const JSZip = loadJSZip();
  const bytes = await fs.readFile(pptxPath);
  return JSZip.loadAsync(bytes);
}

function zipNames(zip) {
  return Object.keys(zip.files).filter((name) => !zip.files[name].dir);
}

async function readZipText(zip, entryName) {
  const entry = zip.file(entryName);
  if (!entry) return "";
  return entry.async("string");
}

async function copyZipEntry(zip, entryName, targetPath) {
  const entry = zip.file(entryName);
  if (!entry) {
    throw new Error(`Zip entry not found: ${entryName}`);
  }
  await fs.mkdir(path.dirname(targetPath), { recursive: true });
  await fs.writeFile(targetPath, await entry.async("nodebuffer"));
}

async function collectFonts(zip, names) {
  const fonts = new Set();
  for (const name of names) {
    if (!/^ppt\/(?:slides|slideMasters|slideLayouts|theme)\/.*\.xml$/.test(name) && !/^ppt\/theme\/.*\.xml$/.test(name)) {
      continue;
    }
    const xml = await readZipText(zip, name);
    for (const match of xml.matchAll(/\btypeface="([^"]+)"/g)) {
      fonts.add(match[1]);
    }
  }
  return [...fonts].sort();
}

function slidesFromPresentation(presentation) {
  if (Array.isArray(presentation.slides?.items)) return presentation.slides.items;
  if (Number.isInteger(presentation.slides?.count) && typeof presentation.slides.getItem === "function") {
    return Array.from({ length: presentation.slides.count }, (_, index) => presentation.slides.getItem(index));
  }
  throw new Error("Could not enumerate imported presentation slides.");
}

function normalizeSlideNumbersInObject(value, slideMap) {
  if (!value || typeof value !== "object") return value;
  if (Array.isArray(value)) {
    return value.map((item) => normalizeSlideNumbersInObject(item, slideMap));
  }

  const output = {};
  for (const [key, item] of Object.entries(value)) {
    if ((key === "slide" || key === "slideNumber" || key === "page" || key === "pageNumber") && Number.isInteger(Number(item))) {
      const mapped = slideMap.get(Number(item));
      output[key] = mapped ?? item;
      continue;
    }
    output[key] = normalizeSlideNumbersInObject(item, slideMap);
  }
  return output;
}

function normalizeInspectNdjson(ndjson) {
  const lines = String(ndjson || "")
    .split(/\r?\n/)
    .filter((line) => line.trim());
  const records = lines.map((line) => JSON.parse(line));
  const slideNumbers = [...new Set(records.map((record) => Number(record.slide)).filter(Number.isInteger))].sort((a, b) => a - b);
  const slideMap = new Map(slideNumbers.map((slide, index) => [slide, index + 1]));
  return records
    .map((record) => JSON.stringify(normalizeSlideNumbersInObject(record, slideMap)))
    .join("\n")
    .concat(records.length > 0 ? "\n" : "");
}

async function main() {
  const argv = process.argv.slice(2);
  if (argv.includes("--help") || argv.includes("-h")) {
    console.log(usage());
    return;
  }

  const artifactToolUtils = await loadArtifactToolUtils();
  const {
    ensureArtifactToolWorkspace,
    importArtifactTool,
    parseArgs,
    requireArg,
    saveBlobToFile,
  } = artifactToolUtils;

  const args = parseArgs(argv);
  const workspaceDir = path.resolve(requireArg(args, "workspace"));
  const pptxPath = path.resolve(requireArg(args, "pptx"));
  const scale = args.scale ? Number.parseFloat(args.scale) : 1;
  const selectedSlides = parseSlideSelection(args.slides);
  const renderPreview = !args["no-preview"];
  const renderLayout = !args["no-layout"];
  if (!Number.isFinite(scale) || scale <= 0) {
    throw new Error("--scale must be a positive number");
  }

  const outDir = args["out-dir"]
    ? path.resolve(workspaceDir, args["out-dir"])
    : path.join(workspaceDir, "template-inspect");
  if (!isWithin(outDir, workspaceDir)) {
    throw new Error(`Refusing to write template inspection outside workspace: ${outDir}`);
  }
  if (path.resolve(outDir) === workspaceDir) {
    throw new Error(
      [
        `Refusing to use the workspace root as template inspection output: ${outDir}`,
        "Omit --out-dir or use a dedicated subdirectory such as --out-dir template-inspect.",
      ].join("\n"),
    );
  }

  const sourceStat = await fs.stat(pptxPath).catch(() => undefined);
  if (!sourceStat?.isFile()) {
    throw new Error(`Missing source PPTX: ${pptxPath}`);
  }

  await ensureArtifactToolWorkspace(workspaceDir);
  const { FileBlob, PresentationFile } = await importArtifactTool(workspaceDir);

  await fs.rm(outDir, { recursive: true, force: true });
  await fs.mkdir(outDir, { recursive: true });
  const slidesDir = path.join(outDir, "source-slides");
  const layoutsDir = path.join(outDir, "layouts");
  const mediaDir = path.join(outDir, "assets", "ppt", "media");
  const inspectPath = path.join(outDir, "template-inspect.ndjson");
  const manifestPath = path.join(outDir, "template-manifest.json");
  if (renderPreview) {
    await fs.mkdir(slidesDir, { recursive: true });
  }
  if (renderLayout) {
    await fs.mkdir(layoutsDir, { recursive: true });
  }

  const presentation = await PresentationFile.importPptx(await FileBlob.load(pptxPath));
  const slides = slidesFromPresentation(presentation);
  const zip = await loadZipArchive(pptxPath);
  const names = zipNames(zip);
  const media = names.filter((name) => name.startsWith("ppt/media/"));
  const slideXmlNames = names.filter((name) => /^ppt\/slides\/slide\d+\.xml$/.test(name));
  const chartNames = names.filter((name) => /^ppt\/(?:charts|embeddings\/charts)\/chart\d+\.xml$/.test(name));

  const slideArtifacts = [];
  for (let index = 0; index < slides.length; index += 1) {
    const slide = slides[index];
    const slideNumber = index + 1;
    if (selectedSlides && !selectedSlides.has(slideNumber)) {
      continue;
    }
    const padded = String(slideNumber).padStart(2, "0");
    const pngPath = path.join(slidesDir, `source-slide-${padded}.png`);
    const layoutPath = path.join(layoutsDir, `source-slide-${padded}.layout.json`);

    let previewRelativePath = undefined;
    if (renderPreview) {
      const preview = await presentation.export({ slide, format: "png", scale });
      await saveBlobToFile(preview, pngPath);
      previewRelativePath = relativeFromWorkspace(workspaceDir, pngPath);
    }

    let layoutRelativePath = undefined;
    if (renderLayout) {
      const layout = await presentation.export({ slide, format: "layout" });
      await saveBlobToFile(layout, layoutPath);
      layoutRelativePath = relativeFromWorkspace(workspaceDir, layoutPath);
    }

    slideArtifacts.push({
      slide: slideNumber,
      previewPath: renderPreview ? pngPath : undefined,
      previewRelativePath,
      layoutPath: renderLayout ? layoutPath : undefined,
      layoutRelativePath,
    });
  }

  const extractedMedia = [];
  for (const entry of media) {
    const target = path.join(mediaDir, path.basename(entry));
    await copyZipEntry(zip, entry, target);
    const stat = await fs.stat(target);
    extractedMedia.push({
      entry,
      path: target,
      relativePath: relativeFromWorkspace(workspaceDir, target),
      bytes: stat.size,
    });
  }

  const inspect = await presentation.inspect({
    kind: "slide,textbox,shape,image,table,chart",
    max_chars: 200000,
  });
  await fs.writeFile(inspectPath, normalizeInspectNdjson(inspect.ndjson), "utf8");

  let tableSlideCount = 0;
  for (const name of slideXmlNames) {
    if ((await readZipText(zip, name)).includes("<a:tbl>")) {
      tableSlideCount += 1;
    }
  }

  const manifest = {
    sourcePptx: pptxPath,
    workspace: workspaceDir,
    outDir,
    generatedAt: new Date().toISOString(),
    slideCount: slides.length,
    renderedSlideCount: slideArtifacts.length,
    renderPreview,
    renderLayout,
    selectedSlides: selectedSlides ? [...selectedSlides] : "all",
    slideArtifacts,
    inspectPath,
    inspectRelativePath: relativeFromWorkspace(workspaceDir, inspectPath),
    inspectTruncated: Boolean(inspect.truncated),
    inspectMetadata: inspect.metadata || {},
    extractedMedia,
    fonts: await collectFonts(zip, names),
    packageParts: {
      mediaCount: media.length,
      slideXmlCount: slideXmlNames.length,
      chartCount: chartNames.length,
      tableSlideCount,
    },
  };
  await writeJson(manifestPath, manifest);
  console.log(manifestPath);
}

main().catch((error) => {
  console.error(error.stack || error.message || String(error));
  console.error(usage());
  process.exit(1);
});
