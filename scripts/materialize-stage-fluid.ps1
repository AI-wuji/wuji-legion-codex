param(
  [Parameter(Mandatory=$true)][string]$OutputDir,
  [string]$SourceDir = ''
)
$ErrorActionPreference = 'Stop'
$root = Split-Path $PSScriptRoot -Parent
if (-not $SourceDir) {
  $lock = Get-Content -Raw -Encoding UTF8 -LiteralPath (Join-Path $root 'sources.lock.json') | ConvertFrom-Json
  $lockedSource = @($lock.sources | Where-Object id -eq 'stage-fluid')
  if ($lockedSource.Count -ne 1) { throw 'sources.lock.json must contain exactly one stage-fluid source' }
  $SourceDir = & (Join-Path $PSScriptRoot 'expand-wuji-path.ps1') -PathValue $lockedSource[0].path -Root $root
}
$source = [IO.Path]::GetFullPath($SourceDir)
foreach ($required in @('stage-fluid.js','stage-fluid.css')) {
  if (-not (Test-Path -LiteralPath (Join-Path $source $required) -PathType Leaf)) {
    throw "Canonical fluid source is missing $required`: $source"
  }
}
$output = [IO.Path]::GetFullPath($OutputDir)
New-Item -ItemType Directory -Path $output -Force | Out-Null
Copy-Item -LiteralPath (Join-Path $source 'stage-fluid.js') -Destination (Join-Path $output 'stage-fluid.js') -Force
$approvedCss = [IO.File]::ReadAllText((Join-Path $source 'stage-fluid.css'), [Text.Encoding]::UTF8)
$stageCss = @'

html, body { min-height: 100%; background: #05070b; }
body { display: grid; place-items: center; overflow: hidden; }
.stage { position: relative; width: min(100vw, 1280px); aspect-ratio: 16 / 9; overflow: hidden; background: #05070b; }
.stage canvas { position: absolute; inset: 0; display: block; width: 100%; height: 100%; }
.stage-content { position: relative; z-index: 1; display: grid; place-items: center; height: 100%; color: rgba(255,255,255,.82); font: 600 28px/1.2 system-ui; pointer-events: none; }
'@
[IO.File]::WriteAllText((Join-Path $output 'stage-fluid.css'), ($approvedCss.TrimEnd() + "`r`n" + $stageCss), [Text.UTF8Encoding]::new($false))
$html = @'
<!doctype html>
<html lang="en">
<head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><link rel="stylesheet" href="./stage-fluid.css"><title>Stage Fluid</title></head>
<body><main class="stage"><canvas aria-hidden="true"></canvas><div class="stage-content">Wuji Stage Fluid</div></main><script src="./stage-fluid.js"></script></body>
</html>
'@
[IO.File]::WriteAllText((Join-Path $output 'index.html'), $html, [Text.UTF8Encoding]::new($false))
Write-Output (Join-Path $output 'index.html')
