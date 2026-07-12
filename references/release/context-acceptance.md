# Context Acceptance

Date: 2026-07-12

Scope: deterministic repository context selection. This is not host token telemetry.

| Query | Budget | Selected | Files scanned | Excerpts |
|---|---:|---:|---:|---:|
| architecture routing evolution boundaries | 4096 B | 3881 B | 110 | 6 |
| presentation slidev behavior verification | 4096 B | 2925 B | 110 | 12 |
| security release audit credentials dependencies | 12288 B | 12099 B | 110 | 12 |
| fix code workerPlan in internal/core/route.go | 2048 B | 1827 B | 110 | 3 |

Reproduce with:

```powershell
./bin/wuji.exe context-select --workspace . --query "<query>" --max-bytes <budget>
./bin/wuji.exe route --query "<query>" --context-artifact "<artifact_path>"
```

Acceptance: every result remained within its hard byte budget, emitted ranked line ranges rather than whole files, and produced a deterministic `WUJI_CONTEXT_CAPSULE_V1` payload, SHA-256, `wuji-context://sha256/...` handle, and verifiable artifact. Loading independently rejects a changed query fingerprint, falsified payload/byte count, quality metadata, workspace escape, excerpt tampering, or changed source file. Generated dependency locks cannot enter the retrieval set, and an explicit code path outranks repeated noise. Code routing without the artifact remains on Aji; a matching artifact below 4096 bytes unlocks only the independent Terra implementation branch when retrieval coverage is at least 60% and at least one code excerpt exists. Real JSON task contracts above 2048 bytes, total replay above 8192 bytes, and parent-context affinity also remain on Aji.

Boundary: model choice is a routing policy (`gpt-5.6-sol` for Aji, Terra for bounded independent implementation, Luna for bounded mechanical extraction). Sol, Terra, and Luna are separate cache domains. Worker prompts use a stable model-local prefix plus deterministic payload and task contract; Terra/Luna may fall directly back to Sol only before generation and make at most two attempts. Execution receipts must report input, cached input and output tokens, retry/failure details, cache domain, and Aji acceptance. This repository can prove handoff bytes and receipt requirements, but provider-side billing and cache accounting remain provider evidence.
