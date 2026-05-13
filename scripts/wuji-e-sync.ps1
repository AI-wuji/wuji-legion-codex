# wuji-e-sync.ps1
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

Write-Log "=== Wuji Backup Sync ==="

# Use robocopy for reliable directory mirroring
$skillSrc = "C:\Users\Administrator\.agents\skills\wuji-legion"
$skillDst = "$BackupRoot\skills\wuji-legion"
if (Test-Path $skillSrc) {
    robocopy $skillSrc $skillDst /MIR /NP /NJH /NJS /NDL > $null 2>&1
    Write-Log "[SKILL] Synced"
}

$wsSrc = "C:\Users\Administrator\Desktop\Hermes"
$wsDst = "$BackupRoot\workspace\Hermes"
if (Test-Path $wsSrc) {
    robocopy $wsSrc $wsDst /MIR /NP /NJH /NJS /NDL /XD node_modules .git __pycache__ > $null 2>&1
    Write-Log "[WORKSPACE] Synced"
}

$cutoff = (Get-Date).AddDays(-30)
Get-ChildItem "$BackupRoot\logs\*.log" | Where-Object { $_.LastWriteTime -lt $cutoff } | Remove-Item -Force
Write-Log "[CLEAN] Done"
Write-Log "=== Complete ==="
