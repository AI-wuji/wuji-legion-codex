# wuji-install.ps1 - install Wuji Legion Codex

$ErrorActionPreference = "Stop"

$REPO = "AI-wuji/wuji-legion-codex"
$SKILL_DIR = Join-Path $env:USERPROFILE ".agents\skills\wuji-legion"
$AGENTS_DST = Join-Path $env:USERPROFILE ".codex\AGENTS.md"
$temp = Join-Path $env:TEMP ("wuji-legion-codex-" + [guid]::NewGuid().ToString("N"))
$CONFIG_PATH = Join-Path $PSScriptRoot "..\config.json"
$INSTALLER_VERSION = "unknown"

if (Test-Path -LiteralPath $CONFIG_PATH) {
    try {
        $config = Get-Content -Raw -LiteralPath $CONFIG_PATH | ConvertFrom-Json
        if ($config.iron_rules_version) {
            $INSTALLER_VERSION = [string]$config.iron_rules_version
        }
    } catch {
    }
}

function Copy-CleanTree {
    param(
        [Parameter(Mandatory=$true)][string]$Source,
        [Parameter(Mandatory=$true)][string]$Destination
    )

    New-Item -ItemType Directory -Force -Path $Destination | Out-Null
    robocopy $Source $Destination /MIR /XD .git __pycache__ output outputs .wuji-errors .wuji-backups /XF *.pyc *.tmp *.log /NFL /NDL /NJH /NJS /NP | Out-Null
    if ($LASTEXITCODE -gt 7) {
        throw "Copy failed: robocopy exit code $LASTEXITCODE"
    }
}

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Wuji Legion Codex Installer v$INSTALLER_VERSION" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

try {
    Write-Host "[1/4] Clone repository..." -ForegroundColor Yellow
    git clone "https://github.com/$REPO.git" $temp 2>$null
    if (-not (Test-Path -LiteralPath (Join-Path $temp "SKILL.md"))) {
        throw "Download failed: SKILL.md not found"
    }

    Write-Host "[2/4] Install skill..." -ForegroundColor Yellow
    Copy-CleanTree -Source $temp -Destination $SKILL_DIR

    Write-Host "[3/4] Install AGENTS.md..." -ForegroundColor Yellow
    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $AGENTS_DST) | Out-Null
    Copy-Item -LiteralPath (Join-Path $temp "GLOBAL_AGENTS.md") -Destination $AGENTS_DST -Force

    Write-Host "[4/4] Verify install..." -ForegroundColor Yellow
    $files = (Get-ChildItem -LiteralPath $SKILL_DIR -Recurse -File).Count
    if ($files -lt 20) {
        throw "Unexpected installed file count: $files"
    }

    Write-Host "OK: Wuji Legion Codex installed." -ForegroundColor Green
    Write-Host "Skill: $SKILL_DIR" -ForegroundColor Cyan
    Write-Host "AGENTS: $AGENTS_DST" -ForegroundColor Cyan
} finally {
    Remove-Item -LiteralPath $temp -Recurse -Force -ErrorAction SilentlyContinue
}
