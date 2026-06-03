param(
    [Parameter(Mandatory = $true)]
    [string]$Workspace,

    [Parameter(Mandatory = $true)]
    [string]$StarterPptx,

    [Parameter(Mandatory = $true)]
    [string]$Map,

    [Parameter(Mandatory = $true)]
    [string]$Out,

    [string]$PreviewDir = "",
    [string]$LayoutDir = "",
    [string]$Report = "",
    [switch]$NoPreview,
    [switch]$NoLayout,
    [string]$Scale = "",
    [string]$SkillDir = "",
    [string]$NodePath = "",
    [string]$NodeModules = "",
    [string]$PythonPath = ""
)

$ErrorActionPreference = "Stop"
. (Join-Path $PSScriptRoot 'wuji-ppt-runtime.ps1')

$runtime = Get-WujiPptRuntime -SkillDir $SkillDir -NodePath $NodePath -NodeModules $NodeModules -PythonPath $PythonPath
$scriptPath = Join-Path $PSScriptRoot "wuji-ppt-template-edit.mjs"
$resolvedWorkspace = [System.IO.Path]::GetFullPath($Workspace)
$null = New-Item -ItemType Directory -Force -Path $resolvedWorkspace
Initialize-WujiArtifactWorkspace -Runtime $runtime -Workspace $resolvedWorkspace
$resolvedStarterPptx = (Resolve-Path -LiteralPath $StarterPptx).Path
$resolvedMap = (Resolve-Path -LiteralPath $Map).Path

$argsList = @("--workspace", $resolvedWorkspace, "--starter-pptx", $resolvedStarterPptx, "--map", $resolvedMap, "--out", [System.IO.Path]::GetFullPath($Out))
if ($PreviewDir) {
    $argsList += @("--preview-dir", [System.IO.Path]::GetFullPath($PreviewDir))
}
if ($LayoutDir) {
    $argsList += @("--layout-dir", [System.IO.Path]::GetFullPath($LayoutDir))
}
if ($Report) {
    $argsList += @("--report", [System.IO.Path]::GetFullPath($Report))
}
if ($NoPreview) {
    $argsList += "--no-preview"
}
if ($NoLayout) {
    $argsList += "--no-layout"
}
if ($Scale) {
    $argsList += @("--scale", $Scale)
}

$code = Invoke-WujiNodeScript -Runtime $runtime -ScriptPath $scriptPath -Arguments $argsList
if ($code -ne 0) {
    throw "wuji-ppt-template-edit failed with exit code $code"
}
