$ErrorActionPreference = 'Stop'
$root = if ($env:WUJI_ROOT) { $env:WUJI_ROOT } else { Split-Path $PSScriptRoot -Parent }
$wuji = Join-Path $root 'bin\wuji.exe'
if (-not (Test-Path -LiteralPath $wuji)) { & (Join-Path $root 'scripts\build.ps1') }
$raw = & $wuji context-select --workspace $root --query 'behavior probe capability verification' --max-bytes 2048
if ($LASTEXITCODE -ne 0) { throw 'context-select failed' }
$result = $raw | ConvertFrom-Json
if ($result.selected_bytes -gt 2048) { throw "context budget exceeded: $($result.selected_bytes)" }
if ($result.excerpts.Count -lt 1) { throw 'context selector returned no evidence' }
$rtk = Join-Path $root 'tools\bin\rtk.exe'
$rtkState = 'not-installed-optional'
if (Test-Path -LiteralPath $rtk) {
  $version = & $rtk --version
  if ($LASTEXITCODE -ne 0) { throw 'installed RTK failed its version probe' }
  $rtkState = $version -join ' '
}
Write-Output "context-budget-ok bytes=$($result.selected_bytes) rtk=$rtkState"
