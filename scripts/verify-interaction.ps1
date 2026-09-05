param([string]$Root = $env:WUJI_ROOT)
$ErrorActionPreference = 'Stop'
. (Join-Path $PSScriptRoot 'sha256.ps1')
if (-not $Root) { $Root = Split-Path $PSScriptRoot -Parent }

$go = Join-Path $env:LOCALAPPDATA 'WujiLegion\go-manual\go\bin\go.exe'
if (-not (Test-Path -LiteralPath $go -PathType Leaf)) { throw "Go runtime is missing: $go" }
$evidenceDir = $env:WUJI_PROBE_EVIDENCE_DIR
if (-not $evidenceDir -or -not (Test-Path -LiteralPath $evidenceDir -PathType Container)) { throw 'WUJI_PROBE_EVIDENCE_DIR is required for a behavior probe' }
$wuji = Join-Path $evidenceDir 'wuji-interaction-probe.exe'

Push-Location $Root
try {
  & $go test ./internal/core -run '^TestResponsePolicy' -count=1
  if ($LASTEXITCODE -ne 0) { throw 'response-policy behavior tests failed' }
  & $go build -buildvcs=false -o $wuji ./cmd/wuji
  if ($LASTEXITCODE -ne 0) { throw 'building the interaction probe binary failed' }
} finally {
  Pop-Location
}

$raw = (& $wuji route --query 'enable action focus: fix this multi-step bug' 2>&1) -join [Environment]::NewLine
if ($LASTEXITCODE -ne 0) { throw "interaction route failed: $raw" }
$route = $raw | ConvertFrom-Json
if ($route.capability -ne 'code') { throw 'response policy displaced the code capability' }
if (-not $route.response_policy.active -or $route.response_policy.activation_reason -ne 'explicit-activation') { throw 'response policy did not activate' }
if (@($route.secondary_capabilities) -notcontains 'interaction') { throw 'interaction overlay was not reported' }
foreach ($id in @('host-safety','explicit-user-instruction','first-action','bounded-steps','single-next-action')) {
  if (@($route.response_policy.directives.id) -notcontains $id) { throw "compiled policy omitted $id" }
}

$explainRaw = (& $wuji response-policy --query 'enable action focus: explain why this failed' 2>&1) -join [Environment]::NewLine
if ($LASTEXITCODE -ne 0) { throw "explanation policy failed: $explainRaw" }
$explain = $explainRaw | ConvertFrom-Json
foreach ($id in @('first-action','bounded-steps','single-next-action')) {
  if (@($explain.directives.id) -contains $id -or @($explain.suppressed) -notcontains $id) { throw "explanation override failed for $id" }
}

$negatedRaw = (& $wuji response-policy --query 'do not stop focus mode' --active 2>&1) -join [Environment]::NewLine
if ($LASTEXITCODE -ne 0) { throw "negated exit policy failed: $negatedRaw" }
$negated = $negatedRaw | ConvertFrom-Json
if (-not $negated.active -or $negated.activation_reason -ne 'session-active') { throw 'negated exit disabled the policy' }

$exitRaw = (& $wuji response-policy --query 'return to normal mode' --active 2>&1) -join [Environment]::NewLine
if ($LASTEXITCODE -ne 0) { throw "exit policy failed: $exitRaw" }
$exit = $exitRaw | ConvertFrom-Json
if ($exit.active -or $exit.activation_reason -ne 'explicit-exit' -or ($null -ne $exit.directives -and @($exit.directives).Count -ne 0)) { throw 'explicit exit did not disable the policy' }

$rulesPath = Join-Path $Root 'capabilities\interaction\references\rules.json'
$rulesHash = (Get-FileHash -Algorithm SHA256 -LiteralPath $rulesPath).Hash.ToLowerInvariant()
if ($route.response_policy.rules_sha256 -ne $rulesHash) { throw 'route rule hash does not match the trusted asset' }

$reportPath = Join-Path $evidenceDir 'action-focus-assertions.json'
$binaryHash = (Get-FileHash -Algorithm SHA256 -LiteralPath $wuji).Hash.ToLowerInvariant()
$report = [ordered]@{
  fixture = 'action-focus-compiler-route-v2'
  verification_scope = 'compiler-and-route-contract'
  source_commit = $route.response_policy.source_commit
  rules_sha256 = $rulesHash
  probe_binary_sha256 = $binaryHash
  domain_capability_preserved = $route.capability
  activation = $route.response_policy.activation_reason
  explanation_suppressed = @($explain.suppressed)
  exit = $exit.activation_reason
}
[IO.File]::WriteAllText($reportPath, ($report | ConvertTo-Json -Compress -Depth 5), [Text.UTF8Encoding]::new($false))
$reportHash = (Get-FileHash -Algorithm SHA256 -LiteralPath $reportPath).Hash.ToLowerInvariant()
Write-Output (@{
  wuji_probe = 'behavior'
  fixture = 'action-focus-compiler-route-v2'
  passed = $true
  evidence = @(@{ id = 'assertions'; path = 'action-focus-assertions.json'; sha256 = $reportHash })
  signature = 'action-focus-compiler-route-v2'
} | ConvertTo-Json -Compress -Depth 5)
