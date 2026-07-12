# Regression Matrix

Date: 2026-07-12

| Surface | Evidence | Result |
|---|---|---|
| Go | format, `go test ./...`, `go vet ./...` | pass |
| Skills | root plus 11 nested Skills validated | pass |
| Capability truth | `scripts/test.ps1 -Full`, 14 capability probes | pass |
| Fusion | lifecycle/evidence audit | pass |
| Optimization | one router; size limits | pass |
| Context | sparse selection and hard budget | pass |
| Migration | 104 unique legacy objects; 52 unique worktree paths | pass |
| Upstreams | 4 explicit verdicts; 3 reviewed HEAD locks; Open Design excluded | pass |
| Presentation | catalog web=109/editable=69; editable PPTX; HTML render; Slidev 52.17.0 build and browser render; moving stage fluid; retained Huashu/humanize/Baoyu entrypoints | pass |
| Dependencies | production audit at severity low | 0 known vulnerabilities |
| Repository hygiene | no repository `node_modules` or `dist` | pass |

Commands:

```powershell
./scripts/verify-presentation.ps1
./scripts/test.ps1
./scripts/test.ps1 -Full
pnpm audit --prod --audit-level low --registry https://registry.npmjs.org
```

The final full run completed in 252.7 seconds with `audit-mode=full verified=14` and `wuji-2.0-tests-ok`.
