$ErrorActionPreference = 'Stop'
$root = Split-Path $PSScriptRoot -Parent
$bin = Join-Path $root 'bin'
New-Item -ItemType Directory -Path $bin -Force | Out-Null
$go = if ($env:WUJI_GO) { [IO.Path]::GetFullPath($env:WUJI_GO) } else { & (Join-Path $PSScriptRoot 'resolve-locked-go.ps1') -Root $root }
if (-not (Test-Path -LiteralPath $go -PathType Leaf)) { throw "Go toolchain is missing: $go" }
Push-Location $root
try {
  & $go build -buildvcs=false -trimpath -o (Join-Path $bin 'wuji.exe') ./cmd/wuji
  if ($LASTEXITCODE -ne 0) { throw 'Go build failed' }
} finally { Pop-Location }
Write-Output (Join-Path $bin 'wuji.exe')
