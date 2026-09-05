---
name: wuji-editable-deck
description: Unified editable presentation Skill combining native artifact-tool PPTX, PowerPoint template following, PPT Master conversion and animation utilities, Huashu HTML-to-PPTX, Baoyu style systems, and narrative QA. Use whenever the deliverable must be an editable PowerPoint or imported Google Slides deck.
---

# Wuji Editable Deck

Treat this as one Skill. Never expose upstream project selection to the user.

## Unified Pipeline

1. Run the shared Humanize planning layer for narrative, speaker intent, media slots, and slide-plan QA when the deck has meaningful structure. Require a staged plan before composition: audience goal, page role, density budget, and critic checks.
2. Use native artifact-tool composition as the default editable PPTX base.
3. Internally use PPT Master atoms for SVG/PPTX conversion, animation XML, template import, image analysis, and visual review.
4. Internally use Huashu HTML-to-PPTX when an approved HTML composition must become editable PowerPoint objects.
5. Use Baoyu, Dashi layout doctrine, Guizang visual/layout checks, Codex Grid, elite brand styles, and retained examples through the unified catalog, not as separate workflows. Image-first sources may supply style or image slots only; they do not change the editable delivery contract.

Read `../../assets/template-catalog.json`; filter `scenario=editable-pptx`. Rebuild it with `../../../../scripts/build-presentation-catalog.ps1` after source upgrades.

## Distilled PPTX Contract

- Treat a source `.pptx` as an editable object graph, not a page image. Preserve native text, shapes, charts, tables, relationships, and template geometry whenever the request is an edit or template-following task.
- Treat imported templates as unknown until inspected. Build a slot map with page roles, object types, coordinates, type scale, capacity, and locked/content-owned regions. Never overwrite a locked region silently.
- Apply non-destructive template editing: keep the original template untouched, address edits by explicit slide/slot identity, and write a new output artifact. A `max_chars` or capacity value is a warning budget, not permission to truncate, add ellipses, or hide overflow.
- For each slide, bind copy to a layout/fill plan. Every visible placeholder must be intentionally overwritten or intentionally retained; reject blank template copy, duplicate layouts, incompatible aspect-ratio images, and layout reuse that makes the deck visually monotonous.
- Use the retained PPT Master atoms for PPTX intake, template analysis, SVG/DrawingML conversion, and visual review. Use its native-enhancement boundary for notes, timings, transitions, and relationships; enhancement is append-oriented and must not regenerate source slides.
- Generator-style output may use the native artifact-tool equivalent of the PptxGenJS contract (slide, text, shape, image, table, chart, deterministic `writeFile` export), but it must pass the same object-editability and render checks.
- Keep image-only and editable outputs separate. A full-slide image, Baoyu-style image deck, or rasterized background can be a deliberate visual layer, but it cannot be reported as editable text/shapes. When reconstructing an image-first design, use a manifest and coordinate contract, then verify the reconstructed native objects against the rendered reference.
- If native chart data is changed, update the chart data source and the visible chart together. Do not fake a chart with a screenshot when the requested result is editable.
- Measure text and paginate before export. A successful file write is insufficient when overflow, clipped text, or rasterized full-slide output remains.

## Completion

Render every slide, inspect full-size output, run overflow and capacity checks, verify editable text/shapes/charts, preserve template fidelity when editing, and deliver one final PPTX. Keep the evidence chain: plan -> composition -> static checks -> render -> critic/visual comparison -> final hash. Browser motion may accompany the deck only when explicitly requested; it never substitutes for the editable deliverable.
