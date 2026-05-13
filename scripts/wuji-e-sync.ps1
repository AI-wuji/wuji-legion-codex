$R="E:\wuji-projects"
$L="$R\logs\sync_$(Get-Date -Format yyyyMMdd).log"
New-Item -Path "$R\logs" -ItemType Directory -Force | Out-Null
$T=Get-Date -Format "yyyy-MM-dd HH:mm:ss"
function Log {
    param([string]$Msg)
    $line = "$T | $Msg"
    Add-Content -Path $L -Value $line -Encoding UTF8
    Write-Host $line
}
Log "=== Sync ==="
$S="C:\Users\Administrator\.agents\skills\wuji-legion"
$D="$R\skills\wuji-legion"
if (Test-Path $S) {
    robocopy $S $D /MIR /NP /NJH /NJS /NDL | Out-Null
    Log "[SKILL] Synced"
}
$W="C:\wuji-projects\wuji-legion-codex"
$X="$R\workspace\wuji-legion-codex"
if (Test-Path $W) {
    robocopy $W $X /MIR /NP /NJH /NJS /NDL /XD node_modules .git __pycache__ | Out-Null
    Log "[WORKSPACE] Synced"
}
$C=(Get-Date).AddDays(-30)
Get-ChildItem "$R\logs\*.log" -ErrorAction SilentlyContinue | Where-Object { $_.LastWriteTime -lt $C } | Remove-Item -Force
Log "[CLEAN] Done"
Log "=== Complete ==="
