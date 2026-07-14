# Change Capsule

Use this capsule only for a cross-module change, a high-risk change, or a change that triggers an external action. Do not require it for a small direct task.

```text
Intent:
Scope out:
Acceptance scenarios:
Verification commands:
Rollback:
```

The capsule bounds the decision and its evidence. It is a distilled change-control aid, not a second workflow, a mandatory proposal phase, or authority to create parallel plans.

`wuji change-capsule` creates a draft without blocking a small task. For a cross-module, high-risk, or external-action change, use `--strict`: it requires an explicit scope-out, at least one acceptance scenario, at least one verification command or evidence handle, and a rollback boundary. Strict output contains machine-readable validation evidence and exits non-zero when a boundary is absent.
