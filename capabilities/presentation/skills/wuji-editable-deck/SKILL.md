---
name: wuji-editable-deck
description: Create or edit editable PPTX or imported Google Slides decks. Keep text, shapes, charts and tables as native objects and verify the rendered result.
---

# Wuji Editable Deck

Deliver an editable deck, not full-slide screenshots. For edits, inspect the source/template, retain native objects and locked regions, and change only requested slots. For a new deck, identify audience, page roles and density before composition; reuse the selected template and native composition tools. Use conversion/animation atoms only when the requested output needs them.

Select the `editable-pptx` entry from `../../assets/template-catalog.json`; use the route's hashed asset and invocation contract. Do not expose source-project selection to the user or substitute an image-only deck for an editable PPTX.

## Distilled PPTX Contract

Keep text, shapes, charts and tables as native editable objects; preserve requested template fidelity; render every slide; and report overflow, broken assets or unverified output as incomplete.

Render every slide and check content, overflow, object editability and template fidelity. A successful file write alone is not completion. Keep the smallest correct workflow; don't require a staged plan for a one-slide edit.

For template intake, reconstruction, chart-data sync or advanced native animation, read [the detailed PPTX contract](references/editable-pptx-contract.md) only when relevant.
