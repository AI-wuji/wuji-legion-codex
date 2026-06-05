function Get-WujiPptRuntime {
    param(
        [string]$SkillDir = "",
        [string]$NodePath = "",
        [string]$NodeModules = "",
        [string]$PythonPath = ""
    )

    $root = Split-Path -Parent $PSScriptRoot
    if ([string]::IsNullOrWhiteSpace($SkillDir)) {
        $presentationsRoot = Join-Path $env:USERPROFILE ".codex\plugins\cache\openai-primary-runtime\presentations"
        $latest = Get-ChildItem -LiteralPath $presentationsRoot -Directory -ErrorAction SilentlyContinue |
            Sort-Object Name -Descending |
            Select-Object -First 1
        if ($latest) {
            $SkillDir = Join-Path $latest.FullName "skills\presentations"
        }
    }

    if ([string]::IsNullOrWhiteSpace($NodePath)) {
        $NodePath = Join-Path $env:USERPROFILE ".cache\codex-runtimes\codex-primary-runtime\dependencies\node\bin\node.exe"
    }

    if ([string]::IsNullOrWhiteSpace($NodeModules)) {
        $NodeModules = Join-Path $env:USERPROFILE ".cache\codex-runtimes\codex-primary-runtime\dependencies\node\node_modules"
    }

    if ([string]::IsNullOrWhiteSpace($PythonPath)) {
        $PythonPath = Join-Path $env:USERPROFILE ".cache\codex-runtimes\codex-primary-runtime\dependencies\python\python.exe"
    }

    if (-not (Test-Path -LiteralPath $SkillDir)) {
        throw "Presentations skill not found: $SkillDir"
    }
    if (-not (Test-Path -LiteralPath $NodePath)) {
        throw "Bundled Node not found: $NodePath"
    }
    if (-not (Test-Path -LiteralPath $NodeModules)) {
        throw "Bundled node_modules not found: $NodeModules"
    }
    if (-not (Test-Path -LiteralPath $PythonPath)) {
        throw "Bundled Python not found: $PythonPath"
    }

    return @{
        Root        = $root
        BinDir      = Join-Path $root ".wuji-tools"
        ScriptsDir  = Join-Path $root "scripts"
        SkillDir    = $SkillDir
        NodePath    = $NodePath
        NodeModules = $NodeModules
        PythonPath  = $PythonPath
    }
}

function Get-WujiGoToolchain {
    param([hashtable]$Runtime)

    $binDir = $Runtime.BinDir
    $manual = Join-Path $binDir 'go-manual\go\bin\go.exe'
    if (Test-Path -LiteralPath $manual) {
        return $manual
    }

    $portable = Get-ChildItem -LiteralPath $binDir -Recurse -Filter go.exe -ErrorAction SilentlyContinue |
        Where-Object { $_.FullName -like '*\go\bin\go.exe' } |
        Sort-Object LastWriteTime -Descending |
        Select-Object -First 1
    if ($portable) {
        return $portable.FullName
    }

    throw "Go toolchain not found under $binDir"
}

function Ensure-WujiNativeUnzip {
    param([hashtable]$Runtime)

    $source = Join-Path $Runtime.Root 'tools\wuji_unzip.go'
    $target = Join-Path $Runtime.BinDir 'unzip.exe'
    if (-not (Test-Path -LiteralPath $source)) {
        throw "Missing native unzip source: $source"
    }

    $needsBuild = -not (Test-Path -LiteralPath $target)
    if (-not $needsBuild) {
        $sourceTime = (Get-Item -LiteralPath $source).LastWriteTimeUtc
        $targetTime = (Get-Item -LiteralPath $target).LastWriteTimeUtc
        if ($sourceTime -gt $targetTime) {
            $needsBuild = $true
        }
    }
    if (-not $needsBuild) {
        return $target
    }

    $go = Get-WujiGoToolchain -Runtime $Runtime
    $goEnvRoot = Join-Path $Runtime.BinDir 'go-env'
    $goCache = Join-Path $goEnvRoot 'cache'
    $goTmp = Join-Path $goEnvRoot 'tmp'
    $goModCache = Join-Path $goEnvRoot 'pkg\mod'
    $goAppData = Join-Path $goEnvRoot 'appdata\roaming'
    $goLocalAppData = Join-Path $goEnvRoot 'appdata\local'
    $null = New-Item -ItemType Directory -Force -Path $Runtime.BinDir, $goEnvRoot, $goCache, $goTmp, $goModCache, $goAppData, $goLocalAppData

    $previousGoCache = $env:GOCACHE
    $previousGoTmpDir = $env:GOTMPDIR
    $previousGoModCache = $env:GOMODCACHE
    $previousGoTelemetry = $env:GOTELEMETRY
    $previousGoEnv = $env:GOENV
    $previousAppData = $env:APPDATA
    $previousLocalAppData = $env:LOCALAPPDATA

    try {
        $env:GOCACHE = $goCache
        $env:GOTMPDIR = $goTmp
        $env:GOMODCACHE = $goModCache
        $env:GOTELEMETRY = 'off'
        $env:GOENV = 'off'
        $env:APPDATA = $goAppData
        $env:LOCALAPPDATA = $goLocalAppData

        & $go build -trimpath -ldflags '-s -w' -o $target $source
        if ($LASTEXITCODE -ne 0 -or -not (Test-Path -LiteralPath $target)) {
            throw "Failed to build native unzip helper at $target"
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

    return $target
}

function Test-WujiSameResolvedPath {
    param(
        [string]$Left,
        [string]$Right
    )

    try {
        $leftPath = [System.IO.Path]::GetFullPath((Resolve-Path -LiteralPath $Left).Path).TrimEnd('\')
        $rightPath = [System.IO.Path]::GetFullPath((Resolve-Path -LiteralPath $Right).Path).TrimEnd('\')
        return $leftPath.Equals($rightPath, [System.StringComparison]::OrdinalIgnoreCase)
    }
    catch {
        return $false
    }
}

function Ensure-WujiWorkspacePackageLink {
    param(
        [string]$Workspace,
        [string]$PackageName,
        [string]$SourcePath
    )

    if (-not (Test-Path -LiteralPath $SourcePath)) {
        return
    }

    $workspaceRoot = [System.IO.Path]::GetFullPath($Workspace)
    $packageSegments = $PackageName -split '/'
    $target = Join-Path (Join-Path $workspaceRoot 'node_modules') ([System.IO.Path]::Combine($packageSegments))
    $parent = Split-Path -Parent $target
    $null = New-Item -ItemType Directory -Force -Path $parent

    if (Test-Path -LiteralPath $target) {
        if (Test-WujiSameResolvedPath -Left $target -Right $SourcePath) {
            return
        }
        $resolvedTarget = [System.IO.Path]::GetFullPath($target)
        if (-not $resolvedTarget.StartsWith($workspaceRoot, [System.StringComparison]::OrdinalIgnoreCase)) {
            throw "Refusing to replace package link outside workspace: $resolvedTarget"
        }
        Remove-Item -LiteralPath $target -Recurse -Force
    }

    try {
        New-Item -ItemType Junction -Path $target -Target $SourcePath | Out-Null
        return
    }
    catch {
        $resolvedTarget = [System.IO.Path]::GetFullPath($target)
        if (-not $resolvedTarget.StartsWith($workspaceRoot, [System.StringComparison]::OrdinalIgnoreCase)) {
            throw "Refusing to copy package source outside workspace: $resolvedTarget"
        }
        Copy-Item -LiteralPath $SourcePath -Destination $target -Recurse -Force
    }
}

function Initialize-WujiArtifactWorkspace {
    param(
        [hashtable]$Runtime,
        [string]$Workspace
    )

    $workspaceRoot = [System.IO.Path]::GetFullPath($Workspace)
    $null = New-Item -ItemType Directory -Force -Path $workspaceRoot

    $packageJsonPath = Join-Path $workspaceRoot 'package.json'
    if (-not (Test-Path -LiteralPath $packageJsonPath)) {
        [System.IO.File]::WriteAllText(
            $packageJsonPath,
            "{`n  `"private`": true,`n  `"type`": `"module`"`n}`n",
            [System.Text.UTF8Encoding]::new($false)
        )
    }

    Ensure-WujiWorkspacePackageLink -Workspace $workspaceRoot -PackageName '@oai/artifact-tool' -SourcePath (Join-Path $Runtime.NodeModules '@oai\artifact-tool')
    Ensure-WujiWorkspacePackageLink -Workspace $workspaceRoot -PackageName 'lucide' -SourcePath (Join-Path $Runtime.NodeModules 'lucide')
}

function Invoke-WujiNodeScript {
    param(
        [hashtable]$Runtime,
        [string]$ScriptPath,
        [string[]]$Arguments
    )

    $previousHome = $env:HOME
    $previousNodePath = $env:NODE_PATH
    $previousPython = $env:PYTHON
    $previousSkillDir = $env:WUJI_PRESENTATIONS_SKILL_DIR
    $previousPath = $env:PATH
    $extraNodePath = Join-Path $Runtime.NodeModules ".pnpm\node_modules"
    $nodePathList = @($Runtime.NodeModules)
    if (Test-Path -LiteralPath $extraNodePath) {
        $nodePathList += $extraNodePath
    }
    $nativeUnzip = Ensure-WujiNativeUnzip -Runtime $Runtime

    try {
        $env:HOME = $env:USERPROFILE
        $env:NODE_PATH = ($nodePathList -join ';')
        $env:PYTHON = $Runtime.PythonPath
        $env:WUJI_PRESENTATIONS_SKILL_DIR = $Runtime.SkillDir
        $env:PATH = "$($Runtime.BinDir);$($Runtime.ScriptsDir);$previousPath"
        & $Runtime.NodePath $ScriptPath @Arguments | Out-Host
        return $LASTEXITCODE
    }
    finally {
        $env:HOME = $previousHome
        $env:NODE_PATH = $previousNodePath
        $env:PYTHON = $previousPython
        $env:WUJI_PRESENTATIONS_SKILL_DIR = $previousSkillDir
        $env:PATH = $previousPath
    }
}
