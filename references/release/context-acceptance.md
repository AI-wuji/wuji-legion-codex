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
./bin/wuji.exe route --query "<query>" --context-artifact "<artifact_path>"
```

Acceptance: every result remained within its hard byte budget, emitted ranked line ranges rather than whole files, and produced a `wuji-context://sha256/...` handle plus a verifiable artifact. Loading independently rejects a changed query fingerprint, falsified byte count, workspace escape, excerpt tampering, or changed source file. Code routing without the artifact remains on Aji; a matching artifact below 4096 bytes unlocks only the independent Terra implementation branch. Task contracts above 2048 bytes, total replay above 8192 bytes, and parent-context affinity also remain on Aji.

Boundary: model choice is a routing policy (`gpt-5.6-sol` for Aji, Terra for bounded independent implementation, Luna for bounded mechanical extraction). Sol, Terra, and Luna are treated as separate cache domains. This repository can prove handoff bytes and execution receipts, but cannot prove provider-side billing or cache accounting.
