# push-to-github.ps1
# 网络正常时运行此脚本推送更新

$repo = "C:\wuji-projects\wuji-legion-codex"

if (-not (Test-Path "$repo\.git")) {
    Write-Host "首次推送: 克隆并覆盖..." -ForegroundColor Yellow
    git clone https://github.com/AI-wuji/wuji-legion-codex.git "$repo\_temp" 2>&1
    if (Test-Path "$repo\_temp\.git") {
        Remove-Item "$repo\_temp\*" -Recurse -Force -ErrorAction SilentlyContinue
        Copy-Item "$repo\SKILL.md","$repo\README.md","$repo\CHANGELOG.md" "$repo\_temp\"
        Copy-Item "$repo\units" "$repo\_temp\" -Recurse
        Copy-Item "$repo\scripts" "$repo\_temp\" -Recurse
        Remove-Item "$repo\_temp\scripts\push-to-github.ps1" -ErrorAction SilentlyContinue
        Set-Location "$repo\_temp"
        git add -A
        git commit -m "V3.0 架构重构: 五大核心部门+省token优化+14新建模"
        git push
        Remove-Item "$repo" -Recurse -Force
        Move-Item "$repo\_temp" "$repo"
        Write-Host "推送完成!" -ForegroundColor Green
    }
} else {
    Set-Location $repo
    Copy-Item "C:\Users\Administrator\.agents\skills\wuji-legion\SKILL.md" ".\" -Force
    Copy-Item "C:\Users\Administrator\.agents\skills\wuji-legion\units\*" ".\units\" -Force
    Copy-Item "C:\Users\Administrator\.agents\skills\wuji-legion\scripts\*" ".\scripts\" -Force
    git add -A
    git commit -m "更新 $(Get-Date -Format yyyy-MM-dd HH:mm)"
    git push
    Write-Host "推送完成!" -ForegroundColor Green
}
