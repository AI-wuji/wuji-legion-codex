param(
  [string]$Root = (Split-Path $PSScriptRoot -Parent),
  [string]$Destination = (Join-Path $env:LOCALAPPDATA 'OfficeCLI\officecli.exe')
)

$ErrorActionPreference = 'Stop'
$lock = Get-Content -Raw -Encoding UTF8 -LiteralPath (Join-Path $Root 'sources.lock.json') | ConvertFrom-Json
$spec = @($lock.office_tools | Where-Object id -eq 'officecli')[0]
if (-not $spec) { throw 'OfficeCLI lock entry is missing' }
$scratch = Join-Path $env:TEMP ('wuji-officecli-install-' + [guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Force -Path $scratch | Out-Null
try {
  $download = Join-Path $scratch $spec.binary
  Invoke-WebRequest -Headers @{ 'User-Agent' = 'Wuji-2.0' } -Uri $spec.url -OutFile $download
  $actual = (Get-FileHash -Algorithm SHA256 -LiteralPath $download).Hash.ToLowerInvariant()
  if ($actual -ne $spec.sha256.ToLowerInvariant()) { throw "OfficeCLI checksum mismatch: expected $($spec.sha256), got $actual" }
  $target = [IO.Path]::GetFullPath($Destination)
  New-Item -ItemType Directory -Force -Path (Split-Path -Parent $target) | Out-Null
  Copy-Item -LiteralPath $download -Destination $target -Force
  $version = @(& $target --version 2>&1) -join ' '
  if ($LASTEXITCODE -ne 0 -or $version -notmatch [regex]::Escape($spec.version)) { throw "OfficeCLI version verification failed: $version" }
  [pscustomobject]@{ installed = $true; version = $spec.version; binary = $target; sha256 = $actual; path_modified = $false; skills_installed = $false; mcp_configured = $false } | ConvertTo-Json -Compress
} finally {
  Remove-Item -LiteralPath $scratch -Recurse -Force -ErrorAction SilentlyContinue
}
