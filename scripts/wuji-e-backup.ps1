# wuji-e-backup.ps1 - E drive sync watcher for Wuji Legion

$ErrorActionPreference = "Stop"

$SkillRoot = Join-Path $env:USERPROFILE ".agents\skills\wuji-legion"
$SyncScript = Join-Path $SkillRoot "scripts\wuji-e-sync.ps1"
$LogDir = "E:\wuji-projects\logs"

if (-not (Test-Path -LiteralPath $SkillRoot)) {
    throw "Skill directory not found: $SkillRoot"
}

New-Item -ItemType Directory -Force -Path $LogDir | Out-Null

$watcher = New-Object System.IO.FileSystemWatcher
$watcher.Path = $SkillRoot
$watcher.IncludeSubdirectories = $true
$watcher.EnableRaisingEvents = $true

$action = {
    $path = $Event.SourceEventArgs.FullPath
    $changeType = $Event.SourceEventArgs.ChangeType
    $log = Join-Path "E:\wuji-projects\logs" ("realtime_{0}.log" -f (Get-Date -Format "yyyyMMdd"))
    "$(Get-Date -Format 'HH:mm:ss') | $changeType | $path" | Add-Content -Path $log -Encoding UTF8
}

Register-ObjectEvent $watcher "Created" -Action $action | Out-Null
Register-ObjectEvent $watcher "Changed" -Action $action | Out-Null
Register-ObjectEvent $watcher "Deleted" -Action $action | Out-Null

Write-Host "Wuji Legion sync watcher is running..." -ForegroundColor Green
Write-Host "Source: $SkillRoot" -ForegroundColor Cyan
Write-Host "Sync script: $SyncScript" -ForegroundColor Cyan
Write-Host "Press Ctrl+C to stop" -ForegroundColor Yellow

while ($true) {
    Start-Sleep -Seconds 300
    if (Test-Path -LiteralPath $SyncScript) {
        & $SyncScript
    }
}
