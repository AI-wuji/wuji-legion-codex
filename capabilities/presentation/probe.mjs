import fs from "node:fs";
import path from "node:path";
import { pathToFileURL } from "node:url";
import { Presentation, PresentationFile } from "@oai/artifact-tool";

const [layoutRoot, outputPath] = process.argv.slice(2);
if (!layoutRoot || !outputPath) throw new Error("layout root and output path are required");
const entry = path.join(layoutRoot, "artifact-tool-compose", "index.mjs");
const { builders } = await import(pathToFileURL(entry).href);
if (!Array.isArray(builders) || builders.length < 20) throw new Error("complete layout builder library not retained");
const deck = Presentation.create({ slideSize: { width: 1280, height: 720 } });
builders.slice(0, 2).forEach((build) => build(deck));
const file = await PresentationFile.exportPptx(deck);
await file.save(outputPath);
if (!fs.existsSync(outputPath) || fs.statSync(outputPath).size < 10_000) throw new Error("PPTX artifact was not created");
console.log(JSON.stringify({ outputPath, sourceBuilders: builders.length, generatedSlides: 2 }));
