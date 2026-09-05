$ErrorActionPreference = 'Stop'
$command = Get-Command lark-cli.ps1 -CommandType ExternalScript -ErrorAction SilentlyContinue | Select-Object -First 1
if (-not $command) { $command = Get-Command lark-cli.exe -CommandType Application -ErrorAction Stop | Select-Object -First 1 }

$version = @(& $command.Source --version 2>&1) -join [Environment]::NewLine
if ($LASTEXITCODE -ne 0 -or $version -notmatch 'lark-cli version\s+([0-9]+\.[0-9]+\.[0-9]+)') {
  throw 'official lark-cli version check failed'
}
$cliVersion = $Matches[1]

$priorErrorAction = $ErrorActionPreference
$ErrorActionPreference = 'Continue'
$status = @(& $command.Source auth status 2>&1) -join [Environment]::NewLine
$statusExitCode = $LASTEXITCODE
$ErrorActionPreference = $priorErrorAction
try { $auth = $status | ConvertFrom-Json -ErrorAction Stop } catch { throw 'official lark-cli returned invalid auth status JSON' }

$authState = 'configured'
if ($statusExitCode -eq 3 -and $auth.ok -eq $false -and $auth.error.type -eq 'config' -and $auth.error.subtype -eq 'not_configured') {
  $authState = 'not-configured'
} elseif ($statusExitCode -ne 0 -or $auth.ok -eq $false) {
  throw "official lark-cli auth status failed: exit=$statusExitCode"
}

Write-Output "feishu-cli-smoke-ok version=$cliVersion auth=$authState"
