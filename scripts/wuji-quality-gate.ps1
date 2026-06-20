<#
.SYNOPSIS
  Wuji Legion quality gate for Go, frontend, Python, PowerShell, and ComfyUI plugins.

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
$secretPatterns = @(
    ('sk-' + 'proj-'),
    ('sk-' + 'live-'),
    ('sk-' + 'ant-'),
    ('gh' + 'p_'),
    ('github' + '_pat_'),
    ('xox' + 'b-'),
    ('xox' + 'p-'),
    ('-----BEGIN ' + 'PRIVATE KEY-----')
)

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

function Find-GoTool {
    param([string]$Name)
    $cmd = Get-Command $Name -ErrorAction SilentlyContinue
    if ($cmd) { return $cmd.Source }
    foreach ($candidate in @(
        (Join-Path (Join-Path $env:LOCALAPPDATA 'WujiLegion\go-manual\go\bin') "$Name.exe"),
        (Join-Path ".\.wuji-tools\go-manual\go\bin" "$Name.exe"),
        (Join-Path ".\.wuji-tools\go\bin" "$Name.exe"),
        (Join-Path ".\.wuji-tools" "$Name.exe")
    )) {
        if (Test-Path -LiteralPath $candidate) {
            return (Resolve-Path -LiteralPath $candidate).Path
        }
    }
    return $null
}

function Get-ProjectFiles {
    param([string]$Filter)
    $roots = @("scripts", "tools", "units", "experts")
    $files = @()
    foreach ($projectRoot in $roots) {
        if (Test-Path -LiteralPath $projectRoot) {
            $files += Get-ChildItem -LiteralPath $projectRoot -Recurse -File -Filter $Filter -ErrorAction SilentlyContinue
        }
    }
    return $files
}

function Get-TrackedRootFiles {
    param([string[]]$Names)
    $files = @()
    foreach ($name in $Names) {
        if (Test-Path -LiteralPath $name) {
            $files += Get-Item -LiteralPath $name
        }
    }
    return $files
}

function Read-JsonUtf8 {
    param([string]$Path)
    return [System.IO.File]::ReadAllText($Path, [System.Text.UTF8Encoding]::new($false)) | ConvertFrom-Json
}

function Resolve-MarkdownTarget {
    param(
        [string]$Target,
        [string]$SourceFile
    )

    $clean = $Target.Trim().Trim('<', '>')
    if ($clean -match '^(https?|mailto):') {
        return $null
    }
    if ($clean -match '^/[A-Za-z]:/') {
        $clean = $clean.TrimStart('/')
    }
    if ($clean -match '^[A-Za-z]:[\\/]') {
        if ($clean -match '^(.*?):\d+$' -and (Test-Path -LiteralPath $Matches[1])) {
            return (Resolve-Path -LiteralPath $Matches[1]).Path
        }
        return $clean
    }
    if ($clean -match '^(.*?):\d+$') {
        $candidate = $Matches[1]
        $absolute = Join-Path (Split-Path -Parent $SourceFile) $candidate
        if (Test-Path -LiteralPath $absolute) {
            return (Resolve-Path -LiteralPath $absolute).Path
        }
    }
    return Join-Path (Split-Path -Parent $SourceFile) $clean
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
    $pkg = Read-JsonUtf8 -Path (Resolve-Path -LiteralPath "package.json")
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

$goFiles = Get-ProjectFiles "*.go" | ForEach-Object { $_.FullName }

if ($goFiles.Count -gt 0 -or (Test-Path "go.mod")) {
    $gofmt = Find-GoTool "gofmt"
    if ($gofmt) {
        if ($goFiles.Count -gt 0) {
            $fmtOutput = New-Object System.Collections.Generic.List[string]
            foreach ($goFile in $goFiles) {
                $one = & $gofmt -l $goFile
                if ($LASTEXITCODE -ne 0) {
                    Add-Result "go:fmt" "FAIL" "gofmt failed for $goFile"
                    $one = @()
                }
                foreach ($item in $one) { $fmtOutput.Add($item) }
            }
            if ($fmtOutput.Count -eq 0) {
                Add-Result "go:fmt" "PASS" "gofmt -l"
            } else {
                Add-Result "go:fmt" "FAIL" (($fmtOutput -join ', ') + " needs gofmt")
            }
        }
    } else {
        Add-Result "go:fmt" "SKIP" "gofmt not found"
    }
    if (Test-Path ".\\scripts\\build-wuji-cli.ps1") {
        Run-Check "go:build:wuji-cli" "powershell" @("-NoProfile", "-File", ".\\scripts\\build-wuji-cli.ps1")
    }
    $govulncheck = Find-GoTool "govulncheck"
    if ($govulncheck -and (Test-Path "go.mod")) {
        Run-Check "go:vulncheck" $govulncheck @("./...")
    } else {
        Add-Result "go:vulncheck" "SKIP" "govulncheck or go.mod not found"
    }
}

if (Test-Path "package.json") {
    Run-PackageScript "typecheck"
    Run-PackageScript "lint"
    Run-PackageScript "test"
    Run-PackageScript "build"
}

$pyFiles = Get-ProjectFiles "*.py"
if ((Test-Path "pyproject.toml") -or (Test-Path "requirements.txt") -or ($pyFiles.Count -gt 0)) {
    if (Test-CommandExists "python") {
        foreach ($pyFile in $pyFiles) {
            Run-Check "python:compile:$($pyFile.Name)" "python" @("-m", "py_compile", $pyFile.FullName)
        }
    } else {
        Add-Result "python:compile" "SKIP" "python not found"
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

$psFiles = Get-ProjectFiles "*.ps1"
foreach ($file in $psFiles) {
    try {
        $null = [System.Management.Automation.Language.Parser]::ParseFile($file.FullName, [ref]$null, [ref]$null)
        Add-Result "powershell:parse:$($file.Name)" "PASS" $file.FullName
    } catch {
        Add-Result "powershell:parse:$($file.Name)" "FAIL" $_.Exception.Message
    }
}

$textFiles = @()
foreach ($filter in @("*.md", "*.txt", "*.json", "*.toml", "*.yaml", "*.yml", "*.ps1", "*.py", "*.go")) {
    $textFiles += Get-ProjectFiles $filter
}
$textFiles += Get-TrackedRootFiles @("README.md", "SKILL.md", "GLOBAL_AGENTS.md", "CHANGELOG.md")
$secretScanExempt = @(
    (Resolve-Path -LiteralPath ".\tools\wuji_cli.go").Path,
    (Resolve-Path -LiteralPath ".\scripts\wuji-quality-gate.ps1").Path
)
$seenTextFiles = @{}
foreach ($file in $textFiles) {
    if ($seenTextFiles.ContainsKey($file.FullName)) { continue }
    $seenTextFiles[$file.FullName] = $true
    if ($secretScanExempt -contains $file.FullName) { continue }
    $content = [System.IO.File]::ReadAllText($file.FullName)
    foreach ($pattern in $secretPatterns) {
        if ($content -like "*$pattern*") {
            Add-Result "secret-scan:$($file.Name)" "FAIL" "$pattern in $($file.FullName)"
            break
        }
    }
}

$markdownFiles = @()
$markdownFiles += Get-ProjectFiles "*.md"
$markdownFiles += Get-TrackedRootFiles @("README.md", "SKILL.md", "GLOBAL_AGENTS.md", "CHANGELOG.md")
$seenMarkdownFiles = @{}
foreach ($file in $markdownFiles) {
    if ($seenMarkdownFiles.ContainsKey($file.FullName)) { continue }
    $seenMarkdownFiles[$file.FullName] = $true
    $content = [System.IO.File]::ReadAllText($file.FullName)
    foreach ($match in [regex]::Matches($content, '\[[^\]]+\]\(([^)]+)\)')) {
        $target = $match.Groups[1].Value
        $resolved = Resolve-MarkdownTarget -Target $target -SourceFile $file.FullName
        if (-not $resolved) { continue }
        if (-not (Test-Path -LiteralPath $resolved)) {
            Add-Result "markdown-link:$($file.Name)" "FAIL" "$target missing from $($file.FullName)"
        }
    }
}

if ((Test-Path -LiteralPath ".\config.json") -and (Test-Path -LiteralPath ".\scripts\wuji-install.ps1")) {
    try {
        $config = Read-JsonUtf8 -Path (Resolve-Path -LiteralPath ".\config.json")
        $install = [System.IO.File]::ReadAllText((Resolve-Path -LiteralPath ".\scripts\wuji-install.ps1"))
        $expected = [string]$config.iron_rules_version
        $readsConfigVersion = $install -match [regex]::Escape('$config.iron_rules_version')
        $hasVersionVariable = $install -match [regex]::Escape('$INSTALLER_VERSION')
        $hasBanner = $install -match [regex]::Escape('Installer v$INSTALLER_VERSION')
        if ($expected -and (-not $readsConfigVersion -or -not $hasVersionVariable -or -not $hasBanner)) {
            Add-Result "version:installer" "FAIL" "installer banner does not match config iron_rules_version=$expected"
        } else {
            Add-Result "version:installer" "PASS" "installer banner matches config iron_rules_version=$expected"
        }
        if ($install -notmatch '\[string\]\$Ref' -or $install -notmatch '40-char commit sha') {
            Add-Result "supply-chain:installer-ref" "FAIL" "installer must require pinned 40-char commit sha"
        } else {
            Add-Result "supply-chain:installer-ref" "PASS" "installer requires pinned commit sha"
        }
        if ($install -match 'git clone\s+"https://github.com/\$REPO\.git"\s+\$temp') {
            Add-Result "supply-chain:installer-moving-branch" "FAIL" "installer must not clone moving default branch"
        } else {
            Add-Result "supply-chain:installer-moving-branch" "PASS" "installer avoids moving default branch clone"
        }
        if ($install -match 'if \(Test-Path -LiteralPath \$ensureScript\)\s*\{' -or $install -match '& powershell[^\r\n]+ensure-wuji-cli[^\r\n]+-Quiet') {
            Add-Result "supply-chain:installer-bootstrap" "FAIL" "installer bootstrap must be behind explicit -Bootstrap"
        } else {
            Add-Result "supply-chain:installer-bootstrap" "PASS" "installer bootstrap is explicit"
        }
        if ($install -match 'Copy-Item[^\r\n]+GLOBAL_AGENTS\.md[^\r\n]+-Destination \$AGENTS_DST[^\r\n]+-Force' -and $install -notmatch 'if \(\$InstallAgents\)') {
            Add-Result "supply-chain:installer-agents" "FAIL" "installer AGENTS write must be behind explicit -InstallAgents"
        } else {
            Add-Result "supply-chain:installer-agents" "PASS" "installer AGENTS write is explicit"
        }
    } catch {
        Add-Result "version:installer" "FAIL" $_.Exception.Message
    }
}

$results | Format-Table -AutoSize

if ($results | Where-Object { $_.status -eq "FAIL" }) {
    exit 1
}
