param([string]$Root = $env:WUJI_ROOT)
$ErrorActionPreference = 'Stop'
. (Join-Path $PSScriptRoot 'sha256.ps1')
if (-not $Root) { $Root = Split-Path $PSScriptRoot -Parent }
$projects = if ($env:WUJI_PROJECTS) { $env:WUJI_PROJECTS } else { [IO.Path]::GetFullPath((Join-Path $Root '..')) }
$ponytail = Join-Path $projects 'wuji-capability-sources\upstream-snapshots\ponytail'
$codexPython = Join-Path $env:USERPROFILE '.cache\codex-runtimes\codex-primary-runtime\dependencies\python\python.exe'
$python = if ($env:WUJI_PYTHON) {
  [IO.Path]::GetFullPath($env:WUJI_PYTHON)
} elseif (Test-Path -LiteralPath $codexPython -PathType Leaf) {
  $codexPython
} else {
  (Get-Command python -CommandType Application -ErrorAction Stop).Source
}
if (-not (Test-Path -LiteralPath $python -PathType Leaf)) { throw "Python runtime is missing: $python" }
foreach ($path in @('LICENSE','AGENTS.md','skills\ponytail\SKILL.md','tests\behavior.test.js','package.json')) {
  if (-not (Test-Path -LiteralPath (Join-Path $ponytail $path))) { throw "ponytail source is incomplete: $path" }
}

$previousPath = $env:Path
try {
  $env:Path = (Split-Path $python -Parent) + [IO.Path]::PathSeparator + $env:Path
  & npm test --prefix $ponytail
  if ($LASTEXITCODE -ne 0) { throw 'ponytail upstream tests failed' }
} finally {
  $env:Path = $previousPath
}

$wuji = Join-Path $Root 'bin\wuji.exe'
if (-not (Test-Path -LiteralPath $wuji)) { & (Join-Path $Root 'scripts\build.ps1') | Out-Null }
$raw = (& $wuji route --query 'implement parser normalization with a Go unit test' 2>&1) -join [Environment]::NewLine
if ($LASTEXITCODE -ne 0) { throw "code route failed: $raw" }
$route = $raw | ConvertFrom-Json
$worker = @($route.workers | Where-Object { $_.model -eq 'gpt-5.6-terra' } | Select-Object -First 1)
if ($route.capability -ne 'code' -or $worker.Count -ne 1 -or -not $worker[0].stable_capability_prefix) { throw 'code route did not emit one bounded Terra worker' }
$prefix = $worker[0].stable_capability_prefix | ConvertFrom-Json
if ($prefix.implementation_doctrine -ne 'ponytail-v3: universal-minimum-correct-task-judgment') { throw 'code route omitted the compact Ponytail doctrine marker' }
foreach ($required in @(
  'trace the actual flow and cite affected file or symbol anchors before choosing',
  'choose the first valid rung: skip, reuse local code, standard library, native platform, installed dependency, one line, minimum code',
  'for bugs, inspect every caller and fix the common root cause once, not each symptom',
  'prefer deletion, fewest files, and the smallest correct diff; no unrequested abstraction, scaffolding, or dependency',
  'for nontrivial logic, name one smallest runnable regression check; trivial one-line edits need no new test',
  'do not weaken validation, error handling, data safety, security, accessibility, or explicit requirements'
)) {
  if ($worker[0].protocol -notcontains $required -or $worker[0].task_contract -notmatch [regex]::Escape($required)) { throw "code route omitted executable Ponytail requirement: $required" }
}

$evidenceDir = $env:WUJI_PROBE_EVIDENCE_DIR
if (-not $evidenceDir -or -not (Test-Path -LiteralPath $evidenceDir -PathType Container)) { throw 'WUJI_PROBE_EVIDENCE_DIR is required for a behavior probe' }
$reportPath = Join-Path $evidenceDir 'code-ponytail-assertions.json'
$report = [ordered]@{
  fixture = 'code-ponytail-fusion-v3'
  source = $ponytail
  source_commit = ((& git -c "safe.directory=$($ponytail.Replace('\','/'))" -C $ponytail rev-parse HEAD) -join '').Trim()
  upstream_tests = 'passed'
  routed_model = $worker[0].model
  doctrine = $prefix.implementation_doctrine
}
[IO.File]::WriteAllText($reportPath, ($report | ConvertTo-Json -Compress), [Text.UTF8Encoding]::new($false))
$reportHash = (Get-FileHash -Algorithm SHA256 -LiteralPath $reportPath).Hash.ToLowerInvariant()
Write-Output (@{
  wuji_probe = 'behavior'
  fixture = 'code-ponytail-fusion-v3'
  passed = $true
  evidence = @(@{ id = 'assertions'; path = 'code-ponytail-assertions.json'; sha256 = $reportHash })
  signature = 'code-ponytail-fusion-v3'
} | ConvertTo-Json -Compress -Depth 5)
