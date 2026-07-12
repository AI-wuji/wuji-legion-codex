param([string]$Root = '')
$ErrorActionPreference = 'Stop'
if (-not $Root) { $Root = Split-Path $PSScriptRoot -Parent }
$sourceLock = Get-Content -Raw -Encoding UTF8 -LiteralPath (Join-Path $Root 'sources.lock.json') | ConvertFrom-Json
$matches = @($sourceLock.toolchains | Where-Object id -eq 'go')
if ($matches.Count -ne 1) { throw 'sources.lock.json must contain exactly one Go toolchain' }
$toolchain = $matches[0]
$go = & (Join-Path $PSScriptRoot 'expand-wuji-path.ps1') -PathValue $toolchain.path -Root $Root
if (-not (Test-Path -LiteralPath $go -PathType Leaf)) { throw "Locked Go toolchain is missing: $go" }
$versionOutput = @(& $go version 2>&1)
$versionExitCode = $LASTEXITCODE
if ($versionExitCode -ne 0 -or ($versionOutput -join ' ') -notmatch [regex]::Escape($toolchain.version)) {
  throw "Locked Go toolchain version mismatch: expected $($toolchain.version)"
}
Write-Output $go
