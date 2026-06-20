# PPTX Master

Mirror source: `kernel-source.json`

## Role

PPTX Master is the only main route for PowerPoint delivery.

Its job is to produce a natively editable `.pptx`, not an HTML deck pretending to be PowerPoint.

Default posture:

- use a native PPTX-first route
- preserve editability as the primary contract
- treat the deck as a design draft with real PowerPoint objects
- keep Go gates and final acceptance outside the generator core

## Upstream replacement decision

Wuji Legion now treats `ppt-master` as the stronger source pool for the PPT main chain.

What is absorbed:

- native editable PPTX as the canonical output
- template as structure-plus-style bundle, not just color skin
- free-design first, template only when explicitly given
- reusable page skeleton thinking
- draft-first mindset: AI removes blank-page work, final polish stays editable in PowerPoint

What is rejected as the default road:

- HTML-first as the normal PPT path
- screenshot-like conversion posing as native PPT authoring
- using browser-rendered layout as the main generator for editable PPTX

## Main chain

```text
source material
-> structure planning
-> design spec
-> page skeleton strategy
-> native editable PPTX authoring
-> preview evidence
-> pptx-audit
-> quality-inspection
```

## Core rules

- Existing `.ppt/.pptx` continuation stays native PPTX-first.
- From-scratch PPT also stays native PPTX-first.
- HTML deck is a separate browser deliverable, not the default PPT road.
- COM refine is final-mile polish only.
- Go gates are gates only, not the main authoring engine.

## What stays from local Wuji

These atoms still strengthen the new main chain:

- `brand-asset-protocol`
- `anti-ai-slop-visual-rules`
- `design-direction-triad`
- `terminal-real-run-verification`
- `pptx-preflight`
- `pptx-batch-gate`
- `pptx-audit`

## Template posture

Template means a reusable structure-and-style bundle.

Rules:

- template use must be explicit
- free design is still allowed when no template is given
- template-following work must not rebuild from a fake blank shell
- fixed-role pages such as cover, TOC, section divider, summary, and ending must respect the template roster

## Hard failures

Fail the route if any of these happens:

- HTML or browser screenshots are used as the primary PPT authoring road
- a slide is delivered as a flat page image while claiming full editability
- template-following is done by visual imitation instead of native continuation
- COM polish turns into the real generator
- Go gate turns into the real generator

## Acceptance focus

Quality inspection checks:

- editable in PowerPoint
- no placeholder residue
- no fake buttons
- no overflow or unreadable density
- visual system matches the requested tone or template
- real PPTX evidence exists, not only explanation
