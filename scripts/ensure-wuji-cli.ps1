param(
    [string]$RepoRoot = (Split-Path -Parent $PSScriptRoot),
    [switch]$Quiet
)

$ErrorActionPreference = "Stop"

$repo = [System.IO.Path]::GetFullPath($RepoRoot)
$binDir = Join-Path $repo ".wuji-tools"
$exe = Join-Path $binDir "wuji-exec-base.exe"
$src = Join-Path $repo "tools\wuji_cli.go"
$buildScript = Join-Path $repo "scripts\build-wuji-cli.ps1"

if (-not (Test-Path -LiteralPath $src)) {
    throw "Missing Go source: $src"
}

if (-not (Test-Path -LiteralPath $buildScript)) {
    throw "Missing build script: $buildScript"
}

$needsBuild = -not (Test-Path -LiteralPath $exe)
if (-not $needsBuild) {
    $srcTime = (Get-Item -LiteralPath $src).LastWriteTimeUtc
    $exeTime = (Get-Item -LiteralPath $exe).LastWriteTimeUtc
    if ($srcTime -gt $exeTime) {
        $needsBuild = $true
    }
}

if ($needsBuild) {
    if (-not $Quiet) {
        Write-Host "Rebuilding wuji-cli..." -ForegroundColor Yellow
    }
    & powershell -NoProfile -ExecutionPolicy Bypass -File $buildScript
    if ($LASTEXITCODE -ne 0) {
        throw "wuji-cli rebuild failed with exit code $LASTEXITCODE"
    }
}

if (-not (Test-Path -LiteralPath $exe)) {
    throw "wuji-cli executable missing after build: $exe"
}

if (-not $Quiet) {
    Write-Host "wuji-cli ready: $exe" -ForegroundColor Green
}
