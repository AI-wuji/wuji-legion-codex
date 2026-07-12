param([switch]$CodebaseMemory)
$ErrorActionPreference = 'Stop'
$root = Split-Path $PSScriptRoot -Parent
$bin = Join-Path $root 'tools\bin'
$lock = Get-Content -Raw -Encoding UTF8 -LiteralPath (Join-Path $root 'sources.lock.json') | ConvertFrom-Json
New-Item -ItemType Directory -Path $bin -Force | Out-Null

function Install-LockedTool([string]$Id) {
  $matches = @($lock.context_tools | Where-Object id -eq $Id)
  if ($matches.Count -ne 1) { throw "sources.lock.json must contain exactly one $Id context tool" }
  $spec = $matches[0]
  $scratch = Join-Path $env:TEMP ('wuji-tool-' + $Id + '-' + [guid]::NewGuid().ToString('N'))
  $archive = Join-Path $scratch $spec.asset
  $extract = Join-Path $scratch 'extract'
  New-Item -ItemType Directory -Path $extract -Force | Out-Null
  try {
    Invoke-WebRequest -Headers @{'User-Agent'='Wuji-2.0'} -Uri $spec.url -OutFile $archive
    $actualHash = (Get-FileHash -Algorithm SHA256 -LiteralPath $archive).Hash.ToLowerInvariant()
    if ($actualHash -ne $spec.sha256.ToLowerInvariant()) {
      throw "$Id archive checksum mismatch: expected $($spec.sha256), got $actualHash"
    }
    Expand-Archive -LiteralPath $archive -DestinationPath $extract -Force
    $binary = Get-ChildItem -LiteralPath $extract -Recurse -File -Filter $spec.binary | Select-Object -First 1
    if (-not $binary) { throw "$Id archive does not contain $($spec.binary)" }
    $target = Join-Path $root $spec.path
    New-Item -ItemType Directory -Path (Split-Path $target -Parent) -Force | Out-Null
    Copy-Item -LiteralPath $binary.FullName -Destination $target -Force
    $versionText = & $target @($spec.version_args) 2>&1
    if ($LASTEXITCODE -ne 0 -or ($versionText -join ' ') -notmatch [regex]::Escape($spec.version)) {
      throw "$Id version verification failed: $($versionText -join ' ')"
    }
    Write-Output "$Id-installed version=$($spec.version)"
  } finally {
    Remove-Item -LiteralPath $scratch -Recurse -Force -ErrorAction SilentlyContinue
  }
}

Install-LockedTool 'rtk'
if ($CodebaseMemory) { Install-LockedTool 'codebase-memory-mcp' }
Write-Output 'context-tools-installed'
