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

## Evolution Commander

The Evolution Commander is capability governance, not a content persona. It must inventory the upstream package, retain execution assets, map overlaps, run the same fixture against upstream and integrated routes, compare artifacts, and then admit, replace, or reject. A name, rule list, or passing registry assertion is not capability evidence.

## Context

The default path uses a small stable prefix plus task-local retrieval. `context-select` emits ranked excerpts and a content-addressed artifact within a hard byte budget. The artifact binds the normalized query fingerprint and excerpt digest; loading it independently verifies the actual byte count, workspace boundary, excerpt hashes, and current source-file hashes. A stale artifact or an artifact selected for a different task cannot unlock code delegation. RTK is the preferred optional command-output filter. Codebase Memory MCP is a read-only cold index for large repositories. Context Mode is useful as a tool-output sandbox but remains optional because its license and host hooks make it unsuitable as copied core code. Heavy temporal GraphRAG remains off the hot path. Web scouting uses GPT/Luna, never Agnes.

## Model Policy

- Aji planning, architecture, merge, and high-risk judgment: `gpt-5.6-sol`, highest available reasoning.
- A bounded independent implementation worker may use `gpt-5.6-terra`; verification that depends on the implementation remains sequential on Aji.
- Broad web scouting uses the default GPT route for the final analysis, while independent source branches use `gpt-5.6-luna`.
- Other mechanical extraction uses `gpt-5.6-luna` when the route emits a compact extraction branch and no project context replay is required.
- `workers[].model` is the executable model id. `model_class` is metadata only; the host must not silently run a worker on the Aji model when a cheaper model is selected.
- If a requested worker model is unavailable, retry using its ordered `fallback_models`; retain the effective model in execution evidence. Aji remains the sole merger and write authority.
- A worker branch is not complete until the host returns `requested_model`, ordered `attempts`, `effective_model`, `result_handle`, `context_handle_ids`, `context_bytes_sent`, `task_contract_bytes`, and `delegation_gate_reason`; route metadata alone cannot prove that a cheaper model or fallback actually ran.

Sol, Terra, and Luna are treated as separate cache domains; cross-model prompt-cache hits are never assumed. The router measures the task contract and context assigned to every worker. It keeps execution on Aji when the contract exceeds 2048 bytes, shared context exceeds 4096 bytes per worker, total replay exceeds 8192 bytes, parent-context affinity is required, or the verified artifact is absent/mismatched. Model switching is never used merely to save tokens when replay cost can erase the model-price advantage.
