# ☯️ 无极军团 — 任务完成提示音
# PowerShell 脚本，播放两短一长的提示音

function Play-TaskCompleteBeep {
    try {
        [Console]::Beep(800, 150)   # 短音 1
        Start-Sleep -Milliseconds 100
        [Console]::Beep(800, 150)   # 短音 2
        Start-Sleep -Milliseconds 100
        [Console]::Beep(1000, 400)  # 长音
        Write-Host "🔔 任务完成！" -ForegroundColor Green
    }
    catch {
        # 静默失败 — 某些环境可能不支持 Beep
    }
}

function Play-ErrorBeep {
    try {
        [Console]::Beep(200, 500)
        [Console]::Beep(150, 500)
        Write-Host "⚠️ 出现问题！" -ForegroundColor Red
    }
    catch {}
}

# 根据参数选择提示音
param([string]$Type = "complete")
if ($Type -eq "complete") { Play-TaskCompleteBeep }
elseif ($Type -eq "error") { Play-ErrorBeep }
