$ErrorActionPreference = "Stop"

$ProjectsRoot = "E:\wuji-projects"
$LogDir = Join-Path $ProjectsRoot "logs"
$LogFile = Join-Path $LogDir ("sync_{0}.log" -f (Get-Date -Format yyyyMMdd))
$SkillSource = Join-Path $env:USERPROFILE ".agents\skills\wuji-legion"
$SkillDest = Join-Path $ProjectsRoot "skills\wuji-legion"
$RepoSource = Join-Path $ProjectsRoot "wuji-legion-codex"
$WorkspaceDest = Join-Path $ProjectsRoot "workspace\wuji-legion-codex"

New-Item -Path $LogDir -ItemType Directory -Force | Out-Null

function Log {
    param([string]$Msg)
    $line = "$(Get-Date -Format 'yyyy-MM-dd HH:mm:ss') | $Msg"
    Add-Content -Path $LogFile -Value $line -Encoding UTF8
    Write-Host $line
}

function Mirror-CleanTree {
    param(
        [Parameter(Mandatory=$true)][string]$Source,
        [Parameter(Mandatory=$true)][string]$Destination
    )

    if (-not (Test-Path -LiteralPath $Source)) {
        Log "SKIP: missing source: $Source"
        return
    }

    New-Item -ItemType Directory -Force -Path $Destination | Out-Null
    robocopy $Source $Destination /MIR /XD .git __pycache__ output outputs .wuji-errors .wuji-backups /XF *.pyc *.tmp *.log /NP /NJH /NJS /NDL | Out-Null
    if ($LASTEXITCODE -gt 7) {
        throw "robocopy failed with exit code $LASTEXITCODE"
    }
    Log "OK: synced: $Source -> $Destination"
}

Log "=== Sync ==="
Mirror-CleanTree -Source $SkillSource -Destination $SkillDest
Mirror-CleanTree -Source $RepoSource -Destination $WorkspaceDest

$cutoff = (Get-Date).AddDays(-30)
Get-ChildItem -LiteralPath $LogDir -Filter "*.log" -File -ErrorAction SilentlyContinue |
    Where-Object { $_.LastWriteTime -lt $cutoff } |
    Remove-Item -Force

Log "CLEAN: old logs checked"
Log "=== Complete ==="
