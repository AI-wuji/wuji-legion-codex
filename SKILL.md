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
- Each `workers` item is executable: use its exact `model`, compact contract and selected context; keep `writes` false unless a disjoint scope is granted.
- `model_class` is only a label. Fallback only for `model-unavailable` or `provider-error-before-generation`, honor `max_attempts`, and record the effective model. Never retry a paid low-quality generation.
- Completion requires every route-declared `execution_evidence_fields` value, including ordered attempts/effective model, payload hashes/bytes, token/cache telemetry, billing baseline/actual cost/savings, and Aji acceptance; route JSON is not execution evidence.
- The default mapping is `sol -> gpt-5.6-sol`, `terra -> gpt-5.6-terra`, and `luna -> gpt-5.6-luna`. A bounded independent code implementation branch may use Terra; dependent verification stays sequential on Aji. Broad independent research and mechanical extraction branches may use Luna.
- Do not run a declared worker in Aji after only printing JSON. No worker means Aji direct execution.
- Sol, Terra and Luna never share assumed cache. Follow `prompt_order`: stable prefix, deterministic `context_payload`, real JSON `task_contract`. Delegate only when payload hashes/bytes match, coverage is at least 6000 BPS, code evidence exists, and replay gates pass. Presentation/writing stay on Aji unless explicit parallel work has `--self-contained-handoff`. A cheaper model alone proves no saving.

Nuwa does not exist in 2.0. Officers are cold review seats. Routine opposition is an internal Aji adversarial pass that reuses current context; launch a real officer only when explicitly requested or when the risk contract requires independent evidence.

## Capability Truth

Use these lifecycle states exactly:

`known -> doctrine-only -> assets-retained -> callable -> behavior-verified -> primary`

Only `callable`, `behavior-verified`, or `primary` may be activated. A manifest flag or direct mount is never callable evidence by itself: the declared smoke/mount probe must execute. Behavior probes must leave content-addressed artifacts in the verifier evidence directory and pass independent hash checks; a receipt, signature, or exit code alone is not fusion. Smoke probes keep a package at `callable`; callable packages are not fused. `primary` is reserved for an Evolution Commander replacement with a verified content-addressed promotion receipt and archived baseline; a manifest cannot self-promote. Say fused only for `behavior-verified` or `primary`. The Evolution Commander owns admission, replacement, regression comparison, and retirement; it does not author ordinary task output.

## Context

Keep the stable prefix small. For code, run `wuji context-select` with the same query and concrete path/symbol anchors, then pass its verified artifact to `wuji route --context-artifact`. Query fingerprint, payload bytes/SHA-256, paths, excerpt/source hashes, coverage and code count must match. Stay on Aji if absent/stale/mismatched, over 4096 context bytes, below 6000 BPS, without code, over 2048 contract bytes, over 8192 replay bytes, or parent context is required. Keep logs and full Skill bodies out. Prefer RTK output filtering; use Codebase Memory MCP only as a read-only cold index. Keep GraphRAG off the default path.

Agnes is image/video-only. Web search uses the default GPT route; broad independent source branches may use Luna.

Read [architecture.md](references/architecture.md) only for architecture/evolution work. Read [capability-contract.md](references/capability-contract.md) only when admitting or repairing a capability.
