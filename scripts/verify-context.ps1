$ErrorActionPreference = 'Stop'
$root = if ($env:WUJI_ROOT) { $env:WUJI_ROOT } else { Split-Path $PSScriptRoot -Parent }
$wuji = Join-Path $root 'bin\wuji.exe'
if (-not (Test-Path -LiteralPath $wuji)) { & (Join-Path $root 'scripts\build.ps1') }
$query = 'fix the code and verify it'
$raw = & $wuji context-select --workspace $root --query $query --max-bytes 2048
if ($LASTEXITCODE -ne 0) { throw 'context-select failed' }
$result = $raw | ConvertFrom-Json
if ($result.selected_bytes -gt 2048) { throw "context budget exceeded: $($result.selected_bytes)" }
if ($result.excerpts.Count -lt 1) { throw 'context selector returned no evidence' }
if (-not $result.query_fingerprint -or -not $result.content_sha256 -or $result.context_handle -ne "wuji-context://sha256/$($result.content_sha256)") {
  throw 'context selector did not return a content-addressed handle'
}
if (-not (Test-Path -LiteralPath $result.artifact_path -PathType Leaf)) { throw 'context artifact is missing' }
$routeRaw = & $wuji route --query $query --context-artifact $result.artifact_path
if ($LASTEXITCODE -ne 0) { throw 'route rejected the generated context artifact' }
$route = $routeRaw | ConvertFrom-Json
if (@($route.workers).Count -ne 1 -or $route.workers[0].model -ne 'gpt-5.6-terra') {
  throw 'verified compact context did not unlock one Terra implementation worker'
}
if ($route.workers[0].context_handles[0] -ne $result.context_handle -or $route.workers[0].allocated_context_bytes -ne $result.selected_bytes) {
  throw 'route did not preserve the verified context handoff'
}
$rtk = Join-Path $root 'tools\bin\rtk.exe'
$rtkState = 'not-installed-optional'
if (Test-Path -LiteralPath $rtk) {
  $version = & $rtk --version
  if ($LASTEXITCODE -ne 0) { throw 'installed RTK failed its version probe' }
  $rtkState = $version -join ' '
}
Write-Output "context-budget-ok bytes=$($result.selected_bytes) rtk=$rtkState"
$evidenceDir = $env:WUJI_PROBE_EVIDENCE_DIR
if (-not $evidenceDir -or -not (Test-Path -LiteralPath $evidenceDir -PathType Container)) {
  throw 'WUJI_PROBE_EVIDENCE_DIR is required for a behavior probe'
}
$reportPath = Join-Path $evidenceDir 'context-assertions.json'
$report = [ordered]@{
  fixture = 'context-selection-budget-v1'
  selected_bytes = [int]$result.selected_bytes
  excerpt_count = @($result.excerpts).Count
  within_budget = ([int]$result.selected_bytes -le 2048)
  context_handle = [string]$result.context_handle
  artifact_path = [string]$result.artifact_path
  delegated_worker_count = @($route.workers).Count
  cross_model_cache_assumed = [bool]$route.delegation_policy.cross_model_cache_assumed
  rtk = $rtkState
}
[IO.File]::WriteAllText($reportPath, ($report | ConvertTo-Json -Compress), [Text.UTF8Encoding]::new($false))
$reportHash = (Get-FileHash -Algorithm SHA256 -LiteralPath $reportPath).Hash.ToLowerInvariant()
Write-Output (@{
  wuji_probe = 'behavior'
  fixture = 'context-selection-budget-v1'
  passed = $true
  evidence = @(@{ id = 'assertions'; path = 'context-assertions.json'; sha256 = $reportHash })
  signature = 'context-selection-contract-v1'
} | ConvertTo-Json -Compress -Depth 5)
