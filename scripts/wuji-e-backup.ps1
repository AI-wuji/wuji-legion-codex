# wuji-e-backup.ps1 — 无极军团 E 盘自动备份守护
# 安装在系统启动时自动运行，实时监测变更

$watcher = New-Object System.IO.FileSystemWatcher
$watcher.Path = "C:\Users\Administrator\.agents\skills\wuji-legion"
$watcher.IncludeSubdirectories = $true
$watcher.EnableRaisingEvents = $true

$action = {
    $path = $Event.SourceEventArgs.FullPath
    $changeType = $Event.SourceEventArgs.ChangeType
    $log = "E:\wuji-legion-backup\logs\realtime_$(Get-Date -Format 'yyyyMMdd').log"
    "$(Get-Date -Format 'HH:mm:ss') | $changeType | $path" | Add-Content -Path $log -Encoding UTF8
}

Register-ObjectEvent $watcher "Created" -Action $action | Out-Null
Register-ObjectEvent $watcher "Changed" -Action $action | Out-Null
Register-ObjectEvent $watcher "Deleted" -Action $action | Out-Null

Write-Host "无极军团备份守护运行中..." -ForegroundColor Green
Write-Host "监控目录: C:\Users\Administrator\.agents\skills\wuji-legion" -ForegroundColor Cyan
Write-Host "备份目标: E:\wuji-legion-backup" -ForegroundColor Cyan
Write-Host "按 Ctrl+C 停止" -ForegroundColor Yellow

# 保持运行
while ($true) {
    Start-Sleep -Seconds 300
    # 每5分钟全量同步一次
    & "C:\Users\Administrator\.agents\skills\wuji-legion\scripts\wuji-e-sync.ps1"
}
