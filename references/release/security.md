# Security Evidence

Date: 2026-07-12

Checks completed:

- `go vet ./...`: passed.
- Go module graph: project module only; no third-party Go dependency.
- Strong credential-pattern review: no real API key, token, or session material found.
- `pnpm audit --prod --audit-level low --registry https://registry.npmjs.org`: no known vulnerabilities.
- Slidev lock: `@slidev/cli=52.17.0`; all DOMPurify paths are overridden to `3.4.12`.
- Full repository audit: credential patterns, hard-coded user paths, legacy runtime pointers, persistent Xiaobai bridge/GUI/daemon/service/shortcut, and smoke-as-fusion claims are rejected.
- Dangerous process and delete sites were reviewed as bounded temporary-directory cleanup, pinned downloads, or controlled subprocesses.

Tool availability limits: `gitleaks`, `semgrep`, `govulncheck`, `trivy`, and PSScriptAnalyzer were not installed, so no result is claimed for them. The local npm mirror has no advisory endpoint; the dependency audit therefore used the official npm registry for that invocation only. `minimumReleaseAge` is not configured; no age-gate claim is made.

Residual risk: future dependency or source updates require rerunning the official-registry audit, tree-hash verification, and full behavior suite before release.
