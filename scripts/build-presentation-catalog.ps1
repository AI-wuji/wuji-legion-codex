param([string]$Output = '')
$ErrorActionPreference = 'Stop'
$root = Split-Path $PSScriptRoot -Parent
if (-not $Output) { $Output = Join-Path $root 'capabilities\presentation\assets\template-catalog.json' }
$sourceLock = Get-Content -Raw -Encoding UTF8 -LiteralPath (Join-Path $root 'sources.lock.json') | ConvertFrom-Json
function Expand-WujiPath([string]$PathValue) {
  return & (Join-Path $PSScriptRoot 'expand-wuji-path.ps1') -PathValue $PathValue -Root $root
}
function Compress-WujiPath([string]$PathValue) {
  $full = [IO.Path]::GetFullPath($PathValue)
  $rootFull = [IO.Path]::GetFullPath($root)
  if ($full.Equals($rootFull, [StringComparison]::OrdinalIgnoreCase) -or $full.StartsWith($rootFull + '\', [StringComparison]::OrdinalIgnoreCase)) {
    $rel = $full.Substring($rootFull.Length).TrimStart('\','/') -replace '\\','/'
    return '${ROOT}/' + $rel
  }
  $projects = $env:WUJI_PROJECTS
  if (-not $projects) { $projects = [IO.Path]::GetFullPath((Join-Path $root '..')) }
  $projectsFull = [IO.Path]::GetFullPath($projects)
  if ($full.StartsWith($projectsFull, [StringComparison]::OrdinalIgnoreCase)) {
    $rel = $full.Substring($projectsFull.Length).TrimStart('\','/') -replace '\\','/'
    return '${WUJI_PROJECTS}/' + $rel
  }
  $user = $env:USERPROFILE
  if ($user) {
    $userFull = [IO.Path]::GetFullPath($user)
    if ($full.StartsWith($userFull, [StringComparison]::OrdinalIgnoreCase)) {
      $rel = $full.Substring($userFull.Length).TrimStart('\','/') -replace '\\','/'
      return '${USERPROFILE}/' + $rel
    }
  }
  return ($full -replace '\\','/')
}
function Get-LockedSourcePath([string]$Id) {
  $matches = @($sourceLock.sources | Where-Object id -eq $Id)
  if ($matches.Count -ne 1) { throw "sources.lock.json must contain exactly one $Id source" }
  $path = Expand-WujiPath $matches[0].path
  if (-not (Test-Path -LiteralPath $path -PathType Container)) { throw "Locked source is missing: $Id ($path)" }
  return $path
}
$rows = @()
function Add-FileRows([string]$Scenario,[string]$Category,[string]$Source,[string]$Base,[string]$Filter,[switch]$Directory) {
  if (-not (Test-Path -LiteralPath $Base)) { return }
  $items = if ($Directory) { Get-ChildItem -LiteralPath $Base -Directory -Filter $Filter } else { Get-ChildItem -LiteralPath $Base -File -Filter $Filter }
  foreach ($item in $items) {
    $id = ($item.BaseName.ToLowerInvariant() -replace '[^a-z0-9]+','-').Trim('-')
    if (-not $id) { $id = ($item.Name.ToLowerInvariant() -replace '[^a-z0-9]+','-').Trim('-') }
    $script:rows += [pscustomobject]@{scenario=$Scenario;category=$Category;id=$id;source=$Source;path=$item.FullName}
  }
}

$html = Get-LockedSourcePath 'html-ppt-skill'
Add-FileRows 'web-deck' 'theme' 'html-ppt' (Join-Path $html 'assets\themes') '*.css'
Add-FileRows 'web-deck' 'layout' 'html-ppt' (Join-Path $html 'templates\single-page') '*.html'
Add-FileRows 'web-deck' 'full-deck' 'html-ppt' (Join-Path $html 'templates\full-decks') '*' -Directory
Add-FileRows 'web-deck' 'effect' 'html-ppt' (Join-Path $html 'assets\animations\fx') '*.js'

$slidev = Get-LockedSourcePath 'slidev-runtime'
Add-FileRows 'web-deck' 'component' 'slidev' (Join-Path $slidev 'components') '*.vue'
$rows += [pscustomobject]@{scenario='web-deck';category='component';id='stage-fluid';source='wuji-stage-fluid';path=(Join-Path $root 'scripts\materialize-stage-fluid.ps1')}

$presentations = Get-ChildItem (Join-Path $env:USERPROFILE '.codex\plugins\cache\openai-primary-runtime\presentations') -Directory | Sort-Object -Property @{ Expression = {
  try { [version]$_.Name } catch { [version]'0.0' }
}; Descending = $true } | Select-Object -First 1
if ($presentations) {
  $compose = Join-Path $presentations.FullName 'skills\presentations\assets\builtin_templates\codex-grid-layout-library\artifact-tool-compose'
  Add-FileRows 'editable-pptx' 'layout' 'openai-presentations' $compose 'slide-*.mjs'
}
$pptMaster = Join-Path (Get-LockedSourcePath 'ppt-master') 'examples'
Add-FileRows 'editable-pptx' 'example' 'ppt-master' $pptMaster 'ppt169_*' -Directory
$baoyuStyles = Join-Path (Get-LockedSourcePath 'baoyu-skills') 'skills\baoyu-slide-deck\references\styles'
Add-FileRows 'editable-pptx' 'style' 'baoyu-slide-deck' $baoyuStyles '*.md'
$huashu = Join-Path (Get-LockedSourcePath 'huashu-design') 'assets\showcases\ppt'
Add-FileRows 'editable-pptx' 'showcase' 'huashu-design' $huashu '*'
$elite = Join-Path $env:USERPROFILE '.agents\skills\elite-powerpoint-designer\templates'
Add-FileRows 'editable-pptx' 'template' 'elite-powerpoint' $elite '*'

$groups = @($rows | Group-Object scenario,category,id | ForEach-Object {
  $first = $_.Group | Select-Object -First 1
  [ordered]@{
    scenario=$first.scenario
    engine=$first.scenario
    category=$first.category
    id=$first.id
    asset_id=('presentation:' + $first.scenario + ':' + $first.category + ':' + $first.id)
    preferred=[ordered]@{source=$first.source;source_version='retained';path=(Compress-WujiPath $first.path)}
    variants=@($_.Group | ForEach-Object { [ordered]@{source=$_.source; source_version='retained'; path=(Compress-WujiPath $_.path)} })
  }
} | Sort-Object scenario,category,id)
$doc = [ordered]@{
  version='2.1'
  rule='One scenario catalog; duplicate normalized ids are grouped as variants instead of parallel templates and each preferred asset has a stable capability-qualified identifier.'
  generator='scripts/build-presentation-catalog.ps1'
  counts=[ordered]@{web_deck=@($groups | Where-Object scenario -eq 'web-deck').Count;editable_pptx=@($groups | Where-Object scenario -eq 'editable-pptx').Count}
  entries=$groups
}
New-Item -ItemType Directory -Force -Path (Split-Path $Output -Parent) | Out-Null
[IO.File]::WriteAllText($Output, ($doc | ConvertTo-Json -Depth 8), [Text.UTF8Encoding]::new($false))
Write-Output $Output
