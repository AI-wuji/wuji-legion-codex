#!/usr/bin/env node

import fs from "node:fs/promises";
import path from "node:path";
import { createRequire } from "node:module";

function usage() {
  return [
    "Usage:",
    "  node scripts/wuji-ppt-native-interactive-demo.mjs --out <file.pptx> [options]",
    "",
    "Options:",
    "  --title <text>     Deck title. Defaults to '真交互 PPTX Demo'.",
    "  --theme <name>     Visual theme. Defaults to 'obsidian'.",
    "",
    "Creates a polished editable PowerPoint deck with native hyperlinks, custom buttons,",
    "and slide-to-slide navigation that remains editable in PowerPoint.",
  ].join("\n");
}

function parseArgs(argv) {
  const args = {};
  for (let i = 0; i < argv.length; i += 1) {
    const token = argv[i];
    if (!token.startsWith("--")) continue;
    const key = token.slice(2);
    const next = argv[i + 1];
    if (typeof next === "string" && !next.startsWith("--")) {
      args[key] = next;
      i += 1;
    } else {
      args[key] = "true";
    }
  }
  return args;
}

function themePalette(name) {
  const theme = String(name || "").toLowerCase();
  if (theme === "light" || theme === "paper") {
    return {
      name: "paper",
      bg: "F5F8FC",
      panel: "FFFFFF",
      panelSoft: "EDF2FB",
      text: "0F172A",
      muted: "52627A",
      line: "D7DEEA",
      accent: "0F62FE",
      accent2: "0E9F6E",
      accent3: "7C3AED",
      accent4: "C2410C",
      accent5: "EA580C",
      chipBg: "EAF1FF",
      glow: "9BC0FF",
    };
  }

  return {
    name: "obsidian",
    bg: "040B16",
    panel: "0B1324",
    panelSoft: "111F35",
    text: "F8FBFF",
    muted: "AEBAD0",
    line: "21324B",
    accent: "66D9FF",
    accent2: "2DE2C5",
    accent3: "C084FC",
    accent4: "FBBF24",
    accent5: "FF7A59",
    chipBg: "102033",
    glow: "7DCBFF",
  };
}

function buttonStyle(theme, accent) {
  return {
    fill: { color: theme.panel, transparency: 0 },
    line: { color: accent, pt: 1.1 },
    shadow: { type: "outer", color: accent, opacity: 0.16, blur: 4, offset: 1, angle: 45 },
  };
}

function addBlob(slide, x, y, w, h, color, transparency) {
  slide.addShape("ellipse", {
    x,
    y,
    w,
    h,
    fill: { color, transparency },
    line: { color, transparency: 100, pt: 0 },
  });
}

function addBackdrop(slide, theme) {
  addBlob(slide, -1.6, -1.1, 4.1, 4.1, theme.accent, 90);
  addBlob(slide, 10.2, -0.7, 3.7, 3.7, theme.accent3, 91);
  addBlob(slide, 8.4, 4.8, 4.9, 4.9, theme.accent2, 93);
  addBlob(slide, 0.4, 5.0, 3.0, 3.0, theme.accent4, 94);
  addBlob(slide, 6.0, 1.5, 3.4, 3.4, theme.accent5, 96);

  slide.addShape("line", {
    x: 0.62,
    y: 0.82,
    w: 12.04,
    h: 0,
    line: { color: theme.line, pt: 1 },
  });
  slide.addShape("line", {
    x: 0.62,
    y: 6.72,
    w: 12.04,
    h: 0,
    line: { color: theme.line, pt: 0.9, transparency: 20 },
  });
  slide.addShape("line", {
    x: 0.62,
    y: 1.02,
    w: 0,
    h: 5.28,
    line: { color: theme.accent, pt: 1.05, transparency: 38 },
  });
}

function addHeader(slide, theme, eyebrow, title, subtitle) {
  slide.addText(eyebrow, {
    x: 0.72,
    y: 0.18,
    w: 2.6,
    h: 0.16,
    fontFace: "Aptos",
    fontSize: 8.8,
    bold: true,
    color: theme.accent,
    charSpace: 1.6,
  });
  slide.addText(title, {
    x: 0.72,
    y: 0.48,
    w: 8.6,
    h: 0.52,
    fontFace: "Aptos Display",
    fontSize: 24.5,
    bold: true,
    color: theme.text,
  });
  slide.addText(subtitle, {
    x: 0.74,
    y: 0.98,
    w: 10.6,
    h: 0.24,
    fontFace: "Microsoft YaHei",
    fontSize: 11.1,
    color: theme.muted,
  });
}

function addChip(slide, theme, x, y, w, label, accent) {
  slide.addShape("roundRect", {
    x,
    y,
    w,
    h: 0.28,
    rectRadius: 0.12,
    fill: { color: accent, transparency: 85 },
    line: { color: accent, pt: 0.8 },
  });
  slide.addText(label, {
    x: x + 0.06,
    y: y + 0.04,
    w: w - 0.12,
    h: 0.12,
    fontFace: "Aptos",
    fontSize: 8.1,
    bold: true,
    color: accent,
    align: "center",
  });
}

function addPillButton(slide, theme, cfg) {
  slide.addShape("roundRect", {
    x: cfg.x,
    y: cfg.y,
    w: cfg.w,
    h: cfg.h ?? 0.42,
    rectRadius: 0.17,
    fill: { color: cfg.fill ?? theme.panel, transparency: cfg.transparency ?? 0 },
    line: { color: cfg.line ?? cfg.accent ?? theme.accent, pt: 1.05 },
    shadow: { type: "outer", color: cfg.line ?? cfg.accent ?? theme.accent, opacity: 0.12, blur: 4, offset: 1, angle: 45 },
    hyperlink: cfg.link,
  });
  slide.addText(cfg.label, {
    x: cfg.x + 0.09,
    y: cfg.y + 0.075,
    w: cfg.w - 0.18,
    h: (cfg.h ?? 0.42) - 0.15,
    fontFace: "Microsoft YaHei",
    fontSize: cfg.fontSize ?? 10.1,
    bold: true,
    color: cfg.text ?? theme.text,
    align: "center",
    valign: "mid",
    hyperlink: cfg.link,
  });
}

function addMetricCard(slide, theme, cfg) {
  slide.addShape("roundRect", {
    x: cfg.x,
    y: cfg.y,
    w: cfg.w,
    h: cfg.h,
    rectRadius: 0.06,
    fill: { color: cfg.fill ?? theme.panel, transparency: 0 },
    line: { color: cfg.line ?? theme.line, pt: 1 },
    shadow: { type: "outer", color: cfg.line ?? theme.line, opacity: 0.12, blur: 5, offset: 1, angle: 45 },
    hyperlink: cfg.link,
  });
  slide.addShape("roundRect", {
    x: cfg.x + 0.1,
    y: cfg.y + 0.1,
    w: cfg.w - 0.2,
    h: 0.13,
    rectRadius: 0.06,
    fill: { color: cfg.accent ?? theme.accent, transparency: 82 },
    line: { color: cfg.accent ?? theme.accent, transparency: 100, pt: 0 },
  });
  slide.addText(cfg.kicker, {
    x: cfg.x + 0.18,
    y: cfg.y + 0.28,
    w: cfg.w - 0.36,
    h: 0.16,
    fontFace: "Aptos",
    fontSize: 8.6,
    bold: true,
    color: cfg.accent ?? theme.accent,
    charSpace: 1.2,
  });
  slide.addText(cfg.title, {
    x: cfg.x + 0.18,
    y: cfg.y + 0.56,
    w: cfg.w - 0.36,
    h: 0.28,
    fontFace: "Aptos Display",
    fontSize: cfg.titleSize ?? 16.4,
    bold: true,
    color: theme.text,
  });
  slide.addText(cfg.body, {
    x: cfg.x + 0.18,
    y: cfg.y + 0.98,
    w: cfg.w - 0.36,
    h: cfg.h - 1.32,
    fontFace: "Microsoft YaHei",
    fontSize: cfg.bodySize ?? 10.4,
    color: theme.muted,
    valign: "mid",
  });
  if (cfg.linkLabel) {
    slide.addText(cfg.linkLabel, {
      x: cfg.x + 0.18,
      y: cfg.y + cfg.h - 0.3,
      w: 1.56,
      h: 0.12,
      fontFace: "Aptos",
      fontSize: 8.5,
      bold: true,
      color: cfg.accent ?? theme.accent,
      hyperlink: cfg.link,
    });
  }
}

function addNav(slide, theme, opts = {}) {
  const y = 6.34;
  addPillButton(slide, theme, {
    x: 10.62,
    y,
    w: 0.72,
    h: 0.36,
    label: "首页",
    accent: theme.accent,
    text: theme.text,
    link: { slide: 1, tooltip: "返回首页" },
  });
  addPillButton(slide, theme, {
    x: 11.42,
    y,
    w: 0.72,
    h: 0.36,
    label: "目录",
    accent: theme.accent2,
    text: theme.text,
    link: { slide: opts.directorySlide ?? 2, tooltip: "跳转到目录" },
  });
  addPillButton(slide, theme, {
    x: 12.22,
    y,
    w: 0.72,
    h: 0.36,
    label: "下一页",
    accent: theme.accent3,
    text: theme.text,
    link: { slide: opts.nextSlide ?? 1, tooltip: "前往下一页" },
  });
}

function addFlowLine(slide, theme, x1, y1, x2, y2, accent) {
  slide.addShape("line", {
    x: x1,
    y: y1,
    w: x2 - x1,
    h: y2 - y1,
    line: { color: accent, pt: 1.15, transparency: 18, beginArrowType: "none", endArrowType: "triangle" },
  });
}

function addMiniNode(slide, x, y, color, label, theme) {
  slide.addShape("ellipse", {
    x,
    y,
    w: 0.38,
    h: 0.38,
    fill: { color, transparency: 0 },
    line: { color, pt: 0.8 },
  });
  slide.addText(label, {
    x: x - 0.16,
    y: y + 0.46,
    w: 0.7,
    h: 0.12,
    fontFace: "Aptos",
    fontSize: 8.2,
    bold: true,
    color: theme.muted,
    align: "center",
  });
}

async function stripNotesParts(pptxPath) {
  const require = createRequire(import.meta.url);
  const JSZip = require("C:/Users/Administrator/.cache/codex-runtimes/codex-primary-runtime/dependencies/node/node_modules/.pnpm/jszip@3.10.1/node_modules/jszip");
  const bytes = await fs.readFile(pptxPath);
  const zip = await JSZip.loadAsync(bytes);

  for (const name of Object.keys(zip.files)) {
    if (name.startsWith("ppt/notesMasters/") || name.startsWith("ppt/notesSlides/")) {
      zip.remove(name);
    }
  }

  const replaceXml = async (entryName, transform) => {
    const entry = zip.file(entryName);
    if (!entry) return;
    const xml = await entry.async("string");
    zip.file(entryName, transform(xml));
  };

  await replaceXml("[Content_Types].xml", (xml) =>
    xml
      .replace(/<Override PartName="\/ppt\/notesMasters\/notesMaster1\.xml"[^>]*\/>/g, "")
      .replace(/<Override PartName="\/ppt\/notesSlides\/notesSlide\d+\.xml"[^>]*\/>/g, "")
      .replace(/\s+<\/Types>/, "</Types>"),
  );
  await replaceXml("ppt/presentation.xml", (xml) =>
    xml.replace(/<p:notesMasterIdLst>[\s\S]*?<\/p:notesMasterIdLst>/g, ""),
  );
  await replaceXml("ppt/_rels/presentation.xml.rels", (xml) =>
    xml.replace(
      /<Relationship Id="rId\d+" Type="http:\/\/schemas\.openxmlformats\.org\/officeDocument\/2006\/relationships\/notesMaster" Target="notesMasters\/notesMaster1\.xml"\/>/g,
      "",
    ),
  );
  for (const name of Object.keys(zip.files).filter((entry) => /^ppt\/slides\/_rels\/slide\d+\.xml\.rels$/.test(entry))) {
    await replaceXml(name, (xml) =>
      xml.replace(
        /<Relationship Id="rId\d+" Type="http:\/\/schemas\.openxmlformats\.org\/officeDocument\/2006\/relationships\/notesSlide" Target="\.\.\/notesSlides\/notesSlide\d+\.xml"\/>/g,
        "",
      ),
    );
  }

  await fs.writeFile(pptxPath, await zip.generateAsync({ type: "nodebuffer", compression: "DEFLATE" }));
}

async function main() {
  const args = parseArgs(process.argv.slice(2));
  if (args.help || args.h) {
    console.log(usage());
    return;
  }

  const out = args.out ? path.resolve(args.out) : "";
  if (!out) {
    throw new Error("Missing --out <file.pptx>.");
  }

  const title = args.title || "真交互 PPTX Demo";
  const theme = themePalette(args.theme || "obsidian");

  const require = createRequire(import.meta.url);
  const PptxGenJS = require("C:/Users/Administrator/.cache/codex-runtimes/codex-primary-runtime/dependencies/node/node_modules/.pnpm/pptxgenjs@4.0.1/node_modules/pptxgenjs");

  const pptx = new PptxGenJS();
  pptx.layout = "LAYOUT_WIDE";
  pptx.author = "Codex";
  pptx.company = "WuJi Legion";
  pptx.subject = "Native interactive PowerPoint";
  pptx.title = title;
  pptx.lang = "zh-CN";
  pptx.theme = {
    headFontFace: "Aptos Display",
    bodyFontFace: "Microsoft YaHei",
    lang: "zh-CN",
  };

  pptx.defineSlideMaster({
    title: "MASTER",
    background: { color: theme.bg },
    objects: [
      {
        shape: "line",
        x: 0.62,
        y: 0.82,
        w: 12.04,
        h: 0,
        line: { color: theme.line, pt: 1 },
      },
      {
        shape: "line",
        x: 0.62,
        y: 6.72,
        w: 12.04,
        h: 0,
        line: { color: theme.line, pt: 0.9, transparency: 20 },
      },
    ],
    slideNumber: { x: 12.86, y: 7.02, color: theme.muted, fontSize: 10 },
  });

  // Slide 1: Hero
  const slide1 = pptx.addSlide("MASTER");
  addBackdrop(slide1, theme);
  addHeader(
    slide1,
    theme,
    "NATIVE INTERACTION / 真交互 PPTX",
    "把按钮、跳转、回流做成真正可编辑的 PowerPoint",
    "不是 HTML 截图，不是整页图片。文字、形状、链接都保留在 .pptx 里，打开后还能继续改。"
  );
  addChip(slide1, theme, 10.42, 0.2, 2.22, "Editable · Native links · Return path", theme.accent);

  slide1.addText("真交互", {
    x: 0.82,
    y: 1.55,
    w: 2.24,
    h: 0.42,
    fontFace: "Aptos",
    fontSize: 16,
    bold: true,
    color: theme.accent,
  });
  slide1.addText("不是静态壳，而是能点的母版", {
    x: 0.82,
    y: 1.95,
    w: 5.6,
    h: 0.88,
    fontFace: "Aptos Display",
    fontSize: 24,
    bold: true,
    color: theme.text,
  });
  slide1.addText("这套 demo 的目标很简单：让观众点击以后真的发生变化，而且变化还留在 PowerPoint 里，不把可编辑性牺牲掉。", {
    x: 0.82,
    y: 2.9,
    w: 5.7,
    h: 0.72,
    fontFace: "Microsoft YaHei",
    fontSize: 11.6,
    color: theme.muted,
  });

  addPillButton(slide1, theme, {
    x: 0.82,
    y: 3.94,
    w: 1.16,
    h: 0.42,
    label: "看目录",
    accent: theme.accent,
    text: theme.text,
    link: { slide: 2, tooltip: "跳到目录页" },
  });
  addPillButton(slide1, theme, {
    x: 2.08,
    y: 3.94,
    w: 1.16,
    h: 0.42,
    label: "看分支",
    accent: theme.accent2,
    text: theme.text,
    link: { slide: 3, tooltip: "跳到分支页" },
  });
  addPillButton(slide1, theme, {
    x: 3.34,
    y: 3.94,
    w: 1.26,
    h: 0.42,
    label: "打开资料",
    accent: theme.accent3,
    text: theme.text,
    link: { url: "https://github.com/gitbrent/pptxgenjs", tooltip: "打开 PptxGenJS GitHub" },
  });
  addPillButton(slide1, theme, {
    x: 4.7,
    y: 3.94,
    w: 1.16,
    h: 0.42,
    label: "看收口",
    accent: theme.accent4,
    text: theme.text,
    link: { slide: 4, tooltip: "跳到收口页" },
  });

  slide1.addShape("roundRect", {
    x: 6.8,
    y: 1.48,
    w: 5.36,
    h: 4.72,
    rectRadius: 0.06,
    fill: { color: theme.panel, transparency: 0 },
    line: { color: theme.accent, pt: 1.1 },
    shadow: { type: "outer", color: theme.accent, opacity: 0.14, blur: 6, offset: 1, angle: 45 },
  });
  addChip(slide1, theme, 7.08, 1.8, 1.18, "PROOF", theme.accent);
  slide1.addText("三种互动，全都留在 PPT 里", {
    x: 7.08,
    y: 2.18,
    w: 4.5,
    h: 0.28,
    fontFace: "Aptos Display",
    fontSize: 21,
    bold: true,
    color: theme.text,
  });
  slide1.addText("目录跳转 / 状态分支 / 外部链接", {
    x: 7.08,
    y: 2.58,
    w: 4.4,
    h: 0.18,
    fontFace: "Microsoft YaHei",
    fontSize: 11.3,
    color: theme.muted,
  });

  slide1.addShape("ellipse", {
    x: 8.62,
    y: 3.0,
    w: 1.54,
    h: 1.54,
    fill: { color: theme.accent, transparency: 5 },
    line: { color: theme.accent, pt: 1.2 },
    shadow: { type: "outer", color: theme.accent, opacity: 0.22, blur: 6, offset: 1, angle: 45 },
  });
  slide1.addText("PPTX", {
    x: 8.97,
    y: 3.43,
    w: 0.88,
    h: 0.18,
    fontFace: "Aptos",
    fontSize: 15,
    bold: true,
    color: "FFFFFF",
    align: "center",
  });
  slide1.addText("可编辑", {
    x: 8.72,
    y: 3.8,
    w: 1.42,
    h: 0.12,
    fontFace: "Aptos",
    fontSize: 9,
    bold: true,
    color: theme.text,
    align: "center",
  });

  addFlowLine(slide1, theme, 7.18, 4.6, 8.54, 4.6, theme.accent2);
  addFlowLine(slide1, theme, 10.26, 4.6, 11.36, 4.6, theme.accent3);
  slide1.addShape("ellipse", {
    x: 7.03,
    y: 4.28,
    w: 0.34,
    h: 0.34,
    fill: { color: theme.accent2, transparency: 0 },
    line: { color: theme.accent2, pt: 0.8 },
  });
  slide1.addShape("ellipse", {
    x: 8.48,
    y: 4.28,
    w: 0.34,
    h: 0.34,
    fill: { color: theme.accent3, transparency: 0 },
    line: { color: theme.accent3, pt: 0.8 },
  });
  slide1.addShape("ellipse", {
    x: 10.12,
    y: 4.28,
    w: 0.34,
    h: 0.34,
    fill: { color: theme.accent4, transparency: 0 },
    line: { color: theme.accent4, pt: 0.8 },
  });

  addMetricCard(slide1, theme, {
    x: 7.08,
    y: 4.82,
    w: 1.56,
    h: 1.02,
    accent: theme.accent2,
    kicker: "JUMP",
    title: "目录跳转",
    body: "一键前往任意章节。"
  });
  addMetricCard(slide1, theme, {
    x: 8.88,
    y: 4.82,
    w: 1.56,
    h: 1.02,
    accent: theme.accent3,
    kicker: "SWITCH",
    title: "状态切换",
    body: "同一主题可分支。"
  });
  addMetricCard(slide1, theme, {
    x: 10.68,
    y: 4.82,
    w: 1.56,
    h: 1.02,
    accent: theme.accent4,
    kicker: "URL",
    title: "外链打开",
    body: "资源页直接连出去。",
  });

  slide1.addText("不是做出“看起来像互动”", {
    x: 7.08,
    y: 6.0,
    w: 4.8,
    h: 0.14,
    fontFace: "Aptos",
    fontSize: 10.2,
    color: theme.muted,
  });
  addNav(slide1, theme, { directorySlide: 2, nextSlide: 2 });

  // Slide 2: Directory / control center
  const slide2 = pptx.addSlide("MASTER");
  addBackdrop(slide2, theme);
  addHeader(
    slide2,
    theme,
    "DIRECTORY / CONTROL CENTER",
    "目录不是列表，而是可点击的控制台",
    "这一页把不同交互拆成三条可视化路径，演示“点完会变”的 PPT 结构。"
  );
  addChip(slide2, theme, 10.46, 0.2, 2.04, "Route map · Buttons · Return path", theme.accent2);

  slide2.addShape("roundRect", {
    x: 0.74,
    y: 1.48,
    w: 7.1,
    h: 4.88,
    rectRadius: 0.06,
    fill: { color: theme.panel, transparency: 0 },
    line: { color: theme.accent2, pt: 1.1 },
  });
  addChip(slide2, theme, 1.0, 1.8, 1.46, "PAGE 02 / 目录", theme.accent2);
  slide2.addText("每个模块都可以直接跳转", {
    x: 1.0,
    y: 2.18,
    w: 5.8,
    h: 0.28,
    fontFace: "Aptos Display",
    fontSize: 22,
    bold: true,
    color: theme.text,
  });
  slide2.addText("这页是一个按钮控制台：左边是入口，右边是路径图。所有按钮都是真链接，继续编辑时它们还在。", {
    x: 1.0,
    y: 2.64,
    w: 6.0,
    h: 0.62,
    fontFace: "Microsoft YaHei",
    fontSize: 11.6,
    color: theme.muted,
  });

  addMetricCard(slide2, theme, {
    x: 1.0,
    y: 3.52,
    w: 6.38,
    h: 0.82,
    accent: theme.accent,
    kicker: "01",
    title: "目录跳转",
    body: "观众点完后，直接进入某一章节页，不需要再翻。"
  });
  addMetricCard(slide2, theme, {
    x: 1.0,
    y: 4.46,
    w: 6.38,
    h: 0.82,
    accent: theme.accent3,
    kicker: "02",
    title: "状态分支",
    body: "同一页的按钮可以导向不同结果，做“如果...就...”的演示。"
  });
  addMetricCard(slide2, theme, {
    x: 1.0,
    y: 5.4,
    w: 6.38,
    h: 0.82,
    accent: theme.accent4,
    kicker: "03",
    title: "外部链接",
    body: "直接连到 GitHub / 文档 / 资源页，适合产品演示和培训。"
  });

  slide2.addShape("roundRect", {
    x: 8.06,
    y: 1.48,
    w: 4.62,
    h: 2.34,
    rectRadius: 0.06,
    fill: { color: theme.panelSoft, transparency: 0 },
    line: { color: theme.accent, pt: 1.05 },
  });
  addChip(slide2, theme, 8.3, 1.78, 1.3, "NAV MAP", theme.accent);
  slide2.addText("按钮不是装饰，是路由", {
    x: 8.3,
    y: 2.2,
    w: 3.5,
    h: 0.22,
    fontFace: "Aptos Display",
    fontSize: 18.5,
    bold: true,
    color: theme.text,
  });
  slide2.addText("你能把每一个卡片都当成一个触发器：点击之后跳到别页、别状态，或者外部网站。", {
    x: 8.3,
    y: 2.58,
    w: 3.9,
    h: 0.56,
    fontFace: "Microsoft YaHei",
    fontSize: 10.8,
    color: theme.muted,
  });
  addPillButton(slide2, theme, {
    x: 8.3,
    y: 3.28,
    w: 1.04,
    label: "主页",
    accent: theme.accent,
    text: theme.text,
    link: { slide: 1, tooltip: "返回首页" },
  });
  addPillButton(slide2, theme, {
    x: 9.44,
    y: 3.28,
    w: 1.04,
    label: "分支",
    accent: theme.accent2,
    text: theme.text,
    link: { slide: 3, tooltip: "进入分支页" },
  });
  addPillButton(slide2, theme, {
    x: 10.58,
    y: 3.28,
    w: 1.04,
    label: "收口",
    accent: theme.accent3,
    text: theme.text,
    link: { slide: 4, tooltip: "进入收口页" },
  });

  slide2.addShape("roundRect", {
    x: 8.06,
    y: 4.0,
    w: 4.62,
    h: 2.36,
    rectRadius: 0.06,
    fill: { color: theme.panel, transparency: 0 },
    line: { color: theme.accent2, pt: 1.05 },
  });
  addChip(slide2, theme, 8.3, 4.3, 1.22, "FLOW", theme.accent2);
  addMiniNode(slide2, 8.48, 5.08, theme.accent, "A", theme);
  addMiniNode(slide2, 9.62, 5.08, theme.accent2, "B", theme);
  addMiniNode(slide2, 10.76, 5.08, theme.accent3, "C", theme);
  addFlowLine(slide2, theme, 8.86, 5.27, 9.46, 5.27, theme.accent);
  addFlowLine(slide2, theme, 10.0, 5.27, 10.6, 5.27, theme.accent2);
  slide2.addText("A = 首页", {
    x: 8.26,
    y: 5.74,
    w: 0.78,
    h: 0.12,
    fontFace: "Aptos",
    fontSize: 8.4,
    color: theme.muted,
    align: "center",
  });
  slide2.addText("B = 分支", {
    x: 9.4,
    y: 5.74,
    w: 0.78,
    h: 0.12,
    fontFace: "Aptos",
    fontSize: 8.4,
    color: theme.muted,
    align: "center",
  });
  slide2.addText("C = 收口", {
    x: 10.54,
    y: 5.74,
    w: 0.78,
    h: 0.12,
    fontFace: "Aptos",
    fontSize: 8.4,
    color: theme.muted,
    align: "center",
  });
  addNav(slide2, theme, { directorySlide: 2, nextSlide: 3 });

  // Slide 3: State switch
  const slide3 = pptx.addSlide("MASTER");
  addBackdrop(slide3, theme);
  addHeader(
    slide3,
    theme,
    "BRANCH / STATE SWITCH",
    "同一页的不同状态，点一下就变成另一个结果",
    "这页演示真正的交互感：左边是状态 A，右边是状态 B，按钮可以带观众切换路径。"
  );
  addChip(slide3, theme, 10.18, 0.2, 2.34, "Branching state · Same slide family", theme.accent3);

  slide3.addShape("roundRect", {
    x: 0.74,
    y: 1.5,
    w: 5.9,
    h: 4.86,
    rectRadius: 0.06,
    fill: { color: theme.panel, transparency: 0 },
    line: { color: theme.accent3, pt: 1.08 },
  });
  slide3.addShape("roundRect", {
    x: 6.84,
    y: 1.5,
    w: 5.9,
    h: 4.86,
    rectRadius: 0.06,
    fill: { color: theme.panelSoft, transparency: 0 },
    line: { color: theme.accent2, pt: 1.08 },
  });
  addChip(slide3, theme, 0.98, 1.8, 1.16, "STATE A", theme.accent3);
  addChip(slide3, theme, 7.08, 1.8, 1.16, "STATE B", theme.accent2);

  slide3.addText("状态 A / 聚焦内容", {
    x: 0.98,
    y: 2.18,
    w: 3.2,
    h: 0.24,
    fontFace: "Aptos Display",
    fontSize: 20.5,
    bold: true,
    color: theme.text,
  });
  slide3.addText("左边更像“内容视图”", {
    x: 0.98,
    y: 2.54,
    w: 3.2,
    h: 0.12,
    fontFace: "Microsoft YaHei",
    fontSize: 11,
    color: theme.muted,
  });
  addMetricCard(slide3, theme, {
    x: 0.98,
    y: 2.9,
    w: 2.42,
    h: 1.0,
    accent: theme.accent3,
    kicker: "01",
    title: "更大标题",
    body: "适合讲观点、放结论、做冲击。"
  });
  addMetricCard(slide3, theme, {
    x: 3.56,
    y: 2.9,
    w: 2.78,
    h: 1.0,
    accent: theme.accent5,
    kicker: "02",
    title: "更少元素",
    body: "把注意力压在一个中心动作上。"
  });
  addMetricCard(slide3, theme, {
    x: 0.98,
    y: 4.08,
    w: 5.36,
    h: 1.2,
    accent: theme.accent4,
    kicker: "TOGGLE",
    title: "按钮一按，直接切到另一种状态",
    body: "状态切换最稳的做法，是跳到另一页的另一套布局。编辑时好维护，演示时也最像“真的变了”。",
  });
  addPillButton(slide3, theme, {
    x: 0.98,
    y: 5.44,
    w: 1.18,
    label: "回首页",
    accent: theme.accent,
    text: theme.text,
    link: { slide: 1, tooltip: "返回首页" },
  });
  addPillButton(slide3, theme, {
    x: 2.28,
    y: 5.44,
    w: 1.18,
    label: "去目录",
    accent: theme.accent2,
    text: theme.text,
    link: { slide: 2, tooltip: "返回目录" },
  });
  addPillButton(slide3, theme, {
    x: 3.58,
    y: 5.44,
    w: 1.18,
    label: "去收口",
    accent: theme.accent3,
    text: theme.text,
    link: { slide: 4, tooltip: "进入收口页" },
  });

  slide3.addText("状态 B / 行动视图", {
    x: 7.08,
    y: 2.18,
    w: 3.2,
    h: 0.24,
    fontFace: "Aptos Display",
    fontSize: 20.5,
    bold: true,
    color: theme.text,
  });
  slide3.addText("右边更像“行动视图”", {
    x: 7.08,
    y: 2.54,
    w: 3.2,
    h: 0.12,
    fontFace: "Microsoft YaHei",
    fontSize: 11,
    color: theme.muted,
  });
  addMetricCard(slide3, theme, {
    x: 7.08,
    y: 2.9,
    w: 2.42,
    h: 1.0,
    accent: theme.accent2,
    kicker: "03",
    title: "更强对比",
    body: "用高亮块把动作感推出来。"
  });
  addMetricCard(slide3, theme, {
    x: 9.66,
    y: 2.9,
    w: 2.92,
    h: 1.0,
    accent: theme.accent,
    kicker: "04",
    title: "更像按钮",
    body: "让观众明确知道“可以点”。"
  });
  addMetricCard(slide3, theme, {
    x: 7.08,
    y: 4.08,
    w: 5.5,
    h: 1.2,
    accent: theme.accent2,
    kicker: "LINK",
    title: "外链、资料、系统页都可以接进来",
    body: "对于产品演示、培训和发布会，这种状态页特别适合承接后续动作。",
  });
  addPillButton(slide3, theme, {
    x: 7.08,
    y: 5.44,
    w: 1.2,
    label: "去首页",
    accent: theme.accent,
    text: theme.text,
    link: { slide: 1, tooltip: "返回首页" },
  });
  addPillButton(slide3, theme, {
    x: 8.4,
    y: 5.44,
    w: 1.2,
    label: "去目录",
    accent: theme.accent2,
    text: theme.text,
    link: { slide: 2, tooltip: "返回目录" },
  });
  addPillButton(slide3, theme, {
    x: 9.72,
    y: 5.44,
    w: 1.4,
    label: "打开文档",
    accent: theme.accent3,
    text: theme.text,
    link: { url: "https://learn.microsoft.com/en-us/office/vba/api/powerpoint.actionsetting.hyperlink", tooltip: "打开 Microsoft Learn" },
  });
  addNav(slide3, theme, { directorySlide: 2, nextSlide: 4 });

  // Slide 4: Close / proof
  const slide4 = pptx.addSlide("MASTER");
  addBackdrop(slide4, theme);
  addHeader(
    slide4,
    theme,
    "CLOSE / PROOF",
    "最终拿到的是能继续编辑的交互成品，不是一次性截图",
    "这页把可交付性说透：字能改、位能移、样式能换、链接还能继续保留。"
  );
  addChip(slide4, theme, 10.36, 0.2, 2.12, "Editable proof · Live links · PPT-safe", theme.accent4);

  slide4.addShape("roundRect", {
    x: 0.74,
    y: 1.48,
    w: 12.0,
    h: 4.88,
    rectRadius: 0.06,
    fill: { color: theme.panel, transparency: 0 },
    line: { color: theme.accent4, pt: 1.08 },
  });
  addChip(slide4, theme, 1.0, 1.8, 1.28, "DELIVERY", theme.accent4);
  slide4.addText("高级感，不靠堆特效", {
    x: 1.0,
    y: 2.18,
    w: 4.8,
    h: 0.26,
    fontFace: "Aptos Display",
    fontSize: 22,
    bold: true,
    color: theme.text,
  });
  slide4.addText("真正的高级，是把编辑性、交互性、视觉层次一起保住。", {
    x: 1.0,
    y: 2.56,
    w: 5.2,
    h: 0.18,
    fontFace: "Microsoft YaHei",
    fontSize: 11.4,
    color: theme.muted,
  });

  addMetricCard(slide4, theme, {
    x: 1.0,
    y: 3.12,
    w: 2.74,
    h: 1.22,
    accent: theme.accent,
    kicker: "EDIT",
    title: "可改字",
    body: "文字仍然是文字，后期可继续编辑。"
  });
  addMetricCard(slide4, theme, {
    x: 3.9,
    y: 3.12,
    w: 2.74,
    h: 1.22,
    accent: theme.accent2,
    kicker: "MOVE",
    title: "可改位",
    body: "卡片、按钮、标题都能再调整。"
  });
  addMetricCard(slide4, theme, {
    x: 6.8,
    y: 3.12,
    w: 2.74,
    h: 1.22,
    accent: theme.accent3,
    kicker: "STYLE",
    title: "可改样式",
    body: "颜色、层级、留白都可按需替换。"
  });
  addMetricCard(slide4, theme, {
    x: 9.7,
    y: 3.12,
    w: 2.74,
    h: 1.22,
    accent: theme.accent4,
    kicker: "LINK",
    title: "可改链接",
    body: "内部跳转和外链都能继续保留。"
  });

  slide4.addShape("roundRect", {
    x: 1.0,
    y: 4.64,
    w: 11.44,
    h: 1.04,
    rectRadius: 0.06,
    fill: { color: theme.bg, transparency: 0 },
    line: { color: theme.line, pt: 1 },
  });
  slide4.addText("如果要把它进一步推到“最高级”，就把同样的结构接到真实业务内容里：产品页、培训页、汇报页、FAQ 页都能用这套交互骨架。", {
    x: 1.24,
    y: 4.9,
    w: 10.88,
    h: 0.22,
    fontFace: "Microsoft YaHei",
    fontSize: 11.3,
    color: theme.text,
    align: "center",
  });
  slide4.addText("这不是静态演示，而是可以继续长出来的交付骨架。", {
    x: 1.24,
    y: 5.28,
    w: 10.88,
    h: 0.14,
    fontFace: "Aptos",
    fontSize: 9.4,
    color: theme.muted,
    align: "center",
  });

  addPillButton(slide4, theme, {
    x: 4.62,
    y: 5.78,
    w: 1.36,
    label: "返回首页",
    accent: theme.accent,
    text: theme.text,
    link: { slide: 1, tooltip: "返回首页" },
  });
  addPillButton(slide4, theme, {
    x: 6.18,
    y: 5.78,
    w: 1.36,
    label: "打开 GitHub",
    accent: theme.accent2,
    text: theme.text,
    link: { url: "https://github.com/gitbrent/pptxgenjs", tooltip: "打开 PptxGenJS GitHub" },
  });
  addNav(slide4, theme, { directorySlide: 2, nextSlide: 1 });

  await fs.mkdir(path.dirname(out), { recursive: true });
  await pptx.writeFile({ fileName: out });
  await stripNotesParts(out);

  console.log(JSON.stringify({ output: out, slides: 4, mode: "native-interactive-demo", theme: theme.name }, null, 2));
}

main().catch((error) => {
  console.error(error.stack || error.message || String(error));
  console.error(usage());
  process.exit(1);
});
