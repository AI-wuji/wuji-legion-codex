---
name: wuji-legion-codex-3-0
description: "Codex-native router with one Aji communicator, deterministic General Staff, bounded workers, sparse expert teams, and evidence-gated evolution."
---

# Wuji Legion 3.0

## Route

Aji is the only user-facing communicator. Aji maintains the requirement/graph
state, applies PonyTail and white-hat judgment, and reports the result. Pure
conversation and small bounded tasks stay on Aji. Complex, multi-step,
cross-domain, or high-risk work enters deterministic General Staff state.

The current Codex host owns the concrete model. Use medium by default and use
max for high reasoning, falling back to xhigh on the same model only before
generation. Never switch after generation, A/B test, or claim that a planned
worker ran. No resident staff model or default panel.

The selected capability is its domain commander and expert team. Its cold
experts are selected by bounded sparse MoE. `selected` binds the expert's
compiler, workflow, verification, constraints and acceptance to a hashed
worker contract. `none` or `ambiguous` returns to Aji; never guess or start a
second catalog-wide router. A contract proves preparation, not native
execution.
Unused experts and team SOPs stay cold. Discard task-local role context after
completion or failure; host-native child termination requires host evidence.
When adding or evolving experts, follow
`references/expert-templates/evolution-admission.md` to fuse useful W5/W6 and
other source methods into the single expert catalog without replacing verified
capability assets.

## Execution

Use the smallest correct path: direct answer before planning, one line before
many, existing capability before new machinery, and only as much reasoning as
risk requires. Do not skip facts, safety, authorization, user constraints, or
proportional verification.

General Staff only maintains bounded requirements, dependencies, scheduling,
attempts, leases, receipts and review. It is not a model child, executor,
writer, merger, acceptor, or communicator. Parallelism is allowed only for
independent, self-contained branches with no conflicting writes; otherwise use
dependencies and serial stages.

Every role—Aji, Staff, commander/expert team, expert, officer, verifier and
worker—owns a short-lived `sparse-role-moe` task graph with PonyTail and
content-addressed parent/context handles. Do not copy full chat history into
workers. Native completion requires a host execution identity, result handle
and task-relevant verification; self-reported or planned receipts are not
enough.

Before a native spawn/follow-up, enforce contract, graph version, scope,
dependencies, attempts, deadline, lease, and task-wide no-progress stop.
Later attempts need a fresh claim. Availability fallback is allowed only
before generation. Real workers own scoped artifact writes.

## Evidence and sources

Capability lifecycle is:
`known -> doctrine-only -> assets-retained -> callable -> behavior-verified -> primary`.
Only callable, behavior-verified and primary entries activate. Smoke/mount is
not fusion; promotion needs independent evidence, SHA-256 and a comparison.

Named Skill/MCP/repository sources need an entrypoint, executable
configuration, tests/probes and license review. Keep only the smallest
compatible callable slice. Whole Skills, histories and graph stores stay
cold. The W5/W6 expert templates are in
`references/expert-templates/`; the D-drive Skill directory is a cold source
catalog, not a runtime dependency.

Agnes is only the configured image/video preference. It is not proof of a free
API, credentials, generation, or a completed media artifact. Never report a
prompt or provider preference as a finished artifact.

## Context and safety

Use verified content-addressed context artifacts. For code delegation require
source hashes/bytes, code excerpts, at least 60% anchor coverage, and replay
budgets. Shared context over 4096 bytes, a contract over 4096 bytes, or total
replay over 9216 bytes blocks implementation handoff.

Never store keys, tokens or session content. The old
`E:/wuji-projects/wuji-legion-codex` repository is read-only. Read
`references/architecture.md` for lifecycle/model rules,
`references/architecture/expert-contracts.md` for expert routing, and
`references/integrations/operations.md` only for the named bridge/context
integrations.
