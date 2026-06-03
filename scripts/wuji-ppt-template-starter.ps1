param(
    [Parameter(Mandatory = $true)]
    [string]$Workspace,

    [Parameter(Mandatory = $true)]
    [string]$Pptx,

    [Parameter(Mandatory = $true)]
    [string]$Map,

    [Parameter(Mandatory = $true)]
    [string]$Out,

    [string]$PreviewDir = "",
    [string]$LayoutDir = "",
    [string]$Inspect = "",
    [string]$ContactSheet = "",
    [string]$Scale = "",
    [string]$SkillDir = "",
    [string]$NodePath = "",
    [string]$NodeModules = "",
    [string]$PythonPath = ""
)

$ErrorActionPreference = "Stop"
. (Join-Path $PSScriptRoot 'wuji-ppt-runtime.ps1')

$runtime = Get-WujiPptRuntime -SkillDir $SkillDir -NodePath $NodePath -NodeModules $NodeModules -PythonPath $PythonPath
$scriptPath = Join-Path $runtime.SkillDir "scripts\prepare_template_starter_deck.mjs"
$resolvedWorkspace = [System.IO.Path]::GetFullPath($Workspace)
$null = New-Item -ItemType Directory -Force -Path $resolvedWorkspace
Initialize-WujiArtifactWorkspace -Runtime $runtime -Workspace $resolvedWorkspace
$resolvedPptx = (Resolve-Path -LiteralPath $Pptx).Path
$resolvedMap = (Resolve-Path -LiteralPath $Map).Path

$argsList = @("--workspace", $resolvedWorkspace, "--pptx", $resolvedPptx, "--map", $resolvedMap, "--out", [System.IO.Path]::GetFullPath($Out))
if ($PreviewDir) {
    $argsList += @("--preview-dir", [System.IO.Path]::GetFullPath($PreviewDir))
}
if ($LayoutDir) {
    $argsList += @("--layout-dir", [System.IO.Path]::GetFullPath($LayoutDir))
}
if ($Inspect) {
    $argsList += @("--inspect", [System.IO.Path]::GetFullPath($Inspect))
}
if ($ContactSheet) {
    $argsList += @("--contact-sheet", [System.IO.Path]::GetFullPath($ContactSheet))
}
if ($Scale) {
    $argsList += @("--scale", $Scale)
}

$code = Invoke-WujiNodeScript -Runtime $runtime -ScriptPath $scriptPath -Arguments $argsList
if ($code -ne 0) {
    throw "prepare_template_starter_deck failed with exit code $code"
}
