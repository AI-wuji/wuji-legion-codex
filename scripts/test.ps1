param([switch]$Full)
$ErrorActionPreference = 'Stop'
$root = Split-Path $PSScriptRoot -Parent
& (Join-Path $PSScriptRoot 'build.ps1')
$go = & (Join-Path $PSScriptRoot 'resolve-locked-go.ps1') -Root $root
$gofmt = Join-Path (Split-Path $go -Parent) 'gofmt.exe'
$codexPython = Join-Path $env:USERPROFILE '.cache\codex-runtimes\codex-primary-runtime\dependencies\python\python.exe'
$python = if ($env:WUJI_PYTHON) {
  [IO.Path]::GetFullPath($env:WUJI_PYTHON)
} elseif (Test-Path -LiteralPath $codexPython -PathType Leaf) {
  $codexPython
} else {
  (Get-Command python -CommandType Application -ErrorAction Stop).Source
}
if (-not (Test-Path -LiteralPath $python -PathType Leaf)) { throw "Python runtime is missing: $python" }
$previousGoTmpDir = $env:GOTMPDIR
$previousGoCache = $env:GOCACHE
$previousPythonUtf8 = $env:PYTHONUTF8
$env:GOTMPDIR = Join-Path $root '.wuji\tmp'
$env:GOCACHE = Join-Path $root '.wuji\gocache'
New-Item -ItemType Directory -Force $env:GOTMPDIR,$env:GOCACHE | Out-Null
Push-Location $root
try {
  $unformatted = @(& $gofmt -l .\cmd .\internal)
  if ($LASTEXITCODE -ne 0 -or $unformatted.Count -gt 0) { throw "Go formatting check failed: $($unformatted -join ', ')" }
  & $go test ./...
  if ($LASTEXITCODE -ne 0) { throw 'Go tests failed' }
  $env:PYTHONUTF8 = '1'
  & $python (Join-Path $env:USERPROFILE '.codex\skills\.system\skill-creator\scripts\quick_validate.py') $root
  if ($LASTEXITCODE -ne 0) { throw 'Skill validation failed' }
  $nestedSkills = Get-ChildItem (Join-Path $root 'capabilities') -Recurse -Filter 'SKILL.md' | Where-Object { $_.FullName -match '\\skills\\' }
  foreach ($skillFile in $nestedSkills) {
    & $python (Join-Path $env:USERPROFILE '.codex\skills\.system\skill-creator\scripts\quick_validate.py') $skillFile.Directory.FullName
    if ($LASTEXITCODE -ne 0) { throw "Nested Skill validation failed: $($skillFile.Directory.FullName)" }
  }
  $auditMode = if ($Full) { 'full' } else { 'fast' }
  & (Join-Path $root 'scripts\audit.ps1') -Mode $auditMode -SkipBuild -SkipGoTest
  if ($LASTEXITCODE -ne 0) { throw 'Wuji audits failed' }
} finally {
  Pop-Location
  $env:GOTMPDIR = $previousGoTmpDir
  $env:GOCACHE = $previousGoCache
  $env:PYTHONUTF8 = $previousPythonUtf8
}
Write-Output 'wuji-3.0-tests-ok'
