---
name: wuji-legion-codex-2-0
description: System-level Codex router with one Aji main brain, cold capability packages, bounded workers, sparse context, verified evolution, and deterministic Go routing.
---

# Wuji Legion 2.0

One chain only:

`Aji -> cold capability mount -> bounded execution -> real verification`

## Runtime

1. Run `wuji route --query "<request>"` when routing is unclear.
2. Mount one user-facing scenario Skill and only its required internal atoms. Do not expose upstream projects as choices or load complete catalogs without need.
3. Treat every request as small. For non-trivial troubleshooting, API/SDK, dependency, framework, routing, cache, architecture, migration, performance, security, or integration work, run one `preflight_workers`: official -> GitHub -> community; <=3 sources/90 seconds; stop on decisive evidence. Skip deterministic edits and offline requests.
4. Run preflight before execution. If evidence changes the approach, discard the stale worker plan and route again. Make the smallest complete change on the active target; do not create compatibility or parallel v2 implementations unless requested.
5. Parallelize only independent branches with compact task contracts. Honor `SecondaryCapabilities`, but Aji remains the only merger and writer.
6. Run the selected capability probe and task-local verification before completion.

## Models

- Aji uses `gpt-5.6-terra` for routing, merge, writes, and completion judgment. Sol is a bounded, read-only escalation worker for explicitly requested high-reasoning judgment; it never becomes the default control plane.
- Route workers are executable. When subagents exist, spawn each with exact `model`, `session_key`, hashes, handles, and write boundary. JSON alone is not execution; report any active-host/`main_model` mismatch.
- Default classes are Terra, Luna, and Sol. Terra is the control plane and handles bounded independent implementation/verification; Luna handles mechanical extraction or broad independent research only when parent-context replay is unnecessary; Sol handles only explicit high-reasoning judgment.
- Select a model once. Luna -> Terra and Terra -> Sol only before generation on `model-unavailable` or `provider-error-before-generation`, within retry limits. Sol gets one attempt/no fallback. Never pay to retry weak output, downgrade, or assume cross-model cache.
- Completion requires route evidence: session, attempts/model/switches, payload hashes/bytes, token/cache/cost telemetry, result handle, and Aji acceptance.
- Delegate only when the deterministic context artifact and task contract pass replay gates. Presentation/writing remain on Aji unless an explicit self-contained handoff passes.

Nuwa does not exist. Officers are cold review seats. Routine opposition is an Aji self-check; a real officer runs only when explicitly requested or required by risk.

## Capability Truth

Lifecycle:

`known -> doctrine-only -> assets-retained -> callable -> behavior-verified -> primary`

Only `callable`, `behavior-verified`, or `primary` may activate. A manifest or mount is not evidence. Smoke proves only callable. Behavior verification requires real probes plus content-addressed artifacts and independent hash checks. `primary` additionally requires an Evolution Commander comparison, archived baseline, and verified promotion receipt. Say fused only for `behavior-verified` or `primary`.

## Context

Keep the stable prefix small. For code, run `wuji context-select` with the same query and concrete path/symbol anchors, then pass its verified artifact to `wuji route --context-artifact`. Query fingerprint, bytes/SHA-256, paths, source/excerpt hashes, coverage, and code count must match.

Stay on Aji when context is absent, stale, mismatched, over 4096 bytes, below 6000 BPS, has no code/content anchor, has a contract over 2048 bytes, exceeds 8192 replay bytes, or needs parent conversation. Keep logs, histories, graph dumps, and full Skill bodies outside the prompt. Prefer RTK filtering; Codebase Memory is a read-only cold index. GraphRAG is off the default path.

## Relation Graphs

Use two bounded cold layers:

- Workspace graph: disposable `workspace -> files -> symbols/tests` index. Query terms/relations before reading candidates. Rebuild stale generations atomically. Enforce relative-path and symlink boundaries, 512 terms/file, 256 refs/term, 64 lookups, 128 candidates, 16 MiB source reads, and bounded graph files.
- Cross-project knowledge graph: query only on `failure`, `reported-failure`, `explicit-reuse`, `capability-miss`, or `verification-trace`. Require canonical `global` or `workspace:<sha256>` scope. Store compact summaries and content-addressed solution/verification objects, replace `(kind,key,scope)`, and cap queries at 12 lookups, 128 candidates, and 10 results.
- Enforce node/byte quotas, TTL, deduplication, revisions, locking, transaction repair, garbage collection, stale-index rebuilds, bounded fan-out, and evidence hash checks.

Pyramid levels are routing indexes, not copies of facts. A hierarchy reduces retrieval expansion but cannot stop storage growth. After a failure is resolved, record root cause and verification once; on a matching later failure, query before experimenting. Never scan all history by default.

## White-Hat Evidence

`internal_adversarial_pass` is only an Aji self-check declaration. An explicit officer must appear as a read-only `officer_workers` task with an independent session, compact contract, content-addressed result, and receipt accepted by Aji. A label, route flag, empty list, or unvalidated receipt means `not executed`.

Agnes is image/video-only. Web scouting uses the default GPT route; bounded independent source branches may use Luna. Explicit provider/model choices win, and keys never enter rules or repositories.

Read `references/architecture.md` only for architecture/evolution work. Read `references/capability-contract.md` only when admitting or repairing a capability.
