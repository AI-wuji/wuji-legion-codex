# Detailed editable PPTX contract

- Treat an input `.pptx` as an object graph. Inspect slide roles, object types, coordinates, typography, capacity and locked/content-owned regions. Preserve original source and write a new output. Capacity is a warning, not permission to truncate, hide or ellipsize content.
- Plan each visible placeholder's content or deliberate retention. Avoid duplicate layouts, bad aspect ratios and repetitive compositions. For a meaningful multi-slide narrative, check audience goal, page role, media slots and density before layout.
- Native editable composition is the default. Use retained PPT Master atoms for SVG/PPTX conversion, template import, notes, timing, relationships and visual review. Use Huashu HTML-to-PPTX only for approved editable conversion. Baoyu, Dashi and Guizang may inform style and QA, not replace native output.
- Keep image-only and editable layers distinct. A full-slide raster may be an intentional background but cannot be reported as editable text or shapes. Reconstruction from an image needs coordinates and native-object comparison against the rendered reference.
- Synchronize chart data sources and visible charts. Measure text and paginate before export. Verify native shapes, text, tables and charts after export.
- Render every slide at full size; check overflow, clipping, content, template fidelity and editability. Record a minimal trace from composition through checks, render and final artifact hash. Optional browser motion never substitutes for the editable deck.
