# Visual Unit

Mirror source: `kernel-source.json`

## Role

Visual unit turns content into visible deliverables. It does not become a separate design system and does not bypass Codex, Go gates, 质检, white-hat, or guard-office.

Covered deliverables:

- editable PPTX
- browser HTML decks
- web and app UI prototypes
- charts, diagrams, infographics, covers, and supporting visuals
- motion-first visual demos when explicitly needed

Single image generation remains routed through `imagegen` when the user mainly wants one image.

## Distilled Huashu Design Atoms

`huashu-design` is absorbed only as source-level atoms:

- `html-native-design-canvas`: use HTML/CSS as a high-fidelity design canvas for UI, web, and HTML decks.
- `brand-asset-protocol`: when real brands, products, people, or venues matter, use official assets, colors, screenshots, and evidence handles before designing.
- `anti-ai-slop-visual-rules`: block generic AI-looking gradients, purposeless cards, vague icons, weak hierarchy, and stock-like composition.
- `design-direction-triad`: when the visual brief is ambiguous, produce three concrete directions before expanding the full artifact.
- `html-deck-to-editable-pptx`: for from-scratch high-fidelity decks, allow HTML-first design and then editable PPTX conversion.
- `motion-stage-sprite-engine`: mount only when the task asks for animation, video, dynamic deck behavior, or staged motion.

These atoms replace weaker local habits. They are not an installed `huashu-design` command, not a second route owner, and not a separate runtime.

## Modes

| Mode | Use When | Hard Gate |
|---|---|---|
| Editable PPTX | Existing deck continuation, template pages, from-scratch PowerPoint | Must be editable; HTML screenshots cannot pretend to be PPTX |
| Interactive PPTX | Buttons, links, menus, zoom, morph, state switching | Must use native PowerPoint mechanisms where the deliverable is PPTX |
| HTML deck | Browser-native presentation, courseware, live visual storytelling | Must preview in a real browser |
| UI prototype | Web, app, component, dashboard, landing page, tool surface | Must verify desktop and mobile states |
| Data visual | Charts, reports, comparisons, structured evidence | Conclusion must be readable at a glance |
| Visual narrative | Cover, infographic, poster, supporting illustration | Image service must support information, not hide weak structure |
| Motion demo | Dynamic presentation, animated scene, product/video explainers | Must provide a motion plan and a previewable source |

## Main Chain

```text
deliverable type
-> design-brief
-> asset/source check when external material matters
-> layout-plan
-> optional design-direction-triad when brief is ambiguous
-> native implementation
-> preview
-> 质检 and white-hat review
```

## Design Brief Requirements

The `design-brief` should be short and executable:

- audience and use case
- density mode
- visual direction
- brand/source constraints
- asset strategy
- interaction or motion requirements
- forbidden visual habits

Do not expand the brief into a long doctrine body. It is a working constraint, not resident context.

## Anti AI Slop Rules

Reject visual output that relies on:

- purple-blue gradients as a default personality
- generic SaaS cards without information structure
- decorative icons that do not clarify meaning
- full-page screenshots or bitmaps pretending to be editable deliverables
- stock-like dark blurred backgrounds when the user needs to inspect real content
- hero-scale text inside compact panels
- style words without layout, density, hierarchy, and asset decisions

## GitHub Trending Source Pool Mapping

- `taste-skill` strengthens `anti-ai-slop-visual-rules`: prefer concrete hierarchy, density, contrast, asset fit, and composition checks over vague style adjectives.
- `opencv-style` strengthens preview inspection: screenshots, exports, or image assets may be checked for blank frames, low contrast, occlusion, and obvious visual regression when relevant.
- These are source-pool lessons only. They do not install a separate design shell or add a new visual atom.

## 2026-06-10 Visual Source Pool Merge

- `Huashu Design 2.0` strengthens `html-native-design-canvas`, `design-direction-triad`, `brand-asset-protocol`, and `motion-stage-sprite-engine`.
- `baoyu-article-illustrator` strengthens article illustration planning: choose illustration type, style, palette, and placement before generating images.
- `ian-xiaohei-illustrations` is a style preset for Chinese article illustrations only when that look fits the brief; it is not a default house style.
- `baoyu-image-gen` strengthens prompt batching and reference-image specification, but does not replace Codex `imagegen`.
- `ppt-keynote` strengthens HTML/Keynote-style visual exploration and `html-deck-to-editable-pptx`; it cannot be used to fake editable PPTX.

All five are source pools. They may replace weak visual habits, but they cannot become a second design system, a resident prompt body, or an independent visual commander.

## Brand Asset Protocol

When a real brand/product/place/person is part of the task:

- intelligence-profile gathers official or high-confidence sources.
- guard-office checks external pages, repos, scripts, downloads, and asset sources before execution.
- visual-profile uses evidence handles and concise summaries, not raw asset dumps.
- compliance-on-demand enters only when license, attribution, publication, or release boundaries matter.

## PPTX Boundary

- Existing `.ppt/.pptx` or template-following work follows [pptx_master.md](E:/wuji-projects/wuji-legion-codex/units/pptx_master.md).
- From-scratch high-fidelity decks may use HTML-first design, then `html-deck-to-editable-pptx`.
- HTML screenshots cannot fake editable PPTX.
- PowerPoint COM/MCP belongs to final refinement only, not the main visual generator.

## 质检 Handoff

Before delivery:

- 质检 checks preview, overflow, readability, editability, mobile/desktop where relevant, and visual coherence.
- White-hat checks false fusion, fake completion, and whether token-saving removed necessary evidence.
- Guard-office checks any external source material or dependency used by the visual chain.

## Forbidden

- installing or invoking `huashu-design` as an independent commander
- using HTML to fake editable PowerPoint
- skipping preview and claiming visual completion
- copying unreviewed external assets into deliverables
- turning design directioning into a pause loop when the task is already specified
- adding decorative complexity that raises token/tool cost without improving first-pass success
