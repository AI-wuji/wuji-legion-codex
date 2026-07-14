param(
  [Parameter(Mandatory = $true)][string]$Capability,
  [string]$Root = $env:WUJI_ROOT
)
$ErrorActionPreference = 'Stop'
if (-not $Root) { $Root = Split-Path $PSScriptRoot -Parent }
$manifestPath = Join-Path $Root "capabilities\$Capability\manifest.json"
if (-not (Test-Path -LiteralPath $manifestPath)) { throw "manifest missing: $manifestPath" }

$manifest = Get-Content -Raw -Encoding UTF8 -LiteralPath $manifestPath | ConvertFrom-Json
if ($manifest.id -ne $Capability) { throw "manifest id mismatch: $($manifest.id) != $Capability" }
if (-not $manifest.primary_skill) { throw "primary_skill missing" }
if (-not $manifest.host_callable -and -not $manifest.direct_mount) {
  throw "callable capability must declare host_callable or direct_mount"
}

$resolved = 0
if ($manifest.sources) {
  foreach ($source in @($manifest.sources)) {
    $matched = $false
    foreach ($glob in @($source.globs)) {
      $projects = $env:WUJI_PROJECTS
      if (-not $projects) { $projects = [IO.Path]::GetFullPath((Join-Path $Root '..')) }
      $expanded = $glob.
        Replace('${ROOT}', $Root).
        Replace('${WUJI_PROJECTS}', $projects).
        Replace('${USERPROFILE}', $env:USERPROFILE)
      $expanded = [Environment]::ExpandEnvironmentVariables($expanded)
      $hits = @(Get-Item -Path $expanded -ErrorAction SilentlyContinue | Where-Object { $_.PSIsContainer })
      if ($hits.Count -gt 0) {
        $path = $hits[0].FullName
        foreach ($required in @($source.required)) {
          $reqHits = @(Get-ChildItem -Path (Join-Path $path $required) -ErrorAction SilentlyContinue)
          if ($reqHits.Count -eq 0) { throw "source $($source.id) missing required $required under $path" }
        }
        $matched = $true
        $resolved++
        break
      }
    }
    if (-not $matched -and $source.priority -eq 'primary') {
      throw "primary source not resolved: $($source.id)"
    }
  }
}

if ($manifest.direct_mount -and $resolved -lt 1) {
  throw "direct_mount capability resolved zero sources: $Capability"
}

# host-callable capabilities without sources: ensure CLI entry exists when primary_skill references wuji.
if ($manifest.host_callable -and $Capability -in @('code','context','evolution')) {
  $wuji = Join-Path $Root 'bin\wuji.exe'
  if (-not (Test-Path -LiteralPath $wuji)) {
    & (Join-Path $Root 'scripts\build.ps1') | Out-Null
  }
  if (-not (Test-Path -LiteralPath $wuji)) { throw "wuji binary missing for host-callable capability $Capability" }
  if ($Capability -eq 'code') {
    $codeQuery = 'fix code workerPlan in internal/core/route.go'
    $rawDirectRoute = (& $wuji route --query $codeQuery 2>&1) -join [Environment]::NewLine
    if ($LASTEXITCODE -ne 0) { throw "wuji direct code route smoke failed: $rawDirectRoute" }
    $directRoute = $rawDirectRoute | ConvertFrom-Json
    if ($directRoute.delegation_decision.reason -ne 'verified-context-artifact-required' -or @($directRoute.workers).Count -ne 1) {
      throw 'code smoke did not create the bounded no-context task judgment'
    }
    $directWorker = @($directRoute.workers)[0]
    if ($directWorker.id -ne 'task-judgment' -or $directWorker.model -ne 'gpt-5.6-sol' -or $directWorker.writes -or $directWorker.context_mode -ne 'task-contract-only' -or $directWorker.allocated_context_bytes -ne 0) {
      throw 'code smoke did not keep the no-context judgment bounded and read-only'
    }
    $rawContext = (& $wuji context-select --workspace $Root --query $codeQuery --max-bytes 2048 2>&1) -join [Environment]::NewLine
    if ($LASTEXITCODE -ne 0) { throw "wuji code context smoke failed: $rawContext" }
    $codeContext = $rawContext | ConvertFrom-Json
    if (-not (Test-Path -LiteralPath $codeContext.artifact_path -PathType Leaf)) {
      throw 'code smoke did not create a context artifact'
    }
    $rawRoute = (& $wuji route --query $codeQuery --context-artifact $codeContext.artifact_path 2>&1) -join [Environment]::NewLine
    if ($LASTEXITCODE -ne 0) { throw "wuji code route smoke failed: $rawRoute" }
    $codeRoute = $rawRoute | ConvertFrom-Json
    if ($codeRoute.capability -ne 'code' -or $codeRoute.primary_skill -ne 'native Codex coding route') {
      throw 'code smoke did not reach the declared native Codex entrypoint'
    }
    if (@($codeRoute.workers).Count -ne 1 -or @($codeRoute.workers | Where-Object model -ne 'gpt-5.6-sol').Count -ne 0) {
      throw 'code smoke did not emit the bounded Sol worker plan'
    }
    $worker = @($codeRoute.workers)[0]
    if (-not $worker.execution_evidence_required -or ($worker.execution_evidence_fields -join ',') -ne 'schema_version,worker_id,requested_model,session_key,host_dispatch_id,write_boundary,attempts,effective_model,model_switch_count,result_handle,stable_prefix_bytes,stable_prefix_sha256,context_handle_ids,context_bytes_sent,context_payload_sha256,task_contract_bytes,task_contract_sha256,delegation_gate_reason,input_tokens,cached_input_tokens,output_tokens,retry_count,accepted_by_aji,attempt_failure_kinds,cache_domain,billing_unit,total_cost_microunits,aji_baseline_microunits,savings_microunits') {
      throw 'code smoke emitted an incomplete execution evidence contract'
    }
    if ($worker.context_mode -ne 'shared-content-addressed-handle' -or $worker.context_handles[0] -ne $codeContext.context_handle -or $worker.context_artifact -ne $codeContext.artifact_path) {
      throw 'code smoke did not hand off the verified content-addressed context'
    }
    if ($worker.allocated_context_bytes -ne $codeContext.payload_bytes -or $worker.context_payload_sha256 -ne $codeContext.payload_sha256 -or -not $worker.context_payload -or $worker.allocated_task_contract_bytes -ne ([Text.Encoding]::UTF8.GetByteCount([string]$worker.task_contract)) -or -not $worker.task_contract_sha256 -or $worker.max_task_contract_bytes -ne 2048) {
      throw 'code smoke did not expose bounded handoff costs'
    }
    if (-not $worker.stable_capability_prefix -or $worker.stable_prefix_bytes -ne ([Text.Encoding]::UTF8.GetByteCount([string]$worker.stable_capability_prefix)) -or -not $worker.stable_prefix_sha256 -or $codeRoute.delegation_decision.estimated_replay_bytes -ne ($worker.stable_prefix_bytes + $worker.allocated_context_bytes + $worker.allocated_task_contract_bytes)) {
      throw 'code smoke omitted the stable prefix from replay costs'
    }
    if (@($worker.fallback_models | Where-Object { $null -ne $_ -and $_ -ne '' }).Count -ne 0 -or $worker.max_attempts -ne 1 -or @($worker.fallback_on | Where-Object { $null -ne $_ -and $_ -ne '' }).Count -ne 0 -or $worker.max_model_switches -ne 0 -or ($worker.prompt_order -join ',') -ne 'stable_capability_prefix,context_payload,task_contract') {
      throw 'code smoke emitted an unsafe fallback or prompt policy'
    }
    if ($codeRoute.delegation_policy.cross_model_cache_assumed -or $codeRoute.delegation_policy.cache_scope -ne 'model-local stable-prefix only' -or $codeRoute.delegation_decision.context_coverage_basis_points -lt 6000 -or $codeRoute.delegation_decision.code_excerpt_count -lt 1 -or $codeRoute.delegation_decision.content_anchor_count -lt 1 -or -not $codeRoute.delegation_decision.allowed -or $codeRoute.execution_lane -ne 'bounded-delegation') {
      throw 'code smoke did not enforce the cross-model cost gate'
    }
  }
  if ($Capability -eq 'evolution') {
    $candidate = Join-Path $Root 'scripts\fixtures\evolution-label-only.json'
    if (-not (Test-Path -LiteralPath $candidate)) { throw "evolution smoke candidate missing: $candidate" }
    $rawResult = (& $wuji evolve --root $Root --candidate $candidate 2>&1) -join [Environment]::NewLine
    if ($LASTEXITCODE -ne 0) { throw "wuji evolve smoke failed: $rawResult" }
    $evolutionResult = $rawResult | ConvertFrom-Json
    if ($evolutionResult.decision -ne 'reject' -or $evolutionResult.applied -or $evolutionResult.candidate_proof.passed) {
      throw 'evolution smoke did not enforce the behavior evidence gate'
    }
    if (Test-Path -LiteralPath (Join-Path $Root 'capabilities\evolution-smoke-candidate')) {
      throw 'evolution smoke unexpectedly mutated the capability registry'
    }
  }
  if ($Capability -eq 'context') {
    # real behavior probe already exists for context; smoke just confirms binary
    if (-not (Test-Path -LiteralPath $wuji)) { throw 'wuji missing' }
  }
}

Write-Output "smoke-ok capability=$Capability sources=$resolved status=$($manifest.status)"
