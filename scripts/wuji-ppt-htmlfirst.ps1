param(
    [Parameter(Mandatory = $true)]
    [string]$Workspace,

    [Parameter(Mandatory = $true)]
    [string]$Html,

    [Parameter(Mandatory = $true)]
    [string]$Out,

    [string]$Title = "",
    [string]$Report = "",
    [string]$NodePath = "",
    [string]$NodeModules = ""
)

$ErrorActionPreference = "Stop"

if ([string]::IsNullOrWhiteSpace($NodePath)) {
    $NodePath = Join-Path $env:USERPROFILE ".cache\codex-runtimes\codex-primary-runtime\dependencies\node\bin\node.exe"
}
if ([string]::IsNullOrWhiteSpace($NodeModules)) {
    $NodeModules = Join-Path $env:USERPROFILE ".cache\codex-runtimes\codex-primary-runtime\dependencies\node\node_modules"
}
if (-not (Test-Path -LiteralPath $NodePath)) {
    throw "Bundled Node not found: $NodePath"
}
if (-not (Test-Path -LiteralPath $NodeModules)) {
    throw "Bundled node_modules not found: $NodeModules"
}

$resolvedWorkspace = [System.IO.Path]::GetFullPath($Workspace)
$null = New-Item -ItemType Directory -Force -Path $resolvedWorkspace
$resolvedHtml = (Resolve-Path -LiteralPath $Html).Path
$scriptPath = Join-Path $PSScriptRoot "wuji-ppt-htmlfirst.mjs"

$previousHome = $env:HOME
$previousNodePath = $env:NODE_PATH
try {
    $env:HOME = $env:USERPROFILE
    $extraNodePath = Join-Path $NodeModules ".pnpm\node_modules"
    $nodePathList = @($NodeModules)
    if (Test-Path -LiteralPath $extraNodePath) {
        $nodePathList += $extraNodePath
    }
    $env:NODE_PATH = ($nodePathList -join ';')
    $argsList = @($scriptPath, "--workspace", $resolvedWorkspace, "--html", $resolvedHtml, "--out", [System.IO.Path]::GetFullPath($Out))
    if ($Title) {
        $argsList += @("--title", $Title)
    }
    if ($Report) {
        $argsList += @("--report", [System.IO.Path]::GetFullPath($Report))
    }
    & $NodePath @argsList
    if ($LASTEXITCODE -ne 0) {
        throw "wuji-ppt-htmlfirst failed with exit code $LASTEXITCODE"
    }
}
finally {
    $env:HOME = $previousHome
    $env:NODE_PATH = $previousNodePath
}
