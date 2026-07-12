$ErrorActionPreference = 'Stop'
$root = Split-Path $PSScriptRoot -Parent
$bin = Join-Path $root 'bin'
New-Item -ItemType Directory -Path $bin -Force | Out-Null
$go = & (Join-Path $PSScriptRoot 'resolve-locked-go.ps1') -Root $root
Push-Location $root
try {
  & $go build -trimpath -o (Join-Path $bin 'wuji.exe') ./cmd/wuji
  if ($LASTEXITCODE -ne 0) { throw 'Go build failed' }
} finally { Pop-Location }
Write-Output (Join-Path $bin 'wuji.exe')
