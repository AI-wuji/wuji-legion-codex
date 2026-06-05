#!/usr/bin/env node

import fs from "node:fs/promises";
import path from "node:path";
import { createRequire } from "node:module";

function usage() {
  return [
    "Usage:",
    "  node scripts/wuji-ppt-template-interactive-demo.mjs --out <file.pptx>",
    "",
    "Creates a template-matched interactive deck with native slide links and external links.",
  ].join("\n");
}

function addGlowButton(slide, theme, cfg) {
  slide.addShape("roundRect", {
    x: cfg.x,
    y: cfg.y,
    w: cfg.w,
    h: cfg.h ?? 0.34,
    rectRadius: 0.12,
    fill: { color: cfg.fill ?? theme.accent, transparency: cfg.transparency ?? 20 },
    line: { color: cfg.line ?? theme.accent, pt: 1.1, transparency: 0 },
    shadow: { type: "outer", color: cfg.line ?? theme.accent, opacity: 0.2, blur: 4, offset: 1, angle: 45 },
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
    x: 9.82,
    y: 0.2,
    w: 2.8,
    h: 0.26,
    fontFace: "Microsoft YaHei",
    fontSize: 15.5,
    bold: true,
    italic: true,
    color: "FFFFFF",
    align: "right",
  });
}

function addBottomBrand(slide, text) {
  slide.addText("―", {
    x: 5.95,
    y: 6.92,
    w: 0.3,
    h: 0.1,
    fontFace: "Microsoft YaHei",
    fontSize: 11,
    bold: true,
    color: "FFFFFF",
    align: "center",
  });
  slide.addText(text, {
    x: 6.18,
    y: 6.89,
    w: 0.96,
    h: 0.16,
    fontFace: "Microsoft YaHei",
    fontSize: 9.4,
    bold: true,
    color: "FFFFFF",
    align: "center",
  });
  slide.addText("―", {
    x: 7.16,
    y: 6.92,
    w: 0.3,
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
  pptx.title = "悦蓝学堂·真交互PPTX";
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

  // Slide 1
  {
    const slide = pptx.addSlide();
    slide.background = { color: bg };
    slide.addImage({ path: bgWave, x: 0, y: 0, w: 13.333, h: 7.5 });
    addTopRightBrand(slide, "悦蓝学堂");
    slide.addText("真交互PPTX：", {
      x: 1.18,
      y: 1.52,
      w: 4.9,
      h: 0.72,
      fontFace: "Microsoft YaHei",
      fontSize: 31,
      bold: true,
      italic: true,
      color: "FFFFFF",
    });
    slide.addText("按钮、跳页、外链，\n都是真能点的 PowerPoint", {
      x: 1.18,
      y: 2.18,
      w: 5.4,
      h: 1.2,
      fontFace: "Microsoft YaHei",
      fontSize: 23,
      bold: true,
      italic: true,
      color: "FFFFFF",
      breakLine: false,
    });
    slide.addText("不是 HTML 截图，不是图片壳。它仍然是可编辑的 .pptx，\n只是把交互、层次和视觉做得更像发布会。", {
      x: 1.2,
      y: 3.68,
      w: 4.9,
      h: 0.64,
      fontFace: "Times New Roman",
      fontSize: 14.5,
      italic: true,
      color: "D8D5E8",
    });

    addGlowButton(slide, {
      accent: accentBlue,
    }, {
      x: 1.18, y: 5.35, w: 1.08, label: "目录", link: { slide: 2 }, text: "FFFFFF", fontSize: 10.2, fill: accentBlue,
    });
    addGlowButton(slide, { accent: accentPink }, {
      x: 2.4, y: 5.35, w: 1.08, label: "分支", link: { slide: 3 }, text: "FFFFFF", fontSize: 10.2, fill: accentPink,
    });
    addGlowButton(slide, { accent: accentPurple }, {
      x: 3.62, y: 5.35, w: 1.28, label: "打开文档", link: { url: "https://github.com/gitbrent/pptxgenjs" }, text: "FFFFFF", fontSize: 10.1, fill: accentPurple,
    });
    addGlowButton(slide, { accent: accentPeach }, {
      x: 5.06, y: 5.35, w: 1.08, label: "收口", link: { slide: 4 }, text: "FFFFFF", fontSize: 10.2, fill: accentPeach,
    });

    addBottomBrand(slide, "悦蓝学堂");
  }

  // Slide 2
  {
    const slide = pptx.addSlide();
    slide.background = { color: bg };
    addTopRightBrand(slide, "悦蓝学堂");
    slide.addText("CONTENT", {
      x: 6.04,
      y: 0.26,
      w: 1.3,
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
      { x: 2.08, label: "目录\n跳转", sub1: "点一下就去不同分支", sub2: "适合章节 / 目录 / 菜单", link: { slide: 3 }, color: accentBlue },
      { x: 9.06, label: "状态\n分支", sub1: "同一页切换不同结果", sub2: "适合 if / 选择 / 演示", link: { slide: 3 }, color: accentPink },
    ];
    circles.forEach((c) => {
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
        y: 1.42,
        w: 1.44,
        h: 0.7,
        fontFace: "Microsoft YaHei",
        fontSize: 20,
        bold: true,
        color: "FFFFFF",
        align: "center",
        valign: "mid",
        hyperlink: c.link,
      });
    });

    slide.addText("掌握万能提示词公式", {
      x: 1.54,
      y: 3.55,
      w: 3.4,
      h: 0.24,
      fontFace: "Microsoft YaHei",
      fontSize: 18,
      bold: true,
      color: accentBlue,
      align: "center",
    });
    slide.addText("学会基础运镜 + 进阶叠加用法", {
      x: 1.06,
      y: 4.14,
      w: 4.32,
      h: 0.24,
      fontFace: "Microsoft YaHei",
      fontSize: 18,
      bold: true,
      color: "FFFFFF",
      align: "center",
    });
    slide.addText("万能提示词公式", {
      x: 8.52,
      y: 3.55,
      w: 3.4,
      h: 0.24,
      fontFace: "Microsoft YaHei",
      fontSize: 18,
      bold: true,
      color: accentPink,
      align: "center",
    });
    slide.addText("镜头运镜技巧", {
      x: 8.9,
      y: 4.14,
      w: 2.66,
      h: 0.24,
      fontFace: "Microsoft YaHei",
      fontSize: 18,
      bold: true,
      color: "FFFFFF",
      align: "center",
    });

    addGlowButton(slide, { accent: accentBlue }, {
      x: 4.08, y: 5.55, w: 1.18, label: "去分支", link: { slide: 3 }, text: "FFFFFF", fontSize: 10.2, fill: accentBlue,
    });
    addGlowButton(slide, { accent: accentPink }, {
      x: 5.44, y: 5.55, w: 1.28, label: "打开GitHub", link: { url: "https://github.com/gitbrent/pptxgenjs" }, text: "FFFFFF", fontSize: 10.0, fill: accentPink,
    });
    addBottomBrand(slide, "悦蓝学堂");
  }

  // Slide 3
  {
    const slide = pptx.addSlide();
    slide.background = { color: bg };
    addTopRightBrand(slide, "悦蓝学堂");
    slide.addImage({ path: ring, x: 3.72, y: 0.66, w: 5.9, h: 5.15 });
    slide.addShape(pptx.ShapeType.ellipse, {
      x: 6.08, y: 2.18, w: 1.15, h: 1.15,
      fill: { color: bg, transparency: 100 },
      line: { color: "FFFFFF", pt: 1.0, transparency: 0 },
    });
    slide.addText("PPTX", {
      x: 6.15,
      y: 2.62,
      w: 0.98,
      h: 0.18,
      fontFace: "Microsoft YaHei",
      fontSize: 14,
      bold: true,
      color: "FFFFFF",
      align: "center",
    });
    slide.addText("点一下，状态就切换", {
      x: 1.08,
      y: 4.8,
      w: 5.7,
      h: 0.34,
      fontFace: "Microsoft YaHei",
      fontSize: 23,
      bold: true,
      italic: true,
      color: "FFFFFF",
    });
    slide.addText("按钮不会只是摆设。它们真的会跳页、回流、打开外链，\n并且整个过程仍然保留在可编辑的 PowerPoint 里。", {
      x: 1.08,
      y: 5.28,
      w: 5.5,
      h: 0.64,
      fontFace: "Times New Roman",
      fontSize: 13.6,
      italic: true,
      color: "D8D5E8",
    });

    addGlowButton(slide, { accent: accentBlue }, {
      x: 1.08, y: 6.06, w: 1.12, label: "回首页", link: { slide: 1 }, text: "FFFFFF", fontSize: 10.2, fill: accentBlue,
    });
    addGlowButton(slide, { accent: accentPink }, {
      x: 2.34, y: 6.06, w: 1.12, label: "看目录", link: { slide: 2 }, text: "FFFFFF", fontSize: 10.2, fill: accentPink,
    });
    addGlowButton(slide, { accent: accentPurple }, {
      x: 3.6, y: 6.06, w: 1.3, label: "打开文档", link: { url: "https://learn.microsoft.com/en-us/office/vba/api/powerpoint.actionsetting.hyperlink" }, text: "FFFFFF", fontSize: 10.0, fill: accentPurple,
    });
    addGlowButton(slide, { accent: accentPeach }, {
      x: 5.06, y: 6.06, w: 1.18, label: "收口", link: { slide: 4 }, text: "FFFFFF", fontSize: 10.2, fill: accentPeach,
    });

    addBottomBrand(slide, "悦蓝学堂");
  }

  // Slide 4
  {
    const slide = pptx.addSlide();
    slide.background = { color: bg };
    slide.addImage({ path: bgWave, x: 0, y: 0, w: 13.333, h: 7.5 });
    slide.addText("THANKS", {
      x: 4.0,
      y: 2.62,
      w: 5.2,
      h: 0.74,
      fontFace: "Microsoft YaHei",
      fontSize: 46,
      bold: true,
      italic: true,
      color: "FFFFFF",
      align: "center",
    });
    slide.addText("FOR WATCHING", {
      x: 5.0,
      y: 3.42,
      w: 3.2,
      h: 0.22,
      fontFace: "Times New Roman",
      fontSize: 17,
      italic: true,
      color: "D8D5E8",
      align: "center",
    });
    addTopRightBrand(slide, "悦蓝学堂");
    addBottomBrand(slide, "悦蓝学堂");
    addGlowButton(slide, { accent: accentBlue }, {
      x: 4.62, y: 5.62, w: 1.2, label: "回首页", link: { slide: 1 }, text: "FFFFFF", fontSize: 10.2, fill: accentBlue,
    });
    addGlowButton(slide, { accent: accentPink }, {
      x: 5.98, y: 5.62, w: 1.44, label: "打开 GitHub", link: { url: "https://github.com/gitbrent/pptxgenjs" }, text: "FFFFFF", fontSize: 10.0, fill: accentPink,
    });
  }

  await fs.mkdir(path.dirname(out), { recursive: true });
  await pptx.writeFile({ fileName: out });
  await stripNotesParts(out);

  console.log(JSON.stringify({ output: out, slides: 4, mode: "template-interactive-demo" }, null, 2));
}

main().catch((error) => {
  console.error(error.stack || error.message || String(error));
  console.error(usage());
  process.exit(1);
});
