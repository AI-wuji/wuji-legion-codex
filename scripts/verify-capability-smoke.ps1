param(
  [Parameter(Mandatory = $true)][string]$Capability,
  [string]$Root = $env:WUJI_ROOT
)
$ErrorActionPreference = 'Stop'
if (-not $Root) { $Root = Split-Path $PSScriptRoot -Parent }
$manifestPath = Join-Path $Root "capabilities\$Capability\manifest.json"
if (-not (Test-Path -LiteralPath $manifestPath)) { throw "manifest missing: $manifestPath" }

$manifest = Get-Content -Raw -Encoding UTF8 -LiteralPath $manifestPath | ConvertFrom-Json
if ($manifest.id -ne $Capability) { throw "manifest id mismatch: $($manifest.id) != $Capability" }
if (-not $manifest.primary_skill) { throw "primary_skill missing" }
if (-not $manifest.host_callable -and -not $manifest.direct_mount) {
  throw "callable capability must declare host_callable or direct_mount"
}

$resolved = 0
if ($manifest.sources) {
  foreach ($source in @($manifest.sources)) {
    $matched = $false
    foreach ($glob in @($source.globs)) {
      $projects = $env:WUJI_PROJECTS
      if (-not $projects) { $projects = [IO.Path]::GetFullPath((Join-Path $Root '..')) }
      $expanded = $glob.
        Replace('${ROOT}', $Root).
        Replace('${WUJI_PROJECTS}', $projects).
        Replace('${USERPROFILE}', $env:USERPROFILE)
      $expanded = [Environment]::ExpandEnvironmentVariables($expanded)
      $hits = @(Get-Item -Path $expanded -ErrorAction SilentlyContinue | Where-Object { $_.PSIsContainer })
      if ($hits.Count -gt 0) {
        $path = $hits[0].FullName
        foreach ($required in @($source.required)) {
          $reqHits = @(Get-ChildItem -Path (Join-Path $path $required) -ErrorAction SilentlyContinue)
          if ($reqHits.Count -eq 0) { throw "source $($source.id) missing required $required under $path" }
        }
        $matched = $true
        $resolved++
        break
      }
    }
    if (-not $matched -and $source.priority -eq 'primary') {
      throw "primary source not resolved: $($source.id)"
    }
  }
}

if ($manifest.direct_mount -and $resolved -lt 1) {
  throw "direct_mount capability resolved zero sources: $Capability"
}

# host-callable capabilities without sources: ensure CLI entry exists when primary_skill references wuji.
if ($manifest.host_callable -and $Capability -in @('code','context','evolution')) {
  $wuji = Join-Path $Root 'bin\wuji.exe'
  if (-not (Test-Path -LiteralPath $wuji)) {
    & (Join-Path $Root 'scripts\build.ps1') | Out-Null
  }
  if (-not (Test-Path -LiteralPath $wuji)) { throw "wuji binary missing for host-callable capability $Capability" }
  if ($Capability -eq 'evolution') {
    & $wuji help | Out-Null
    if ($LASTEXITCODE -ne 0) { throw 'wuji help failed' }
  }
  if ($Capability -eq 'context') {
    # real behavior probe already exists for context; smoke just confirms binary
    if (-not (Test-Path -LiteralPath $wuji)) { throw 'wuji missing' }
  }
}

Write-Output "smoke-ok capability=$Capability sources=$resolved status=$($manifest.status)"