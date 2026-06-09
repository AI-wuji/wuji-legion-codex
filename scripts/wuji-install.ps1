# wuji-install.ps1 - install Wuji Legion Codex

param(
    [string]$Ref = "",
    [switch]$Bootstrap,
    [switch]$InstallAgents
)

$ErrorActionPreference = "Stop"

$REPO = "AI-wuji/wuji-legion-codex"
$SKILL_DIR = Join-Path $env:USERPROFILE ".agents\skills\wuji-legion"
$AGENTS_DST = Join-Path $env:USERPROFILE ".codex\AGENTS.md"
$temp = Join-Path $env:TEMP ("wuji-legion-codex-" + [guid]::NewGuid().ToString("N"))
$CONFIG_PATH = Join-Path $PSScriptRoot "..\config.json"
$INSTALLER_VERSION = "unknown"

function Read-JsonUtf8 {
    param([string]$Path)
    return [System.IO.File]::ReadAllText($Path, [System.Text.UTF8Encoding]::new($false)) | ConvertFrom-Json
}

if (Test-Path -LiteralPath $CONFIG_PATH) {
    try {
        $config = Read-JsonUtf8 -Path $CONFIG_PATH
        if ($config.iron_rules_version) {
            $INSTALLER_VERSION = [string]$config.iron_rules_version
        }
    } catch {
    }
}

if ([string]::IsNullOrWhiteSpace($Ref) -or $Ref -notmatch '^[0-9a-fA-F]{40}$') {
    throw "Supply-chain gate: install requires -Ref <40-char commit sha>. Moving branches and tags are forbidden."
}

function Copy-CleanTree {
    param(
        [Parameter(Mandatory=$true)][string]$Source,
        [Parameter(Mandatory=$true)][string]$Destination
    )

    New-Item -ItemType Directory -Force -Path $Destination | Out-Null
    robocopy $Source $Destination /MIR `
        /XD .git __pycache__ output outputs .wuji-errors .wuji-backups .wuji-tools .env .cache node_modules feedback .codex .agents .opensquilla `
        /XF *.pyc *.tmp *.log .env .env.* *.pem *.key *.pfx *.p12 *token* *cookie* *credential* `
        /NFL /NDL /NJH /NJS /NP | Out-Null
    if ($LASTEXITCODE -gt 7) {
        throw "Copy failed: robocopy exit code $LASTEXITCODE"
    }
}

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Wuji Legion Codex Installer v$INSTALLER_VERSION" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

try {
    Write-Host "[1/4] Fetch pinned commit..." -ForegroundColor Yellow
    git init $temp | Out-Null
    git -C $temp remote add origin "https://github.com/$REPO.git"
    git -C $temp fetch --depth 1 origin $Ref
    git -C $temp checkout --detach FETCH_HEAD | Out-Null
    $checkedCommit = (git -C $temp rev-parse HEAD).Trim()
    if ($checkedCommit.ToLowerInvariant() -ne $Ref.ToLowerInvariant()) {
        throw "Pinned commit mismatch: expected $Ref got $checkedCommit"
    }
    if (-not (Test-Path -LiteralPath (Join-Path $temp "SKILL.md"))) {
        throw "Download failed: SKILL.md not found"
    }

    Write-Host "[2/4] Install skill mirror..." -ForegroundColor Yellow
    Copy-CleanTree -Source $temp -Destination $SKILL_DIR

    Write-Host "[3/4] Optional AGENTS.md install..." -ForegroundColor Yellow
    if ($InstallAgents) {
        New-Item -ItemType Directory -Force -Path (Split-Path -Parent $AGENTS_DST) | Out-Null
        Copy-Item -LiteralPath (Join-Path $temp "GLOBAL_AGENTS.md") -Destination $AGENTS_DST -Force
        Write-Host "AGENTS installed: $AGENTS_DST" -ForegroundColor Cyan
    } else {
        Write-Host "Skipped AGENTS.md. Use -InstallAgents to write it." -ForegroundColor DarkGray
    }

    Write-Host "[4/4] Verify install..." -ForegroundColor Yellow
    $files = (Get-ChildItem -LiteralPath $SKILL_DIR -Recurse -File).Count
    if ($files -lt 20) {
        throw "Unexpected installed file count: $files"
    }

    $ensureScript = Join-Path $SKILL_DIR "scripts\ensure-wuji-cli.ps1"
    if ($Bootstrap) {
        if (-not (Test-Path -LiteralPath $ensureScript)) {
            throw "Bootstrap requested but ensure-wuji-cli.ps1 is missing"
        }
        Write-Host "Bootstrapping Go execution base..." -ForegroundColor Yellow
        & powershell -NoProfile -ExecutionPolicy Bypass -File $ensureScript -RepoRoot $SKILL_DIR -Quiet
    } else {
        Write-Host "Skipped Go bootstrap. Use -Bootstrap to run it." -ForegroundColor DarkGray
    }

    Write-Host "OK: Wuji Legion Codex installed at pinned commit $checkedCommit." -ForegroundColor Green
    Write-Host "Skill: $SKILL_DIR" -ForegroundColor Cyan
} finally {
    Remove-Item -LiteralPath $temp -Recurse -Force -ErrorAction SilentlyContinue
}
