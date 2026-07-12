param(
  [ValidateSet('fast','full')]
  [string]$Mode = 'full'
)
$ErrorActionPreference = 'Stop'
$root = Split-Path $PSScriptRoot -Parent
$wuji = Join-Path $root 'bin\wuji.exe'
if (-not (Test-Path -LiteralPath $wuji)) { & (Join-Path $root 'scripts\build.ps1') }

function Get-SourceTreeHash([string]$Path) {
  $full = [IO.Path]::GetFullPath($Path).TrimEnd('\','/')
  $lines = @(Get-ChildItem -LiteralPath $full -Recurse -File -Force | Where-Object {
    $_.FullName -notmatch '\\.git\\'
  } | ForEach-Object {
    $relative = $_.FullName.Substring($full.Length).TrimStart('\','/') -replace '\\','/'
    $hash = (Get-FileHash -Algorithm SHA256 -LiteralPath $_.FullName).Hash.ToLowerInvariant()
    "$relative`t$hash"
  } | Sort-Object)
  $payload = ($lines -join "`n") + "`n"
  $sha = [Security.Cryptography.SHA256]::Create()
  try {
    return ([BitConverter]::ToString($sha.ComputeHash([Text.Encoding]::UTF8.GetBytes($payload))) -replace '-','').ToLowerInvariant()
  } finally { $sha.Dispose() }
}

$catalogPath = Join-Path $root 'capabilities\presentation\assets\template-catalog.json'
$catalogHashBefore = (Get-FileHash -Algorithm SHA256 -LiteralPath $catalogPath).Hash
$capabilities = if ($Mode -eq 'fast') {
  @('code','code-review','context','evolution','frontend','image','search','writing')
} else {
  @('all')
}
$verification = @()
foreach ($cap in $capabilities) {
  $chunk = & $wuji verify --capability $cap | ConvertFrom-Json
  if ($LASTEXITCODE -ne 0) { throw "fusion-audit failed while verifying $cap" }
  $verification += @($chunk)
}
if (@($verification | Where-Object { -not $_.passed }).Count -gt 0) {
  throw 'fusion-audit failed: a capability claim exceeds its evidence'
}
Write-Output "audit-mode=$Mode verified=$($verification.Count)"
$catalogHashAfter = (Get-FileHash -Algorithm SHA256 -LiteralPath $catalogPath).Hash
if ($catalogHashAfter -ne $catalogHashBefore) { throw 'fusion-audit failed: capability verification modified the presentation catalog' }

$skillBytes = (Get-Item -LiteralPath (Join-Path $root 'SKILL.md')).Length
$agentBytes = (Get-Item -LiteralPath (Join-Path $root 'AGENTS.md')).Length
$sourceFiles = Get-ChildItem -LiteralPath $root -Recurse -File | Where-Object {
  $_.FullName -notmatch '\\.git\\|\\.wuji\\|\\tools\\bin\\|\\bin\\'
}
$sourceBytes = ($sourceFiles | Measure-Object Length -Sum).Sum
if ($skillBytes -gt 6000 -or $agentBytes -gt 5000 -or $sourceBytes -gt 1000000) {
  throw "optimization-audit failed: skill=$skillBytes agents=$agentBytes source=$sourceBytes"
}
$nestedSkillFiles = Get-ChildItem (Join-Path $root 'capabilities') -Recurse -Filter 'SKILL.md' | Where-Object { $_.FullName -match '\\skills\\' }
if (@($nestedSkillFiles | Where-Object Length -gt 6000).Count -gt 0) {
  throw 'context-bloat-audit failed: a selected scenario Skill exceeds 6KB'
}
$manifestBytes = (Get-ChildItem (Join-Path $root 'capabilities') -Recurse -Filter 'manifest.json' | Measure-Object Length -Sum).Sum
if ($manifestBytes -gt 65536) { throw "context-bloat-audit failed: manifests=$manifestBytes" }
if (Test-Path -LiteralPath (Join-Path $root 'capabilities\nuwa')) {
  throw 'optimization-audit failed: Nuwa capability returned'
}
$legacyLedger = Get-Content -Raw -Encoding UTF8 -LiteralPath (Join-Path $root 'migration\legacy-verdict-ledger.json') | ConvertFrom-Json
$worktreeLedger = Get-Content -Raw -Encoding UTF8 -LiteralPath (Join-Path $root 'migration\legacy-worktree-ledger.json') | ConvertFrom-Json
$validLifecycle = @('known','doctrine-only','assets-retained','callable','behavior-verified','primary')
if ($legacyLedger.total -ne 104 -or @($legacyLedger.entries).Count -ne 104 -or @($legacyLedger.entries.object | Sort-Object -Unique).Count -ne 104) {
  throw 'fusion-audit failed: the complete 104-object legacy verdict ledger is missing or duplicated'
}
if (@($legacyLedger.entries | Where-Object { $_.status_2_0 -notin $validLifecycle }).Count -gt 0) {
  throw 'fusion-audit failed: legacy verdict ledger contains an invalid lifecycle claim'
}
if ($worktreeLedger.total -ne 52 -or @($worktreeLedger.entries).Count -ne 52 -or @($worktreeLedger.entries.path | Sort-Object -Unique).Count -ne 52) {
  throw 'fusion-audit failed: the complete legacy worktree ledger is missing or duplicated'
}
if (@($worktreeLedger.entries | Where-Object { $_.action_2_0 -notin @('distill','implement','exclude') }).Count -gt 0) {
  throw 'fusion-audit failed: legacy worktree ledger contains an invalid action'
}
if (@($worktreeLedger.entries | Where-Object { $_.path -match '(?i)(bridge|gui|shortcut)' -and $_.action_2_0 -eq 'implement' }).Count -gt 0) {
  throw 'fusion-audit failed: a legacy resident bridge or GUI was admitted into 2.0'
}
$sourceLock = Get-Content -Raw -Encoding UTF8 -LiteralPath (Join-Path $root 'sources.lock.json') | ConvertFrom-Json
$upstreamReview = Get-Content -Raw -Encoding UTF8 -LiteralPath (Join-Path $root 'migration\upstream-review.json') | ConvertFrom-Json
$expectedUpstreamDecisions = [ordered]@{
  'ppt-master' = 'update'
  'huashu-design' = 'update-and-distill'
  'frontend-slides' = 'update-cold-only'
  'open-design' = 'exclude-new-runtime'
}
if (@($upstreamReview.entries).Count -ne $expectedUpstreamDecisions.Count -or
    @($upstreamReview.entries.source | Sort-Object -Unique).Count -ne $expectedUpstreamDecisions.Count) {
  throw 'fusion-audit failed: upstream review must contain exactly four unique source decisions'
}
foreach ($sourceId in $expectedUpstreamDecisions.Keys) {
  $reviewRows = @($upstreamReview.entries | Where-Object source -eq $sourceId)
  $lockRows = @($sourceLock.sources | Where-Object id -eq $sourceId)
  if ($reviewRows.Count -ne 1 -or $lockRows.Count -ne 1 -or $reviewRows[0].decision -ne $expectedUpstreamDecisions[$sourceId]) {
    throw "fusion-audit failed: invalid upstream decision for $sourceId"
  }
  if ($sourceId -eq 'open-design') {
    if ($lockRows[0].commit -ne $reviewRows[0].locked -or $lockRows[0].commit -eq $reviewRows[0].reviewed_head) {
      throw 'fusion-audit failed: excluded Open Design runtime was admitted'
    }
  } elseif ($lockRows[0].commit -ne $reviewRows[0].reviewed_head) {
    throw "fusion-audit failed: reviewed upstream update was not locked for $sourceId"
  }
}
foreach ($source in $sourceLock.sources) {
  $sourcePath = & (Join-Path $PSScriptRoot 'expand-wuji-path.ps1') -PathValue $source.path -Root $root
  if (-not (Test-Path -LiteralPath $sourcePath)) { throw "fusion-audit failed: pinned cold source missing $($source.id) ($sourcePath)" }
  $gitMetadata = Join-Path $sourcePath '.git'
  $hasGitMetadata = (Test-Path -LiteralPath $gitMetadata -PathType Leaf) -or (Test-Path -LiteralPath (Join-Path $gitMetadata 'HEAD') -PathType Leaf)
  $actualCommit = ''
  if ($hasGitMetadata) {
    $gitOutput = @(& git -C $sourcePath rev-parse HEAD 2>$null)
    $gitExitCode = $LASTEXITCODE
    $commitLines = @($gitOutput | ForEach-Object { ([string]$_).Trim() } | Where-Object { $_ })
    if ($commitLines.Count -eq 1) { $actualCommit = $commitLines[0] }
  } else {
    $gitExitCode = -1
  }
  if ($gitExitCode -eq 0 -and $actualCommit) {
    if ($actualCommit -ne $source.commit) {
      throw "fusion-audit failed: pinned commit mismatch for $($source.id), expected $($source.commit), got $actualCommit"
    }
  } elseif ($source.tree_sha256) {
    $actualTreeHash = Get-SourceTreeHash $sourcePath
    if ($actualTreeHash -ne $source.tree_sha256) {
      throw "fusion-audit failed: source snapshot hash mismatch for $($source.id)"
    }
  } elseif (@($source.evidence_files).Count -gt 0) {
    foreach ($evidence in $source.evidence_files) {
      $relative = [string]$evidence.path
      if ([IO.Path]::IsPathRooted($relative) -or (($relative -split '[\\/]') -contains '..')) {
        throw "fusion-audit failed: unsafe evidence path for $($source.id): $relative"
      }
      if ($evidence.sha256 -notmatch '^[a-fA-F0-9]{64}$') {
        throw "fusion-audit failed: invalid evidence hash for $($source.id): $relative"
      }
      $evidencePath = Join-Path $sourcePath $relative
      if (-not (Test-Path -LiteralPath $evidencePath -PathType Leaf)) {
        throw "fusion-audit failed: source evidence missing for $($source.id): $relative"
      }
      $actualHash = (Get-FileHash -Algorithm SHA256 -LiteralPath $evidencePath).Hash.ToLowerInvariant()
      if ($actualHash -ne $evidence.sha256) {
        throw "fusion-audit failed: source evidence hash mismatch for $($source.id): $relative"
      }
    }
  } else {
    throw "fusion-audit failed: source lacks commit, tree, or file evidence $($source.id)"
  }
}
foreach ($toolchain in $sourceLock.toolchains) {
  $toolchainPath = & (Join-Path $PSScriptRoot 'expand-wuji-path.ps1') -PathValue $toolchain.path -Root $root
  if (-not (Test-Path -LiteralPath $toolchainPath)) { throw "fusion-audit failed: toolchain missing $($toolchain.id)" }
  $versionText = & $toolchainPath version 2>&1
  if (($versionText -join ' ') -notmatch [regex]::Escape($toolchain.version)) { throw "fusion-audit failed: wrong $($toolchain.id) version" }
}
foreach ($tool in $sourceLock.context_tools) {
  if ($tool.sha256 -notmatch '^[a-fA-F0-9]{64}$' -or -not $tool.url.StartsWith('https://')) {
    throw "fusion-audit failed: invalid locked context tool metadata $($tool.id)"
  }
  $toolPath = Join-Path $root $tool.path
  if (-not (Test-Path -LiteralPath $toolPath)) {
    if ($tool.optional) { continue }
    throw "fusion-audit failed: required context tool missing $($tool.id)"
  }
  $toolVersion = & $toolPath @($tool.version_args) 2>&1
  if ($LASTEXITCODE -ne 0 -or ($toolVersion -join ' ') -notmatch [regex]::Escape($tool.version)) {
    throw "fusion-audit failed: wrong installed context tool version $($tool.id)"
  }
}
$searchManifest = Get-Content -Raw -Encoding UTF8 -LiteralPath (Join-Path $root 'capabilities\search\manifest.json')
if ($searchManifest -match 'agnes') { throw 'optimization-audit failed: Agnes returned to search routing' }
$imageManifest = Get-Content -Raw -Encoding UTF8 -LiteralPath (Join-Path $root 'capabilities\image\manifest.json') | ConvertFrom-Json
$imageLifecycleRank = @{'known'=0;'doctrine-only'=1;'assets-retained'=2;'callable'=3;'behavior-verified'=4;'primary'=5}
if ($imageManifest.probe.fixture -match '^xiaobai-image2-' -and $imageLifecycleRank[[string]$imageManifest.status] -gt $imageLifecycleRank.callable) {
  throw 'fusion-audit failed: a provider-only Xiaobai probe promoted the whole image capability'
}
if ($imageManifest.probe.kind -eq 'smoke' -and $imageLifecycleRank[[string]$imageManifest.status] -gt $imageLifecycleRank.callable) {
  throw 'fusion-audit failed: an image smoke probe was presented as behavioral fusion'
}

$runtimeTextExtensions = @('.ps1','.go','.json','.md','.yaml','.yml','.js','.cjs','.mjs','.ts','.vue','.css','.html')
$runtimeTextFiles = @($sourceFiles | Where-Object {
  $_.Extension.ToLowerInvariant() -in $runtimeTextExtensions -and
  $_.FullName -notmatch '\\migration\\' -and
  $_.FullName -ne (Join-Path $root 'AGENTS.md') -and
  $_.FullName -ne (Join-Path $root 'scripts\audit.ps1') -and
  $_.FullName -ne (Join-Path $root 'scripts\build-legacy-ledger.ps1')
})
$legacyRuntimePattern = '(?i)(?:E:[\\/]+wuji-projects[\\/]+wuji-legion-codex(?:[\\/]|$)|\$\{WUJI_PROJECTS\}/slidev-demo|PPT的image2美化|smoke-fluid-(?:reference|script|style))'
$hardcodedUserPattern = '(?i)(?:[A-Z]:[\\/]+Users[\\/]+[^$\\/{]+[\\/]|AppData[\\/]+(?:Local|Roaming)[\\/])'
$residentXiaobaiPattern = '(?i)(?:xiaobai.{0,40}(?:bridge|gui|daemon|service|shortcut)|(?:bridge|gui|daemon|service|shortcut).{0,40}xiaobai)'
$secretPatterns = @(
  '(?i)\bsk-[A-Za-z0-9_-]{16,}\b',
  '\bgh[pousr]_[A-Za-z0-9]{20,}\b',
  '\bAKIA[0-9A-Z]{16}\b',
  '(?i)\bBearer\s+[A-Za-z0-9._-]{24,}\b'
)
foreach ($file in $runtimeTextFiles) {
  $text = [IO.File]::ReadAllText($file.FullName, [Text.Encoding]::UTF8)
  if ($text -match $legacyRuntimePattern) { throw "fusion-audit failed: active runtime points to a retired legacy source ($($file.FullName))" }
  if ($text -match $hardcodedUserPattern) { throw "optimization-audit failed: hardcoded user directory found ($($file.FullName))" }
  if ($file.Extension -ne '.md' -and $text -match $residentXiaobaiPattern) { throw "optimization-audit failed: resident Xiaobai control surface found ($($file.FullName))" }
  foreach ($secretPattern in $secretPatterns) {
    if ($text -match $secretPattern) { throw "security-audit failed: credential-like literal found ($($file.FullName))" }
  }
}

$context = & $wuji context-select --workspace $root --query 'capability behavior verification context budget' --max-bytes 4096 | ConvertFrom-Json
if ($LASTEXITCODE -ne 0 -or $context.selected_bytes -gt 4096 -or $context.excerpts.Count -lt 1) {
  throw 'context-bloat-audit failed'
}

$activeRoot = Join-Path $env:USERPROFILE '.agents\skills'
$activeWuji = @(Get-ChildItem -LiteralPath $activeRoot -Directory -ErrorAction SilentlyContinue | Where-Object { $_.Name -like 'wuji-legion*' })
if ($activeWuji.Count -ne 1 -or $activeWuji[0].Name -ne 'wuji-legion-codex-2-0') {
  throw "optimization-audit failed: expected one active Wuji entry, found $($activeWuji.Name -join ',')"
}
$activeNuwa = @(Get-ChildItem -LiteralPath $activeRoot -Directory -ErrorAction SilentlyContinue | Where-Object { $_.Name -match 'nuwa' })
if ($activeNuwa.Count -ne 0) { throw "optimization-audit failed: active Nuwa entries found $($activeNuwa.Name -join ',')" }

$webRoute = & $wuji route --query 'build a Slidev web presentation with stage fluid' | ConvertFrom-Json
$pptxRoute = & $wuji route --query 'create an editable PPTX' | ConvertFrom-Json
$writingRoute = & $wuji route --query 'translate this article' | ConvertFrom-Json
$searchRoute = & $wuji route --query 'search the web for the latest solution' | ConvertFrom-Json
$codeRoute = & $wuji route --query 'fix the code and verify it' | ConvertFrom-Json
if ($webRoute.primary_skill -ne 'wuji-web-deck' -or $webRoute.engine -ne 'web-deck' -or ($webRoute.mounted_sources.id -contains 'ppt-master-complete')) { throw 'fusion-audit failed: web presentation scenario is not consolidated' }
if ($pptxRoute.primary_skill -ne 'wuji-editable-deck' -or $pptxRoute.engine -ne 'editable-pptx' -or ($pptxRoute.mounted_sources.id -contains 'slidev-runtime-complete')) { throw 'fusion-audit failed: editable presentation scenario is not consolidated' }
if ($writingRoute.primary_skill -ne 'wuji-writing-suite' -or $writingRoute.engine -ne 'translation') { throw 'fusion-audit failed: writing suite leaked source selection' }
if ($searchRoute.provider -ne 'default-gpt-search' -or @($searchRoute.workers | Where-Object model_class -eq 'agnes').Count -gt 0) { throw 'optimization-audit failed: Agnes returned to search' }
if (@($searchRoute.workers).Count -ne 3 -or @($searchRoute.workers | Where-Object model -ne 'gpt-5.6-luna').Count -ne 0) { throw 'optimization-audit failed: research workers lost Luna model assignment' }
if (@($searchRoute.workers | Where-Object { ($_.fallback_models -join ',') -ne 'gpt-5.6-terra,gpt-5.6-sol' }).Count -ne 0) { throw 'optimization-audit failed: research worker fallback order changed' }
if (@($codeRoute.workers).Count -ne 2 -or @($codeRoute.workers | Where-Object model -ne 'gpt-5.6-terra').Count -ne 0) { throw 'optimization-audit failed: code workers lost Terra model assignment' }
if (@($codeRoute.workers | Where-Object { ($_.fallback_models -join ',') -ne 'gpt-5.6-luna,gpt-5.6-sol' }).Count -ne 0) { throw 'optimization-audit failed: code worker fallback order changed' }
if ($codeRoute.model_policy.class_models.terra -ne 'gpt-5.6-terra' -or $codeRoute.model_policy.class_models.luna -ne 'gpt-5.6-luna') { throw 'optimization-audit failed: executable model policy is incomplete' }

[pscustomobject]@{
  fusion_audit = 'pass'
  optimization_audit = 'pass'
  context_bloat_audit = 'pass'
  active_router_count = $activeWuji.Count
  skill_bytes = $skillBytes
  agents_bytes = $agentBytes
  source_bytes = $sourceBytes
  context_selected_bytes = $context.selected_bytes
  legacy_objects_reclassified = $legacyLedger.total
  legacy_worktree_paths_reclassified = $worktreeLedger.total
  upstream_decisions_verified = @($upstreamReview.entries).Count
  pinned_cold_sources = @($sourceLock.sources).Count
  pinned_context_tools = @($sourceLock.context_tools).Count
  nested_skill_count = @($nestedSkillFiles).Count
  manifest_bytes = $manifestBytes
} | ConvertTo-Json
