# Wuji Legion 2.0 Architecture

## Decision

Wuji 2.0 is a system-level Codex Skill, not another agent host. Codex executes; the Go CLI supplies deterministic routing, package verification, sparse context selection, and evolution gates.

## One Brain

Aji and the former General Staff are one strongest reasoning surface. Aji owns requirement normalization, dependency-aware fan-out, final merge, write authority, and completion judgment. Nuwa is removed because staffing belongs to routing and capability ownership, while evolution belongs to the Evolution Commander.

## Experts

Use three layers, none resident in the hot prompt:

1. A small domain owner contract names the capability and acceptance boundary.
2. A complete cold package retains upstream Skills, scripts, assets, templates, UI, and entrypoints.
3. A bounded task worker receives only its branch contract, selected context handles, and the mounted package.

Large domains expose one scenario-oriented suite, or a very small number when output formats have genuinely different runtimes. Presentation exposes exactly two: `wuji-web-deck` and `wuji-editable-deck`. HTML-PPT, Slidev, stage fluid, PPT Master, Huashu, Humanize PPT, and Baoyu are internal templates, components, scripts, or shared planning layers. Users never choose upstream projects.

Parallelize independent branches. Keep dependency edges sequential. Workers never merge each other and never become a second command chain.

## Task Lifecycle

Every request starts as one small task. The router may emit two ordered stages, but they remain inside the single Codex execution chain:

1. `preflight_workers`: one bounded Luna prior-art scan for non-trivial solution work, ordered official -> GitHub -> community, capped at 3 sources and 90 seconds. Deterministic edits and explicit offline requests skip it.
2. `workers`: bounded execution branches selected only after preflight completes. Preflight and execution never run in parallel. If evidence changes the approach, the host discards the stale execution plan and routes again.
3. Aji merge and verification: after routing, the current Codex host creates every eligible read-only branch as a native child agent with the route's exact `model`, `session_key`, and `fallback_models`. Aji remains the only writer and completion judge.

This is not a second scheduler. The Go CLI declares deterministic stage contracts; Codex performs the actual calls. Except for simple self-contained Q&A, a task is routed actively. A route is effective only when the host runs the declared model and returns a task-relevant result handle for Aji to accept.

## Evolution Commander

The Evolution Commander is capability governance, not a content persona. It must inspect an upstream candidate's source entrypoints, executable scripts/configuration, tests or probes, and license; retain execution assets where compatible; map overlaps; run the same fixture against upstream and integrated routes; compare artifacts; and then admit, replace, or reject. A README, name, rule list, or passing registry assertion is not capability evidence.

## Context

The default path uses a small stable prefix plus task-local retrieval. `context-select` excludes generated lockfiles, ranks unique path/symbol anchors instead of repeated token counts, and emits ranked excerpts plus a deterministic `WUJI_CONTEXT_CAPSULE_V1` payload within a hard byte budget. The content-addressed artifact binds the normalized query fingerprint, retrieval terms, payload SHA-256, excerpt digest, coverage, and code-excerpt count; loading it independently verifies all byte counts, the workspace boundary, excerpt hashes, and current source-file hashes. A stale artifact, low-coverage artifact, non-code artifact, or artifact selected for a different task cannot unlock code delegation. RTK is the preferred optional command-output filter. Codebase Memory MCP is a read-only cold index for large repositories. Context Mode is useful as a tool-output sandbox but remains optional because its license and host hooks make it unsuitable as copied core code. Heavy temporal GraphRAG remains off the hot path. Web scouting uses GPT/Luna, never Agnes.

## Model Policy

- Aji routing, merge, writes, and completion judgment: `gpt-5.6-terra`, using the configured reasoning effort.
- A bounded independent high-reasoning implementation or review worker may use `gpt-5.6-sol`; mechanical verification remains a Luna branch only when it is self-contained, otherwise it stays sequential on Aji.
- Broad web scouting uses the default GPT route for the final analysis, while independent source branches use `gpt-5.6-luna`.
- Other mechanical extraction uses `gpt-5.6-luna` when the route emits a compact extraction branch and no project context replay is required.
- Sol is reserved for an explicit, bounded high-reasoning judgment worker. It is read-only, has one attempt, and returns evidence and options for Aji on Terra to merge.
- `preflight_workers[].model` and `workers[].model` are executable model ids. `model_class` is metadata only; the host must not silently run a worker on the Aji model when a cheaper model is selected, or report route JSON as execution.
- In the GPT 5.6 policy, Terra never appears in a worker route. Sol and Luna each have one attempt and no automatic fallback; a generated or unavailable branch returns its evidence to Aji. Aji remains the sole merger and write authority. `gpt-5.4`, `gpt-5.4-mini`, and any other GPT worker model are rejected before dispatch. Explicit Grok or other non-GPT provider/model choices are outside this GPT policy.
- Each worker gets a deterministic task `session_key`; model selection happens once at task start and remains sticky inside that session. GPT 5.6 workers do not switch models before or after generation.
- A worker branch is not complete until the host returns native-agent identity, requested model, session key, retry/failure kind, Aji acceptance, result/context handles and bytes, task-contract bytes, cache domain, and delegation reason; route metadata, `wuji dispatch`, and `codex exec` cannot prove a requested model ran. Generic interactive output and sandbox-blocked/unavailable evidence are rejected. Provider-side effective model, token, cache, and billing fields stay explicitly unavailable unless the native host exposes them independently.

Sol, Terra, and Luna are separate cache domains; cross-model prompt-cache hits are never assumed. Every worker receives a stable model-local capability prefix followed by its deterministic context payload and JSON task contract. The router measures those actual payloads and keeps execution on Aji when retrieval coverage is below 60%, no code excerpt exists, a contract exceeds 2048 bytes, shared context exceeds 4096 bytes per worker, total replay exceeds 8192 bytes, parent-context affinity is required, or the verified artifact is absent/mismatched. Presentation and writing remain direct by default because their worker handoffs usually depend on parent context. Model switching is never used merely to save tokens when replay cost can erase the model-price advantage.

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

Routine opposition is an Aji adversarial pass. A real white-hat officer is a cold independent review seat and is started only by an explicit request or a high-risk contract. `internal_adversarial_pass: true` is not execution evidence. The host must run the officer, retain a content-addressed review result, and include its status in completion evidence; otherwise the correct status is `not executed`.
