param()

$ErrorActionPreference = 'Stop'

$root = Split-Path -Parent $PSScriptRoot
$fixture = Join-Path $root 'outputs\tests\wuji-sync'
$mirror = Join-Path $env:USERPROFILE '.agents\skills\wuji-sync-test'
$syncScript = Join-Path $PSScriptRoot 'sync-active-skill.ps1'
$genExperts = Join-Path $PSScriptRoot 'gen_experts.py'

function Assert-True {
    param(
        [bool]$Condition,
        [string]$Message
    )
    if (-not $Condition) {
        throw $Message
    }
}

function Resolve-Python {
    $python = Get-Command python -ErrorAction SilentlyContinue
    if ($python) {
        return @{
            Command = $python.Source
            Prefix = @()
        }
    }
    $py = Get-Command py -ErrorAction SilentlyContinue
    if ($py) {
        return @{
            Command = $py.Source
            Prefix = @("-3")
        }
    }
    throw "Python 3 is required for sync regression tests."
}

function Invoke-Python {
    param(
        [string[]]$ScriptArgs
    )
    $python = Resolve-Python
    & $python.Command @($python.Prefix + $ScriptArgs)
    if ($LASTEXITCODE -ne 0) {
        throw "Python command failed: $($ScriptArgs -join ' ')"
    }
}

function Hash-Of {
    param([string]$Path)
    return (Get-FileHash -Algorithm SHA256 -LiteralPath $Path).Hash
}

function Get-CriticalMirrorFiles {
    param([string]$RepoRoot)

    $files = @(
        'SKILL.md',
        'GLOBAL_AGENTS.md',
        'kernel-source.json',
        'fusion-matrix.json',
        'hotpath-manifest.json',
        'residual-entrypoints.json',
        'experts\INDEX.md'
    )

    foreach ($dir in @('security', 'expedition', 'prompt', 'comfyui')) {
        $expertDir = Join-Path $RepoRoot ("experts\" + $dir)
        Assert-True (Test-Path -LiteralPath $expertDir) "missing expert directory: experts\\$dir"
        Get-ChildItem -LiteralPath $expertDir -Filter *.md | ForEach-Object {
            $files += ("experts\" + $dir + "\" + $_.Name)
        }
    }

    return $files
}

try {
if (Test-Path -LiteralPath $fixture) {
    Remove-Item -LiteralPath $fixture -Recurse -Force
}
New-Item -ItemType Directory -Force -Path $fixture | Out-Null
if (Test-Path -LiteralPath $mirror) {
    Remove-Item -LiteralPath $mirror -Recurse -Force
}

$genExpertsSource = [System.IO.File]::ReadAllText($genExperts, [System.Text.UTF8Encoding]::new($false))
Assert-True (-not $genExpertsSource.Contains('WUJI_ALLOW_LEGACY_GEN_EXPERTS')) 'legacy generator escape hatch must stay removed'
Assert-True (-not $genExpertsSource.Contains('EXPERTS: list[dict]')) 'legacy hardcoded expert payload must stay removed'
Assert-True ($genExpertsSource.Contains('checked-in cards remain the source for role text')) 'new generator contract text missing'

Invoke-Python -ScriptArgs @($genExperts)

& powershell -NoProfile -ExecutionPolicy Bypass -File $syncScript -RepoRoot $root -Destination $mirror
if ($LASTEXITCODE -ne 0) {
    throw "sync-active-skill failed exit=$LASTEXITCODE"
}

$criticalFiles = Get-CriticalMirrorFiles -RepoRoot $root

foreach ($relPath in $criticalFiles) {
    $sourcePath = Join-Path $root $relPath
    $mirrorPath = Join-Path $mirror $relPath
    Assert-True (Test-Path -LiteralPath $mirrorPath) "missing mirror file: $relPath"
    Assert-True ((Hash-Of $sourcePath) -eq (Hash-Of $mirrorPath)) "hash mismatch after sync: $relPath"
}

$mirrorIndex = [System.IO.File]::ReadAllText((Join-Path $mirror 'experts\INDEX.md'), [System.Text.UTF8Encoding]::new($false))
Assert-True ($mirrorIndex.Contains('## Specialized Entrances')) 'mirror expert index missing specialized section'
Assert-True ($mirrorIndex.Contains('Specialized Entrances')) 'mirror expert index missing specialized-entrypoint classification'

$mirrorSkill = [System.IO.File]::ReadAllText((Join-Path $mirror 'SKILL.md'), [System.Text.UTF8Encoding]::new($false))
$mirrorAgents = [System.IO.File]::ReadAllText((Join-Path $mirror 'GLOBAL_AGENTS.md'), [System.Text.UTF8Encoding]::new($false))
 $mirrorReadme = [System.IO.File]::ReadAllText((Join-Path $mirror 'README.md'), [System.Text.UTF8Encoding]::new($false))
 $mirrorDistill = [System.IO.File]::ReadAllText((Join-Path $mirror 'units\distillation.md'), [System.Text.UTF8Encoding]::new($false))
Assert-True ($mirrorSkill.Contains('default GPT route')) 'mirror SKILL missing Agnes fallback-to-default-GPT rule'
Assert-True ($mirrorAgents.Contains('default GPT route')) 'mirror GLOBAL_AGENTS missing Agnes fallback-to-default-GPT rule'
Assert-True ($mirrorSkill.Contains('simplest effective path first')) 'mirror SKILL missing simplest-effective-path rule'
Assert-True ($mirrorSkill.Contains('Agnes do broad low-cost scouting first') -or $mirrorSkill.Contains('broad low-cost scouting first')) 'mirror SKILL missing Agnes scouting rule'
Assert-True ($mirrorAgents.Contains('scope, target surface, finish line, out-of-scope exclusions, and completion evidence')) 'mirror GLOBAL_AGENTS missing goal boundary lock rule'
Assert-True ($mirrorReadme.Contains('Superpowers` is gap-fill-only') -or $mirrorReadme.Contains('Superpowers is gap-fill-only')) 'mirror README missing Superpowers gap-fill ruling'
Assert-True ($mirrorReadme.Contains('Agency Agents` is rejected') -or $mirrorReadme.Contains('Agency Agents is rejected')) 'mirror README missing Agency Agents reject ruling'
Assert-True ($mirrorDistill.Contains('Headroom`: replace-level') -or $mirrorDistill.Contains('Headroom: replace-level')) 'mirror distillation missing Headroom replace ruling'
Assert-True ($mirrorDistill.Contains('gstack`: reject') -or $mirrorDistill.Contains('gstack: reject')) 'mirror distillation missing gstack reject ruling'

Write-Host 'PASS test-wuji-sync'
}
finally {
    if (Test-Path -LiteralPath $mirror) {
        Remove-Item -LiteralPath $mirror -Recurse -Force
    }
    if (Test-Path -LiteralPath $fixture) {
        Remove-Item -LiteralPath $fixture -Recurse -Force
    }
    $testsRoot = Join-Path $root 'outputs\tests'
    if ((Test-Path -LiteralPath $testsRoot) -and -not (Get-ChildItem -LiteralPath $testsRoot -Force | Select-Object -First 1)) {
        Remove-Item -LiteralPath $testsRoot -Force -ErrorAction SilentlyContinue
    }
}
