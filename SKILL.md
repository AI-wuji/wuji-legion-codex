---
name: wuji-legion-codex-2-0
description: System-level Codex router for fast, accurate task execution with one Aji main brain, cold-mounted complete capability packages, bounded parallel experts, behavior-verified evolution, sparse project context, and a deterministic Go base. Use for normal Codex work when Wuji should route the request, mount the right Skill/plugin/MCP, execute the smallest complete change, and verify the real result without loading the whole legion.
---

# Wuji Legion 2.0

Use one chain only:

`Aji (reason and merge) -> cold capability mount -> direct/parallel execution -> real verification`

## Runtime

1. Run `wuji route --query "<request>"` when routing is not obvious.
2. Select one user-facing scenario Skill, then mount only its normalized catalog entries and required internal atoms. Never expose upstream projects as choices, replace capability with prose, or load every source package.
3. Execute the smallest complete change on the active target. Do not create compatibility pages or parallel implementations unless explicitly requested.
4. Use parallel workers only for independent branches with compact task contracts. Honor SecondaryCapabilities for multi-intent requests without switching write authority away from Aji. Mount primary sources by default; mount secondary/optional atoms only when named or when the user asks for full capability packages. Aji alone merges and writes the final decision.
5. Run the selected package probe plus task-local verification before claiming completion.

## Model Execution

The route result has two different levels of model use:

- `main_model` is the Aji control-plane model: `gpt-5.6-sol` owns routing, architecture, merge, writes, and completion judgment.
- Every item in `workers` is an executable delegation, not a description. Spawn it with the exact `model` value, pass only its compact contract and selected context handles, and keep `writes` false unless the route explicitly grants a disjoint write scope.
- `model_class` is only a label. The host must use `model`; if that model is unavailable, retry in the listed `fallback_models` order and record the effective model in the task result.
- A worker is complete only with an execution receipt containing `requested_model`, ordered `attempts`, `effective_model`, `result_handle`, `context_handle_ids`, `context_bytes_sent`, `task_contract_bytes`, and `delegation_gate_reason`; route JSON alone is not execution evidence.
- The default mapping is `sol -> gpt-5.6-sol`, `terra -> gpt-5.6-terra`, and `luna -> gpt-5.6-luna`. A bounded independent code implementation branch may use Terra; dependent verification stays sequential on Aji. Broad independent research and mechanical extraction branches may use Luna.
- Do not execute a route-declared worker branch in Aji after merely printing the route JSON. If no worker is emitted, the task stays on the Aji model by design.
- Never assume prompt-cache sharing across Sol, Terra, and Luna. Delegation is allowed only when the route's measured task contract, verified context artifact, and total replay estimate pass the declared byte gates; model price alone is not evidence of savings.

Nuwa does not exist in 2.0. Officers are cold review seats. Routine opposition is an internal Aji adversarial pass that reuses current context; launch a real officer only when explicitly requested or when the risk contract requires independent evidence.

## Capability Truth

Use these lifecycle states exactly:

`known -> doctrine-only -> assets-retained -> callable -> behavior-verified -> primary`

Only `callable`, `behavior-verified`, or `primary` may be activated. A manifest flag or direct mount is never callable evidence by itself: the declared smoke/mount probe must execute. Behavior probes must leave content-addressed artifacts in the verifier evidence directory and pass independent hash checks; a receipt, signature, or exit code alone is not fusion. Smoke probes keep a package at `callable`; callable packages are not fused. `primary` is reserved for an Evolution Commander replacement with a verified content-addressed promotion receipt and archived baseline; a manifest cannot self-promote. Say fused only for `behavior-verified` or `primary`. The Evolution Commander owns admission, replacement, regression comparison, and retirement; it does not author ordinary task output.

## Context

Keep the stable prefix small. For codebase work, run `wuji context-select` with the same task query, then pass its verified `artifact_path` to `wuji route --context-artifact`. The artifact is valid only while its query fingerprint, actual byte count, workspace paths, excerpt hashes, and source-file hashes still match. Code stays on Aji when the artifact is absent, stale, mismatched, larger than 4096 bytes, when the task contract exceeds 2048 bytes, when total replay exceeds 8192 bytes, or when parent context is required. Keep raw logs and full Skill bodies out of the prompt. Use RTK for supported command-output compression when installed; use Codebase Memory MCP only as a read-only cold index on large repositories. Do not load GraphRAG or persistent graph dumps on the default path.

Agnes is image/video-only. Web search uses the default GPT route; broad independent source branches may use Luna.

Read [architecture.md](references/architecture.md) only for architecture/evolution work. Read [capability-contract.md](references/capability-contract.md) only when admitting or repairing a capability.
