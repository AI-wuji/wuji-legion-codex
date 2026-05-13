# wuji-install.ps1 — 无极军团一键安装引导

$REPO = "AI-wuji/wuji-legion-codex"
$SKILL_DIR = "$env:USERPROFILE\.agents\skills\wuji-legion"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "   Wuji Legion Installer v1.0" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

Write-Host "[1/4] Creating skill directory..." -ForegroundColor Yellow
New-Item -ItemType Directory -Force -Path $SKILL_DIR | Out-Null

Write-Host "[2/4] Downloading from github.com/$REPO ..." -ForegroundColor Yellow
$temp = "$env:TEMP\wuji-legion-codex"
Remove-Item -Recurse -Force $temp -ErrorAction SilentlyContinue
git clone "https://github.com/$REPO.git" $temp 2>$null

if (Test-Path "$temp\SKILL.md") {
    Copy-Item -Path "$temp\*" -Destination $SKILL_DIR -Recurse -Force
    Write-Host "  [OK] Downloaded successfully" -ForegroundColor Green
} else {
    Write-Host "  [FAIL] Cannot download from GitHub" -ForegroundColor Red
    Write-Host "  Please check network or manually clone:"
    Write-Host "  git clone https://github.com/$REPO.git"
    exit 1
}

Write-Host "[3/4] Checking E drive backup..." -ForegroundColor Yellow
$eBackup = "E:\wuji-legion-backup\skills\wuji-legion"
if (Test-Path $eBackup) {
    Write-Host "  [OK] E drive backup found at $eBackup" -ForegroundColor Green
    $restore = Read-Host "  Restore from E drive backup? (y/N)"
    if ($restore -eq "y" -or $restore -eq "Y") {
        Copy-Item -Path "$eBackup\*" -Destination $SKILL_DIR -Recurse -Force
        Write-Host "  [OK] Restored from E drive backup" -ForegroundColor Green
    }
} else {
    Write-Host "  [SKIP] No E drive backup found" -ForegroundColor Yellow
}

Write-Host "[4/4] Verifying installation..." -ForegroundColor Yellow
$files = (Get-ChildItem $SKILL_DIR -Recurse -File).Count
if ($files -gt 5) {
    Write-Host "  [OK] $files files installed" -ForegroundColor Green
} else {
    Write-Host "  [WARN] Only $files files - may be incomplete" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Installation complete!" -ForegroundColor Green
Write-Host "  Now tell Codex: 阿极，恢复" -ForegroundColor White
Write-Host "========================================" -ForegroundColor Cyan
