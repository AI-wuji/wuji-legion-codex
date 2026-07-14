---
name: wuji-code-review-suite
description: Native, line-anchored code review for changed code, pull requests, and review-only requests.
---

# Wuji Code Review

Review the target diff or explicitly named files directly. Keep this read-only
unless the user explicitly asks for a fix.

1. Establish the review scope with Git status/diff or the supplied files.
2. Read the surrounding implementation and applicable tests before reporting a
   concern. Run focused existing tests when they can validate a claim.
3. Report only concrete bugs, regressions, security risks, or missing tests.
   Give every finding a file and line reference, ordered by severity.
4. If there are no findings, say so and identify any remaining verification
   gap. Do not turn style preferences into defects.

The retained Open Code Review CLI is an optional external integration because
it needs its own configured LLM credential. Do not install it, configure it,
or route user work through it unless the user explicitly requests that tool.
