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

  $pptMaster = Get-LockedSourcePath 'ppt-master'
  $pptMasterScripts = Join-Path $pptMaster 'skills\ppt-master\scripts'
  $compileTargets = @(
    (Join-Path $pptMasterScripts 'pptx_to_svg.py'),
    (Join-Path $pptMasterScripts 'svg_to_pptx.py'),
    (Join-Path $pptMasterScripts 'pptx_animations.py')
  )
  $previousPycache = $env:PYTHONPYCACHEPREFIX
  $env:PYTHONPYCACHEPREFIX = Join-Path $scratch 'pycache'
  & $python -m py_compile @compileTargets
  if ($LASTEXITCODE -ne 0) { throw 'PPT Master executable scripts failed compilation' }
  $env:PYTHONPYCACHEPREFIX = $previousPycache
  $pptMasterExamples = @(Get-ChildItem (Join-Path $pptMaster 'examples') -Directory).Count
  if ($pptMasterExamples -lt 20) { throw "PPT Master examples missing: $pptMasterExamples" }
  Write-Output "ppt-master-callable examples=$pptMasterExamples"

  $huashu = Get-LockedSourcePath 'huashu-design'
  & $node --check (Join-Path $huashu 'scripts\export_deck_pptx.mjs')
  if ($LASTEXITCODE -ne 0) { throw 'Huashu editable PPTX exporter failed syntax probe' }
  & $node --check (Join-Path $huashu 'scripts\render-video.js')
  if ($LASTEXITCODE -ne 0) { throw 'Huashu motion renderer failed syntax probe' }
  Write-Output 'huashu-ppt-and-motion-entrypoints-ok'

  $slidev = Get-LockedSourcePath 'slidev-runtime'
  $slidevWork = Join-Path $scratch 'slidev'
  Copy-Item -LiteralPath $slidev -Destination $slidevWork -Recurse
  $pnpmCommand = Get-Command pnpm.cmd -ErrorAction SilentlyContinue
  if (-not $pnpmCommand) { $pnpmCommand = Get-Command pnpm -ErrorAction SilentlyContinue }
  if (-not $pnpmCommand) { throw 'pnpm is required for the locked Slidev behavior probe' }
  Push-Location $slidevWork
  try {
    & $pnpmCommand.Source install --frozen-lockfile
    if ($LASTEXITCODE -ne 0) { throw 'Slidev dependency installation failed' }
    & $pnpmCommand.Source run build
    if ($LASTEXITCODE -ne 0) { throw 'Slidev build failed' }
  } finally { Pop-Location }
  $slidevIndex = Join-Path $slidevWork 'dist\index.html'
  if (-not (Test-Path -LiteralPath $slidevIndex) -or (Get-Item $slidevIndex).Length -lt 1000) { throw 'Slidev output is missing or blank' }
  $slidevShot = Join-Path $scratch 'slidev.png'
  & $node $browserProbe $slidevIndex $slidevShot slidev
  if ($LASTEXITCODE -ne 0) { throw 'Slidev browser behavior probe failed' }
  Write-Output "slidev-built-and-rendered bytes=$((Get-Item $slidevIndex).Length)"

  $fluid = Join-Path $scratch 'stage-fluid'
  & (Join-Path $root 'scripts\materialize-stage-fluid.ps1') -OutputDir $fluid
  if ($LASTEXITCODE -ne 0) { throw 'Fluid background materialization failed' }
  & $node --check (Join-Path $fluid 'stage-fluid.js')
  if ($LASTEXITCODE -ne 0) { throw 'Fluid background script failed syntax probe' }
  $fluidShot = Join-Path $scratch 'fluid.png'
  & $node $browserProbe (Join-Path $fluid 'index.html') $fluidShot pointer
  if ($LASTEXITCODE -ne 0) { throw 'Fluid background browser behavior probe failed' }
  Write-Output 'stage-fluid-rendered-and-moving'

  $humanize = Get-LockedSourcePath 'humanize-ppt'
  $humanizeOut = Join-Path $scratch 'humanize-ppt'
  & $python (Join-Path $humanize 'scripts\smoke_check.py') --out $humanizeOut
  if ($LASTEXITCODE -ne 0) { throw 'Humanize PPT behavior smoke failed' }
  foreach ($required in @('deck_brief.md','ast_outline.md','slide_plan.json','router_plan.json','run_manifest.json','outputs\qa\qa_report.md')) {
    if (-not (Test-Path -LiteralPath (Join-Path $humanizeOut $required))) { throw "Humanize PPT missing $required" }
  }
  Write-Output 'humanize-ppt-ast-plan-and-qa-ok'

  $baoyuDeck = Join-Path (Get-LockedSourcePath 'baoyu-skills') 'skills\baoyu-slide-deck'
  $baoyuStyles = @(Get-ChildItem (Join-Path $baoyuDeck 'references\styles') -Filter '*.md').Count
  $baoyuScripts = @(Get-ChildItem (Join-Path $baoyuDeck 'scripts') -Filter '*.ts').Count
  if ($baoyuStyles -lt 15 -or $baoyuScripts -lt 2) { throw "Baoyu slide-deck incomplete styles=$baoyuStyles scripts=$baoyuScripts" }
  Write-Output "baoyu-slide-deck-retained styles=$baoyuStyles scripts=$baoyuScripts"
  $env:NODE_PATH = $previousNodePath
  $env:CHROME_PATH = $previousChromePath
} finally {
  Remove-Item -LiteralPath $scratch -Recurse -Force -ErrorAction SilentlyContinue
  $env:HOME = $previousHome
  if (Get-Variable previousNodePath -ErrorAction SilentlyContinue) { $env:NODE_PATH = $previousNodePath }
  if (Get-Variable previousChromePath -ErrorAction SilentlyContinue) { $env:CHROME_PATH = $previousChromePath }
  if (Get-Variable previousPycache -ErrorAction SilentlyContinue) { $env:PYTHONPYCACHEPREFIX = $previousPycache }
}
