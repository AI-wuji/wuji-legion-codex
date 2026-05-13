# wuji-e-sync.ps1 - Wuji Legion E-Drive Backup Sync
# Usage: powershell wuji-e-sync.ps1
param([switch]$Quiet)

$BackupRoot = "E:\wuji-legion-backup"
$LogFile = Join-Path $BackupRoot "logs\sync_$(Get-Date -Format yyyyMMdd).log"
$Timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"

function Write-Log {
    param([string]$Msg)
    $line = "$Timestamp | $Msg"
    Add-Content -Path $LogFile -Value $line -Encoding UTF8
    if (-not $Quiet) { Write-Host $line }
}

Write-Log "=== Wuji Legion Backup Sync ==="

# 1. Sync skill directory
$skillSrc = "C:\Users\Administrator\.agents\skills\wuji-legion"
$skillDst = Join-Path $BackupRoot "skills\wuji-legion"
if (Test-Path $skillSrc) {
    New-Item -ItemType Directory -Force -Path $skillDst | Out-Null
    Copy-Item -Path "$skillSrc\*" -Destination $skillDst -Recurse -Force
    Write-Log "[SKILL] Synced"
}

# 2. Sync workspace
$wsSrc = "C:\Users\Administrator\Desktop\Hermes"
$wsDst = Join-Path $BackupRoot "workspace\Hermes"
if (Test-Path $wsSrc) {
    New-Item -ItemType Directory -Force -Path $wsDst | Out-Null
    Copy-Item -Path "$wsSrc\*" -Destination $wsDst -Recurse -Force -Exclude @("node_modules", ".git", "__pycache__")
    Write-Log "[WORKSPACE] Synced"
}

# 3. Clean logs older than 30 days
$cutoff = (Get-Date).AddDays(-30)
Get-ChildItem "$BackupRoot\logs\*.log" | Where-Object { $_.LastWriteTime -lt $cutoff } | Remove-Item -Force
Write-Log "[CLEAN] Done"

Write-Log "=== Backup Complete ==="
