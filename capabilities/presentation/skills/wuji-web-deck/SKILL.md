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

## Unified Assets

Read `../../assets/template-catalog.json`; filter `scenario=web-deck` and choose the smallest matching template/component set. The catalog merges HTML themes/layouts/effects, Slidev components, presenter assets, and stage components while retaining source provenance.

Materialize fluid with `../../../../scripts/materialize-stage-fluid.ps1`. Rebuild the catalog with `../../../../scripts/build-presentation-catalog.ps1` after source upgrades.

## Completion

Build the real active deck, open it in a browser, verify 16:9 framing, navigation, presenter behavior, interactive components, nonblank pixels, and console errors. Keep source files editable and remove superseded pages.
