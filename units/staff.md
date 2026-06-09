# Staff Runtime Mirror

Mirror source: `kernel-source.json`

## Role

Staff runtime owns layer 1 and coordinates layer 2:

1. `task-routing`
2. `capability-mount`

It is the only route owner in the fused kernel.

## Layer 1

`task-routing` decides only:

- state
- owner profile
- oversight chain
- closeout policy
- execution budget

## Execution Budget

Staff runtime applies `execution_budget_contract` before expanding scope:

- `FAST_REPLY`: direct answer or discussion; no tool gate unless evidence is needed.
- `LIGHT_TASK`: small scoped owner task; no full-legion scan or full-suite run.
- `STRUCTURAL_TASK`: router, kernel, officer, gate, multi-file, or root-cause work; targeted gates first and at most one full final verification when required.
- `RELEASE_TASK`: explicit full scan, broad cleanup, release, or final completion claim; full audit once at final.

Routing must bind current scope before expansion. Officer sidecars stay on demand and exit after merge.

## Layer 2

`capability-mount` decides only:

- which distilled atoms to mount
- which plugin or MCP surface is justified
- which cold capability stays cold
- when `parallel-hypothesis-fanout` is justified as need-driven sidecar evidence

Mount policy is always `minimal-gap-first`.

## Distilled atoms

Staff runtime mounts `distilled_atoms` only after owner selection.

It must prefer gap-fill or replacement over stacking:

- mount only the missing capability
- keep low-frequency atoms cold
- keep `source_lineage_atoms` separate from fused Wuji `distilled_atoms`

## Forbidden

Staff runtime must not:

- become a second execution engine
- keep OpenSquilla as a visible external system
- stack abilities just because they exist
- reopen finished work with management ceremony

## Optimization posture

Routing quality is measured by:

- fewer unnecessary mounts
- lower resident context
- higher first-pass success
- lower total rework cost

Parallel sidecars may inspect competing causes or disjoint slices concurrently without a Wuji-imposed numeric cap, but staff runtime keeps one main route, one merge decision, and one closeout verdict. Close each sidecar immediately after its result is merged.
