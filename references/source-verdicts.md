# Context Source Verdicts

Evidence checked on 2026-07-12. Full commit-level decisions are recorded in `migration/upstream-review.json`.

| Project | Verdict | Use in 2.0 |
| --- | --- | --- |
| Aider repo map | Doctrine plus local implementation | PageRank-style relevance inspired compact path/symbol selection; no Aider runtime shell. |
| RTK v0.43.0 | Optional primary output filter | Apache-2.0, Windows binary, verified by installer/probe; never required for correctness. |
| Codebase Memory MCP v0.9.0 | Optional cold index | MIT, Windows binary, useful for large repositories; read-only and on-demand, not memory authority. |
| Context Mode v1.0.169 | Optional host integration | Strong tool-output sandbox and FTS/BM25 ideas; do not copy ELv2 code or make it a required core. |
| Graphiti/GraphRAG | Rejected from hot path | Useful for temporal knowledge products, too heavy for routine project retrieval without benchmark proof. |
| PPT Master | Update after behavior regression | Upstream adds material PPTX fidelity and quality-gate work; keep it cold and expose it only through the editable-deck scenario. |
| Huashu Design | Update and distill | Keep the complete package, while routing only design reasoning and execution guards through current scenarios. |
| Frontend Slides | Update cold source only | YAML fixes improve retained templates but do not create a new runtime surface. |
| Open Design | Reject new runtime | Its current daemon, auth, provider, and agent platform would become a second host; no new runtime is admitted. |

Repository rules cannot shrink Codex's system/tool/Skill catalog or an already huge conversation. Project config therefore adds early body compaction, a tool-output cap, and a history byte cap; a fresh 2.0 task is still required to remove an inherited 200k+ outer prefix.
