# Graphify code-only pilot

This is a cold, optional adapter for `Graphify-Labs/graphify` commit `26b02b5e3430e4ab85dd7e72c7b98836d8e65c48` / PyPI `graphifyy==0.9.63` (Apache-2.0 with MIT notice). It is not a primary workspace index and does not install hooks, rules, watchers, MCP configuration, LLM providers, or global state.

`scripts/graphify-pilot.ps1` requires PowerShell 7. It copies only explicitly selected workspace files into a disposable directory and invokes upstream `graphify extract ... --code-only --no-cluster --max-workers 1`. The adapter accepts at most 32 files and 1 MiB by default, times out after 20 seconds, rejects lexical escapes and reparse points in the selected path chain, passes process arguments without shell joining, limits each captured process stream to 64 KiB, caps raw graph parsing at 16 MiB, and projects at most 256 nodes, 512 edges, and 16 KiB. Projected edges are retained only when both endpoint IDs remain in the bounded node projection; `truncated.edges` reports edges dropped for either endpoint or count limits. Every selected file carries its raw SHA-256 so consumers can reject stale results. Relations are compact pointers and are labelled `EXTRACTED`; the pilot performs no inference.

Example:

```powershell
./scripts/graphify-pilot.ps1 -Workspace . -File @('cmd/wuji/main.go','internal/router/router.go')
```

The adapter returns `status: unavailable` with exit code 3 when Graphify is absent or fails, `status: timeout` with exit code 4 on its deadline, and `status: output-limit` with exit code 5 when captured output is excessive. It never falls back to a home-grown parser. `scripts/test-graphify-pilot.ps1` uses a fake process boundary to verify paths containing spaces, selection, known call pointers, deletion and reparse-point escape rejection, dense graph caps, hashing, cleanup, timeout, and noisy process output; this validates the adapter, not upstream Graphify.

On the 2026-09-19 host, an actual pinned-runtime installation was attempted only under `.wuji/graphify-pilot-*`. It was unavailable: no `uv` or normal Python exists, and the bundled Python 3.14.5 lacks `zlib`, `ensurepip`, and `pip`. Therefore upstream extraction is **not behavior-verified** on this host and the candidate remains cold/unavailable. Re-run the example with an isolated, working Python environment containing exactly `graphifyy==0.9.63` before making any upstream capability claim.
