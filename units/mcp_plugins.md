# MCP and Plugin Mirror

Mirror source: `kernel-source.json`

## Role

MCP and plugins belong to layer 2 only: `capability-mount`.

They are tools, not route owners.

## Admission order

Use this order:

1. owner native ability
2. local tool or Go gate
3. approved plugin
4. approved MCP
5. on-demand discovery only if a real gap remains

There is no standing MCP candidate shelf anymore. MCP review is task-scoped and ad hoc.

## Rules

- do not mount by default just because something is installed
- do not let MCP or plugins become a second router
- do not let external OpenSquilla surfaces reappear as runtime authority
- networked or high-risk surfaces stay gated
- host availability does not equal legion admission
- only plugins explicitly admitted by the main chain may appear in route `plugin_candidates`

## Distilled posture

Only keep the parts that improve:

- task completion
- evidence quality
- lower total token cost
- lower rework

Anything that mainly adds orchestration noise should stay out.

## Huashu Design Boundary

`huashu-design` is not mounted as a plugin or MCP server.

Only its distilled visual atoms may be mounted by `visual-profile`:

- `html-native-design-canvas`
- `brand-asset-protocol`
- `anti-ai-slop-visual-rules`
- `design-direction-triad`
- `native-pptx-master-route`
- `motion-stage-sprite-engine`

Any external assets, scripts, npm packages, repositories, or browser automation brought in for those atoms must pass guard-office first.
