---
name: wuji-web-deck
description: Unified browser-native presentation Skill combining HTML deck themes, animations, presenter mode, Slidev Markdown/Vue components, interactive stage fluid, and narrative planning. Use for dynamic HTML presentations, developer talks, browser playback, presenter views, interactive components, or web-first slide experiences.
---

# Wuji Web Deck

Treat this as one Skill. Never ask the user to choose an upstream project.

## Scenario Routing

- Use the HTML deck runtime for theme-driven browser slides, presenter mode, reorder/drag workflows, lightweight animation, and export-ready standalone decks.
- Use the Slidev runtime internally for Markdown authoring, Vue components, live code, developer talks, and component-heavy interaction.
- Mount stage fluid as a component inside the real 16:9 stage, never as a second page or whole-window effect.
- Run the shared Humanize planner for multi-slide narrative, speaker intent, media slots, and QA. It is a layer, not a separate output engine.
- Govern layouts through a registry and explicit slide metadata. A slide declares its layout, content slots, image aspect expectations, and density budget; free-form HTML is allowed only outside governed layouts.

## Unified Assets

Read `../../assets/template-catalog.json`; filter `scenario=web-deck` and choose the smallest matching template/component set. The catalog merges HTML themes/layouts/effects, Slidev components, presenter assets, and stage components while retaining source provenance.

Materialize fluid with `../../../../scripts/materialize-stage-fluid.ps1`. Rebuild the catalog with `../../../../scripts/build-presentation-catalog.ps1` after source upgrades.

## Distilled Web Presentation Contract

- Keep one real 16:9 stage. Keyboard navigation, stable hash/deep links, presenter behavior, and export-ready source are part of the deck contract, distilled from reveal.js and Slidev practice.
- Validate the layout registry before rendering: every declared layout exists, every visible copy slot is filled or explicitly marked intentional, `contentLocked` layouts are rejected or replaced, and a multi-slide deck does not silently reuse one layout for every page.
- Validate image slots by aspect ratio and object-fit policy. Do not stretch media to fill a slot; crop or replace it according to the declared slot contract and record the choice in the deck manifest.
- Keep typography and density measurable: use `fillPlan`, `maxChars`, visible item counts, page roles, and type scale as preflight constraints. These are capacity signals, never truncation instructions.
- Markdown/Vue/component-heavy authoring may use the retained Slidev runtime internally. Do not expose Slidev, reveal.js, or Marp as competing user-facing routes.
- Markdown-to-PPTX or HTML-to-PPTX exports are accepted only when the resulting artifact remains editable and passes the shared PPTX object and render checks; a screenshot or flattened page is not an editable deliverable.
- Browser verification must include nonblank pixels, navigation/presenter behavior where applicable, no console errors, and a responsive 16:9 frame. Large or animated decks must not block the browser interaction loop. When motion is unavailable, the deck must still render a complete static fallback.

## Completion

Build the real active deck, open it in a browser, verify 16:9 framing, navigation, presenter behavior, interactive components, nonblank pixels, and console errors. Keep source files editable and remove superseded pages.
