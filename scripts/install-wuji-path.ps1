param(
  [string]$Root = '',
  [switch]$User
)
$ErrorActionPreference = 'Stop'
if (-not $Root) { $Root = Split-Path $PSScriptRoot -Parent }
$bin = Join-Path $Root 'bin'
if (-not (Test-Path -LiteralPath (Join-Path $bin 'wuji.exe'))) {
  & (Join-Path $Root 'scripts\build.ps1') | Out-Null
}
$target = if ($User) {
  [Environment]::GetEnvironmentVariable('Path', 'User')
} else {
  $env:Path
}
$parts = @($target -split ';' | Where-Object { $_ -and $_.Trim() -ne '' })
if ($parts -notcontains $bin) {
  $parts = @($bin) + $parts
  $joined = ($parts -join ';').TrimEnd(';')
  if ($User) {
    [Environment]::SetEnvironmentVariable('Path', $joined, 'User')
    Write-Output "Added user PATH entry: $bin (restart shells to pick up)"
  } else {
    $env:Path = $joined
    Write-Output "Added session PATH entry: $bin"
  }
} else {
  Write-Output "PATH already contains: $bin"
}
if (Get-Command wuji -ErrorAction SilentlyContinue) {
  Write-Output ("wuji -> " + (Get-Command wuji).Source)
} else {
  Write-Output "wuji not yet visible in this shell; call: $bin\wuji.exe"
}