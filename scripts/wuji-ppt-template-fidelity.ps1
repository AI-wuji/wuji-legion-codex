param(
    [Parameter(Mandatory = $true)]
    [string]$Workspace,

    [Parameter(Mandatory = $true)]
    [string]$FinalPptx,

    [string]$Map = "",
    [string]$StarterPptx = "",
    [string]$StarterLayoutDir = "",
    [string]$FinalLayoutDir = "",
    [string]$EditDir = "",
    [string]$AgentLog = "",
    [string]$SkillDir = "",
    [string]$NodePath = "",
    [string]$NodeModules = "",
    [string]$PythonPath = ""
)

$ErrorActionPreference = "Stop"
. (Join-Path $PSScriptRoot 'wuji-ppt-runtime.ps1')

$runtime = Get-WujiPptRuntime -SkillDir $SkillDir -NodePath $NodePath -NodeModules $NodeModules -PythonPath $PythonPath
$scriptPath = Join-Path $PSScriptRoot "wuji-ppt-template-fidelity.mjs"
$resolvedWorkspace = [System.IO.Path]::GetFullPath($Workspace)
$null = New-Item -ItemType Directory -Force -Path $resolvedWorkspace
$resolvedFinalPptx = (Resolve-Path -LiteralPath $FinalPptx).Path

$argsList = @("--workspace", $resolvedWorkspace, "--final-pptx", $resolvedFinalPptx)
if ($Map) {
    $argsList += @("--map", [System.IO.Path]::GetFullPath($Map))
}
if ($StarterPptx) {
    $argsList += @("--starter-pptx", [System.IO.Path]::GetFullPath($StarterPptx))
}
if ($StarterLayoutDir) {
    $argsList += @("--starter-layout-dir", [System.IO.Path]::GetFullPath($StarterLayoutDir))
}
if ($FinalLayoutDir) {
    $argsList += @("--final-layout-dir", [System.IO.Path]::GetFullPath($FinalLayoutDir))
}
if ($EditDir) {
    $argsList += @("--edit-dir", [System.IO.Path]::GetFullPath($EditDir))
}
if ($AgentLog) {
    $argsList += @("--agent-log", [System.IO.Path]::GetFullPath($AgentLog))
}

$code = Invoke-WujiNodeScript -Runtime $runtime -ScriptPath $scriptPath -Arguments $argsList
if ($code -ne 0) {
    throw "wuji-ppt-template-fidelity failed with exit code $code"
}
