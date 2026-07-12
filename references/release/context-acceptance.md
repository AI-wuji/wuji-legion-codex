# Context Acceptance

Date: 2026-07-12

Scope: deterministic repository context selection. This is not host token telemetry.

| Query | Budget | Selected | Files scanned | Excerpts |
|---|---:|---:|---:|---:|
| architecture routing evolution boundaries | 4096 B | 4071 B | 97 | 2 |
| presentation slidev behavior verification | 4096 B | 4008 B | 97 | 2 |
| security release audit credentials dependencies | 12288 B | 12238 B | 97 | 7 |

Reproduce with:

```powershell
./bin/wuji.exe context-select --workspace . --query "<query>" --max-bytes <budget>
```

Acceptance: every result remained within its hard byte budget, emitted ranked line ranges rather than whole files, and reported the policy `rank before read`, `hard byte budget`, and `keep raw logs out of context`. The full audit independently measured `context_selected_bytes=4039` for its 4096-byte fixture.

Boundary: model choice is a routing policy (`gpt-5.6-sol` for Aji, Terra for bounded implementation/verification, Luna for mechanical extraction). This repository cannot prove provider-side model billing or token accounting.
