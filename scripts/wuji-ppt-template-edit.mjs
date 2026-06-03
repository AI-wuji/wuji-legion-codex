#!/usr/bin/env node

import fs from "node:fs/promises";
import path from "node:path";
import { pathToFileURL } from "node:url";

function usage() {
  return [
    "Usage:",
    "  node scripts/wuji-ppt-template-edit.mjs --workspace <dir> --starter-pptx <starter.pptx> --map <template-frame-map.json> --out <final.pptx> [options]",
    "",
    "Options:",
    "  --preview-dir <dir>   Render edited slide PNGs. Defaults to <workspace>/preview/final.",
    "  --layout-dir <dir>    Write edited slide layout JSON. Defaults to <workspace>/layout/final.",
    "  --report <file>       Optional JSON report path. Defaults to <workspace>/template-edit-report.json.",
    "  --scale <n>           Render scale. Defaults to 1.",
    "  --no-preview         Skip final slide PNG export.",
    "  --no-layout          Skip final slide layout JSON export.",
    "",
    "Imports template-starter.pptx with artifact-tool, edits inherited elements",
    "in place using template-frame-map.json editTargets, then exports a final PPTX.",
  ].join("\n");
}

function isWithin(child, parent) {
  const relative = path.relative(parent, child);
  return relative === "" || (!relative.startsWith("..") && !path.isAbsolute(relative));
}

async function readJson(filePath) {
  return JSON.parse(await fs.readFile(filePath, "utf8"));
}

async function writeJson(filePath, value) {
  await fs.mkdir(path.dirname(filePath), { recursive: true });
  await fs.writeFile(filePath, `${JSON.stringify(value, null, 2)}\n`, "utf8");
}

async function loadArtifactToolUtils() {
  const skillDir = process.env.WUJI_PRESENTATIONS_SKILL_DIR;
  if (!skillDir) {
    throw new Error("Missing WUJI_PRESENTATIONS_SKILL_DIR. Invoke this script via wuji-ppt-template-edit.ps1.");
  }

  const modulePath = path.join(skillDir, "scripts", "artifact_tool_utils.mjs");
  try {
    return await import(pathToFileURL(modulePath).href);
  } catch (error) {
    throw new Error(`Failed to load artifact_tool_utils from ${modulePath}: ${error.message}`);
  }
}

function slidesFromPresentation(presentation) {
  if (Array.isArray(presentation.slides?.items)) return presentation.slides.items;
  if (Number.isInteger(presentation.slides?.count) && typeof presentation.slides.getItem === "function") {
    return Array.from({ length: presentation.slides.count }, (_, index) => presentation.slides.getItem(index));
  }
  throw new Error("Could not enumerate imported presentation slides.");
}

function targetIds(target) {
  const values = [
    target.shapeId,
    ...(Array.isArray(target.shapeIds) ? target.shapeIds : []),
    target.sourceElementId,
    ...(Array.isArray(target.sourceElementIds) ? target.sourceElementIds : []),
    target.elementId,
    ...(Array.isArray(target.elementIds) ? target.elementIds : []),
  ];
  return [...new Set(values.map((value) => String(value || "").trim()).filter(Boolean))];
}

function clone(value) {
  return value === undefined ? undefined : JSON.parse(JSON.stringify(value));
}

function textLinesFromTarget(target) {
  if (typeof target.text === "string") {
    return target.text.replace(/\r/g, "").split("\n");
  }
  if (Array.isArray(target.lines)) {
    return target.lines.map((line) => String(line ?? ""));
  }
  if (Array.isArray(target.textLines)) {
    return target.textLines.map((line) => String(line ?? ""));
  }
  return undefined;
}

function inlineNodesFromRuns(runs) {
  return runs.map((run) => ({ textRun: clone(run) }));
}

function paragraphsFromTextTarget(existingParagraphs, target) {
  if (Array.isArray(target.paragraphs) && target.paragraphs.length > 0) {
    return clone(target.paragraphs);
  }

  const templateParagraph = clone(existingParagraphs?.[0] || {});
  const templateRun = clone(templateParagraph.runs?.[0] || {});

  if (Array.isArray(target.runs) && target.runs.length > 0) {
    const runs = clone(target.runs);
    templateParagraph.runs = runs;
    templateParagraph.inlineNodes = inlineNodesFromRuns(runs);
    return [templateParagraph];
  }

  const lines = textLinesFromTarget(target);
  if (!lines) return undefined;

  return lines.map((line) => {
    const paragraph = clone(templateParagraph);
    const run = clone(templateRun);
    run.text = line;
    paragraph.runs = [run];
    paragraph.inlineNodes = inlineNodesFromRuns(paragraph.runs);
    return paragraph;
  });
}

function normalizedPosition(value) {
  if (!value || typeof value !== "object") return undefined;
  const left = Number(value.left ?? value.x);
  const top = Number(value.top ?? value.y);
  const width = Number(value.width ?? value.w);
  const height = Number(value.height ?? value.h);
  const position = {};
  if (Number.isFinite(left)) position.left = left;
  if (Number.isFinite(top)) position.top = top;
  if (Number.isFinite(width)) position.width = width;
  if (Number.isFinite(height)) position.height = height;
  return Object.keys(position).length > 0 ? position : undefined;
}

function resolveShape(slide, id) {
  try {
    return slide.shapes?.getById?.(id);
  } catch {
    return undefined;
  }
}

function resolveTable(slide, id) {
  try {
    return slide.tables?.getById?.(id);
  } catch {
    return undefined;
  }
}

function resolveImage(slide, id) {
  const images = Array.isArray(slide.images?.items) ? slide.images.items : [];
  return images.find((image) => image?.id === id);
}

function resolveElement(slide, id) {
  try {
    return slide.elements?.getById?.(id);
  } catch {
    return undefined;
  }
}

function resolveTargetObject(slide, id) {
  const shape = resolveShape(slide, id);
  if (shape) return { kind: "shape", item: shape };

  const table = resolveTable(slide, id);
  if (table) return { kind: "table", item: table };

  const image = resolveImage(slide, id);
  if (image) return { kind: "image", item: image };

  const element = resolveElement(slide, id);
  if (element) {
    if (typeof element.replace === "function" && typeof element.setImageReference === "function") {
      return { kind: "image", item: element };
    }
    if (typeof element.position === "object" && typeof element.toProto === "function") {
      return { kind: "shape", item: element };
    }
    return { kind: "element", item: element };
  }

  return undefined;
}

function applyTextUpdate(shape, target) {
  const textModel = shape?.text;
  const lines = textLinesFromTarget(target);

  if (textModel && typeof textModel.set === "function") {
    if (typeof target.text === "string") {
      textModel.set(target.text);
      return true;
    }
    if (lines) {
      textModel.set(lines.join("\n"));
      return true;
    }
    if (Array.isArray(target.runs) && target.runs.length > 0) {
      const content = target.runs.map((run) => String(run?.text ?? "")).join("");
      textModel.set(content);
      return true;
    }
  }

  const proto = shape.toProto();
  const paragraphs = paragraphsFromTextTarget(proto.paragraphs, target);
  if (!paragraphs) return false;
  proto.paragraphs = paragraphs;
  shape.data = proto;
  return true;
}

function applyShapePosition(shape, target) {
  const position =
    normalizedPosition(target.position) ||
    normalizedPosition(target.frame) ||
    normalizedPosition(target.bbox);
  if (!position) return false;
  shape.position = { ...shape.position, ...position };
  return true;
}

function applyShapeDelete(shape) {
  shape.delete();
}

function imageReplaceOptions(target) {
  const options = {};
  if (typeof target.alt === "string") options.alt = target.alt;
  if (typeof target.fit === "string") options.fit = target.fit;
  if (typeof target.prompt === "string") options.prompt = target.prompt;
  if (typeof target.imagePath === "string" && target.imagePath.trim()) {
    options.path = path.resolve(target.imagePath.trim());
  } else if (typeof target.path === "string" && target.path.trim()) {
    options.path = path.resolve(target.path.trim());
  } else if (typeof target.data === "string" && target.data.trim()) {
    options.data = target.data.trim();
  }
  return Object.keys(options).length > 0 ? options : undefined;
}

function imageSourceForAdd(target) {
  if (typeof target.imagePath === "string" && target.imagePath.trim()) {
    return { uri: path.resolve(target.imagePath.trim()) };
  }
  if (typeof target.path === "string" && target.path.trim()) {
    return { uri: path.resolve(target.path.trim()) };
  }
  if (typeof target.data === "string" && target.data.trim()) {
    return { dataUrl: target.data.trim() };
  }
  return undefined;
}

function imageFrameFromExisting(image, target) {
  const explicit =
    normalizedPosition(target.position) ||
    normalizedPosition(target.frame) ||
    normalizedPosition(target.bbox);
  if (explicit) return explicit;
  if (image?.position && typeof image.position === "object") {
    return normalizedPosition(image.position);
  }
  return undefined;
}

function applyImageReplaceByClone(slide, image, target) {
  if (!slide?.images || typeof slide.images.add !== "function") {
    return false;
  }
  const source = imageSourceForAdd(target);
  if (!source) {
    return false;
  }
  const frame = imageFrameFromExisting(image, target);
  if (!frame) {
    return false;
  }

  const fit = typeof target.fit === "string" && target.fit.trim()
    ? target.fit.trim()
    : (typeof image.fit === "string" && image.fit.trim() ? image.fit.trim() : "cover");
  const alt = typeof target.alt === "string"
    ? target.alt
    : (typeof image.alt === "string" ? image.alt : "");
  const name = typeof image.name === "string" && image.name.trim() ? image.name : undefined;
  const replacement = slide.images.add({ ...source, fit, alt, name });
  replacement.position = { ...frame };

  const crop = target.crop && typeof target.crop === "object" ? target.crop : image?.crop;
  if (crop && typeof crop === "object") {
    replacement.crop = crop;
  }

  if (typeof image.delete === "function") {
    image.delete();
  }
  return true;
}

function applyImageUpdate(image, target) {
  let changed = false;
  const replace = imageReplaceOptions(target);
  if (replace) {
    image.replace(replace);
    changed = true;
  } else if (typeof target.imageReferenceId === "string" && target.imageReferenceId.trim()) {
    image.setImageReference(target.imageReferenceId.trim());
    changed = true;
  }

  if (typeof target.alt === "string") {
    image.alt = target.alt;
    changed = true;
  }
  const width = Number(target.width ?? target.w);
  if (Number.isFinite(width) && width > 0) {
    image.width = width;
    changed = true;
  }
  const height = Number(target.height ?? target.h);
  if (Number.isFinite(height) && height > 0) {
    image.height = height;
    changed = true;
  }
  const crop = target.crop;
  if (crop && typeof crop === "object") {
    image.crop = crop;
    changed = true;
  }
  const fit = target.fit;
  if (typeof fit === "string" && fit.trim()) {
    image.fit = fit.trim();
    changed = true;
  }
  return changed;
}

function applyDelete(item) {
  if (typeof item.delete === "function") {
    item.delete();
    return true;
  }
  return false;
}

function applyTarget(slide, target) {
  const ids = targetIds(target);
  if (ids.length === 0) {
    return {
      applied: false,
      warnings: ["target_has_no_inherited_id"],
      operations: [],
    };
  }

  const warnings = [];
  const operations = [];
  let applied = false;
  const action = String(target.action || "keep").trim().toLowerCase();

  for (const id of ids) {
    const resolved = resolveTargetObject(slide, id);
    if (!resolved) {
      warnings.push(`missing_target:${id}`);
      continue;
    }

    if (action === "keep") {
      operations.push({ id, kind: resolved.kind, action: "keep" });
      continue;
    }

    if (action === "delete") {
      if (!applyDelete(resolved.item)) {
        warnings.push(`delete_not_supported:${resolved.kind}:${id}`);
        continue;
      }
      applied = true;
      operations.push({ id, kind: resolved.kind, action });
      continue;
    }

    if (resolved.kind === "shape") {
      let changed = false;
      if (["rewrite", "fill-placeholder", "replace", "rewrite-and-reposition"].includes(action)) {
        changed = applyTextUpdate(resolved.item, target) || changed;
      }
      if (action === "rewrite-and-reposition") {
        changed = applyShapePosition(resolved.item, target) || changed;
      }
      if (!changed) {
        warnings.push(`shape_action_noop:${action}:${id}`);
        continue;
      }
      applied = true;
      operations.push({ id, kind: resolved.kind, action });
      continue;
    }

    if (resolved.kind === "image") {
      if (!["replace", "rewrite", "fill-placeholder"].includes(action)) {
        warnings.push(`unsupported_image_action:${action}:${id}`);
        continue;
      }
      const cloned = applyImageReplaceByClone(slide, resolved.item, target);
      if (!cloned && !applyImageUpdate(resolved.item, target)) {
        warnings.push(`image_action_noop:${action}:${id}`);
        continue;
      }
      applied = true;
      operations.push({ id, kind: resolved.kind, action, mode: cloned ? "clone-add-delete" : "in-place" });
      continue;
    }

    if (resolved.kind === "table") {
      warnings.push(`table_edit_not_implemented:${id}`);
      continue;
    }

    warnings.push(`unsupported_target_kind:${resolved.kind}:${id}`);
  }

  return { applied, warnings, operations };
}

async function renderOutputs(presentation, slides, previewDir, layoutDir, scale, artifactToolUtils, options = {}) {
  const { padSlideNumber, saveBlobToFile } = artifactToolUtils;
  const renderPreview = options.renderPreview !== false;
  const renderLayout = options.renderLayout !== false;
  if (renderPreview) {
    await fs.mkdir(previewDir, { recursive: true });
  }
  if (renderLayout) {
    await fs.mkdir(layoutDir, { recursive: true });
  }

  const rendered = [];
  for (let index = 0; index < slides.length; index += 1) {
    const slideNumber = index + 1;
    const padded = padSlideNumber(slideNumber);
    const slide = slides[index];

    let previewPath = undefined;
    if (renderPreview) {
      previewPath = path.join(previewDir, `slide-${padded}.png`);
      const preview = await presentation.export({ slide, format: "png", scale });
      await saveBlobToFile(preview, previewPath);
    }

    let layoutPath = undefined;
    if (renderLayout) {
      layoutPath = path.join(layoutDir, `slide-${padded}.layout.json`);
      const layout = await presentation.export({ slide, format: "layout" });
      await saveBlobToFile(layout, layoutPath);
    }

    rendered.push({
      slide: slideNumber,
      previewPath,
      layoutPath,
    });
  }
  return rendered;
}

async function writeEvidence(workspaceDir, starterPptxPath, outPath) {
  const evidenceDir = path.join(workspaceDir, "template-edit-evidence");
  const evidencePath = path.join(evidenceDir, "template-edit-evidence.mjs");
  const contents = [
    "// Template-following artifact-tool evidence for fidelity scan.",
    `const starterPptxPath = ${JSON.stringify(starterPptxPath)};`,
    `const finalPptxPath = ${JSON.stringify(outPath)};`,
    "async function templateEditEvidence(PresentationFile, FileBlob) {",
    "  const presentation = await PresentationFile.importPptx(await FileBlob.load(starterPptxPath));",
    "  const pptx = await PresentationFile.exportPptx(presentation);",
    "  return { presentation, pptx, finalPptxPath };",
    "}",
    "",
  ].join("\n");
  await fs.mkdir(evidenceDir, { recursive: true });
  await fs.writeFile(evidencePath, contents, "utf8");
  return { evidenceDir, evidencePath };
}

async function main() {
  const argv = process.argv.slice(2);
  if (argv.includes("--help") || argv.includes("-h")) {
    console.log(usage());
    return;
  }

  const artifactToolUtils = await loadArtifactToolUtils();
  const { ensureArtifactToolWorkspace, importArtifactTool, parseArgs, requireArg } = artifactToolUtils;
  const args = parseArgs(argv);
  if (args.help) {
    console.log(usage());
    return;
  }

  const workspaceDir = path.resolve(requireArg(args, "workspace"));
  const starterPptxPath = path.resolve(requireArg(args, "starter-pptx"));
  const mapPath = path.resolve(requireArg(args, "map"));
  const outPath = path.resolve(requireArg(args, "out"));
  const previewDir = args["preview-dir"]
    ? path.resolve(args["preview-dir"])
    : path.join(workspaceDir, "preview", "final");
  const layoutDir = args["layout-dir"]
    ? path.resolve(args["layout-dir"])
    : path.join(workspaceDir, "layout", "final");
  const reportPath = args.report
    ? path.resolve(args.report)
    : path.join(workspaceDir, "template-edit-report.json");
  const scale = args.scale ? Number.parseFloat(args.scale) : 1;
  const renderPreview = !args["no-preview"];
  const renderLayout = !args["no-layout"];

  if (!Number.isFinite(scale) || scale <= 0) {
    throw new Error("--scale must be a positive number");
  }
  for (const writePath of [previewDir, layoutDir, reportPath, outPath]) {
    if (!isWithin(writePath, workspaceDir) && writePath !== outPath) {
      throw new Error(`Refusing to write outside workspace: ${writePath}`);
    }
  }

  await fs.mkdir(workspaceDir, { recursive: true });
  await ensureArtifactToolWorkspace(workspaceDir);

  const { FileBlob, PresentationFile } = await importArtifactTool(workspaceDir);
  const starterPptx = await FileBlob.load(starterPptxPath);
  const presentation = await PresentationFile.importPptx(starterPptx);
  const slides = slidesFromPresentation(presentation);
  const map = await readJson(mapPath);
  const outputSlides = Array.isArray(map.outputSlides) ? map.outputSlides : [];
  if (outputSlides.length !== slides.length) {
    throw new Error(`Starter slide count ${slides.length} does not match map outputSlides ${outputSlides.length}.`);
  }

  const appliedTargets = [];
  const warnings = [];
  for (const entry of outputSlides) {
    const slideIndex = Number(entry.outputSlide) - 1;
    const slide = slides[slideIndex];
    if (!slide) {
      throw new Error(`Starter slide missing for outputSlide ${entry.outputSlide}.`);
    }
    const targets = Array.isArray(entry.editTargets) ? entry.editTargets : [];
    for (const [targetIndex, target] of targets.entries()) {
      const result = applyTarget(slide, target);
      if (result.warnings.length > 0) {
        for (const warning of result.warnings) {
          warnings.push({
            outputSlide: entry.outputSlide,
            targetIndex,
            warning,
          });
        }
      }
      appliedTargets.push({
        outputSlide: entry.outputSlide,
        targetIndex,
        action: target.action || "keep",
        sourceSlide: entry.sourceSlide,
        operations: result.operations,
        applied: result.applied,
      });
    }
  }

  const rendered = await renderOutputs(
    presentation,
    slides,
    previewDir,
    layoutDir,
    scale,
    artifactToolUtils,
    { renderPreview, renderLayout },
  );
  await fs.mkdir(path.dirname(outPath), { recursive: true });
  const pptx = await PresentationFile.exportPptx(presentation);
  await pptx.save(outPath);
  const evidence = await writeEvidence(workspaceDir, starterPptxPath, outPath);

  const report = {
    workspace: workspaceDir,
    starterPptx: starterPptxPath,
    mapPath,
    output: outPath,
    slideCount: slides.length,
    renderPreview,
    renderLayout,
    appliedTargets,
    warnings,
    rendered,
    evidence,
  };
  await writeJson(reportPath, report);
  console.log(JSON.stringify({ reportPath, output: outPath, warnings: warnings.length }, null, 2));
}

main().catch((error) => {
  console.error(error.stack || error.message || String(error));
  console.error(usage());
  process.exit(1);
});
