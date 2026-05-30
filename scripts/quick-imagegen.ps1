param(
  [Parameter(Mandatory=$true)]
  [string]$Prompt,

  [string]$Out = "",

  [string]$Size = "1024x1024",

  [string]$Quality = "medium",

  [string]$Model = "gpt-image-2"
)

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
$imagegen = Join-Path $env:USERPROFILE ".codex\skills\imagegen\scripts\image_gen.ps1"

if (-not (Test-Path -LiteralPath $imagegen)) {
  throw "imagegen PowerShell runner not found: $imagegen"
}

if ([string]::IsNullOrWhiteSpace($Out)) {
  $outDir = Join-Path $repoRoot "output\imagegen"
  New-Item -ItemType Directory -Force -Path $outDir | Out-Null
  $stamp = Get-Date -Format "yyyyMMdd-HHmmss"
  $Out = Join-Path $outDir "image-$stamp.png"
} else {
  $outParent = Split-Path -Parent $Out
  if ($outParent) {
    New-Item -ItemType Directory -Force -Path $outParent | Out-Null
  }
}

powershell.exe -NoProfile -ExecutionPolicy Bypass -File $imagegen `
  -Prompt $Prompt `
  -Model $Model `
  -Quality $Quality `
  -Size $Size `
  -Out $Out `
  -Force

if (-not (Test-Path -LiteralPath $Out)) {
  throw "imagegen finished without creating output: $Out"
}

Write-Output $Out
