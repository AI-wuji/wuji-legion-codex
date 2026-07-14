param(
  [string]$LegacyMatrix = '',
  [string]$LegacyRoot = '',
  [string]$VerdictOutput = '',
  [string]$WorktreeOutput = ''
)
$ErrorActionPreference = 'Stop'
$root = Split-Path $PSScriptRoot -Parent
if (-not $LegacyRoot) { $LegacyRoot = [IO.Path]::GetFullPath((Join-Path $root '..\wuji-legion-codex')) }
if (-not $LegacyMatrix) { $LegacyMatrix = Join-Path $LegacyRoot 'fusion-matrix.json' }
if (-not $VerdictOutput) { $VerdictOutput = Join-Path $root 'migration\legacy-verdict-ledger.json' }
if (-not $WorktreeOutput) { $WorktreeOutput = Join-Path $root 'migration\legacy-worktree-ledger.json' }
if (-not (Test-Path -LiteralPath $LegacyMatrix -PathType Leaf)) { throw "Legacy matrix not found: $LegacyMatrix" }
if (-not (Test-Path -LiteralPath (Join-Path $LegacyRoot '.git'))) { throw "Legacy repository not found: $LegacyRoot" }

$legacy = Get-Content -Raw -Encoding UTF8 -LiteralPath $LegacyMatrix | ConvertFrom-Json

$behavior = @{
  'rtk'='context'; 'huashu-design'='presentation'; 'ppt-master'='presentation';
  'html-ppt-skill'='presentation'; 'ppt-keynote'='presentation';
  'humanize-ppt'='presentation'; 'open-code-review'='code-review'
}
$callable = @{
  'hyperframes'='video'; 'humanizer-zh'='writing'; 'khazix-writer'='writing';
  'baoyu-article-illustrator'='visual'; 'baoyu-translate'='writing'; 'taste-skill'='visual';
  'baoyu-electron-extract'='search'; 'baoyu-url-to-markdown'='search';
  'baoyu-youtube-transcript'='search'; 'baoyu-post-to-x/wechat/weibo'='writing';
  'baoyu-image-gen'='image'; 'codebase-memory-mcp'='context';
  'dynamic-electronic-board-sources'='visual'; 'Agnes AI'='image+video'
}
$assets = @{
  'baoyu-skills'='multi-capability'; 'frontend-slides'='presentation'
}
$updatedSources = @('ppt-master','huashu-design','frontend-slides')

$entries = foreach ($item in @($legacy.object_verdicts)) {
  $name = [string]$item.object
  $legacyStatus = [string]$item.runtime_status
  $status = 'known'
  $action = 'exclude'
  $capability = 'excluded'
  $evidence = @('legacy fusion-matrix rationale and boundary')
  $ruling = 'Excluded because it does not fill a current 2.0 capability gap.'

  if ($behavior.ContainsKey($name) -or $callable.ContainsKey($name)) {
    $status = 'assets-retained'
    $action = 'retain-cold'
    $capability = 'cold-reference'
    $evidence = @('sources.lock.json','references/source-execution-status.md')
    $ruling = 'The scenario adapter may be verified separately; this upstream object remains cold until it has its own callable adapter and behavior evidence.'
  } elseif ($assets.ContainsKey($name)) {
    $status = 'assets-retained'
    $action = if ($updatedSources -contains $name) { 'update' } else { 'retain-cold' }
    $capability = $assets[$name]
    $evidence = @('sources.lock.json')
    $ruling = 'Retained as a cold evidence package and not exposed as a second user-facing workflow.'
  } elseif ($legacyStatus -eq 'landed') {
    $status = 'doctrine-only'
    $action = 'distill'
    $capability = 'core-doctrine'
    $evidence = @('SKILL.md','references/architecture.md')
    $ruling = 'Only the useful rule survived; the legacy landed label had no complete callable package.'
  } elseif ($legacyStatus -eq 'restricted-boundary') {
    $status = 'doctrine-only'
    $action = 'distill'
    $capability = 'boundary-doctrine'
    $evidence = @('SKILL.md','references/capability-contract.md')
    $ruling = 'The restriction remains doctrine; no restricted runtime or external write surface was admitted.'
  } elseif ($legacyStatus -eq 'source-pool-only' -and [string]$item.fusion_mode -eq 'gap-fill') {
    $status = 'doctrine-only'
    $action = 'distill'
    $capability = 'core-doctrine'
    $evidence = @('SKILL.md','references/architecture.md')
    $ruling = 'Useful execution doctrine was distilled without importing the upstream runtime.'
  }

  [pscustomobject][ordered]@{
    object = $name
    legacy_status = $legacyStatus
    legacy_mode = [string]$item.fusion_mode
    action_2_0 = $action
    status_2_0 = $status
    capability = $capability
    legacy_surfaces = @($item.landed_surfaces)
    boundary = [string]$item.boundary
    evidence = $evidence
    ruling = $ruling
  }
}

$verdictDocument = [ordered]@{
  version = '2.0'
  generated_from = 'legacy fusion-matrix object_verdicts'
  rule = 'Every legacy object is re-decided under 2.0; no legacy landed label or runtime shell survives by default.'
  total = @($entries).Count
  legacy_counts = [ordered]@{}
  action_counts = [ordered]@{}
  lifecycle_counts = [ordered]@{}
  entries = @($entries | Sort-Object object)
}
foreach ($group in @($entries | Group-Object legacy_status | Sort-Object Name)) { $verdictDocument.legacy_counts[$group.Name] = $group.Count }
foreach ($group in @($entries | Group-Object action_2_0 | Sort-Object Name)) { $verdictDocument.action_counts[$group.Name] = $group.Count }
foreach ($group in @($entries | Group-Object status_2_0 | Sort-Object Name)) { $verdictDocument.lifecycle_counts[$group.Name.Replace('-','_')] = $group.Count }

$statusLines = @(& git -C $LegacyRoot -c core.quotepath=false status --porcelain=v1 --untracked-files=all)
if ($LASTEXITCODE -ne 0) { throw 'Unable to read legacy worktree status' }
$worktreeEntries = foreach ($line in $statusLines) {
  if ([string]::IsNullOrWhiteSpace($line) -or $line.Length -lt 4) { continue }
  $state = $line.Substring(0,2)
  $path = $line.Substring(3).Trim()
  $action = 'distill'
  $target = '2.0 core manifests, Skills, tests, and references'
  $ruling = 'Superseded by the smaller 2.0 kernel; only current behavior and doctrine are retained.'

  if ($path -match '(?i)xiaobai-image2-generate\.ps1$') {
    $action = 'implement'; $target = 'capabilities/image/providers/xiaobai-image2'
    $ruling = 'Reimplemented as a task-scoped direct client with no resident bridge or stored credential.'
  } elseif ($path -match '(?i)xiaobai-image2-bridge\.md$') {
    $action = 'distill'; $target = 'capabilities/image/providers/xiaobai-image2/README.md'
    $ruling = 'Operational contract retained; local bridge and GUI instructions removed.'
  } elseif ($path -match '(?i)(xiaobai-image2-bridge|xiaobai-image2-gui|start-xiaobai|stop-xiaobai|create-xiaobai|inspect-xiaobai|test-xiaobai)') {
    $action = 'exclude'; $target = 'none'
    $ruling = 'Resident server, GUI, shortcut, and compatibility endpoints would create a second control surface.'
  } elseif ($path -match '(?i)push-to-github\.ps1$') {
    $action = 'exclude'; $target = 'none'
    $ruling = 'Repository publishing is an external write and is not a standing system capability.'
  } elseif ($path -match '(?i)(wuji-ppt-elite-showcase|wuji-ppt-motion-carrier)') {
    $action = 'distill'; $target = 'scripts/verify-presentation.ps1'
    $ruling = 'Useful rendering checks are represented by the bounded presentation behavior probe; showcase-only carriers are not runtime features.'
  } elseif ($path -match '^(experts|units)/') {
    $action = 'distill'; $target = 'capabilities/*/skills and references/architecture.md'
    $ruling = 'Long-lived personas and units were collapsed into cold scenario Skills and compact doctrine.'
  }

  [pscustomobject][ordered]@{
    path = $path.Replace('\\','/')
    legacy_state = $state
    action_2_0 = $action
    target_2_0 = $target
    ruling = $ruling
  }
}
$worktreeDocument = [ordered]@{
  version = '2.0'
  generated_from = 'legacy git worktree status'
  rule = 'Uncommitted legacy work is evidence, not a patch queue. Each path is distilled, reimplemented, or excluded.'
  total = @($worktreeEntries).Count
  action_counts = [ordered]@{}
  entries = @($worktreeEntries | Sort-Object path)
}
foreach ($group in @($worktreeEntries | Group-Object action_2_0 | Sort-Object Name)) { $worktreeDocument.action_counts[$group.Name] = $group.Count }

New-Item -ItemType Directory -Force -Path (Split-Path $VerdictOutput -Parent),(Split-Path $WorktreeOutput -Parent) | Out-Null
[IO.File]::WriteAllText($VerdictOutput, ($verdictDocument | ConvertTo-Json -Depth 10 -Compress), [Text.UTF8Encoding]::new($false))
[IO.File]::WriteAllText($WorktreeOutput, ($worktreeDocument | ConvertTo-Json -Depth 8 -Compress), [Text.UTF8Encoding]::new($false))
Write-Output $VerdictOutput
Write-Output $WorktreeOutput
