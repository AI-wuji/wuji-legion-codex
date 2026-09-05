# Wuji Legion 3.0 Architecture

## Decision

Wuji 3.0 is a system-level Codex Skill, not another agent host. Aji understands the user's natural-language request and applies universal PonyTail judgment; Codex executes; the Go CLI supplies deterministic routing, package verification, sparse context selection, and evolution gates.

## One Brain

Terra Aji is the sole user-facing communicator, requirement-table/graph maintainer, universal PonyTail judgment center, capability composer, and final reporter. General Staff remains a deterministic task-state and scheduling mechanism, not a resident model child. It does not execute task work, write artifacts, merge results, or accept completion. The native host dispatches only required execution nodes; real execution plus independent verification evidence determines completion. Terra staff, Aji merge/accept/execute, and Nuwa are not part of this architecture.

## Experts

Use three layers, none resident in the hot prompt:

1. A small domain owner contract names the capability and verification boundary.
2. A complete cold package retains upstream Skills, scripts, assets, templates, UI, and entrypoints.
3. A bounded task worker receives only its branch contract, selected context handles, and the mounted package.

Large domains expose one scenario-oriented suite, or a very small number when output formats have genuinely different runtimes. Presentation exposes exactly two: `wuji-web-deck` and `wuji-editable-deck`. HTML-PPT, Slidev, stage fluid, PPT Master, Huashu, Humanize PPT, and Baoyu are internal templates, components, scripts, or shared planning layers. Users never choose upstream projects.

Parallelize independent branches. Keep dependency edges sequential. Workers never merge each other and never become a second command chain.

## Task Lifecycle

Every request starts as one small task. Except for simple self-contained Q&A and explicit non-GPT provider mode, the router emits ordered stages inside the single Codex execution chain:

1. `general_staff_state`: the deterministic General Staff mechanism consumes the bounded requirement snapshot and maintains the ordered task graph, scheduling decisions, receipts, failures, and requirement review. No staff model worker is created for this stage.
2. `preflight_workers`: when required, one bounded prior-art scan runs before solution work, ordered official -> GitHub -> community, capped by default at 3 sources and 90 seconds. An explicit full-web or comprehensive research request routes to the existing search capability instead. Deterministic edits and explicit offline requests skip preflight.
3. `workers`: bounded execution branches run only after preflight completes. Preflight and execution never run in parallel. If evidence changes the approach, the deterministic task graph invalidates the stale plan and affected descendants.
4. Execution and verification: the host creates each eligible execution node with its exact `model`, `session_key`, and read-only or scoped-artifact-write boundary, and records independent behavior evidence. A deterministic evidence verifier binds active requirements, successful execution artifacts, and verification handles. Aji reports the verified outcome; no conversational role is credited with execution, merging, acceptance, or artifact writing.

The Go CLI declares deterministic stage contracts; Codex performs the actual calls. Except for pure conversation, a task is routed actively. User stops, insertions, and edits increment the requirement and task graph versions while reusing the staff instance and session key. Only affected nodes and descendants are cancelled or invalidated; stale graph versions, attempts, and late receipts are rejected. Rebuild occurs only after a whole-graph veto or task-identity change. A route is effective only when the host runs the declared model and returns task-relevant evidence.

## Evolution Commander

The Evolution Commander is capability governance, not a content persona. It must inspect an upstream candidate's source entrypoints, executable scripts/configuration, tests or probes, and license; retain execution assets where compatible; map overlaps; run the same fixture against upstream and integrated routes; compare artifacts; and then admit, replace, or reject. A README, name, rule list, or passing registry assertion is not capability evidence.

## Context

The default path uses a small stable prefix plus task-local retrieval. `context-select` excludes generated lockfiles, ranks unique path/symbol anchors instead of repeated token counts, and emits ranked excerpts plus a deterministic `WUJI_CONTEXT_CAPSULE_V1` payload within a hard byte budget. The content-addressed artifact binds the normalized query fingerprint, retrieval terms, payload SHA-256, excerpt digest, coverage, and code-excerpt count; loading it independently verifies all byte counts, the workspace boundary, excerpt hashes, and current source-file hashes. A stale artifact, low-coverage artifact, non-code artifact, or artifact selected for a different task cannot unlock code delegation. RTK is the preferred optional command-output filter. Codebase Memory MCP is a read-only cold index for large repositories. Context Mode is useful as a tool-output sandbox but remains optional because its license and host hooks make it unsuitable as copied core code. Heavy temporal GraphRAG remains off the hot path. Web scouting uses GPT/Luna, never Agnes.

## Model Policy

- Default and explicitly selected GPT models share one chain: Terra Aji owns communication, PonyTail judgment, capability composition, the requirement table/graph, and final reporting; deterministic General Staff maintains task state and scheduling.
- Aji uses `gpt-5.6-terra` by default and falls back to `gpt-5.6-sol` when Terra is unavailable; Luna is never Aji's default. Execution nodes use task-appropriate models and may follow their declared availability chain only before generation and only for declared availability failures.
- `preflight_workers[].model` and `workers[].model` are executable model ids. General Staff is deterministic state, not a model worker. Execution nodes use a task-appropriate exact model and may follow the declared Luna -> Terra -> Sol availability chain only before generation starts; `gpt-5.4`, `gpt-5.4-mini`, and other unapproved GPT worker models are rejected before dispatch.
- An explicit non-GPT provider/model preserves the existing capability/provider mode. It does not produce GPT hierarchy worker contracts; its provider-defined execution path retains the same separation between user communication and task execution.
- Each worker gets a deterministic task `session_key`; model selection remains sticky after generation starts. No A/B test, quality retry, or post-generation model switch is allowed.
- A worker branch is not complete until the host returns native-agent identity, requested model, session key, retry/failure kind, result/context handles and bytes, task-contract bytes, cache domain, delegation reason, and independent verification evidence; route metadata, `wuji dispatch`, and `codex exec` cannot prove a requested model ran. Generic interactive output and sandbox-blocked/unavailable evidence are rejected. Provider-side effective model, token, cache, and billing fields stay explicitly unavailable unless the native host exposes them independently.

Sol, Terra, and Luna are separate cache domains; cross-model prompt-cache hits are never assumed. Every worker receives a stable model-local capability prefix followed by its deterministic context payload and JSON task contract. The router withholds execution-node dispatch when retrieval coverage is below 60%, no code excerpt exists, a contract exceeds 2048 bytes, shared context exceeds 4096 bytes per worker, total replay exceeds 8192 bytes, parent-context affinity is required, or the verified artifact is absent/mismatched. Presentation and writing wait for an appropriately scoped execution-node contract when their handoffs depend on parent context. Model switching is never used merely to save tokens when replay cost can erase the model-price advantage.

## Bounded Relation Graphs

The project uses a pyramid-shaped retrieval path, but not a pyramid-shaped duplicate database:

```text
workspace / explicit knowledge scope
        -> bounded index terms and relation keys
        -> candidate files or verified node locations
        -> source/evidence reads on demand
```

The workspace graph is derived and disposable. It stores file metadata, source hashes, bounded retrieval terms, and a small set of symbol/test relations. It is regenerated as a disposable snapshot when the index is absent or stale. A workspace query performs at most 64 index lookups; a knowledge query performs at most 12 index lookups, reads at most 128 candidate nodes, and returns at most 10 matches. Each workspace file contributes at most 512 terms, each workspace term and knowledge reverse index retains at most 256 references, and workspace fallback scans at most 512 source files. These limits are deliberately boring: they make worst-case behavior visible and testable.

The experience graph is not consulted for every task. It is an incident and reuse index, activated only by a failure, a reported failure, explicit reuse, a capability miss, or a verification trace. Each node is keyed by `(kind, normalized key, explicit scope)`, updated in place for the same identity, and linked to a compact summary, solution location, root cause when applicable, and a local verification artifact whose current SHA-256 is checked again at query time. Raw transcripts, secrets, and full solution bodies do not belong in the graph.

This design incorporates the useful parts of public GraphRAG, Graphiti, and hierarchical code-retrieval work: community or scope summaries for routing, temporal/provenance validity, and bottom-up candidate selection. It rejects their heavier graph database and broad indexing behavior from the default hot path. A hierarchy reduces expansion and prompt size; it does not by itself bound fact growth. That requires replacement, deduplication, stale invalidation, evidence expiry/retirement policy, and hard I/O/result budgets.

## Independent Opposition

Simple tasks receive staff evidence review without staff acceptance; medium tasks receive internal QA only; large/high-risk tasks receive one composite-MoE independent officer by default, with a governance-risk audit section in the same review. There is no default panel. An internal adversarial pass is not independent-officer evidence, and no capability is complete without real behavior verification and independent hashes.
