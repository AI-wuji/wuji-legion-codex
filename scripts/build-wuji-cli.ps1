param(
    [string]$Output
)

$ErrorActionPreference = 'Stop'

$root = Split-Path -Parent $PSScriptRoot
$src = Join-Path $root 'tools\wuji_cli.go'
$binDir = Join-Path $root '.wuji-tools'
$cacheRoot = Join-Path $env:LOCALAPPDATA 'WujiLegion\go-cache'
if (-not $Output) {
    $Output = Join-Path $binDir 'wuji-exec-base.exe'
}
$Output = [System.IO.Path]::GetFullPath($Output)
$outputName = [System.IO.Path]::GetFileNameWithoutExtension($Output)
if (-not $outputName) {
    $outputName = 'default'
}
$outputName = ($outputName -replace '[^A-Za-z0-9._-]', '_')

function Find-Go {
    $externalManual = Join-Path $env:LOCALAPPDATA 'WujiLegion\go-manual\go\bin\go.exe'
    if (Test-Path -LiteralPath $externalManual) {
        return $externalManual
    }

    $manual = Join-Path $binDir 'go-manual\go\bin\go.exe'
    if (Test-Path -LiteralPath $manual) {
        return $manual
    }

    foreach ($candidate in @(
        (Join-Path $binDir 'go\bin\go.exe'),
        (Join-Path $binDir 'go1.25.4.windows-amd64\go\bin\go.exe')
    )) {
        if (Test-Path -LiteralPath $candidate) {
            return $candidate
        }
    }

    $cmd = Get-Command go -ErrorAction SilentlyContinue
    if ($cmd) { return $cmd.Source }

    throw 'Go toolchain not found. Install Go or place go*.windows-amd64.zip under .wuji-tools.'
}

if (-not (Test-Path -LiteralPath $src)) {
    throw "Missing Go source: $src"
}
New-Item -ItemType Directory -Force -Path (Split-Path -Parent $Output) | Out-Null
$go = Find-Go
$goEnvRoot = Join-Path $cacheRoot (Join-Path 'go-env' $outputName)
$goCache = Join-Path $goEnvRoot 'cache'
$goTmp = Join-Path $goEnvRoot 'tmp'
$goModCache = Join-Path $goEnvRoot 'pkg\mod'
New-Item -ItemType Directory -Force -Path $goEnvRoot, $goCache, $goTmp, $goModCache | Out-Null

$previousGoCache = $env:GOCACHE
$previousGoTmpDir = $env:GOTMPDIR
$previousGoModCache = $env:GOMODCACHE
$previousGoTelemetry = $env:GOTELEMETRY
$previousGoEnv = $env:GOENV
$previousAppData = $env:APPDATA
$previousLocalAppData = $env:LOCALAPPDATA
$goAppData = Join-Path $goEnvRoot 'appdata\roaming'
$goLocalAppData = Join-Path $goEnvRoot 'appdata\local'
New-Item -ItemType Directory -Force -Path $goAppData, $goLocalAppData | Out-Null

try {
    $env:GOCACHE = $goCache
    $env:GOTMPDIR = $goTmp
    $env:GOMODCACHE = $goModCache
    $env:GOTELEMETRY = 'off'
    $env:GOENV = 'off'
    $env:APPDATA = $goAppData
    $env:LOCALAPPDATA = $goLocalAppData

    & $go build -trimpath -ldflags '-s -w' -o $Output $src
    if ($LASTEXITCODE -ne 0) {
        throw "go build failed with exit code $LASTEXITCODE"
    }
}
finally {
    $env:GOCACHE = $previousGoCache
    $env:GOTMPDIR = $previousGoTmpDir
    $env:GOMODCACHE = $previousGoModCache
    $env:GOTELEMETRY = $previousGoTelemetry
    $env:GOENV = $previousGoEnv
    $env:APPDATA = $previousAppData
    $env:LOCALAPPDATA = $previousLocalAppData
}
Write-Host "Built: $Output"

$staleBins = @(
    'wuji-cli-test.exe',
    'wuji-cli-dormant.exe',
    'wuji-cli-verify.exe'
)
foreach ($name in $staleBins) {
    $stalePath = Join-Path $binDir $name
    if ((Test-Path -LiteralPath $stalePath) -and ([System.IO.Path]::GetFullPath($stalePath) -ne $Output)) {
        Remove-Item -LiteralPath $stalePath -Force -ErrorAction SilentlyContinue
    }
}

if ([System.IO.Path]::GetFileName($Output) -eq 'wuji-exec-base.exe') {
    $shim = Join-Path ([System.IO.Path]::GetDirectoryName($Output)) 'wuji-cli.cmd'
    $shimContent = @"
@echo off
setlocal
powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0..\scripts\ensure-wuji-cli.ps1" -RepoRoot "%~dp0.." -Quiet
if errorlevel 1 exit /b %errorlevel%
"%~dp0wuji-exec-base.exe" %*
"@
    $currentShim = ''
    if (Test-Path -LiteralPath $shim) {
        try {
            $currentShim = [System.IO.File]::ReadAllText($shim, [System.Text.ASCIIEncoding]::new())
        }
        catch {
            $currentShim = ''
        }
    }
    if ($currentShim -ne $shimContent) {
        [System.IO.File]::WriteAllText($shim, $shimContent, [System.Text.ASCIIEncoding]::new())
    }
    Write-Host "Shim: $shim"

    $psShim = Join-Path ([System.IO.Path]::GetDirectoryName($Output)) 'wuji-cli.ps1'
$psShimContent = @'
param(
    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]]$WujiArgs
)
$repoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot ".."))
& "$repoRoot\scripts\ensure-wuji-cli.ps1" -RepoRoot $repoRoot -Quiet
if (-not $?) {
    exit 1
}
& "$PSScriptRoot\wuji-exec-base.exe" @WujiArgs
exit $LASTEXITCODE
'@
    $currentPsShim = ''
    if (Test-Path -LiteralPath $psShim) {
        try {
            $currentPsShim = [System.IO.File]::ReadAllText($psShim, [System.Text.UTF8Encoding]::new($false))
        }
        catch {
            $currentPsShim = ''
        }
    }
    if ($currentPsShim -ne $psShimContent) {
        [System.IO.File]::WriteAllText($psShim, $psShimContent, [System.Text.UTF8Encoding]::new($false))
    }
    Write-Host "PowerShell Shim: $psShim"
}
