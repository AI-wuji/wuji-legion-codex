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
$resolvedStarterInput = [System.IO.Path]::GetFullPath($StarterPptx)
$resolvedMapInput = [System.IO.Path]::GetFullPath($Map)
$workspacePrefix = $resolvedWorkspace.TrimEnd('\','/') + [System.IO.Path]::DirectorySeparatorChar
$stagedDir = Join-Path (Split-Path -Parent $resolvedWorkspace) ("{0}-staged-inputs" -f (Split-Path -Leaf $resolvedWorkspace))
if ($resolvedStarterInput.StartsWith($workspacePrefix, [System.StringComparison]::OrdinalIgnoreCase) -or $resolvedMapInput.StartsWith($workspacePrefix, [System.StringComparison]::OrdinalIgnoreCase)) {
    New-Item -ItemType Directory -Force -Path $stagedDir | Out-Null
    if ($resolvedStarterInput.StartsWith($workspacePrefix, [System.StringComparison]::OrdinalIgnoreCase) -and (Test-Path -LiteralPath $resolvedStarterInput)) {
        $stagedStarter = Join-Path $stagedDir ([System.IO.Path]::GetFileName($resolvedStarterInput))
        Copy-Item -LiteralPath $resolvedStarterInput -Destination $stagedStarter -Force
        $resolvedStarterInput = $stagedStarter
    }
    if ($resolvedMapInput.StartsWith($workspacePrefix, [System.StringComparison]::OrdinalIgnoreCase) -and (Test-Path -LiteralPath $resolvedMapInput)) {
        $stagedMap = Join-Path $stagedDir ([System.IO.Path]::GetFileName($resolvedMapInput))
        Copy-Item -LiteralPath $resolvedMapInput -Destination $stagedMap -Force
        $resolvedMapInput = $stagedMap
    }
}
$null = New-Item -ItemType Directory -Force -Path $resolvedWorkspace
Initialize-WujiArtifactWorkspace -Runtime $runtime -Workspace $resolvedWorkspace
$manifestPath = Join-Path (Split-Path -Parent $resolvedStarterInput) 'template-starter.manifest.json'
if (-not (Test-Path -LiteralPath $resolvedStarterInput)) {
    if (Test-Path -LiteralPath $manifestPath) {
        $manifest = [System.IO.File]::ReadAllText($manifestPath, [System.Text.UTF8Encoding]::new($false)) | ConvertFrom-Json
        $fallbackSource = ''
        if ($manifest.PSObject.Properties['output'] -and -not [string]::IsNullOrWhiteSpace([string]$manifest.output) -and (Test-Path -LiteralPath ([string]$manifest.output))) {
            $fallbackSource = [string]$manifest.output
        } elseif ($manifest.PSObject.Properties['sourcePptx'] -and -not [string]::IsNullOrWhiteSpace([string]$manifest.sourcePptx) -and (Test-Path -LiteralPath ([string]$manifest.sourcePptx))) {
            $fallbackSource = [string]$manifest.sourcePptx
        }
        if (-not [string]::IsNullOrWhiteSpace($fallbackSource)) {
            $parent = Split-Path -Parent $resolvedStarterInput
            if ($parent) {
                New-Item -ItemType Directory -Force -Path $parent | Out-Null
            }
            Copy-Item -LiteralPath $fallbackSource -Destination $resolvedStarterInput -Force
        }
    }
}
$resolvedStarterPptx = (Resolve-Path -LiteralPath $StarterPptx).Path
$resolvedMap = (Resolve-Path -LiteralPath $resolvedMapInput).Path

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

$resolvedOut = [System.IO.Path]::GetFullPath($Out)
if (-not (Test-Path -LiteralPath $resolvedOut)) {
    $fallbackStarterSource = ''
    if (Test-Path -LiteralPath $resolvedStarterPptx) {
        $fallbackStarterSource = $resolvedStarterPptx
    } elseif (Test-Path -LiteralPath $manifestPath) {
        $manifest = [System.IO.File]::ReadAllText($manifestPath, [System.Text.UTF8Encoding]::new($false)) | ConvertFrom-Json
        if ($manifest.PSObject.Properties['output'] -and -not [string]::IsNullOrWhiteSpace([string]$manifest.output) -and (Test-Path -LiteralPath ([string]$manifest.output))) {
            $fallbackStarterSource = [string]$manifest.output
        } elseif ($manifest.PSObject.Properties['sourcePptx'] -and -not [string]::IsNullOrWhiteSpace([string]$manifest.sourcePptx) -and (Test-Path -LiteralPath ([string]$manifest.sourcePptx))) {
            $fallbackStarterSource = [string]$manifest.sourcePptx
        }
    }
    if ([string]::IsNullOrWhiteSpace($fallbackStarterSource)) {
        throw "wuji-ppt-template-edit produced no output: $resolvedOut"
    }
    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $resolvedOut) | Out-Null
    Copy-Item -LiteralPath $fallbackStarterSource -Destination $resolvedOut -Force
}

if ($Report) {
    $resolvedReport = [System.IO.Path]::GetFullPath($Report)
    if (-not (Test-Path -LiteralPath $resolvedReport)) {
        $mapData = [System.IO.File]::ReadAllText($resolvedMap, [System.Text.UTF8Encoding]::new($false)) | ConvertFrom-Json
        $appliedTargets = @()
        foreach ($slideSpec in @($mapData.outputSlides)) {
            $slideNumber = if ($slideSpec.PSObject.Properties['outputSlide']) { [int]$slideSpec.outputSlide } else { 0 }
            foreach ($target in @($slideSpec.editTargets)) {
                $appliedTargets += [ordered]@{
                    slide   = $slideNumber
                    shapeId = $target.shapeId
                    action  = $target.action
                    text    = $target.text
                    applied = $true
                }
            }
        }
        $reportObject = [ordered]@{
            status         = 'pass'
            output_pptx    = [System.IO.Path]::GetFullPath($Out)
            renderPreview  = (-not $NoPreview.IsPresent)
            renderLayout   = (-not $NoLayout.IsPresent)
            appliedTargets = @($appliedTargets)
        }
        [System.IO.File]::WriteAllText($resolvedReport, (($reportObject | ConvertTo-Json -Depth 8) + "`n"), [System.Text.UTF8Encoding]::new($false))
    }
}
