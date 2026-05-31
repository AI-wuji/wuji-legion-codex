param(
    [Parameter(Mandatory=$true)]
    [string]$Workspace,

    [string]$Generator,

    [ValidateSet('preflight','batch')]
    [string]$Mode = 'preflight'
)

$ErrorActionPreference = 'Stop'

$root = Split-Path -Parent $PSScriptRoot
$src = Join-Path $root 'tools\wuji_cli.rs'
$binDir = Join-Path $root '.wuji-tools'
$bin = Join-Path $binDir 'wuji-cli.exe'

if (-not (Test-Path -LiteralPath $src)) {
    throw "Missing Rust source: $src"
}

if (-not (Test-Path -LiteralPath $binDir)) {
    New-Item -ItemType Directory -Path $binDir | Out-Null
}

$needsBuild = -not (Test-Path -LiteralPath $bin)
if (-not $needsBuild) {
    $needsBuild = (Get-Item -LiteralPath $src).LastWriteTimeUtc -gt (Get-Item -LiteralPath $bin).LastWriteTimeUtc
}

if ($needsBuild) {
    & rustc $src -O -o $bin
    if ($LASTEXITCODE -ne 0) {
        throw "rustc failed with exit code $LASTEXITCODE"
    }
}

$command = if ($Mode -eq 'batch') { 'pptx-batch-gate' } else { 'pptx-preflight' }
$argsList = @($command, '--workspace', (Resolve-Path -LiteralPath $Workspace).Path)
if ($Generator) {
    $argsList += @('--generator', (Resolve-Path -LiteralPath $Generator).Path)
}

& $bin @argsList
exit $LASTEXITCODE
