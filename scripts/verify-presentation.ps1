$ErrorActionPreference = 'Stop'
$root = if ($env:WUJI_ROOT) { $env:WUJI_ROOT } else { Split-Path $PSScriptRoot -Parent }
$sourceLock = Get-Content -Raw -Encoding UTF8 -LiteralPath (Join-Path $root 'sources.lock.json') | ConvertFrom-Json
function Get-LockedSourcePath([string]$Id) {
  $matches = @($sourceLock.sources | Where-Object id -eq $Id)
  if ($matches.Count -ne 1) { throw "sources.lock.json must contain exactly one $Id source" }
  $path = & (Join-Path $PSScriptRoot 'expand-wuji-path.ps1') -PathValue $matches[0].path -Root $root
  if (-not (Test-Path -LiteralPath $path -PathType Container)) { throw "Locked source is missing: $Id ($path)" }
  return $path
}
function Add-ProbeArtifact([string]$Id, [string]$Source, [string]$Name) {
  if (-not (Test-Path -LiteralPath $Source -PathType Leaf) -or (Get-Item -LiteralPath $Source).Length -lt 1) {
    throw "Probe evidence is missing or empty: $Id ($Source)"
  }
  $target = Join-Path $env:WUJI_PROBE_EVIDENCE_DIR $Name
  if ([IO.Path]::GetFullPath($Source) -ne [IO.Path]::GetFullPath($target)) {
    Copy-Item -LiteralPath $Source -Destination $target -Force
  }
  [pscustomobject]@{
    id = $Id
    path = $Name
    sha256 = (Get-FileHash -Algorithm SHA256 -LiteralPath $target).Hash.ToLowerInvariant()
  }
}
$skillRoot = Join-Path $env:USERPROFILE '.codex\plugins\cache\openai-primary-runtime\presentations'
$version = Get-ChildItem -LiteralPath $skillRoot -Directory | Sort-Object -Property @{ Expression = {
  try { [version]$_.Name } catch { [version]'0.0' }
}; Descending = $true } | Select-Object -First 1
if (-not $version) { throw 'Presentations runtime is not installed' }
$skill = Join-Path $version.FullName 'skills\presentations'
$layoutRoot = Join-Path $skill 'assets\builtin_templates\codex-grid-layout-library'
$setup = Join-Path $skill 'container_tools\setup_artifact_tool_workspace.mjs'
$node = Join-Path $env:USERPROFILE '.cache\codex-runtimes\codex-primary-runtime\dependencies\node\bin\node.exe'
if (-not (Test-Path -LiteralPath $node)) { throw 'Bundled Node runtime is missing' }
$previousHome = $env:HOME
$env:HOME = $env:USERPROFILE
$nodeModules = Join-Path $env:USERPROFILE '.cache\codex-runtimes\codex-primary-runtime\dependencies\node\node_modules'
$python = Join-Path $env:USERPROFILE '.cache\codex-runtimes\codex-primary-runtime\dependencies\python\python.exe'
$evidenceDir = $env:WUJI_PROBE_EVIDENCE_DIR
if (-not $evidenceDir -or -not (Test-Path -LiteralPath $evidenceDir -PathType Container)) {
  throw 'WUJI_PROBE_EVIDENCE_DIR is required for a behavior probe'
}

$scratch = Join-Path $env:TEMP ('wuji-2-ppt-probe-' + [guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $scratch | Out-Null
try {
  $catalogPath = Join-Path $scratch 'template-catalog.json'
  & (Join-Path $root 'scripts\build-presentation-catalog.ps1') -Output $catalogPath | Out-Null
  $catalog = Get-Content -Raw -Encoding UTF8 -LiteralPath $catalogPath | ConvertFrom-Json
  if ($catalog.counts.web_deck -lt 100 -or $catalog.counts.editable_pptx -lt 60) { throw 'Unified presentation catalog is incomplete' }
  Write-Output "presentation-catalog-unified web=$($catalog.counts.web_deck) editable=$($catalog.counts.editable_pptx)"
  & $node $setup --workspace $scratch
  if ($LASTEXITCODE -ne 0) { throw 'Artifact-tool workspace setup failed' }
  Copy-Item -LiteralPath (Join-Path $root 'capabilities\presentation\probe.mjs') -Destination (Join-Path $scratch 'probe.mjs')
  $probeLayoutRoot = Join-Path $scratch 'layout'
  New-Item -ItemType Directory -Path $probeLayoutRoot | Out-Null
  Copy-Item -LiteralPath (Join-Path $layoutRoot 'artifact-tool-compose') -Destination $probeLayoutRoot -Recurse
  $pptx = Join-Path $scratch 'behavior-probe.pptx'
  & $node (Join-Path $scratch 'probe.mjs') $probeLayoutRoot $pptx
  if ($LASTEXITCODE -ne 0) { throw 'Presentation behavior probe failed' }
  Add-Type -AssemblyName System.IO.Compression.FileSystem
  $zip = [IO.Compression.ZipFile]::OpenRead($pptx)
  try {
    $slides = @($zip.Entries | Where-Object { $_.FullName -match '^ppt/slides/slide\d+\.xml$' })
    if ($slides.Count -ne 2) { throw "Expected 2 generated slides, got $($slides.Count)" }
    $shapeCount = 0
    foreach ($entry in $slides) {
      $reader = [IO.StreamReader]::new($entry.Open())
      try { $xml = $reader.ReadToEnd() } finally { $reader.Dispose() }
      $shapeCount += ([regex]::Matches($xml, '<p:(sp|pic|graphicFrame)>')).Count
    }
    if ($shapeCount -lt 6) { throw "Editable object evidence is too weak: $shapeCount" }
  } finally { $zip.Dispose() }
  Write-Output "pptx-created slides=2 editable-shapes=$shapeCount"

  $htmlPpt = Get-LockedSourcePath 'html-ppt-skill'
  $themeCount = @(Get-ChildItem (Join-Path $htmlPpt 'assets\themes') -Filter '*.css').Count
  $templateCount = @(Get-ChildItem (Join-Path $htmlPpt 'templates') -Recurse -Filter '*.html').Count
  $fxCount = @(Get-ChildItem (Join-Path $htmlPpt 'assets\animations\fx') -Filter '*.js').Count
  if ($themeCount -lt 30 -or $templateCount -lt 40 -or $fxCount -lt 15) {
    throw "html-ppt asset retention failed themes=$themeCount templates=$templateCount fx=$fxCount"
  }
  $previousNodePath = $env:NODE_PATH
  $previousChromePath = $env:CHROME_PATH
  $env:CHROME_PATH = 'C:\Program Files\Google\Chrome\Application\chrome.exe'
  $playwrightCoreModules = Get-ChildItem (Join-Path $nodeModules '.pnpm') -Directory -Filter 'playwright-core@*' | ForEach-Object {
    if ($_.Name -match '^playwright-core@(?<version>\d+(?:\.\d+){1,3})(?:_|$)') {
      [pscustomobject]@{ Directory = $_; Version = [version]$Matches.version }
    }
  } | Sort-Object Version -Descending | Select-Object -First 1 | ForEach-Object { Join-Path $_.Directory.FullName 'node_modules' }
  if (-not $playwrightCoreModules) { throw 'Playwright Core runtime is missing' }
  $env:NODE_PATH = @($nodeModules, $playwrightCoreModules) -join ';'
  $browserProbe = Join-Path $root 'capabilities\presentation\probe-browser.cjs'
  $htmlShot = Join-Path $scratch 'html-ppt.png'
  & $node $browserProbe (Join-Path $htmlPpt 'templates\animation-showcase.html') $htmlShot presenter
  if ($LASTEXITCODE -ne 0) { throw 'html-ppt browser behavior probe failed' }
  Write-Output "html-ppt-rendered themes=$themeCount templates=$templateCount fx=$fxCount"

  $fluid = Join-Path $scratch 'stage-fluid'
  & (Join-Path $root 'scripts\materialize-stage-fluid.ps1') -OutputDir $fluid
  if ($LASTEXITCODE -ne 0) { throw 'Fluid background materialization failed' }
  & $node --check (Join-Path $fluid 'stage-fluid.js')
  if ($LASTEXITCODE -ne 0) { throw 'Fluid background script failed syntax probe' }
  $fluidShot = Join-Path $scratch 'fluid.png'
  & $node $browserProbe (Join-Path $fluid 'index.html') $fluidShot pointer
  if ($LASTEXITCODE -ne 0) { throw 'Fluid background browser behavior probe failed' }
  Write-Output 'stage-fluid-rendered-and-moving'

  # DashiAI is the only secondary presentation package selected automatically
  # by a narrow semantic trigger, so its real output belongs in this probe.
  & (Join-Path $root 'scripts\verify-dashiai-ppt.ps1') -Root $root | Out-Host
  if ($LASTEXITCODE -ne 0) { throw 'DashiAI behavior probe failed' }

  $assertionsPath = Join-Path $evidenceDir 'presentation-assertions.json'
  $assertions = [ordered]@{
    fixture = 'unified-presentation-artifact-v1'
    catalog_web_deck = [int]$catalog.counts.web_deck
    catalog_editable_pptx = [int]$catalog.counts.editable_pptx
    pptx_slides = 2
    pptx_editable_shapes = [int]$shapeCount
    html_themes = $themeCount
    html_templates = $templateCount
    html_effects = $fxCount
    stage_fluid_rendered = $true
    dashiai_rendered = $true
  }
  [IO.File]::WriteAllText($assertionsPath, ($assertions | ConvertTo-Json -Compress), [Text.UTF8Encoding]::new($false))
  $probeEvidence = @(
    (Add-ProbeArtifact 'assertions' $assertionsPath 'presentation-assertions.json')
    (Add-ProbeArtifact 'pptx' $pptx 'behavior-probe.pptx')
    (Add-ProbeArtifact 'html-render' $htmlShot 'html-ppt.png')
    (Add-ProbeArtifact 'fluid-render' $fluidShot 'fluid.png')
    (Add-ProbeArtifact 'dashiai-assertions' (Join-Path $evidenceDir 'dashiai-ppt-assertions.json') 'dashiai-ppt-assertions.json')
  )
  $probeReceipt = [ordered]@{
    wuji_probe = 'behavior'
    fixture = 'unified-presentation-artifact-v1'
    passed = $true
    evidence = $probeEvidence
    signature = 'unified-presentation-contract-v1'
  } | ConvertTo-Json -Compress -Depth 5
  $env:NODE_PATH = $previousNodePath
  $env:CHROME_PATH = $previousChromePath
} finally {
  Remove-Item -LiteralPath $scratch -Recurse -Force -ErrorAction SilentlyContinue
  $env:HOME = $previousHome
  if (Get-Variable previousNodePath -ErrorAction SilentlyContinue) { $env:NODE_PATH = $previousNodePath }
  if (Get-Variable previousChromePath -ErrorAction SilentlyContinue) { $env:CHROME_PATH = $previousChromePath }
}
Write-Output $probeReceipt
