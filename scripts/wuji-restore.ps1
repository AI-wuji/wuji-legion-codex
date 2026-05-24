<#
.SYNOPSIS
  无极军团 — 全量恢复脚本
  系统重装/Codex重装后，一键还原所有配置
#>

$REPO = "https://github.com/AI-wuji/wuji-legion-codex.git"
$REPO_NAME = "wuji-legion-codex"
$WORK_DIR = "E:\wuji-projects\$REPO_NAME"
$SKILL_DIR = "$env:USERPROFILE\.agents\skills\wuji-legion"
$AGENTS_DST = "$env:USERPROFILE\.codex\AGENTS.md"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  无极军团 v5.5 — 全量恢复" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# === STEP 1: 克隆/拉取仓库 ===
Write-Host "[1/5] 获取仓库..." -ForegroundColor Yellow
New-Item -ItemType Directory -Force -Path $WORK_DIR -ErrorAction SilentlyContinue | Out-Null
Set-Location $WORK_DIR
if (Test-Path ".git") {
    git pull 2>$null
    Write-Host "  [OK] 已更新仓库" -ForegroundColor Green
} else {
    Set-Location "E:\wuji-projects"
    git clone $REPO 2>$null
    if (Test-Path "$WORK_DIR\SKILL.md") {
        Write-Host "  [OK] 已克隆仓库" -ForegroundColor Green
    } else {
        Write-Host "  [FAIL] 无法克隆，请检查网络" -ForegroundColor Red
        exit 1
    }
}

# === STEP 2: 安装全局规则 ===
Write-Host "[2/5] 安装全局规则 AGENTS.md ..." -ForegroundColor Yellow
New-Item -ItemType Directory -Force -Path "$env:USERPROFILE\.codex" -ErrorAction SilentlyContinue | Out-Null
Copy-Item -Path "$WORK_DIR\GLOBAL_AGENTS.md" -Destination $AGENTS_DST -Force
if (Test-Path $AGENTS_DST) {
    Write-Host "  [OK] 全局规则已安装（铁律+白帽纠察+MoE+Cache）" -ForegroundColor Green
} else {
    Write-Host "  [FAIL] 安装失败" -ForegroundColor Red
}

# === STEP 3: 安装无极军团Skill ===
Write-Host "[3/5] 安装无极军团 Skill ..." -ForegroundColor Yellow
Remove-Item -Recurse -Force $SKILL_DIR -ErrorAction SilentlyContinue
$excludeList = @("node_modules", ".git", "__pycache__", ".gitignore")
Copy-Item -Path "$WORK_DIR\*" -Destination $SKILL_DIR -Recurse -Force -Exclude $excludeList
$skillFiles = (Get-ChildItem $SKILL_DIR -Recurse -File).Count
Write-Host "  [OK] $skillFiles 个文件已安装到 $SKILL_DIR" -ForegroundColor Green

# === STEP 4: 恢复E盘工作目录 ===
Write-Host "[4/5] 恢复 E 盘工作目录 ..." -ForegroundColor Yellow
$eDir = "E:\wuji-projects\$REPO_NAME"
if (-not (Test-Path $eDir)) {
    New-Item -ItemType Directory -Force -Path $eDir | Out-Null
    Copy-Item -Path "$WORK_DIR\*" -Destination $eDir -Recurse -Force -Exclude $excludeList
    Write-Host "  [OK] 工作目录已恢复" -ForegroundColor Green
} else {
    Write-Host "  [SKIP] E 盘目录已存在" -ForegroundColor Yellow
}

# === STEP 5: 列出需要手动重装的插件 ===
Write-Host "[5/5] 需要手动安装的插件（Codex Desktop → 设置 → 插件）:" -ForegroundColor Yellow
$plugins = @(
    @{Name="GitHub"; Pri="⭐"; Desc="PR/Issue/CI管理"},
    @{Name="Supabase"; Pri="⭐"; Desc="数据库+后端"},
    @{Name="Vercel"; Pri="⭐"; Desc="前端部署"},
    @{Name="Figma"; Pri="⭐"; Desc="UI设计→代码"},
    @{Name="HeyGen"; Pri="⭐"; Desc="AI数字人视频"},
    @{Name="Sentry"; Pri="高"; Desc="错误追踪"},
    @{Name="CodeRabbit"; Pri="高"; Desc="AI代码审查"},
    @{Name="Notion"; Pri="高"; Desc="文档协作"},
    @{Name="Remotion"; Pri="中"; Desc="程序化视频"},
    @{Name="Cloudinary"; Pri="中"; Desc="媒体管理"},
    @{Name="Linear"; Pri="中"; Desc="任务跟踪"},
    @{Name="Canva"; Pri="低"; Desc="轻量设计"},
    @{Name="CircleCI"; Pri="低"; Desc="CI/CD"},
    @{Name="Hugging Face"; Pri="低"; Desc="AI模型/数据集"},
    @{Name="Readwise"; Pri="低"; Desc="知识管理"},
    @{Name="Game Studio"; Pri="低"; Desc="浏览器游戏"},
    @{Name="Superpowers"; Pri="低"; Desc="TDD/调试工作流"}
)
$plugins | ForEach-Object { Write-Host "  $($_.Pri) $($_.Name) — $($_.Desc)" }

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  恢复完成！" -ForegroundColor Green
Write-Host "  全局规则 + 无极军团 Skill 已就绪" -ForegroundColor Green
Write-Host "  插件需手动安装（详见 units/plugins.md）" -ForegroundColor Yellow
Write-Host "========================================" -ForegroundColor Cyan
