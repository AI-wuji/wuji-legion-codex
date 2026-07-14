$ErrorActionPreference = 'Stop'
$status = @(& lark-cli auth status 2>&1) -join [Environment]::NewLine
if ($LASTEXITCODE -ne 0) { throw 'official lark-cli auth status failed' }
try { $auth = $status | ConvertFrom-Json -ErrorAction Stop } catch { throw 'official lark-cli returned invalid auth status' }
if (-not $auth.brand -or -not $auth.identities) { throw 'official lark-cli auth status is incomplete' }
Write-Output "feishu-auth-smoke-ok brand=$($auth.brand)"
