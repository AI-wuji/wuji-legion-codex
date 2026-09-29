import fs from "node:fs";
import { Presentation, PresentationFile } from "@oai/artifact-tool";

const [outputPath] = process.argv.slice(2);
if (!outputPath) throw new Error("output path is required");

const deck = Presentation.create({ slideSize: { width: 1280, height: 720 } });
for (let index = 0; index < 2; index += 1) {
  const slide = deck.slides.add();
  slide.background.fill = index === 0 ? "#F8FAFC" : "#FFF7ED";

  const title = slide.shapes.add({
    geometry: "textbox",
    position: { left: 72, top: 48, width: 1136, height: 64 },
    fill: "none",
    line: { fill: "none", width: 0 },
  });
  title.text = `Wuji presentation probe ${index + 1}`;
  title.text.style = {
    typeface: "Aptos",
    fontSize: 32,
    bold: true,
    color: "#142735",
    autoFit: "none",
  };

  const surface = slide.shapes.add({
    geometry: "roundRect",
    position: { left: 72, top: 160, width: 520, height: 360 },
    fill: index === 0 ? "#FFFFFF" : "#FFFBEB",
    line: { style: "solid", fill: "#CBD5E1", width: 1 },
    borderRadius: "rounded-2xl",
  });
  surface.text = "Native editable surface";
  surface.text.style = {
    typeface: "Aptos",
    fontSize: 24,
    bold: true,
    color: "#1E293B",
    autoFit: "none",
  };

  const detail = slide.shapes.add({
    geometry: "textbox",
    position: { left: 112, top: 260, width: 440, height: 150 },
    fill: "none",
    line: { fill: "none", width: 0 },
  });
  detail.text = "This slide contains native text and shape objects.";
  detail.text.style = {
    typeface: "Aptos",
    fontSize: 22,
    color: "#475569",
    autoFit: "none",
  };

  const badge = slide.shapes.add({
    geometry: "roundRect",
    position: { left: 720, top: 210, width: 360, height: 96 },
    fill: index === 0 ? "#DBEAFE" : "#FED7AA",
    line: { style: "solid", fill: "#93C5FD", width: 1 },
    borderRadius: "rounded-full",
  });
  badge.text = "Editable object verified";
  badge.text.style = {
    typeface: "Aptos",
    fontSize: 22,
    bold: true,
    color: "#1E3A8A",
    autoFit: "none",
  };
}

const file = await PresentationFile.exportPptx(deck);
await file.save(outputPath);
if (!fs.existsSync(outputPath) || fs.statSync(outputPath).size < 10_000) throw new Error("PPTX artifact was not created");
console.log(JSON.stringify({ outputPath, generatedSlides: 2, nativeObjectsPerSlide: 4 }));
