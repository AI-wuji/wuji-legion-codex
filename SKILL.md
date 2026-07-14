---
name: wuji-legion-codex-2-0
description: System-level Codex router: one Aji brain, cold capability packages, bounded workers, sparse context, and verified evolution.
---

# Wuji Legion 2.0

One chain: `Aji -> cold capability mount -> bounded execution -> real verification`.

## Runtime

1. Keep simple self-contained Q&A on Aji. For all other tasks run `wuji route --query "<request>"`; route JSON, mount, `wuji dispatch`, and `codex exec` are not execution evidence.
2. Mount one scenario Skill plus needed atoms. Automatic sources need callable lifecycle, entrypoint, and activation. Pass a verified entrypoint to the selected native worker; selection, mount, preparation, and injection are not execution. Accept only a native agent ID plus command/tool/artifact result. Keep other catalogs cold; use `wuji source-audit` for repair/admission.
3. Every task is small. For non-trivial troubleshooting, API/SDK, dependency, framework, routing, cache, architecture, migration, performance, security, or integration work, run one preflight in order: official -> GitHub -> community; at most 3 sources/90 seconds; stop on decisive evidence. Skip deterministic or offline work.
4. Complete preflight before workers. If it changes the approach, discard the old plan and reroute. Change only the active target, with the smallest complete fix; do not create compatibility or parallel v2 paths unless asked.
5. When `route.change_capsule.required`, create a strict `wuji change-capsule` with scope-out, acceptance, verification, and rollback before edits. A passing capsule is evidence, not a second workflow.
6. Parallelize independent branches only, with compact contracts. Aji alone merges and writes. Run the selected capability probe and task-local verification before completion.

## Models

- Aji is `gpt-5.6-terra`: routing, merge, writes, and completion only. Terra never appears in a worker route.
- A worker is real only when this Codex host creates a native read-only child with its exact `model`, `session_key`, and `fallback_models`, then returns agent identity, requested model, result handle, and Aji acceptance. The Go CLI prepares native-host contracts only; external `codex exec` is an explicit local compatibility diagnostic and is never execution evidence. Do not claim execution from a plan, source injection, generic/unavailable output, or CLI output.
- Luna is context-free mechanical/research work; Sol is bounded high-reasoning review or judgment. GPT workers are only `gpt-5.6-luna` or `gpt-5.6-sol`, one model and one attempt, no automatic fallback or switching. Explicit non-GPT choices bypass this policy.
- Sol, Terra, and Luna do not share prompt cache. Delegate only after replay gates pass; presentation/writing remain on Aji unless the handoff is explicitly self-contained. Provider billing/cache telemetry is unavailable unless the host exposes it.

For code, trace the flow, then use Ponytail's first valid rung: skip, reuse local code, standard library, native platform, installed dependency, then minimum code. Fix the common root cause, reject unrequested abstractions, preserve safety/requirements, and run one proportionate check for nontrivial logic.

Nuwa does not exist. Officers are cold: routine opposition is Aji self-check; a real officer needs explicit/risk-required independent execution evidence.

## Capability Truth

`known -> doctrine-only -> assets-retained -> callable -> behavior-verified -> primary`

Only `callable`, `behavior-verified`, and `primary` activate. A manifest or mount is not evidence; smoke proves callable only. Behavior requires a real probe, content-addressed artifacts, and independent hashes; primary also needs comparison, baseline, and promotion receipt. A source without semantic activation plus entrypoint is cold reference material. Say "fused" only for behavior-verified/primary.

For a named external Skill, MCP, or repo, inspect source entrypoints, scripts/config, tests/probes, and license, not README claims. Retain the smallest compatible callable slice and behavior-verify it; otherwise record rejection. An MCP without a proven host adapter and invocation stays cold/unavailable.

## Context And Graphs

Keep the stable prefix small. For code use `wuji context-select` with the same query and concrete path/symbol anchors, then pass its verified artifact to `wuji route --context-artifact`. Fingerprint, bytes/SHA-256, paths, excerpt hashes, coverage, and code count must match.

Stay on Aji for absent/stale/mismatched context, >4096 bytes, <6000 BPS, no code/content anchor, contract >2048 bytes, replay >8192 bytes, or parent-context affinity. Keep logs, histories, graphs, and full Skill bodies out of prompts. RTK is preferred; Codebase Memory is a cold read-only index; GraphRAG is off by default.

Use bounded cold graphs only: a disposable workspace graph (`workspace -> files -> symbols/tests`) with 512 terms/file, 256 refs/term, 64 lookups, 128 candidates, and 16 MiB reads; and a cross-project graph only for failure/reuse events, with 12 lookups, 128 candidates, and 10 results. Store compact, scoped, content-addressed pointers with root cause and verification hash, never transcripts. Pyramid levels are routing indexes, not copies; enforce quotas, TTL, replacement/deduplication, locking, repair, GC, stale rebuilds, and bounded fan-out. Query experience only on matching failure/reuse, never every task.

`internal_adversarial_pass` is not white-hat execution. An explicit officer needs a compact read-only task, content-addressed result, and Aji-accepted receipt; otherwise report `not executed`.

Agnes is image/video-only. Web scouting uses default GPT, with bounded Luna source branches. Feishu links/tasks mount the official `feishu-lark` Skill and default to read-only. Explicit provider/model wins; never store keys in rules or repositories. Read `references/architecture.md` for architecture/evolution and `references/capability-contract.md` for capability admission/repair.
