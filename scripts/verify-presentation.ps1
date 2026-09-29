$ErrorActionPreference = 'Stop'
. (Join-Path $PSScriptRoot 'sha256.ps1')
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
$editableSkill = Join-Path $root 'capabilities\presentation\skills\wuji-editable-deck\SKILL.md'
$webSkill = Join-Path $root 'capabilities\presentation\skills\wuji-web-deck\SKILL.md'
$candidateReview = Join-Path $root 'references\presentation-external-candidate-review.md'
foreach ($requiredContract in @(
  @{ Path = $editableSkill; Marker = '## Distilled PPTX Contract' },
  @{ Path = $webSkill; Marker = '## Distilled Web Presentation Contract' },
  @{ Path = $candidateReview; Marker = '## GitHub Candidates' },
  @{ Path = $candidateReview; Marker = '## Candidate Distillation Matrix' },
  @{ Path = $candidateReview; Marker = 'returned no verifiable exact PPT Skill/repository' }
)) {
  if (-not (Test-Path -LiteralPath $requiredContract.Path -PathType Leaf)) { throw "Presentation contract file is missing: $($requiredContract.Path)" }
  $contractText = Get-Content -Raw -Encoding UTF8 -LiteralPath $requiredContract.Path
  if (-not $contractText.Contains($requiredContract.Marker)) { throw "Presentation contract marker is missing: $($requiredContract.Marker)" }
}
$pptMasterRoot = Get-LockedSourcePath 'ppt-master'
foreach ($requiredPptMasterFile in @(
  'skills\ppt-master\SKILL.md',
  'skills\ppt-master\scripts\pptx_intake.py',
  'skills\ppt-master\scripts\pptx_to_svg.py',
  'skills\ppt-master\scripts\svg_to_pptx.py',
  'skills\ppt-master\scripts\pptx_template_import.py',
  'skills\ppt-master\scripts\pptx_delivery_check.py',
  'skills\ppt-master\scripts\native_payloads.py',
  'LICENSE'
)) {
  if (-not (Test-Path -LiteralPath (Join-Path $pptMasterRoot $requiredPptMasterFile) -PathType Leaf)) { throw "PPT Master retained entrypoint is missing: $requiredPptMasterFile" }
}
$pptMasterLicense = Get-Content -Raw -Encoding UTF8 -LiteralPath (Join-Path $pptMasterRoot 'LICENSE')
if (-not $pptMasterLicense.Contains('MIT License')) { throw 'PPT Master license evidence is not MIT' }
$version = Get-ChildItem -LiteralPath $skillRoot -Directory | Sort-Object -Property @{ Expression = {
  try { [version]$_.Name } catch { [version]'0.0' }
}; Descending = $true } | Select-Object -First 1
if (-not $version) { throw 'Presentations runtime is not installed' }
$skill = Join-Path $version.FullName 'skills\presentations'
$renderPresentation = Join-Path $skill 'container_tools\render_presentation.mjs'
$runtimeApi = Join-Path $skill 'artifact_tool_docs\API_QUICK_START.md'
foreach ($runtimeFile in @($renderPresentation, $runtimeApi)) {
  if (-not (Test-Path -LiteralPath $runtimeFile -PathType Leaf)) { throw "Latest presentations runtime file is missing: $runtimeFile" }
}
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
  $probeNodeModules = Join-Path $scratch 'node_modules'
  New-Item -ItemType Junction -Path $probeNodeModules -Target $nodeModules | Out-Null
  $previousRuntimeNodeModules = $env:RUNTIME_NODE_MODULES
  $env:RUNTIME_NODE_MODULES = $nodeModules
	$wuji = Join-Path $root 'bin\wuji.exe'
	if (-not (Test-Path -LiteralPath $wuji -PathType Leaf)) { throw 'wuji binary is required for presentation fusion selection' }
	$invoke = Join-Path $root 'capabilities\presentation\scripts\invoke-presentation.ps1'
	$invocations = @()
	foreach ($selection in @(
	  @{ Engine = 'editable-pptx'; Compatibility = 'editable-pptx,default'; Evidence = 'editable-invocation.json' },
	  @{ Engine = 'web-deck'; Compatibility = 'web-deck,default'; Evidence = 'web-invocation.json' },
	  @{ Engine = 'web-deck'; Compatibility = 'web-deck,stage-fluid'; Evidence = 'fluid-invocation.json' }
	)) {
	  $rawContract = (& $wuji asset-select --root $root --capability presentation --domain $selection.Engine --compatibility $selection.Compatibility 2>&1) -join [Environment]::NewLine
	  if ($LASTEXITCODE -ne 0) { throw "presentation fusion selection failed for $($selection.Compatibility): $rawContract" }
	  $contract = $rawContract | ConvertFrom-Json
	  if (-not $contract.asset_id -or -not $contract.asset_sha256 -or -not $contract.entrypoint_sha256 -or $contract.asset_bytes -le 0) { throw "presentation fusion contract is incomplete for $($selection.Engine)" }
	  $record = Join-Path $scratch $selection.Evidence
	  & $invoke -Engine $selection.Engine -AssetId $contract.asset_id -AssetPath $contract.asset_path -AssetSHA256 $contract.asset_sha256 -AssetBytes $contract.asset_bytes -Output $record | Out-Null
	  if ($LASTEXITCODE -ne 0) { throw "presentation fusion invocation failed for $($selection.Compatibility)" }
	  $proof = Get-Content -Raw -Encoding UTF8 -LiteralPath $record | ConvertFrom-Json
	  if ($proof.contract_asset_sha256 -ne $contract.asset_sha256 -or $proof.contract_asset_bytes -ne $contract.asset_bytes -or -not $proof.selected_catalog_sha256) { throw "presentation fusion invocation evidence is incomplete for $($selection.Compatibility)" }
	  $invocations += [pscustomobject]@{ Contract = $contract; Path = $record; Evidence = [IO.Path]::GetFileName($record) }
	}
  $catalogPath = Join-Path $scratch 'template-catalog.json'
  & (Join-Path $root 'scripts\build-presentation-catalog.ps1') -Output $catalogPath | Out-Null
  $catalog = Get-Content -Raw -Encoding UTF8 -LiteralPath $catalogPath | ConvertFrom-Json
  if ($catalog.counts.web_deck -lt 100 -or $catalog.counts.editable_pptx -lt 1) { throw 'Unified presentation catalog is incomplete' }
  $catalogSources = @($catalog.entries | ForEach-Object { $_.preferred.source } | Sort-Object -Unique)
  foreach ($requiredCatalogSource in @('ppt-master', 'baoyu-slide-deck', 'huashu-design')) {
    if ($catalogSources -notcontains $requiredCatalogSource) { throw "Unified presentation catalog is missing source: $requiredCatalogSource" }
  }
  Write-Output "presentation-catalog-unified web=$($catalog.counts.web_deck) editable=$($catalog.counts.editable_pptx)"
  Copy-Item -LiteralPath (Join-Path $root 'capabilities\presentation\probe.mjs') -Destination (Join-Path $scratch 'probe.mjs')
  $pptx = Join-Path $scratch 'behavior-probe.pptx'
  & $node (Join-Path $scratch 'probe.mjs') $pptx
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
  $renderDir = Join-Path $scratch 'rendered-pptx'
  & $node $renderPresentation --input $pptx --output_dir $renderDir --scale 1 | Out-Null
  if ($LASTEXITCODE -ne 0) { throw 'Latest presentation runtime render failed' }
  $renderedSlides = @(Get-ChildItem -LiteralPath $renderDir -Filter 'slide-*.png' -File)
  if ($renderedSlides.Count -ne 2) { throw "Expected 2 rendered slides, got $($renderedSlides.Count)" }
  Write-Output "pptx-rendered slides=$($renderedSlides.Count)"

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
  $playwrightPackage = Join-Path $nodeModules 'playwright'
  $playwrightCorePackage = Join-Path $nodeModules 'playwright-core'
  if (-not (Test-Path -LiteralPath (Join-Path $playwrightPackage 'package.json') -PathType Leaf) -or
      -not (Test-Path -LiteralPath (Join-Path $playwrightCorePackage 'package.json') -PathType Leaf)) {
    throw 'Playwright runtime is missing'
  }
  $env:NODE_PATH = $nodeModules
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
    distilled_contract_verified = $true
    candidate_matrix_verified = $true
    ppt_master_entrypoints_verified = $true
    catalog_web_deck = [int]$catalog.counts.web_deck
    catalog_editable_pptx = [int]$catalog.counts.editable_pptx
    pptx_slides = 2
    pptx_editable_shapes = [int]$shapeCount
    html_themes = $themeCount
    html_templates = $templateCount
    html_effects = $fxCount
    stage_fluid_rendered = $true
    dashiai_rendered = $true
		fusion_asset_contracts_verified = $true
		fusion_asset_ids = @($invocations | ForEach-Object { $_.Contract.asset_id })
		fusion_invocation_hashes = @($invocations | ForEach-Object { (Get-FileHash -Algorithm SHA256 -LiteralPath $_.Path).Hash.ToLowerInvariant() })
  }
  [IO.File]::WriteAllText($assertionsPath, ($assertions | ConvertTo-Json -Compress), [Text.UTF8Encoding]::new($false))
  $probeEvidence = @(
    (Add-ProbeArtifact 'assertions' $assertionsPath 'presentation-assertions.json')
    (Add-ProbeArtifact 'pptx' $pptx 'behavior-probe.pptx')
    (Add-ProbeArtifact 'html-render' $htmlShot 'html-ppt.png')
    (Add-ProbeArtifact 'fluid-render' $fluidShot 'fluid.png')
    (Add-ProbeArtifact 'dashiai-assertions' (Join-Path $evidenceDir 'dashiai-ppt-assertions.json') 'dashiai-ppt-assertions.json')
		( Add-ProbeArtifact 'editable-invocation' $invocations[0].Path $invocations[0].Evidence )
		( Add-ProbeArtifact 'web-invocation' $invocations[1].Path $invocations[1].Evidence )
		( Add-ProbeArtifact 'fluid-invocation' $invocations[2].Path $invocations[2].Evidence )
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
  if (Get-Variable previousRuntimeNodeModules -ErrorAction SilentlyContinue) { $env:RUNTIME_NODE_MODULES = $previousRuntimeNodeModules }
}
Write-Output $probeReceipt
