# Context Source Verdicts

Evidence checked on 2026-07-12. Full commit-level decisions are recorded in `migration/upstream-review.json`.

| Project | Verdict | Use in 2.0 |
| --- | --- | --- |
| Aider repo map | Doctrine plus local implementation | PageRank-style relevance inspired compact path/symbol selection; no Aider runtime shell. |
| RTK v0.50.0 | Optional primary output filter | Apache-2.0, Windows binary, verified by installer/probe; never required for correctness. |
| Codebase Memory MCP v0.11.0 | Optional cold index | MIT, Windows binary, useful for large repositories; read-only and on-demand, not memory authority. |
| Context Mode v1.0.169 | Optional host integration | Strong tool-output sandbox and FTS/BM25 ideas; tracked in [compatibility contract](integrations/context-mode-compatibility.md); do not copy ELv2 code or make it a required core. |
| Whittle | Optional lossless output-filter prior art | Apache-2.0; useful fail-open, source-pass-through, omission-marker, and measured-token guarantees; no daemon, hook, or model router is required by 3.0. |
| Context Ledger | Restorable commit-boundary pointer prior art | MIT; useful git-pointer rehydration principle; existing Wuji content-addressed execution/knowledge artifacts cover the durable evidence boundary, while `context-recall` covers cold tool observations. |
| codex-agent-mem | Advisory local-first continuity prior art | Apache-2.0; useful scope/session isolation and unchanged-pack hashes; do not introduce its second SQLite/MCP state authority or automatic startup injection. |
| Headroom / Distil | Cache-aware compression prior art | Apache-2.0 / source-reviewed; useful “freeze cached prefix” and reversible expansion principles, but host/provider cache receipts are unavailable to the core CLI, so no cache-saving claim is made. |
| RTK | Optional command-output filter prior art | Apache-2.0; useful narrow filters and ownership-safe hook installation; never a required global rewrite or correctness dependency. |
| jev-routing | Negative routing evidence | Archived MIT; model-dependent tool steering did not generalize, so 3.0 keeps deterministic routing and only borrows bounded context reduction ideas. |
| Graphiti/GraphRAG | Rejected from hot path | Useful for temporal knowledge products, too heavy for routine project retrieval without benchmark proof. |
| PPT Master | Update after behavior regression | Upstream adds material PPTX fidelity and quality-gate work; keep it cold and expose it only through the editable-deck scenario. |
| Huashu Design | Update and distill | Keep the complete package, while routing only design reasoning and execution guards through current scenarios. |
| Frontend Slides | Update cold source only | YAML fixes improve retained templates but do not create a new runtime surface. |
| Open Design | Reject new runtime | Its current daemon, auth, provider, and agent platform would become a second host; no new runtime is admitted. |

Repository rules cannot shrink Codex's system/tool/Skill catalog or an already huge conversation. Project config therefore adds early body compaction, a tool-output cap, and a history byte cap; a fresh 2.0 task is still required to remove an inherited 200k+ outer prefix.
