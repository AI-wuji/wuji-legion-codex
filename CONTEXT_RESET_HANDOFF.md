# Wuji Legion Context Reset Handoff

Purpose: start a fresh Codex thread without carrying the current long outer context.

Use this file only as a short seed. Do not paste old transcripts, full reports, screenshots, logs, or long articles.

## Current State

- Repo: `E:/wuji-projects/wuji-legion-codex`
- Runtime: Codex-only, single main chain, Go `wuji-cli` deterministic base.
- OpenSquilla is absorbed as atoms; it is not an external executor or second router.
- Independent officers are explicit and on demand: `white-hat`, `guard-office`, `root-cause-officer`, `audit`, `quality-inspection`, `performance-benchmark-on-demand`, `compliance-on-demand`.
- Nuwa is a cold capability-gap lens, not an officer or router.
- `route-task` now emits `capability_mounts.distilled_atom_evidence`, binding each mounted atom to `fusion-matrix.json` evidence: `source_pool`, `decision`, `owner`, `reason`, `fusion_policy`, and residency.
- `officer-run` is now a real executable officer merge path and writes explicit seat verdicts plus main-chain merge evidence under `.wuji-tools/officers/`.
- `context-pack` now carries distilled atom evidence into dynamic assembly instead of only atom names.
- Current core repo gates pass: `fusion-audit`, `optimization-audit`, `context-bloat-audit`, `scripts/test-wuji-officers.ps1`, `scripts/test-wuji-cli.ps1`.
- Current root-cause verdict exists at `.wuji-tools/officers/root-cause-report.json`.

## Active Problem

Backend rows show cached/blue-hit volume escalating from about 180k-229k to about 750k-784k, with current `runtime-context-audit` reporting:

- cached p95 about `783600`
- input p95 about `786860`
- fresh input p95 about `34771`
- uncached p95 about `34771`

Diagnosis: the current thread/outer context is too long and has entered emergency bloat. High cache hit rate is not enough because cached input still has cost. Repo rule changes cannot shrink an already-loaded thread.

## New Thread Operating Rules

- Keep the stable prefix small; do not replay old history.
- Load only one owner profile, one selected skill, and triggered officers.
- Return evidence handles and short summaries, not full logs or long source dumps.
- Use `rg` with cold directories excluded. Do not search `.wuji-tools`, `outputs`, `.git`, `.codex`, `.agents`, or raw session/auth paths unless explicitly required.
- For token/cost/cache claims, use numeric-only `outputs/runtime-usage.jsonl` and run `runtime-context-audit`.
- If cached/blue-hit again exceeds 30k p95, stop broad work and refresh this handoff instead of continuing the same long thread.
- If cached/blue-hit reaches about 240k+ p95 or total input reaches about 256k+ p95, treat the current thread as emergency-only: no same-thread optimization, refresh the handoff and continue from a new thread.

## Immediate Next Task

Continue optimizing Wuji Legion for low total token cost and high first-pass success:

1. start from this handoff in a fresh thread,
2. do not replay old discussion history,
3. verify `outputs/runtime-context-audit-report.json` in the new thread after a short seed-only run,
4. keep officers and skills on-demand,
5. enforce cold-search excludes and targeted audits only,
6. avoid full-legion scans unless the user explicitly asks for release-level verification.

## Evidence Handles

- `kernel-source.json`
- `hotpath-manifest.json`
- `config.json`
- `tools/wuji_cli.go`
- `.wuji-tools/officers/main-chain-merge.json`
- `.wuji-tools/officers/root-cause-report.json`
- `outputs/runtime-context-audit-report.json`
- `outputs/bench-report.json`
- `outputs/context-bloat-audit-report.json`
- `outputs/fusion-audit-report.json`
- `outputs/optimization-audit-report.json`
