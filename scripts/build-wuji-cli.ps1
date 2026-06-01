param(
    [string]$Output
)

$ErrorActionPreference = 'Stop'

$root = Split-Path -Parent $PSScriptRoot
$src = Join-Path $root 'tools\wuji_cli.go'
$binDir = Join-Path $root '.wuji-tools'
if (-not $Output) {
    $Output = Join-Path $binDir 'wuji-exec-base.exe'
}
$Output = [System.IO.Path]::GetFullPath($Output)

function Find-Go {
    $cmd = Get-Command go -ErrorAction SilentlyContinue
    if ($cmd) { return $cmd.Source }

    $portable = Get-ChildItem -LiteralPath $binDir -Recurse -Filter go.exe -ErrorAction SilentlyContinue |
        Where-Object { $_.FullName -like '*\go\bin\go.exe' } |
        Sort-Object FullName |
        Select-Object -First 1
    if ($portable) { return $portable.FullName }

    $zip = Get-ChildItem -LiteralPath $binDir -Filter 'go*.windows-amd64.zip' -File -ErrorAction SilentlyContinue |
        Sort-Object LastWriteTime -Descending |
        Select-Object -First 1
    if ($zip) {
        $dest = Join-Path $binDir ($zip.BaseName)
        if (-not (Test-Path -LiteralPath (Join-Path $dest 'go\bin\go.exe'))) {
            Expand-Archive -LiteralPath $zip.FullName -DestinationPath $dest -Force
        }
    }

    throw 'Go toolchain not found. Install Go or place go*.windows-amd64.zip under .wuji-tools.'
}

if (-not (Test-Path -LiteralPath $src)) {
    throw "Missing Go source: $src"
}
New-Item -ItemType Directory -Force -Path (Split-Path -Parent $Output) | Out-Null
$go = Find-Go
& $go build -trimpath -ldflags '-s -w' -o $Output $src
if ($LASTEXITCODE -ne 0) {
    throw "go build failed with exit code $LASTEXITCODE"
}
Write-Host "Built: $Output"

if ([System.IO.Path]::GetFileName($Output) -eq 'wuji-exec-base.exe') {
    $shim = Join-Path ([System.IO.Path]::GetDirectoryName($Output)) 'wuji-cli.cmd'
    $shimContent = "@echo off`r`n`"%~dp0wuji-exec-base.exe`" %*`r`n"
    [System.IO.File]::WriteAllText($shim, $shimContent, [System.Text.ASCIIEncoding]::new())
    Write-Host "Shim: $shim"
}
