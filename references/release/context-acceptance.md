# Context Acceptance

Date: 2026-07-12

Scope: deterministic repository context selection. This is not host token telemetry.

| Query | Budget | Selected | Files scanned | Excerpts |
|---|---:|---:|---:|---:|
| architecture routing evolution boundaries | 4096 B | 3881 B | 110 | 6 |
| presentation slidev behavior verification | 4096 B | 2925 B | 110 | 12 |
| security release audit credentials dependencies | 12288 B | 12099 B | 110 | 12 |
| fix code workerPlan in internal/core/route.go | 4096 B | 1827 B | 110 | 3 |

Reproduce with:

```powershell
./bin/wuji.exe context-select --workspace . --query "<query>" --max-bytes <budget>
./bin/wuji.exe route --query "<query>" --context-artifact "<artifact_path>"
```

Acceptance: every result remained within its hard byte budget, emitted ranked line ranges rather than whole files, and produced a deterministic `WUJI_CONTEXT_CAPSULE_V1` payload, SHA-256, `wuji-context://sha256/...` handle, and verifiable artifact. Loading independently rejects a changed query fingerprint, falsified payload/byte count, quality metadata, workspace escape, excerpt tampering, or changed source file. Generated dependency locks cannot enter the retrieval set, and an explicit code path outranks repeated noise. Code routing without the artifact remains on Aji; a matching artifact below 4096 bytes unlocks only the independent host-selected implementation branch when retrieval coverage is at least 60% and at least one code excerpt exists. Real JSON task contracts above 4096 bytes, total replay above 9216 bytes, and parent-context affinity also remain on Aji.

Boundary: Aji is the sole user-facing communicator, requirement-graph maintainer, PonyTail judgment center, and final reporter. The project uses the Codex host-selected model; Luna/Terra/Sol are reasoning roles mapped to low/medium/max, with high-reasoning availability fallback max -> xhigh on the same host model before generation only. Every non-pure-conversation task uses deterministic General Staff state and scheduling, not a resident model worker; staff never executes task work, writes artifacts, merges results, or accepts completion. Task execution nodes use their declared reasoning role and availability chain, select once before generation, and keep a sticky session after generation begins. Non-trivial solution tasks emit one ordered search preflight capped at 3 sources and 90 seconds; it must finish before execution workers, and changed evidence invalidates the stale plan. Worker prompts use a stable model-local prefix plus deterministic payload and task contract. Schema v2 execution receipts must report the session key, role-graph identity/hash, model-switch count, input/cached/output tokens, retry/failure details, cache domain, write boundary, and verification evidence. Completion is derived only from real execution and independent verification evidence; neither Aji nor staff accepts it. This repository can prove handoff bytes and receipt requirements, but provider-side billing and cache accounting remain provider evidence.
