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
$resolvedPptxInput = [System.IO.Path]::GetFullPath($Pptx)
$resolvedMapInput = [System.IO.Path]::GetFullPath($Map)
$workspacePrefix = $resolvedWorkspace.TrimEnd('\','/') + [System.IO.Path]::DirectorySeparatorChar
$stagedDir = Join-Path (Split-Path -Parent $resolvedWorkspace) ("{0}-staged-inputs" -f (Split-Path -Leaf $resolvedWorkspace))
if ($resolvedPptxInput.StartsWith($workspacePrefix, [System.StringComparison]::OrdinalIgnoreCase) -or $resolvedMapInput.StartsWith($workspacePrefix, [System.StringComparison]::OrdinalIgnoreCase)) {
    New-Item -ItemType Directory -Force -Path $stagedDir | Out-Null
    if ($resolvedPptxInput.StartsWith($workspacePrefix, [System.StringComparison]::OrdinalIgnoreCase) -and (Test-Path -LiteralPath $resolvedPptxInput)) {
        $stagedPptx = Join-Path $stagedDir ([System.IO.Path]::GetFileName($resolvedPptxInput))
        Copy-Item -LiteralPath $resolvedPptxInput -Destination $stagedPptx -Force
        $resolvedPptxInput = $stagedPptx
    }
    if ($resolvedMapInput.StartsWith($workspacePrefix, [System.StringComparison]::OrdinalIgnoreCase) -and (Test-Path -LiteralPath $resolvedMapInput)) {
        $stagedMap = Join-Path $stagedDir ([System.IO.Path]::GetFileName($resolvedMapInput))
        Copy-Item -LiteralPath $resolvedMapInput -Destination $stagedMap -Force
        $resolvedMapInput = $stagedMap
    }
}
$null = New-Item -ItemType Directory -Force -Path $resolvedWorkspace
Initialize-WujiArtifactWorkspace -Runtime $runtime -Workspace $resolvedWorkspace
$resolvedPptx = (Resolve-Path -LiteralPath $resolvedPptxInput).Path
$resolvedMap = (Resolve-Path -LiteralPath $resolvedMapInput).Path

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

$resolvedOut = [System.IO.Path]::GetFullPath($Out)
if (-not (Test-Path -LiteralPath $resolvedOut)) {
    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $resolvedOut) | Out-Null
    Copy-Item -LiteralPath $resolvedPptx -Destination $resolvedOut -Force
}

$resolvedOut = [System.IO.Path]::GetFullPath($Out)
if (-not (Test-Path -LiteralPath $resolvedOut)) {
    $parent = Split-Path -Parent $resolvedOut
    if ($parent) {
        New-Item -ItemType Directory -Force -Path $parent | Out-Null
    }
    Copy-Item -LiteralPath $resolvedPptx -Destination $resolvedOut -Force
}
