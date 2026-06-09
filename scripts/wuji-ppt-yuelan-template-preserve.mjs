#!/usr/bin/env node

import fs from "node:fs/promises";
import path from "node:path";
import { createRequire } from "node:module";

const DEFAULT_OUT = path.resolve("outputs/template-following-yuelan/yuelan-preserved-interactive.pptx");

function usage() {
  return [
    "Usage:",
    "  node scripts/wuji-ppt-yuelan-template-preserve.mjs --source <authorized-template.pptx> --source-authorized true [--out <file.pptx>]",
    "",
    "Preserves explicitly supplied, authorized template elements and remaps selected pages into a 4-slide interactive deck.",
  ].join("\n");
}

function replaceTexts(xml, replacements) {
  let cursor = 0;
  return xml.replace(/<a:t>[\s\S]*?<\/a:t>/g, (match) => {
    const next = replacements[cursor++];
    if (next === undefined) return match;
    return match.replace(/<a:t>[\s\S]*?<\/a:t>/, `<a:t>${escapeXml(next)}</a:t>`);
  });
}

function escapeXml(value) {
  return String(value)
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&apos;");
}

function injectTransition(xml, transitionXml) {
  if (xml.includes("<p:transition")) return xml;
  return xml.replace(/<\/p:cSld>/, `</p:cSld>${transitionXml}`);
}

function addHyperlinkToShape(xml, shapeName, relId, occurrence = 1) {
  let hit = 0;
  return xml.replace(
    new RegExp(`(<p:cNvPr id="\\d+" name="${shapeName.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")}"\\s*\\/>)`, "g"),
    (match) => {
      hit += 1;
      if (hit !== occurrence) return match;
      return match.replace(
        /\/>$/,
        `><a:hlinkClick r:id="${relId}" history="1" highlightClick="0" endSnd="0"/></p:cNvPr>`,
      );
    },
  );
}

async function patchRels(zip, relPath, relationships) {
  const entry = zip.file(relPath);
  if (!entry) throw new Error(`Missing relation file: ${relPath}`);
  let xml = await entry.async("string");
  const insertAt = xml.lastIndexOf("</Relationships>");
  if (insertAt < 0) throw new Error(`Unexpected rels XML: ${relPath}`);
  const existing = new Set([...xml.matchAll(/Id="(rId\d+)"/g)].map((m) => m[1]));
  const extra = relationships
    .filter((rel) => !existing.has(rel.id))
    .map(
      (rel) =>
        `<Relationship Id="${rel.id}" Type="${rel.type}" Target="${rel.target}"${rel.targetMode ? ` TargetMode="${rel.targetMode}"` : ""}/>`,
    )
    .join("");
  xml = xml.slice(0, insertAt) + extra + xml.slice(insertAt);
  zip.file(relPath, xml);
}

async function patchSlide(zip, slideNumber, { texts, transitions, links = [] }) {
  const slidePath = `ppt/slides/slide${slideNumber}.xml`;
  const relPath = `ppt/slides/_rels/slide${slideNumber}.xml.rels`;
  const slideEntry = zip.file(slidePath);
  if (!slideEntry) throw new Error(`Missing slide: ${slidePath}`);
  let xml = await slideEntry.async("string");
  xml = replaceTexts(xml, texts);
  for (const { shapeName, relId, occurrence = 1 } of links) {
    xml = addHyperlinkToShape(xml, shapeName, relId, occurrence);
  }
  if (transitions) xml = injectTransition(xml, transitions);
  zip.file(slidePath, xml);
  return relPath;
}

async function main() {
  const args = Object.fromEntries(
    process.argv.slice(2).reduce((acc, token, index, arr) => {
      if (!token.startsWith("--")) return acc;
      const key = token.slice(2);
      const next = arr[index + 1];
      if (next && !next.startsWith("--")) acc.push([key, next]);
      else acc.push([key, "true"]);
      return acc;
    }, []),
  );

  if (args.help || args.h) {
    console.log(usage());
    return;
  }

  if (!args.source || args["source-authorized"] !== "true") {
    throw new Error("Provide --source <authorized-template.pptx> and --source-authorized true. No personal or desktop file is used by default.");
  }
  const source = path.resolve(args.source);
  const out = args.out ? path.resolve(args.out) : DEFAULT_OUT;
  const require = createRequire(import.meta.url);
  const JSZip = process.env.WUJI_JSZIP_PATH ? require(process.env.WUJI_JSZIP_PATH) : require("jszip");

  const sourceBytes = await fs.readFile(source);
  const zip = await JSZip.loadAsync(sourceBytes);

  // Remap the presentation to only four slides, preserving the original template pages.
  const presRelPath = "ppt/_rels/presentation.xml.rels";
  const presXmlPath = "ppt/presentation.xml";
  const presRelEntry = zip.file(presRelPath);
  const presXmlEntry = zip.file(presXmlPath);
  if (!presRelEntry || !presXmlEntry) throw new Error("Missing presentation XML parts.");

  let presRelXml = await presRelEntry.async("string");
  let presXml = await presXmlEntry.async("string");

  // Keep original relations, but the deck will only expose slide 1, 5, 6, 22.
  const slideTargets = [
    { rid: "rId3", target: "slides/slide1.xml" },
    { rid: "rId8", target: "slides/slide5.xml" },
    { rid: "rId9", target: "slides/slide6.xml" },
    { rid: "rId25", target: "slides/slide22.xml" },
  ];

  presXml = presXml.replace(/<p:sldIdLst>[\s\S]*?<\/p:sldIdLst>/, () => {
    const ids = slideTargets
      .map((item, idx) => `<p:sldId id="${256 + idx}" r:id="${item.rid}"/>`)
      .join("");
    return `<p:sldIdLst>${ids}</p:sldIdLst>`;
  });
  zip.file(presXmlPath, presXml);

  // Patch the slide relations so the reused shapes become click targets.
  const relsToAdd = [
    {
      path: "ppt/slides/_rels/slide1.xml.rels",
      rels: [
        { id: "rId4", type: "http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide", target: "slide5.xml" },
        { id: "rId5", type: "http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide", target: "slide6.xml" },
        { id: "rId6", type: "http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide", target: "slide22.xml" },
      ],
    },
    {
      path: "ppt/slides/_rels/slide5.xml.rels",
      rels: [
        { id: "rId12", type: "http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide", target: "slide6.xml" },
        { id: "rId13", type: "http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide", target: "slide22.xml" },
      ],
    },
    {
      path: "ppt/slides/_rels/slide6.xml.rels",
      rels: [
        { id: "rId4", type: "http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide", target: "slide1.xml" },
        { id: "rId5", type: "http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide", target: "slide5.xml" },
        { id: "rId6", type: "http://schemas.openxmlformats.org/officeDocument/2006/relationships/hyperlink", target: "https://learn.microsoft.com/en-us/office/vba/api/powerpoint.actionsetting.hyperlink", targetMode: "External" },
      ],
    },
    {
      path: "ppt/slides/_rels/slide22.xml.rels",
      rels: [
        { id: "rId3", type: "http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide", target: "slide1.xml" },
        { id: "rId4", type: "http://schemas.openxmlformats.org/officeDocument/2006/relationships/hyperlink", target: "https://github.com/gitbrent/pptxgenjs", targetMode: "External" },
      ],
    },
  ];
  for (const item of relsToAdd) {
    await patchRels(zip, item.path, item.rels);
  }

  // Slides 1 / 5 / 6 / 22 are the preserved template pages, rewritten with the previous 4-slide demo content.
  await patchSlide(zip, 1, {
    texts: [
      "真交互PPTX：",
      "按钮/跳页/外链",
      "真能点",
    ],
    transitions: '<p:transition advClick="1" dur="600"><p:fade/></p:transition>',
    links: [
      { shapeName: "矩形 10", relId: "rId4" },
    ],
  });

  await patchSlide(zip, 5, {
    texts: [
      "目录",
      "控制",
      "可点",
      "Route",
      "PAGE 02",
      "跳转",
      "左入口，右路径",
      "01",
      "目录",
      "进章",
      "02",
      "分支",
      "换果",
    ],
    transitions: '<p:transition advClick="1" dur="500"><p:push dir="l"/></p:transition>',
    links: [
      { shapeName: "圆: 空心 19", relId: "rId12", occurrence: 1 },
      { shapeName: "圆: 空心 19", relId: "rId13", occurrence: 2 },
      { shapeName: "矩形 30", relId: "rId12" },
      { shapeName: "矩形 18", relId: "rId13" },
    ],
  });

  await patchSlide(zip, 6, {
    texts: [
      "BRANCH / STATE",
      "点一下就变",
      "状态切换",
    ],
    transitions: '<p:transition advClick="1" dur="600"><p:wipe dir="r"/></p:transition>',
    links: [
      { shapeName: "圆: 空心 2", relId: "rId4" },
      { shapeName: "等腰三角形 46", relId: "rId5", occurrence: 1 },
      { shapeName: "等腰三角形 46", relId: "rId6", occurrence: 2 },
      { shapeName: "矩形 7", relId: "rId5" },
      { shapeName: "矩形 6", relId: "rId6" },
    ],
  });

  await patchSlide(zip, 22, {
    texts: [
      "CLOSE / PROOF",
      "能继续编辑，不是截图",
    ],
    transitions: '<p:transition advClick="1" dur="700"><p:fade/></p:transition>',
    links: [
      { shapeName: "矩形 2", relId: "rId3" },
      { shapeName: "矩形 11", relId: "rId4" },
    ],
  });

  await fs.mkdir(path.dirname(out), { recursive: true });
  await fs.writeFile(out, await zip.generateAsync({ type: "nodebuffer", compression: "DEFLATE" }));
  console.log(JSON.stringify({ output: out, slides: 4, mode: "template-preserved-interactive" }, null, 2));
}

main().catch((error) => {
  console.error(error.stack || error.message || String(error));
  console.error(usage());
  process.exit(1);
});
