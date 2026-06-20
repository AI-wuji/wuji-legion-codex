# Development Mirror

Mirror source: `kernel-source.json`

## Scope

Development handles ordinary software implementation. It does not self-certify review, release, or independent oversight.

## Core Flow

`read surface -> lock acceptance -> inspect impact -> search prior art -> implement -> verify -> hand to oversight if triggered`

## Rules

- Search prior art before inventing from scratch when uncertain.
- Lock the acceptance target before broad implementation.
- Inspect affected surface before editing cross-file behavior.
- Prefer first-pass success over fast partial output.
- Use `root-cause-radar` before repairing diagnosable failures.
- Use `patch-debt-root-cure` to remove workaround chains instead of stacking another patch.
- Use code-structure memory only as a guarded, read-only, on-demand accelerator
  for impact mapping. `codebase-memory-mcp` may inform `code-map-before-edit`
  only after source/build, secret-filter, stale-index, and benchmark gates pass.
- Use loop discipline only when bounded: explicit budget, dedupe, stop
  condition, failure evidence, and real verification. No autonomous or
  unbounded coding loop may become a second execution chain.
- Keep execution inside the same scoped goal until verification is complete or truly blocked.
- For page or screen changes, use `target-page-in-place-replacement`: modify the active target surface, migrate the real route/imports/tests/assets, remove superseded page files, and verify the actual opened route. Do not create a parallel v2, compatibility page, wrapper, or duplicate route unless the user explicitly asks for a fallback.
- Use `active-route-entrypoint-verification` before claiming a page replacement is complete.

## Distilled Atoms

- `version-doc-mcp`
- `disciplined-debug-loop`
- `root-cause-radar`
- `parallel-hypothesis-fanout`
- `patch-debt-root-cure`
- `terminal-real-run-verification`
- `code-map-before-edit`
- `target-page-in-place-replacement`
- `active-route-entrypoint-verification`

## Verification

Use the smallest verification tier that preserves first-pass success and evidence. Real browser, command, local program, export, or artifact checks are preferred when available.
