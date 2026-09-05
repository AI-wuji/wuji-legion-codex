$ErrorActionPreference = 'Stop'
. (Join-Path $PSScriptRoot 'sha256.ps1')
$root = if ($env:WUJI_ROOT) { $env:WUJI_ROOT } else { Split-Path $PSScriptRoot -Parent }
$wuji = Join-Path $root 'bin\wuji.exe'
if (-not (Test-Path -LiteralPath $wuji)) { & (Join-Path $root 'scripts\build.ps1') }
$query = 'fix code workerPlan in internal/core/route.go'
$raw = & $wuji context-select --workspace $root --query $query --max-bytes 2048
if ($LASTEXITCODE -ne 0) { throw 'context-select failed' }
$result = $raw | ConvertFrom-Json
if ($result.selected_bytes -gt 2048) { throw "context budget exceeded: $($result.selected_bytes)" }
if ($result.excerpts.Count -lt 1) { throw 'context selector returned no evidence' }
if ($result.coverage_basis_points -lt 6000 -or $result.code_excerpt_count -lt 1 -or $result.content_anchor_count -lt 1) { throw 'context selector returned low-quality code evidence' }
if ($result.selected_bytes -ne $result.payload_bytes -or -not $result.payload_sha256) { throw 'context selector did not expose deterministic prompt payload metadata' }
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
$worker = @($route.workers)[0]
if (-not $worker.context_payload -or $worker.context_payload_sha256 -ne $result.payload_sha256 -or ([Text.Encoding]::UTF8.GetByteCount([string]$worker.context_payload)) -ne $result.payload_bytes) {
  throw 'route did not send the deterministic context payload'
}
if (-not $worker.task_contract -or $worker.allocated_task_contract_bytes -ne ([Text.Encoding]::UTF8.GetByteCount([string]$worker.task_contract)) -or -not $worker.task_contract_sha256) {
  throw 'route did not send a measured task contract'
}
if (-not $worker.stable_capability_prefix -or $worker.stable_prefix_bytes -ne ([Text.Encoding]::UTF8.GetByteCount([string]$worker.stable_capability_prefix)) -or -not $worker.stable_prefix_sha256) {
  throw 'route did not send a measured stable capability prefix'
}
if ($route.delegation_decision.estimated_replay_bytes -ne ($worker.stable_prefix_bytes + $worker.source_execution_bytes + $worker.allocated_context_bytes + $worker.allocated_task_contract_bytes)) {
  throw 'route replay estimate omitted a prompt component'
}
$expectedPromptOrder = @('stable_capability_prefix')
if ($worker.source_execution_bytes -gt 0) {
  if (@($worker.source_execution).Count -lt 1 -or @($worker.source_execution | Where-Object { -not $_.source_id -or -not $_.entrypoint_sha256 -or $_.entrypoint_bytes -le 0 }).Count -gt 0) {
    throw 'route emitted an incomplete source execution contract'
  }
  $expectedPromptOrder += 'source_execution'
} elseif (@($worker.source_execution | Where-Object { $_.source_id -or $_.entrypoint -or $_.entrypoint_bytes -gt 0 }).Count -gt 0) {
  throw 'route emitted an empty source execution contract'
}
$expectedPromptOrder += @('context_payload','task_contract')
if (($worker.prompt_order -join ',') -ne ($expectedPromptOrder -join ',') -or @($worker.fallback_models | Where-Object { $null -ne $_ -and $_ -ne '' }).Count -ne 0 -or $worker.max_attempts -ne 1 -or @($worker.fallback_on | Where-Object { $null -ne $_ -and $_ -ne '' }).Count -ne 0 -or $worker.max_model_switches -ne 0) {
  throw 'route emitted an unsafe prompt or retry policy'
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
  cache_scope = [string]$route.delegation_policy.cache_scope
  coverage_basis_points = [int]$result.coverage_basis_points
  code_excerpt_count = [int]$result.code_excerpt_count
  content_anchor_count = [int]$result.content_anchor_count
  payload_sha256 = [string]$result.payload_sha256
  stable_prefix_bytes = [int]$worker.stable_prefix_bytes
  source_execution_bytes = [int]$worker.source_execution_bytes
  task_contract_bytes = [int]$worker.allocated_task_contract_bytes
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
