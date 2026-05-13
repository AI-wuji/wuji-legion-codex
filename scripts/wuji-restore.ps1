
$BACKUP_ROOT = "E:\wuji-legion-backup"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Wuji Legion Disaster Recovery" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

if (-not (Test-Path $BACKUP_ROOT)) {
    Write-Host "[FAIL] Backup not found at $BACKUP_ROOT" -ForegroundColor Red
    exit 1
}

$skillCount = (Get-ChildItem "$BACKUP_ROOT\skills" -Recurse -File).Count
$wsCount = (Get-ChildItem "$BACKUP_ROOT\workspace" -Recurse -File).Count
$lastSync = (Get-ChildItem "$BACKUP_ROOT\logs\*.log" | Sort-Object LastWriteTime -Descending | Select-Object -First 1).LastWriteTime

Write-Host "Backup found at: $BACKUP_ROOT" -ForegroundColor Green
Write-Host "  Skills: $skillCount files" -ForegroundColor Gray
Write-Host "  Workspace: $wsCount files" -ForegroundColor Gray
Write-Host "  Last sync: $lastSync" -ForegroundColor Gray
Write-Host ""

$confirm = Read-Host "Restore all files? This will OVERWRITE current files. (y/N)"
if ($confirm -ne "y" -and $confirm -ne "Y") {
    Write-Host "Restore cancelled." -ForegroundColor Yellow
    exit
}

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
