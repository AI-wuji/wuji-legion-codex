param(
  [Parameter(Mandatory = $true)][string]$PathValue,
  [string]$Root = ''
)
if (-not $Root) {
  if ($env:WUJI_ROOT) { $Root = $env:WUJI_ROOT }
  else { $Root = Split-Path $PSScriptRoot -Parent }
}
$projects = $env:WUJI_PROJECTS
if (-not $projects) { $projects = [IO.Path]::GetFullPath((Join-Path $Root '..')) }
$value = [string]$PathValue
$value = $value.Replace('${ROOT}', $Root)
$value = $value.Replace('${WUJI_PROJECTS}', $projects)
$value = $value.Replace('${USERPROFILE}', $env:USERPROFILE)
$value = $value.Replace('${LOCALAPPDATA}', $env:LOCALAPPDATA)
$value = [Environment]::ExpandEnvironmentVariables($value)
return [IO.Path]::GetFullPath($value)
