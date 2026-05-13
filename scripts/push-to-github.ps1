# push-to-github.ps1 - One-click push Wuji Legion to GitHub
# Usage: 
#   1. Get a GitHub token from https://github.com/settings/tokens (repo scope)
#   2. Run: .\scripts\push-to-github.ps1
#   3. Paste token when prompted

$token = Read-Host "Enter GitHub Personal Access Token" -AsSecureString
$bstr = [System.Runtime.InteropServices.Marshal]::SecureStringToBSTR($token)
$tokenPlain = [System.Runtime.InteropServices.Marshal]::PtrToStringAuto($bstr)

# Create repo via API
$body = @{
    name = "wuji-legion-codex"
    description = "Wuji Legion for Codex CLI - True Parallel Multi-Agent Combat System"
    homepage = "https://ai-wuji.github.io/wuji-legion-codex/"
} | ConvertTo-Json

$result = curl.exe -s --ssl-no-revoke -X POST "https://api.github.com/user/repos" `
    -H "Authorization: token $tokenPlain" `
    -H "Content-Type: application/json" `
    -d $body
Write-Host "Repo created: $result" -ForegroundColor Cyan

# Configure git credentials for this push only
git remote remove origin 2>$null
git remote add origin "https://github.com/AI-wuji/wuji-legion-codex.git"

# Use git credential helper for the push
$cred = @"
protocol=https
host=github.com
username=AI-wuji
password=$tokenPlain
"@
$cred | git credential approve 2>$null

# Push
git push -u origin master
Write-Host "Done! https://github.com/AI-wuji/wuji-legion-codex" -ForegroundColor Green
