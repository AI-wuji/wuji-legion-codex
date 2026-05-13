# wuji-restore.ps1 - Wuji Legion Disaster Recovery
# 从 E 盘备份恢复整个系统
# Usage: powershell wuji-restore.ps1

$BACKUP_ROOT = "E:\wuji-legion-backup"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Wuji Legion Disaster Recovery" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# 1. Verify backup exists
if (-not (Test-Path $BACKUP_ROOT)) {
    Write-Host "[FAIL] Backup not found at $BACKUP_ROOT" -ForegroundColor Red
    exit 1
}

# 2. Show backup info
$skillCount = (Get-ChildItem "$BACKUP_ROOT\skills" -Recurse -File).Count
$wsCount = (Get-ChildItem "$BACKUP_ROOT\workspace" -Recurse -File).Count
$lastSync = (Get-ChildItem "$BACKUP_ROOT\logs\*.log" | Sort-Object LastWriteTime -Descending | Select-Object -First 1).LastWriteTime

Write-Host "Backup found at: $BACKUP_ROOT" -ForegroundColor Green
Write-Host "  Skills: $skillCount files" -ForegroundColor Gray
Write-Host "  Workspace: $wsCount files" -ForegroundColor Gray
Write-Host "  Last sync: $lastSync" -ForegroundColor Gray
Write-Host ""

# 3. Confirm restore
$confirm = Read-Host "Restore all files? This will OVERWRITE current files. (y/N)"
if ($confirm -ne "y" -and $confirm -ne "Y") {
    Write-Host "Restore cancelled." -ForegroundColor Yellow
    exit
}

# 4. Restore skills
Write-Host "Restoring skills..." -ForegroundColor Yellow
$skillSrc = "$BACKUP_ROOT\skills\wuji-legion"
$skillDst = "C:\Users\Administrator\.agents\skills\wuji-legion"
if (Test-Path $skillSrc) {
    New-Item -ItemType Directory -Force -Path $skillDst | Out-Null
    Copy-Item -Path "$skillSrc\*" -Destination $skillDst -Recurse -Force
    Write-Host "[OK] Skills restored" -ForegroundColor Green
} else {
    Write-Host "[SKIP] No skill backup found" -ForegroundColor Yellow
}

# 5. Restore workspace
Write-Host "Restoring workspace..." -ForegroundColor Yellow
$wsSrc = "$BACKUP_ROOT\workspace\Hermes"
$wsDst = "C:\Users\Administrator\Desktop\Hermes"
if (Test-Path $wsSrc) {
    New-Item -ItemType Directory -Force -Path $wsDst | Out-Null
    Copy-Item -Path "$wsSrc\*" -Destination $wsDst -Recurse -Force -Exclude @("node_modules", ".git", "__pycache__")
    Write-Host "[OK] Workspace restored" -ForegroundColor Green
} else {
    Write-Host "[SKIP] No workspace backup found" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Restore complete!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
