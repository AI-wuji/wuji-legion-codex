param(
    [string]$RepoRoot = (Split-Path -Parent $PSScriptRoot),
    [string]$Destination = "",
    [switch]$InstallAgents,
    [switch]$Bootstrap
)

$ErrorActionPreference = "Stop"

$repo = [System.IO.Path]::GetFullPath($RepoRoot)
if ([string]::IsNullOrWhiteSpace($Destination)) {
    $Destination = Join-Path $env:USERPROFILE ".agents\skills\wuji-legion"
}
$destinationPath = [System.IO.Path]::GetFullPath($Destination)
$agentsDestination = Join-Path $env:USERPROFILE ".codex\AGENTS.md"
$criticalMirrorFiles = @(
    "SKILL.md",
    "GLOBAL_AGENTS.md",
    "kernel-source.json",
    "fusion-matrix.json",
    "hotpath-manifest.json",
    "residual-entrypoints.json",
    "experts\INDEX.md"
)

foreach ($dir in @("security", "expedition", "prompt", "comfyui")) {
    $expertDir = Join-Path $repo ("experts\" + $dir)
    if (-not (Test-Path -LiteralPath $expertDir)) {
        throw "Missing expert directory: $expertDir"
    }
    Get-ChildItem -LiteralPath $expertDir -Filter *.md | ForEach-Object {
        $criticalMirrorFiles += ("experts\" + $dir + "\" + $_.Name)
    }
}

function Ensure-UnderRoot {
    param(
        [string]$Path,
        [string]$Root,
        [string]$Label
    )
    $fullPath = [System.IO.Path]::GetFullPath($Path).TrimEnd('\')
    $fullRoot = [System.IO.Path]::GetFullPath($Root).TrimEnd('\')
    if (-not $fullPath.StartsWith($fullRoot, [System.StringComparison]::OrdinalIgnoreCase)) {
        throw "$Label must stay under $fullRoot"
    }
}

function Copy-CleanTree {
    param(
        [Parameter(Mandatory = $true)][string]$Source,
        [Parameter(Mandatory = $true)][string]$Target
    )

    New-Item -ItemType Directory -Force -Path $Target | Out-Null
    robocopy $Source $Target /MIR `
        /XD .git __pycache__ output outputs .wuji-errors .wuji-backups .wuji-tools .env .cache node_modules feedback .codex .agents .opensquilla `
        /XF *.pyc *.tmp *.log .env .env.* *.pem *.key *.pfx *.p12 *token* *cookie* *credential* `
        /NFL /NDL /NJH /NJS /NP | Out-Null
    if ($LASTEXITCODE -gt 7) {
        throw "Copy failed: robocopy exit code $LASTEXITCODE"
    }
}

function Resolve-Python {
    $python = Get-Command python -ErrorAction SilentlyContinue
    if ($python) {
        return @{
            Command = $python.Source
            Prefix = @()
        }
    }
    $py = Get-Command py -ErrorAction SilentlyContinue
    if ($py) {
        return @{
            Command = $py.Source
            Prefix = @("-3")
        }
    }
    throw "Python 3 is required to rebuild expert mirrors before sync."
}

function Invoke-ExpertIndexRebuild {
    param(
        [Parameter(Mandatory = $true)][string]$RootPath,
        [Parameter(Mandatory = $true)][string]$Label
    )

    $pythonCommand = Resolve-Python
    $scriptPath = Join-Path $RootPath "scripts\gen_experts.py"
    if (-not (Test-Path -LiteralPath $scriptPath)) {
        throw "Missing gen_experts.py under $RootPath"
    }

    Write-Host "Rebuilding expert index for $Label..." -ForegroundColor Yellow
    & $pythonCommand.Command @($pythonCommand.Prefix + @($scriptPath))
    if ($LASTEXITCODE -ne 0) {
        throw "Expert index rebuild failed for ${Label}: exit code $LASTEXITCODE"
    }
}

function Assert-HashMatch {
    param(
        [Parameter(Mandatory = $true)][string]$SourcePath,
        [Parameter(Mandatory = $true)][string]$TargetPath,
        [Parameter(Mandatory = $true)][string]$Label
    )

    if (-not (Test-Path -LiteralPath $SourcePath)) {
        throw "Missing source file for ${Label}: $SourcePath"
    }
    if (-not (Test-Path -LiteralPath $TargetPath)) {
        throw "Missing target file for ${Label}: $TargetPath"
    }
    $sourceHash = (Get-FileHash -Algorithm SHA256 $SourcePath).Hash
    $targetHash = (Get-FileHash -Algorithm SHA256 $TargetPath).Hash
    if ($sourceHash -ne $targetHash) {
        throw "Post-sync hash mismatch for $Label"
    }
}

if (-not (Test-Path -LiteralPath (Join-Path $repo "SKILL.md"))) {
    throw "Missing source SKILL.md under $repo"
}
if (-not (Test-Path -LiteralPath (Join-Path $repo "GLOBAL_AGENTS.md"))) {
    throw "Missing source GLOBAL_AGENTS.md under $repo"
}

$skillRoot = Join-Path $env:USERPROFILE ".agents\skills"
Ensure-UnderRoot -Path $destinationPath -Root $skillRoot -Label "Destination"

Write-Host "Syncing Wuji Legion skill mirror..." -ForegroundColor Yellow
Write-Host "Source: $repo" -ForegroundColor DarkGray
Write-Host "Target: $destinationPath" -ForegroundColor DarkGray
Write-Host "Global rule sync includes Agnes mirror fallback policy: fail over to default GPT, never to silent local substitution." -ForegroundColor DarkGray

Invoke-ExpertIndexRebuild -RootPath $repo -Label "source repo"

Copy-CleanTree -Source $repo -Target $destinationPath

Invoke-ExpertIndexRebuild -RootPath $destinationPath -Label "destination mirror"

if ($InstallAgents) {
    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $agentsDestination) | Out-Null
    Copy-Item -LiteralPath (Join-Path $repo "GLOBAL_AGENTS.md") -Destination $agentsDestination -Force
    Write-Host "AGENTS updated: $agentsDestination" -ForegroundColor Cyan
}

if ($Bootstrap) {
    $ensureScript = Join-Path $destinationPath "scripts\ensure-wuji-cli.ps1"
    if (-not (Test-Path -LiteralPath $ensureScript)) {
        throw "Bootstrap requested but ensure-wuji-cli.ps1 is missing in $destinationPath"
    }
    & powershell -NoProfile -ExecutionPolicy Bypass -File $ensureScript -RepoRoot $destinationPath -Quiet
}

foreach ($relPath in $criticalMirrorFiles) {
    Assert-HashMatch `
        -SourcePath (Join-Path $repo $relPath) `
        -TargetPath (Join-Path $destinationPath $relPath) `
        -Label $relPath
}

Write-Host "OK: active skill mirror synced." -ForegroundColor Green
Write-Host "Critical mirror files verified: $($criticalMirrorFiles.Count)" -ForegroundColor Cyan
