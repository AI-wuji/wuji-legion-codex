#!/usr/bin/env node

import fs from "node:fs/promises";
import path from "node:path";
import { createRequire } from "node:module";

function usage() {
  return [
    "Usage:",
    "  node scripts/wuji-ppt-template-interactive-demo-v2.mjs --out <file.pptx>",
    "",
    "Creates a template-matched interactive deck with real internal slide links and external links.",
  ].join("\n");
}

function addGlowButton(slide, theme, cfg) {
  slide.addShape("roundRect", {
    x: cfg.x - 0.04,
    y: cfg.y - 0.03,
    w: cfg.w + 0.08,
    h: (cfg.h ?? 0.34) + 0.06,
    rectRadius: 0.14,
    fill: { color: cfg.fill ?? theme.accent, transparency: 74 },
    line: { color: cfg.fill ?? theme.accent, transparency: 100, pt: 0 },
  });
  slide.addShape("roundRect", {
    x: cfg.x,
    y: cfg.y,
    w: cfg.w,
    h: cfg.h ?? 0.34,
    rectRadius: 0.12,
    fill: { color: cfg.fill ?? theme.accent, transparency: cfg.transparency ?? 18 },
    line: { color: cfg.line ?? theme.accent, pt: 1.05, transparency: 0 },
    shadow: { type: "outer", color: cfg.line ?? theme.accent, opacity: 0.24, blur: 4, offset: 1, angle: 45 },
    hyperlink: cfg.link,
  });
  slide.addText(cfg.label, {
    x: cfg.x,
    y: cfg.y + 0.02,
    w: cfg.w,
    h: (cfg.h ?? 0.34) - 0.04,
    fontFace: "Microsoft YaHei",
    fontSize: cfg.fontSize ?? 10.5,
    bold: true,
    italic: Boolean(cfg.italic),
    color: cfg.text ?? "FFFFFF",
    align: "center",
    valign: "mid",
    hyperlink: cfg.link,
  });
}

function addTopRightBrand(slide, text) {
  slide.addText(text, {
    x: 9.66,
    y: 0.18,
    w: 2.85,
    h: 0.28,
    fontFace: "Microsoft YaHei",
    fontSize: 15.5,
    bold: true,
    italic: true,
    color: "FFFFFF",
    align: "right",
  });
}

function addBottomBrand(slide, text) {
  slide.addText("—", {
    x: 5.9,
    y: 6.88,
    w: 0.2,
    h: 0.1,
    fontFace: "Microsoft YaHei",
    fontSize: 11,
    bold: true,
    color: "FFFFFF",
    align: "center",
  });
  slide.addText(text, {
    x: 6.15,
    y: 6.85,
    w: 0.96,
    h: 0.14,
    fontFace: "Microsoft YaHei",
    fontSize: 9.4,
    bold: true,
    color: "FFFFFF",
    align: "center",
  });
  slide.addText("—", {
    x: 7.1,
    y: 6.88,
    w: 0.2,
    h: 0.1,
    fontFace: "Microsoft YaHei",
    fontSize: 11,
    bold: true,
    color: "FFFFFF",
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

async function injectTransitions(pptxPath) {
  const require = createRequire(import.meta.url);
  const JSZip = require("C:/Users/Administrator/.cache/codex-runtimes/codex-primary-runtime/dependencies/node/node_modules/.pnpm/jszip@3.10.1/node_modules/jszip");
  const bytes = await fs.readFile(pptxPath);
  const zip = await JSZip.loadAsync(bytes);

  const transitions = [
    { slide: 1, xml: '<p:transition advClick="1" dur="600"><p:fade/></p:transition>' },
    { slide: 2, xml: '<p:transition advClick="1" dur="500"><p:push dir="l"/></p:transition>' },
    { slide: 3, xml: '<p:transition advClick="1" dur="600"><p:wipe dir="r"/></p:transition>' },
    { slide: 4, xml: '<p:transition advClick="1" dur="700"><p:fade/></p:transition>' },
  ];

  for (const { slide, xml: transitionXml } of transitions) {
    const name = `ppt/slides/slide${slide}.xml`;
    const entry = zip.file(name);
    if (!entry) continue;
    const slideXml = await entry.async("string");
    if (slideXml.includes("<p:transition")) {
      continue;
    }
    const patched = slideXml.replace(
      /(<p:cSld[\s\S]*?<\/p:cSld>)/,
      `$1${transitionXml}`,
    );
    zip.file(name, patched);
  }

  await fs.writeFile(pptxPath, await zip.generateAsync({ type: "nodebuffer", compression: "DEFLATE" }));
}

async function main() {
  const args = Object.fromEntries(
    process.argv.slice(2).reduce((acc, token, index, arr) => {
      if (!token.startsWith("--")) return acc;
      const key = token.slice(2);
      const next = arr[index + 1];
      if (next && !next.startsWith("--")) {
        acc.push([key, next]);
      } else {
        acc.push([key, "true"]);
      }
      return acc;
    }, []),
  );

  if (args.help || args.h) {
    console.log(usage());
    return;
  }

  const out = args.out ? path.resolve(args.out) : "";
  if (!out) throw new Error("Missing --out <file.pptx>.");

  const mediaDir = path.resolve("E:/wuji-projects/wuji-legion-codex/outputs/template-following-yuelan/media");
  const bgWave = path.join(mediaDir, "image1.png");
  const ring = path.join(mediaDir, "image2.png");

  const require = createRequire(import.meta.url);
  const PptxGenJS = require("C:/Users/Administrator/.cache/codex-runtimes/codex-primary-runtime/dependencies/node/node_modules/.pnpm/pptxgenjs@4.0.1/node_modules/pptxgenjs");

  const pptx = new PptxGenJS();
  pptx.layout = "LAYOUT_WIDE";
  pptx.author = "Codex";
  pptx.company = "WuJi Legion";
  pptx.subject = "Template-matched interactive PowerPoint";
  pptx.title = "悦蓝模板交互式PPTX";
  pptx.lang = "zh-CN";
  pptx.theme = {
    headFontFace: "Microsoft YaHei",
    bodyFontFace: "Microsoft YaHei",
    lang: "zh-CN",
  };

  const bg = "05002B";
  const accentBlue = "52C7FF";
  const accentPink = "D45CFF";
  const accentPurple = "7E4BFF";
  const accentPeach = "F2A37E";
  const brandTop = "AIGC全能视频创作营";
  const brandBottom = "悦蓝学堂";

  // Slide 1: hero
  {
    const slide = pptx.addSlide();
    slide.background = { color: bg };
    slide.addImage({ path: bgWave, x: 0, y: 0, w: 13.333, h: 7.5 });
    addTopRightBrand(slide, brandTop);
    slide.addText("真交互PPT：\n按钮、跳页、外链，\n都是真能点的 PowerPoint", {
      x: 1.18,
      y: 1.46,
      w: 5.4,
      h: 1.42,
      fontFace: "Microsoft YaHei",
      fontSize: 30,
      bold: true,
      italic: true,
      color: "FFFFFF",
      breakLine: false,
    });
    slide.addText("不是 HTML 截图，不是图片壳。它仍然是可编辑的 .pptx，\n只是把交互、层次和视觉做得更像发布会。", {
      x: 1.2,
      y: 3.48,
      w: 5.1,
      h: 0.72,
      fontFace: "Times New Roman",
      fontSize: 14.2,
      italic: true,
      color: "D8D5E8",
    });
    addGlowButton(slide, { accent: accentBlue }, {
      x: 1.18, y: 5.36, w: 1.08, label: "目录", link: { slide: 2 }, text: "FFFFFF", fontSize: 10.2, fill: accentBlue,
    });
    addGlowButton(slide, { accent: accentPink }, {
      x: 2.4, y: 5.36, w: 1.08, label: "分支", link: { slide: 3 }, text: "FFFFFF", fontSize: 10.2, fill: accentPink,
    });
    addGlowButton(slide, { accent: accentPurple }, {
      x: 3.62, y: 5.36, w: 1.28, label: "打开文档", link: { url: "https://github.com/gitbrent/pptxgenjs" }, text: "FFFFFF", fontSize: 10.1, fill: accentPurple,
    });
    addGlowButton(slide, { accent: accentPeach }, {
      x: 5.06, y: 5.36, w: 1.08, label: "收口", link: { slide: 4 }, text: "FFFFFF", fontSize: 10.2, fill: accentPurple,
    });
    addBottomBrand(slide, brandBottom);
  }

  // Slide 2: content
  {
    const slide = pptx.addSlide();
    slide.background = { color: bg };
    addTopRightBrand(slide, brandTop);
    slide.addText("CONTENT", {
      x: 6.03,
      y: 0.26,
      w: 1.32,
      h: 0.2,
      fontFace: "Times New Roman",
      fontSize: 21,
      italic: true,
      color: "FFFFFF",
      align: "center",
    });
    slide.addShape(pptx.ShapeType.line, {
      x: 0.82, y: 0.5, w: 4.26, h: 0, line: { color: "69617E", pt: 0.9, transparency: 35 },
    });
    slide.addShape(pptx.ShapeType.line, {
      x: 4.7, y: 0.5, w: 0.26, h: 0, line: { color: "69617E", pt: 0.9, transparency: 35 },
    });
    slide.addShape(pptx.ShapeType.line, {
      x: 4.7, y: 0.56, w: 0.26, h: 0, line: { color: "69617E", pt: 0.9, transparency: 35 },
    });
    slide.addShape(pptx.ShapeType.line, {
      x: 4.7, y: 0.62, w: 0.26, h: 0, line: { color: "69617E", pt: 0.9, transparency: 35 },
    });

    const circles = [
      {
        x: 2.0,
        label: "目录\n跳转",
        sub1: "掌握可编辑交互结构",
        sub2: "学会基础跳转 + 进阶分支用法",
        link: { slide: 3 },
        color: accentBlue,
      },
      {
        x: 9.1,
        label: "状态\n分支",
        sub1: "万能链接设计",
        sub2: "按钮 / 跳页 / 外链技巧",
        link: { slide: 3 },
        color: accentPink,
      },
    ];

    circles.forEach((c) => {
      slide.addShape(pptx.ShapeType.ellipse, {
        x: c.x - 0.12,
        y: 1.02,
        w: 2.04,
        h: 2.04,
        fill: { color: c.color, transparency: 87 },
        line: { color: c.color, pt: 0, transparency: 100 },
      });
      slide.addShape(pptx.ShapeType.ellipse, {
        x: c.x,
        y: 1.14,
        w: 1.8,
        h: 1.8,
        fill: { color: bg, transparency: 100 },
        line: { color: c.color, pt: 2.4, transparency: 0 },
        shadow: { type: "outer", color: c.color, opacity: 0.22, blur: 4, offset: 1, angle: 45 },
        hyperlink: c.link,
      });
      slide.addText(c.label, {
        x: c.x + 0.18,
        y: 1.38,
        w: 1.44,
        h: 0.74,
        fontFace: "Microsoft YaHei",
        fontSize: 19,
        bold: true,
        color: "FFFFFF",
        align: "center",
        valign: "mid",
        hyperlink: c.link,
      });
    });

    slide.addText(circles[0].sub1, {
      x: 1.52,
      y: 3.52,
      w: 3.16,
      h: 0.24,
      fontFace: "Microsoft YaHei",
      fontSize: 18,
      bold: true,
      color: accentBlue,
      align: "center",
    });
    slide.addText(circles[0].sub2, {
      x: 0.98,
      y: 4.12,
      w: 4.36,
      h: 0.24,
      fontFace: "Microsoft YaHei",
      fontSize: 18,
      bold: true,
      color: "FFFFFF",
      align: "center",
    });
    slide.addText(circles[1].sub1, {
      x: 8.62,
      y: 3.52,
      w: 3.0,
      h: 0.24,
      fontFace: "Microsoft YaHei",
      fontSize: 18,
      bold: true,
      color: accentPink,
      align: "center",
    });
    slide.addText(circles[1].sub2, {
      x: 8.78,
      y: 4.12,
      w: 2.75,
      h: 0.24,
      fontFace: "Microsoft YaHei",
      fontSize: 18,
      bold: true,
      color: "FFFFFF",
      align: "center",
    });

    addGlowButton(slide, { accent: accentBlue }, {
      x: 4.08, y: 5.58, w: 1.18, label: "去分支", link: { slide: 3 }, text: "FFFFFF", fontSize: 10.2, fill: accentBlue,
    });
    addGlowButton(slide, { accent: accentPink }, {
      x: 5.44, y: 5.58, w: 1.28, label: "打开GitHub", link: { url: "https://github.com/gitbrent/pptxgenjs" }, text: "FFFFFF", fontSize: 10.0, fill: accentPink,
    });
    addBottomBrand(slide, brandBottom);
  }

  // Slide 3: interactive state switch
  {
    const slide = pptx.addSlide();
    slide.background = { color: bg };
    addTopRightBrand(slide, brandTop);

    slide.addShape(pptx.ShapeType.ellipse, {
      x: 4.86,
      y: 1.62,
      w: 3.0,
      h: 3.0,
      fill: { color: accentPink, transparency: 88 },
      line: { color: accentPink, pt: 0, transparency: 100 },
    });
    slide.addShape(pptx.ShapeType.ellipse, {
      x: 5.0,
      y: 1.72,
      w: 2.75,
      h: 2.75,
      fill: { color: accentBlue, transparency: 92 },
      line: { color: accentBlue, pt: 0, transparency: 100 },
    });
    slide.addImage({ path: ring, x: 3.45, y: 0.56, w: 6.2, h: 5.4 });
    slide.addShape(pptx.ShapeType.ellipse, {
      x: 6.1, y: 2.02, w: 1.22, h: 1.22,
      fill: { color: bg, transparency: 100 },
      line: { color: "FFFFFF", pt: 1.0, transparency: 0 },
    });
    slide.addText("PPTX", {
      x: 6.16,
      y: 2.49,
      w: 1.1,
      h: 0.18,
      fontFace: "Microsoft YaHei",
      fontSize: 14,
      bold: true,
      color: "FFFFFF",
      align: "center",
    });

    slide.addText("点一下，状态就切换", {
      x: 1.1,
      y: 4.76,
      w: 5.6,
      h: 0.32,
      fontFace: "Microsoft YaHei",
      fontSize: 24,
      bold: true,
      italic: true,
      color: "FFFFFF",
    });
    slide.addText("按钮不会只是摆设。它们真能跳页、回流、打开外链，\n而且整个过程仍然留在可编辑的 PowerPoint 里。", {
      x: 1.08,
      y: 5.22,
      w: 5.55,
      h: 0.7,
      fontFace: "Times New Roman",
      fontSize: 13.8,
      italic: true,
      color: "D8D5E8",
    });

    addGlowButton(slide, { accent: accentBlue }, {
      x: 1.08, y: 6.05, w: 1.12, label: "回首页", link: { slide: 1 }, text: "FFFFFF", fontSize: 10.2, fill: accentBlue,
    });
    addGlowButton(slide, { accent: accentPink }, {
      x: 2.34, y: 6.05, w: 1.12, label: "看目录", link: { slide: 2 }, text: "FFFFFF", fontSize: 10.2, fill: accentPink,
    });
    addGlowButton(slide, { accent: accentPurple }, {
      x: 3.6, y: 6.05, w: 1.3, label: "打开文档", link: { url: "https://learn.microsoft.com/en-us/office/vba/api/powerpoint.actionsetting.hyperlink" }, text: "FFFFFF", fontSize: 10.0, fill: accentPurple,
    });
    addGlowButton(slide, { accent: accentPeach }, {
      x: 5.06, y: 6.05, w: 1.18, label: "收口", link: { slide: 4 }, text: "FFFFFF", fontSize: 10.2, fill: accentPurple,
    });

    addBottomBrand(slide, brandBottom);
  }

  // Slide 4: close
  {
    const slide = pptx.addSlide();
    slide.background = { color: bg };

    slide.addShape(pptx.ShapeType.ellipse, {
      x: 9.7,
      y: 4.95,
      w: 2.3,
      h: 2.3,
      fill: { color: accentPink, transparency: 90 },
      line: { color: accentPink, pt: 0, transparency: 100 },
    });
    slide.addShape(pptx.ShapeType.ellipse, {
      x: 0.65,
      y: -0.25,
      w: 2.15,
      h: 2.15,
      fill: { color: accentBlue, transparency: 92 },
      line: { color: accentBlue, pt: 0, transparency: 100 },
    });

    // Use the template's wave asset as the upper and lower cropped bands.
    slide.addImage({ path: bgWave, x: -0.05, y: -2.45, w: 13.42, h: 4.95 });
    slide.addImage({ path: bgWave, x: -0.05, y: 4.95, w: 13.42, h: 4.95 });

    slide.addText("THANKS", {
      x: 3.98,
      y: 2.56,
      w: 5.35,
      h: 0.76,
      fontFace: "Microsoft YaHei",
      fontSize: 45,
      bold: true,
      italic: true,
      color: "FFFFFF",
      align: "center",
    });
    slide.addText("FOR WATCHING", {
      x: 5.02,
      y: 3.34,
      w: 3.2,
      h: 0.22,
      fontFace: "Times New Roman",
      fontSize: 17,
      italic: true,
      color: "D8D5E8",
      align: "center",
    });
    addTopRightBrand(slide, brandTop);
    addBottomBrand(slide, brandBottom);

    addGlowButton(slide, { accent: accentBlue }, {
      x: 4.58, y: 5.6, w: 1.2, label: "回首页", link: { slide: 1 }, text: "FFFFFF", fontSize: 10.2, fill: accentBlue,
    });
    addGlowButton(slide, { accent: accentPink }, {
      x: 5.95, y: 5.6, w: 1.44, label: "打开 GitHub", link: { url: "https://github.com/gitbrent/pptxgenjs" }, text: "FFFFFF", fontSize: 10.0, fill: accentPink,
    });
  }

  await fs.mkdir(path.dirname(out), { recursive: true });
  await pptx.writeFile({ fileName: out });
  await stripNotesParts(out);
  await injectTransitions(out);

  console.log(JSON.stringify({ output: out, slides: 4, mode: "template-interactive-demo-v2" }, null, 2));
}

main().catch((error) => {
  console.error(error.stack || error.message || String(error));
  console.error(usage());
  process.exit(1);
});
