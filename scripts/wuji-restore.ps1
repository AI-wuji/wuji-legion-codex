<#
.SYNOPSIS
  Restore Wuji Legion Codex.

.DESCRIPTION
  Restores the repository, skill directory, and global AGENTS.md after reinstalling Codex or Windows.
#>

$ErrorActionPreference = "Stop"

$REPO = "https://github.com/AI-wuji/wuji-legion-codex.git"
$REPO_NAME = "wuji-legion-codex"
$PROJECTS_ROOT = "E:\wuji-projects"
$WORK_DIR = Join-Path $PROJECTS_ROOT $REPO_NAME
$SKILL_DIR = Join-Path $env:USERPROFILE ".agents\skills\wuji-legion"
$AGENTS_DST = Join-Path $env:USERPROFILE ".codex\AGENTS.md"

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
Write-Host "  Wuji Legion Codex v9.5 Restore" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

Write-Host "[1/4] Get repository..." -ForegroundColor Yellow
New-Item -ItemType Directory -Force -Path $PROJECTS_ROOT | Out-Null
if (Test-Path -LiteralPath (Join-Path $WORK_DIR ".git")) {
    Set-Location $WORK_DIR
    git pull --ff-only
} else {
    git clone $REPO $WORK_DIR
}

if (-not (Test-Path -LiteralPath (Join-Path $WORK_DIR "SKILL.md"))) {
    throw "Repository is incomplete: SKILL.md not found"
}

Write-Host "[2/4] Install AGENTS.md..." -ForegroundColor Yellow
New-Item -ItemType Directory -Force -Path (Split-Path -Parent $AGENTS_DST) | Out-Null
Copy-Item -LiteralPath (Join-Path $WORK_DIR "GLOBAL_AGENTS.md") -Destination $AGENTS_DST -Force

Write-Host "[3/4] Install skill..." -ForegroundColor Yellow
Copy-CleanTree -Source $WORK_DIR -Destination $SKILL_DIR

Write-Host "[4/4] Verify restore..." -ForegroundColor Yellow
$skillFiles = (Get-ChildItem -LiteralPath $SKILL_DIR -Recurse -File).Count
if ($skillFiles -lt 20) {
    throw "Unexpected restored file count: $skillFiles"
}

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Restore complete." -ForegroundColor Green
Write-Host "  Skill: $SKILL_DIR" -ForegroundColor Cyan
Write-Host "  AGENTS: $AGENTS_DST" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
