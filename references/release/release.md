# Wuji Legion 2.0 Baseline

Date: 2026-07-12
Decision: ready for the first Git baseline after final hygiene and size audits.

Completed scope:

- One Codex host, one Aji router/merger/write authority; no Nuwa, second router, default council, bridge, GUI, daemon, or second control plane.
- Legacy evidence was reviewed read-only and classified item by item. Useful doctrine was distilled, missing approved behavior was implemented, and rejected architecture was excluded. Nothing was copied wholesale.
- Complete ledgers cover 104 legacy objects and 52 legacy worktree paths with unique actions and evidence.
- Presentation cold sources are current and portable: built-in Slidev 52.17.0 and approved stage-fluid assets are locked by content hash and behavior verified from temporary copies.
- Open Design's new daemon/auth/provider platform remains excluded. Xiaobai remains an explicit per-call image provider and does not elevate the whole image capability.
- Capability labels are bounded by real evidence; smoke proves only callable status.

Release gates:

- Full behavior suite: pass.
- Fusion, optimization, and context-bloat audits: pass.
- Dependency audit: 0 known production vulnerabilities.
- Source budget before these reports: 985,508 / 1,000,000 bytes. Final audit must remain below the limit.
- No credentials, generated dependency trees, build output, or temporary browser artifacts may be committed.

Excluded follow-on work: Open Design daemon/auth/provider platform, a new desktop control plane, and any unrequested product expansion. These are outside 2.0 completion rather than unfinished work.
