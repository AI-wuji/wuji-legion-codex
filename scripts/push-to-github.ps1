param(
    [string]$Message = "Update Wuji Legion Codex $(Get-Date -Format 'yyyy-MM-dd HH:mm')",
    [switch]$NoPush
)

$ErrorActionPreference = "Stop"

$repo = Resolve-Path -LiteralPath (Join-Path $PSScriptRoot "..")
Set-Location $repo

if (-not (Test-Path -LiteralPath ".git")) {
    throw "Current directory is not a Git repository: $repo"
}

$status = git status --short
if (-not $status) {
    Write-Host "SKIP: no changes to commit." -ForegroundColor Yellow
    exit 0
}

git add -A
git diff --cached --quiet
if ($LASTEXITCODE -eq 0) {
    Write-Host "SKIP: no staged changes." -ForegroundColor Yellow
    exit 0
}

git commit -m $Message

if ($NoPush) {
    Write-Host "OK: committed without push: $Message" -ForegroundColor Green
    exit 0
}

git push
Write-Host "OK: committed and pushed: $Message" -ForegroundColor Green
