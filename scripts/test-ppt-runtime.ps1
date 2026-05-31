param(
  [string]$SkillDir = "",
  [string]$NodePath = "",
  [string]$NodeModules = ""
)

$ErrorActionPreference = "Stop"

if ([string]::IsNullOrWhiteSpace($SkillDir)) {
  $presentationsRoot = Join-Path $env:USERPROFILE ".codex\plugins\cache\openai-primary-runtime\presentations"
  $latest = Get-ChildItem -LiteralPath $presentationsRoot -Directory -ErrorAction SilentlyContinue |
    Sort-Object Name -Descending |
    Select-Object -First 1
  if ($latest) {
    $SkillDir = Join-Path $latest.FullName "skills\presentations"
  }
}

if ([string]::IsNullOrWhiteSpace($NodePath)) {
  $NodePath = Join-Path $env:USERPROFILE ".cache\codex-runtimes\codex-primary-runtime\dependencies\node\bin\node.exe"
}

if ([string]::IsNullOrWhiteSpace($NodeModules)) {
  $NodeModules = Join-Path $env:USERPROFILE ".cache\codex-runtimes\codex-primary-runtime\dependencies\node\node_modules"
}

if (-not (Test-Path -LiteralPath $NodePath)) {
  throw "Bundled Node not found: $NodePath"
}

if (-not (Test-Path -LiteralPath $NodeModules)) {
  throw "Bundled node_modules not found: $NodeModules"
}

if (-not (Test-Path -LiteralPath $SkillDir)) {
  throw "Presentations skill not found: $SkillDir"
}

$env:HOME = $env:USERPROFILE
$env:NODE_PATH = $NodeModules

$PresentationJsx = Join-Path $NodeModules "@oai\artifact-tool\dist\presentation-jsx\index.mjs"
if (-not (Test-Path -LiteralPath $PresentationJsx)) {
  throw "presentation-jsx entry not found: $PresentationJsx"
}

$PresentationJsxUrl = ([System.Uri]$PresentationJsx).AbsoluteUri
$probe = @"
const mod = await import('$PresentationJsxUrl');
if (!mod.Fragment || !mod.createRef) throw new Error('presentation-jsx import failed');
console.log('artifact-tool-ok');
"@

$probePath = Join-Path $env:TEMP "wuji-ppt-runtime-probe.mjs"
Set-Content -LiteralPath $probePath -Value $probe -Encoding UTF8

try {
  & $NodePath $probePath
  if ($LASTEXITCODE -ne 0) {
    throw "artifact-tool probe failed with exit code $LASTEXITCODE"
  }
}
finally {
  Remove-Item -LiteralPath $probePath -Force -ErrorAction SilentlyContinue
}

$unzip = Get-Command unzip -ErrorAction SilentlyContinue
if ($unzip) {
  Write-Output "unzip-found: $($unzip.Source)"
}
else {
  Write-Output "unzip-not-found: use PowerShell/.NET ZipArchive fallback"
}

Write-Output "ppt-runtime-ok"
