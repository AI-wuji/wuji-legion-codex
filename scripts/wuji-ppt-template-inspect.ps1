param(
    [Parameter(Mandatory = $true)]
    [string]$Workspace,

    [Parameter(Mandatory = $true)]
    [string]$Pptx,

    [string]$OutDir = "",
    [string]$Slides = "",
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
$scriptPath = Join-Path $PSScriptRoot "wuji-ppt-template-inspect.mjs"
$resolvedWorkspace = [System.IO.Path]::GetFullPath($Workspace)
$null = New-Item -ItemType Directory -Force -Path $resolvedWorkspace
Initialize-WujiArtifactWorkspace -Runtime $runtime -Workspace $resolvedWorkspace
$resolvedPptx = (Resolve-Path -LiteralPath $Pptx).Path

$argsList = @("--workspace", $resolvedWorkspace, "--pptx", $resolvedPptx)
if ($OutDir) {
    $argsList += @("--out-dir", $OutDir)
}
if ($Slides) {
    $argsList += @("--slides", $Slides)
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
    throw "wuji-ppt-template-inspect failed with exit code $code"
}
