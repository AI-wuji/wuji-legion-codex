$ErrorActionPreference = 'Stop'
$root = if ($env:WUJI_ROOT) { $env:WUJI_ROOT } else { Split-Path $PSScriptRoot -Parent }
$sourceLock = Get-Content -Raw -Encoding UTF8 -LiteralPath (Join-Path $root 'sources.lock.json') | ConvertFrom-Json
$sourceMatches = @($sourceLock.sources | Where-Object id -eq 'open-code-review')
if ($sourceMatches.Count -ne 1) { throw 'sources.lock.json must contain exactly one open-code-review source' }
$source = & (Join-Path $PSScriptRoot 'expand-wuji-path.ps1') -PathValue $sourceMatches[0].path -Root $root
if (-not (Test-Path -LiteralPath $source -PathType Container)) { throw "Locked source is missing: open-code-review ($source)" }
$go = & (Join-Path $PSScriptRoot 'resolve-locked-go.ps1') -Root $root
$cache = Join-Path $root '.wuji\open-code-review'
$previousGoCache = $env:GOCACHE
$previousGoModCache = $env:GOMODCACHE
$previousGoTmpDir = $env:GOTMPDIR
$previousGoProxy = $env:GOPROXY
$env:GOCACHE = Join-Path $cache 'build'
$env:GOMODCACHE = Join-Path $cache 'modules'
$env:GOTMPDIR = Join-Path $cache 'tmp'
$env:GOPROXY = if ($env:GOPROXY) { $env:GOPROXY } else { 'https://goproxy.cn,direct' }
New-Item -ItemType Directory -Force $env:GOCACHE,$env:GOMODCACHE,$env:GOTMPDIR | Out-Null
$out = Join-Path $env:TEMP ('opencodereview-' + [guid]::NewGuid().ToString('N') + '.exe')
Push-Location $source
try {
  & $go build -trimpath -o $out ./cmd/opencodereview
  if ($LASTEXITCODE -ne 0) { throw 'Open Code Review build failed' }
  $help = & $out --help 2>&1
  if ($LASTEXITCODE -ne 0 -or ($help -join "`n") -notmatch 'review|config|usage') { throw 'Open Code Review CLI probe failed' }
  Write-Output "open-code-review-built bytes=$((Get-Item $out).Length)"
} finally {
  Pop-Location
  Remove-Item -LiteralPath $out -Force -ErrorAction SilentlyContinue
  $env:GOCACHE = $previousGoCache
  $env:GOMODCACHE = $previousGoModCache
  $env:GOTMPDIR = $previousGoTmpDir
  $env:GOPROXY = $previousGoProxy
}
