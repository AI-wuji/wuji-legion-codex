---
name: wuji-editable-deck
description: Unified editable presentation Skill combining native artifact-tool PPTX, PowerPoint template following, PPT Master conversion and animation utilities, Huashu HTML-to-PPTX, Baoyu style systems, and narrative QA. Use whenever the deliverable must be an editable PowerPoint or imported Google Slides deck.
---

# Wuji Editable Deck

Treat this as one Skill. Never expose upstream project selection to the user.

## Unified Pipeline

1. Run the shared Humanize planning layer for narrative, speaker intent, media slots, and slide-plan QA when the deck has meaningful structure.
2. Use native artifact-tool composition as the default editable PPTX base.
3. Internally use PPT Master atoms for SVG/PPTX conversion, animation XML, template import, image analysis, and visual review.
4. Internally use Huashu HTML-to-PPTX when an approved HTML composition must become editable PowerPoint objects.
5. Use Baoyu, Codex Grid, elite brand styles, and retained examples through the unified catalog, not as separate workflows.

Read `../../assets/template-catalog.json`; filter `scenario=editable-pptx`. Rebuild it with `../../../../scripts/build-presentation-catalog.ps1` after source upgrades.

## Completion

Render every slide, inspect full-size output, run overflow checks, verify editable text/shapes/charts, preserve template fidelity when editing, and deliver one final PPTX. Browser motion may accompany the deck only when explicitly requested; it never substitutes for the editable deliverable.
