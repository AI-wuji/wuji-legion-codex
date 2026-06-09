# HTML Slides Master

Mirror source: `kernel-source.json`

## Role

HTML Slides Master produces browser-native presentation artifacts. It does not pretend to be PowerPoint and does not become an external design runtime.

Use it for:

- high-fidelity HTML decks
- live browser presentations
- courseware or product demos that need motion, interaction, or a strong visual stage
- first-pass visual exploration before an editable PPTX conversion

Do not use it for:

- continuing an existing PowerPoint template
- pretending HTML screenshots are editable PPTX
- avoiding the Presentations plugin when the user explicitly needs true PowerPoint editing

## Distilled Sources

From `huashu-design`, only these atoms remain:

- `html-native-design-canvas`
- `design-direction-triad`
- `anti-ai-slop-visual-rules`
- `brand-asset-protocol`
- `html-deck-to-editable-pptx`
- `motion-stage-sprite-engine`

From prior HTML/PPT skills, keep:

- fixed 16:9 stage when the artifact is a deck
- browser preview as acceptance evidence
- notes separated from audience-visible content
- density modes for speaker-led and reading-first decks
- optional conversion to editable PPTX only when the route is from-scratch

## Main Chain

```text
content input
-> audience + purpose + density
-> design-brief
-> 3 visible directions when ambiguous
-> selected direction
-> section outline
-> full HTML deck
-> browser preview
-> 质检
-> optional editable PPTX conversion
```

## Direction Policy

When the user gives only a vague visual request, use `design-direction-triad`:

- one safe direction
- one bold direction
- one wildcard direction

If the user has already locked a clear style, skip directioning and execute.

## Stage Rules

- A deck uses a fixed 16:9 slide stage.
- A slide must fit in one viewport without internal scrolling.
- Speaker notes stay outside audience-visible text.
- Mobile preview can scale the stage, but must not rearrange slide content into a different document.

## Motion Rules

Mount `motion-stage-sprite-engine` only when motion is required:

- animated presentation
- live demo
- video/GIF export
- dynamic dashboard or product reveal

Motion plan must name:

- stage objects
- timing rhythm
- interaction or autoplay behavior
- acceptable static degradation if exporting to editable PPTX

## Anti AI Slop Rules

Reject:

- generic purple gradient deck defaults
- decorative card grids without narrative structure
- icons used as filler
- unreadable tiny text
- repeated hero sections inside every slide
- template-like pages with no point of view
- animation that does not clarify sequence or attention

## 质检 Checklist

- browser preview exists
- fixed stage renders correctly
- no overflow, overlap, or hidden content
- notes are separated
- style is explicit and coherent
- motion, if required, is visible in the live artifact
- editable PPTX conversion, if used, is not misrepresented as preserving HTML animation
