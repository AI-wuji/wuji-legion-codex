<#
.SYNOPSIS
  Wuji Legion quality gate for Rust, frontend, Python, PowerShell, and ComfyUI plugins.

.DESCRIPTION
  Detects project type and runs non-destructive quality checks when the matching tools/scripts exist.
  It never claims success for a missing tool; missing checks are reported as SKIP.
#>

param(
    [string]$Path = "."
)

$ErrorActionPreference = "Stop"
$root = Resolve-Path -LiteralPath $Path
Set-Location $root

$results = New-Object System.Collections.Generic.List[object]

function Add-Result {
    param(
        [string]$Name,
        [string]$Status,
        [string]$Detail
    )
    $results.Add([pscustomobject]@{
        name = $Name
        status = $Status
        detail = $Detail
    })
}

function Test-CommandExists {
    param([string]$Name)
    return $null -ne (Get-Command $Name -ErrorAction SilentlyContinue)
}

function Run-Check {
    param(
        [string]$Name,
        [string]$Command,
        [string[]]$Arguments = @()
    )
    try {
        & $Command @Arguments
        if ($LASTEXITCODE -eq 0 -or $null -eq $LASTEXITCODE) {
            Add-Result $Name "PASS" "$Command $($Arguments -join ' ')"
        } else {
            Add-Result $Name "FAIL" "$Command exited with $LASTEXITCODE"
        }
    } catch {
        Add-Result $Name "FAIL" $_.Exception.Message
    }
}

function Run-PackageScript {
    param([string]$Script)
    if (-not (Test-Path "package.json")) { return }
    $pkg = Get-Content -Raw "package.json" | ConvertFrom-Json
    if ($pkg.scripts.PSObject.Properties.Name -contains $Script) {
        if (Test-CommandExists "pnpm") {
            Run-Check "frontend:$Script" "pnpm" @($Script)
        } elseif (Test-CommandExists "npm") {
            Run-Check "frontend:$Script" "npm" @("run", $Script)
        } else {
            Add-Result "frontend:$Script" "SKIP" "npm/pnpm not found"
        }
    } else {
        Add-Result "frontend:$Script" "SKIP" "package.json has no script '$Script'"
    }
}

if (Test-Path "Cargo.toml") {
    if (Test-CommandExists "cargo") {
        Run-Check "rust:fmt" "cargo" @("fmt", "--check")
        Run-Check "rust:check" "cargo" @("check", "--all-targets")
        Run-Check "rust:clippy" "cargo" @("clippy", "--all-targets", "--", "-D", "warnings")
        Run-Check "rust:test" "cargo" @("test")
        if (Test-CommandExists "cargo-audit") {
            Run-Check "rust:audit" "cargo" @("audit")
        } else {
            Add-Result "rust:audit" "SKIP" "cargo-audit not found"
        }
        Run-Check "rust:deny" "cargo" @("deny", "check")
    } else {
        Add-Result "rust" "SKIP" "cargo not found"
    }
}

if (Test-Path "package.json") {
    Run-PackageScript "typecheck"
    Run-PackageScript "lint"
    Run-PackageScript "test"
    Run-PackageScript "build"
}

if ((Test-Path "pyproject.toml") -or (Test-Path "requirements.txt") -or (Get-ChildItem -Filter "*.py" -File -ErrorAction SilentlyContinue)) {
    if (Test-CommandExists "python") {
        Run-Check "python:compileall" "python" @("-m", "compileall", ".")
    } else {
        Add-Result "python:compileall" "SKIP" "python not found"
    }
}

if (Test-Path "__init__.py") {
    if (Test-CommandExists "python") {
        $smoke = @'
import importlib.util
from pathlib import Path
p = Path("__init__.py")
spec = importlib.util.spec_from_file_location("wuji_plugin_smoke", p)
m = importlib.util.module_from_spec(spec)
spec.loader.exec_module(m)
assert hasattr(m, "NODE_CLASS_MAPPINGS"), "missing NODE_CLASS_MAPPINGS"
print("ComfyUI plugin import smoke OK")
'@
        $tmp = New-TemporaryFile
        Set-Content -LiteralPath $tmp -Value $smoke -Encoding UTF8
        Run-Check "comfyui:import-smoke" "python" @($tmp.FullName)
        Remove-Item -LiteralPath $tmp -Force -ErrorAction SilentlyContinue
    }
}

$psFiles = Get-ChildItem -Recurse -File -Filter "*.ps1" -ErrorAction SilentlyContinue
foreach ($file in $psFiles) {
    try {
        $null = [System.Management.Automation.Language.Parser]::ParseFile($file.FullName, [ref]$null, [ref]$null)
        Add-Result "powershell:parse:$($file.Name)" "PASS" $file.FullName
    } catch {
        Add-Result "powershell:parse:$($file.Name)" "FAIL" $_.Exception.Message
    }
}

$results | Format-Table -AutoSize

if ($results | Where-Object { $_.status -eq "FAIL" }) {
    exit 1
}
