param(
  [ValidateSet('fast','full')]
  [string]$Mode = 'full'
)
$ErrorActionPreference = 'Stop'
$utf8 = [Text.UTF8Encoding]::new($false)
[Console]::OutputEncoding = $utf8
$OutputEncoding = $utf8
$root = Split-Path $PSScriptRoot -Parent
$wuji = Join-Path $root 'bin\wuji.exe'
& (Join-Path $root 'scripts\build.ps1') | Out-Null
if ($LASTEXITCODE -ne 0 -or -not (Test-Path -LiteralPath $wuji)) { throw 'fusion-audit failed: could not build the audit binary' }
$goTemp = Join-Path $root '.wuji\go-test-tmp'
New-Item -ItemType Directory -Force -Path $goTemp | Out-Null
$env:GOTMPDIR = $goTemp
$go = & (Join-Path $root 'scripts\resolve-locked-go.ps1') -Root $root
& $go test -p 1 -count=1 ./internal/core ./cmd/wuji | Out-Null
if ($LASTEXITCODE -ne 0) { throw 'fusion-audit failed: automatic source semantic routing regression' }

function Invoke-WujiJson([string[]]$CliArgs) {
  $raw = @(& $wuji @CliArgs 2>&1)
  if ($LASTEXITCODE -ne 0) { throw "wuji command failed: $($CliArgs -join ' ')" }
  try {
    return (($raw -join [Environment]::NewLine) | ConvertFrom-Json -ErrorAction Stop)
  } catch {
    throw "wuji emitted invalid JSON: $($CliArgs -join ' ')"
  }
}

function Test-OfficeCliBehavior {
  $evidenceDir = Join-Path $root ('.wuji\officecli-audit-' + [guid]::NewGuid().ToString('N'))
  New-Item -ItemType Directory -Force -Path $evidenceDir | Out-Null
  $probe = Join-Path $root 'scripts\verify-officecli.ps1'
  $probeStderr = Join-Path $evidenceDir 'probe.stderr.log'
  try {
    $priorErrorAction = $ErrorActionPreference
    $ErrorActionPreference = 'Continue'
    $raw = @(& powershell -NoProfile -ExecutionPolicy Bypass -File $probe -Root $root -EvidenceDir $evidenceDir 2> $probeStderr)
    $probeExitCode = $LASTEXITCODE
    $ErrorActionPreference = $priorErrorAction
    $stderr = if (Test-Path -LiteralPath $probeStderr) { Get-Content -Raw -LiteralPath $probeStderr } else { '' }
    if ($probeExitCode -ne 0) { throw "OfficeCLI behavior probe failed: $($raw -join [Environment]::NewLine)$stderr" }
    try {
      $receipt = (($raw -join [Environment]::NewLine) | ConvertFrom-Json -ErrorAction Stop)
    } catch {
      throw 'OfficeCLI behavior probe did not emit a valid JSON receipt'
    }
    $expected = @(
      'officecli-assertions.json',
      'officecli-probe.docx.html',
      'officecli-probe.docx.json',
      'officecli-probe.xlsx.html',
      'officecli-probe.xlsx.json'
    )
    if ($receipt.wuji_probe -ne 'behavior' -or $receipt.fixture -ne 'officecli-stateless-v1' -or -not $receipt.passed -or $receipt.evidence_dir -ne $evidenceDir) {
      throw 'OfficeCLI behavior probe receipt is incomplete'
    }
    $artifacts = @($receipt.evidence)
    $actualNames = @($artifacts | ForEach-Object { [string]$_.path } | Sort-Object -Unique)
    $expectedNames = @($expected | Sort-Object)
    if ($artifacts.Count -ne $expectedNames.Count -or $actualNames.Count -ne $expectedNames.Count -or (($actualNames -join ',') -ne ($expectedNames -join ','))) {
      throw 'OfficeCLI behavior probe evidence set is incomplete'
    }
    foreach ($artifact in $artifacts) {
      $name = [string]$artifact.path
      if ([IO.Path]::GetFileName($name) -ne $name -or $artifact.sha256 -notmatch '^[a-fA-F0-9]{64}$') {
        throw "OfficeCLI behavior probe emitted unsafe evidence metadata: $name"
      }
      $path = Join-Path $evidenceDir $name
      if (-not (Test-Path -LiteralPath $path -PathType Leaf)) { throw "OfficeCLI behavior evidence is missing: $name" }
      $actual = (Get-FileHash -Algorithm SHA256 -LiteralPath $path).Hash.ToLowerInvariant()
      if ($actual -ne ([string]$artifact.sha256).ToLowerInvariant()) { throw "OfficeCLI behavior evidence hash mismatch: $name" }
    }
    return $receipt
  } finally {
    Remove-Item -LiteralPath $evidenceDir -Recurse -Force -ErrorAction SilentlyContinue
  }
}

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
  @('code','code-review','context','data','documents','evolution','feishu','frontend','image','knowledge','search','security','video','visual','writing')
} else {
  @('all')
}
$verification = @()
foreach ($cap in $capabilities) {
  $chunk = Invoke-WujiJson @('verify', '--capability', $cap)
  $verification += @($chunk)
}
if (@($verification | Where-Object { -not $_.passed }).Count -gt 0) {
  throw 'fusion-audit failed: a capability claim exceeds its evidence'
}
foreach ($result in @($verification | Where-Object { $_.effective_status -in @('behavior-verified','primary') })) {
  if (-not $result.probe_evidence -or @($result.probe_evidence.evidence).Count -lt 1) {
    throw "fusion-audit failed: behavioral capability lacks verified artifact evidence $($result.capability)"
  }
}
foreach ($result in @($verification | Where-Object { $_.claimed_status -eq 'primary' })) {
  if ($result.effective_status -ne 'primary') {
    throw "fusion-audit failed: primary capability lacks verified promotion evidence $($result.capability)"
  }
}
Write-Output "audit-mode=$Mode verified=$($verification.Count)"
$sourceAudit = @(Invoke-WujiJson @('source-audit'))
if (@($sourceAudit | Where-Object { $_.state -eq 'unavailable' }).Count -gt 0) {
  throw 'fusion-audit failed: retained source is unavailable'
}
if (@($sourceAudit | Where-Object {
  $_.state -eq 'auto-selectable' -and
  ([string]::IsNullOrWhiteSpace($_.entrypoint) -or $_.lifecycle -notin @('callable','behavior-verified','primary'))
}).Count -gt 0) {
  throw 'fusion-audit failed: automatic source lacks an executable routing contract'
}
if (@($sourceAudit | Where-Object {
  $_.state -eq 'cold-retained' -and $_.lifecycle -in @('callable','behavior-verified','primary')
}).Count -gt 0) {
  throw 'fusion-audit failed: callable source was silently demoted to cold retention'
}
$officeCliBehavior = Test-OfficeCliBehavior
# The presentation capability verifier already runs the DashiAI behavior
# probe inside a fresh evidence directory and independently hashes its
# assertion artifact. Do not invoke that probe a second time without the
# verifier-owned evidence environment.
$catalogHashAfter = (Get-FileHash -Algorithm SHA256 -LiteralPath $catalogPath).Hash
if ($catalogHashAfter -ne $catalogHashBefore) { throw 'fusion-audit failed: capability verification modified the presentation catalog' }

$skillBytes = (Get-Item -LiteralPath (Join-Path $root 'SKILL.md')).Length
$agentBytes = (Get-Item -LiteralPath (Join-Path $root 'AGENTS.md')).Length
$sourceFiles = Get-ChildItem -LiteralPath $root -Recurse -File | Where-Object {
  $_.FullName -notmatch '\\.git\\|\\.wuji\\|\\tools\\bin\\|\\bin\\'
}
$sourceBytes = ($sourceFiles | Measure-Object Length -Sum).Sum
$maxSourceBytes = 1572864
if ($skillBytes -gt 6000 -or $agentBytes -gt 5000 -or $sourceBytes -gt $maxSourceBytes) {
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
foreach ($tool in @($sourceLock.office_tools)) {
  if ($tool.id -ne 'officecli' -or $tool.version -notmatch '^\d+\.\d+\.\d+$' -or $tool.sha256 -notmatch '^[a-fA-F0-9]{64}$' -or -not $tool.url.StartsWith('https://') -or $tool.license -ne 'Apache-2.0') {
    throw 'fusion-audit failed: invalid locked OfficeCLI metadata'
  }
  $toolPath = & (Join-Path $PSScriptRoot 'expand-wuji-path.ps1') -PathValue $tool.destination -Root $root
  if (-not (Test-Path -LiteralPath $toolPath -PathType Leaf)) { throw 'fusion-audit failed: verified OfficeCLI binary is missing' }
  $actualHash = (Get-FileHash -Algorithm SHA256 -LiteralPath $toolPath).Hash.ToLowerInvariant()
  if ($actualHash -ne $tool.sha256.ToLowerInvariant()) { throw 'fusion-audit failed: installed OfficeCLI checksum mismatch' }
  $toolVersion = @(& $toolPath --version 2>&1) -join ' '
  if ($LASTEXITCODE -ne 0 -or $toolVersion -notmatch [regex]::Escape($tool.version)) { throw 'fusion-audit failed: wrong installed OfficeCLI version' }
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

$context = Invoke-WujiJson @('context-select', '--workspace', $root, '--query', 'capability behavior verification context budget', '--max-bytes', '4096')
if ($LASTEXITCODE -ne 0 -or $context.selected_bytes -gt 4096 -or $context.excerpts.Count -lt 1) {
  throw 'context-bloat-audit failed'
}
if (-not $context.query_fingerprint -or -not $context.content_sha256 -or $context.context_handle -ne "wuji-context://sha256/$($context.content_sha256)" -or -not (Test-Path -LiteralPath $context.artifact_path -PathType Leaf)) {
  throw 'context-bloat-audit failed: content-addressed artifact contract is incomplete'
}
if ($context.retrieval_mode -notlike 'workspace-graph*' -or $context.graph_lookups -lt 1 -or $context.indexed_files -lt 1 -or $context.candidate_files -ge $context.indexed_files) {
  throw 'context-bloat-audit failed: workspace graph did not narrow candidates before source reads'
}
$graph = Invoke-WujiJson @('graph-sync', '--workspace', $root)
if ($LASTEXITCODE -ne 0 -or $graph.max_terms_per_file -ne 512 -or $graph.max_refs_per_term -ne 256 -or $graph.max_lookups -ne 64) {
  throw 'context-bloat-audit failed: workspace graph hard limits are missing'
}

$activeRoot = Join-Path $env:USERPROFILE '.agents\skills'
$activeWuji = @(Get-ChildItem -LiteralPath $activeRoot -Directory -ErrorAction SilentlyContinue | Where-Object { $_.Name -like 'wuji-legion*' })
if ($activeWuji.Count -ne 1 -or $activeWuji[0].Name -ne 'wuji-legion-codex-2-0') {
  throw "optimization-audit failed: expected one active Wuji entry, found $($activeWuji.Name -join ',')"
}
$activeNuwa = @(Get-ChildItem -LiteralPath $activeRoot -Directory -ErrorAction SilentlyContinue | Where-Object { $_.Name -match 'nuwa' })
if ($activeNuwa.Count -ne 0) { throw "optimization-audit failed: active Nuwa entries found $($activeNuwa.Name -join ',')" }
$activeSkill = Join-Path $activeWuji[0].FullName 'SKILL.md'
$activeAgents = Join-Path $activeWuji[0].FullName 'AGENTS.md'
if ((Get-FileHash -Algorithm SHA256 -LiteralPath $activeSkill).Hash -ne (Get-FileHash -Algorithm SHA256 -LiteralPath (Join-Path $root 'SKILL.md')).Hash -or
    (Get-FileHash -Algorithm SHA256 -LiteralPath $activeAgents).Hash -ne (Get-FileHash -Algorithm SHA256 -LiteralPath (Join-Path $root 'AGENTS.md')).Hash) {
  throw 'optimization-audit failed: active Wuji registration is stale or points at a different contract'
}

$webRoute = Invoke-WujiJson @('route', '--query', 'build a Slidev web presentation with stage fluid')
$dashiRoute = Invoke-WujiJson @('route', '--query', 'create an editable HTML deck')
$pptxRoute = Invoke-WujiJson @('route', '--query', 'create an editable PPTX')
$writingRoute = Invoke-WujiJson @('route', '--query', 'translate this article')
$searchRoute = Invoke-WujiJson @('route', '--query', 'search the web for the latest solution')
$feishuRoute = Invoke-WujiJson @('route', '--query', '读取飞书文档 https://my.feishu.cn/wiki/IqXbwaplZiEGGekoTarctLjsnPe')
$codeQuery = 'fix code workerPlan in internal/core/route.go'
$codeDirectRoute = Invoke-WujiJson @('route', '--query', $codeQuery)
$codeContext = Invoke-WujiJson @('context-select', '--workspace', $root, '--query', $codeQuery, '--max-bytes', '2048')
$codeRoute = Invoke-WujiJson @('route', '--query', $codeQuery, '--context-artifact', $codeContext.artifact_path)
$serialSearchRoute = Invoke-WujiJson @('route', '--query', 'research the web serial only')
$officeCliRoute = Invoke-WujiJson @('route', '--query', 'export Excel structure as JSON')
$ordinaryDocumentRoute = Invoke-WujiJson @('route', '--query', 'create a Word document')
$officePptxRoute = Invoke-WujiJson @('route', '--query', 'export PPTX structure')
if ($webRoute.primary_skill -ne 'wuji-web-deck' -or $webRoute.engine -ne 'web-deck' -or ($webRoute.mounted_sources.id -contains 'ppt-master-complete')) { throw 'fusion-audit failed: web presentation scenario is not consolidated' }
if ($dashiRoute.primary_skill -ne 'wuji-web-deck' -or $dashiRoute.engine -ne 'web-deck' -or ($dashiRoute.mounted_sources.id -notcontains 'wuji-dashiai-deck-adapter')) { throw 'fusion-audit failed: DashiAI narrow semantic route is not automatically mounted' }
if ($pptxRoute.primary_skill -ne 'wuji-editable-deck' -or $pptxRoute.engine -ne 'editable-pptx' -or ($pptxRoute.mounted_sources.id -contains 'slidev-runtime-complete')) { throw 'fusion-audit failed: editable presentation scenario is not consolidated' }
if ($writingRoute.primary_skill -ne 'wuji-writing-suite' -or $writingRoute.engine -ne 'translation') { throw 'fusion-audit failed: writing suite leaked source selection' }
if ($searchRoute.provider -ne 'default-gpt-search' -or @($searchRoute.workers | Where-Object model_class -eq 'agnes').Count -gt 0) { throw 'optimization-audit failed: Agnes returned to search' }
if ($feishuRoute.capability -ne 'feishu' -or $feishuRoute.primary_skill -ne 'feishu-lark' -or @($feishuRoute.mounted_sources | Where-Object id -eq 'official-lark-cli-skill').Count -ne 1) { throw 'fusion-audit failed: Feishu did not automatically select the official read-first capability' }
if ($officeCliRoute.capability -ne 'documents' -or $officeCliRoute.primary_skill -ne 'wuji-document-suite' -or @($officeCliRoute.mounted_sources | Where-Object id -eq 'officecli-stateless-adapter').Count -ne 1) { throw 'fusion-audit failed: OfficeCLI did not mount only for its narrow document route' }
if (@($ordinaryDocumentRoute.mounted_sources | Where-Object id -eq 'officecli-stateless-adapter').Count -ne 0) { throw 'fusion-audit failed: OfficeCLI mounted for an ordinary document task' }
if ($officePptxRoute.capability -ne 'presentation' -or @($officePptxRoute.mounted_sources | Where-Object id -eq 'officecli-stateless-adapter').Count -ne 0) { throw 'fusion-audit failed: OfficeCLI leaked into PPTX routing' }
if (@($searchRoute.workers).Count -ne 3 -or @($searchRoute.workers | Where-Object model -ne 'gpt-5.6-luna').Count -ne 0) { throw 'optimization-audit failed: research workers lost Luna model assignment' }
if (@($searchRoute.workers | Where-Object { @($_.fallback_models | Where-Object { $null -ne $_ -and $_ -ne '' }).Count -ne 0 -or $_.max_attempts -ne 1 -or @($_.fallback_on | Where-Object { $null -ne $_ -and $_ -ne '' }).Count -ne 0 -or $_.max_model_switches -ne 0 }).Count -ne 0) { throw 'optimization-audit failed: research worker received an automatic model fallback' }
if ($serialSearchRoute.parallel -or $serialSearchRoute.delegation_decision.reason -ne 'serial-task-reasoning' -or @($serialSearchRoute.workers).Count -ne 1) { throw 'optimization-audit failed: serial research did not remain one bounded task judgment' }
$serialSearchWorker = @($serialSearchRoute.workers)[0]
if ($serialSearchWorker.id -ne 'task-judgment' -or $serialSearchWorker.model -ne 'gpt-5.6-sol' -or $serialSearchWorker.writes -or $serialSearchWorker.context_mode -ne 'task-contract-only') { throw 'optimization-audit failed: serial research judgment is not bounded and read-only' }
if ($codeDirectRoute.delegation_decision.reason -ne 'verified-context-artifact-required' -or @($codeDirectRoute.workers).Count -ne 1) { throw 'optimization-audit failed: code did not receive its bounded no-context judgment' }
$codeDirectWorker = @($codeDirectRoute.workers)[0]
if ($codeDirectWorker.id -ne 'task-judgment' -or $codeDirectWorker.model -ne 'gpt-5.6-sol' -or $codeDirectWorker.writes -or $codeDirectWorker.context_mode -ne 'task-contract-only' -or $codeDirectWorker.allocated_context_bytes -ne 0) { throw 'optimization-audit failed: code no-context judgment is not bounded and read-only' }
if (@($codeRoute.workers).Count -ne 1 -or @($codeRoute.workers | Where-Object model -ne 'gpt-5.6-sol').Count -ne 0) { throw 'optimization-audit failed: code worker lost bounded Sol assignment' }
if (@($codeRoute.workers | Where-Object { @($_.fallback_models | Where-Object { $null -ne $_ -and $_ -ne '' }).Count -ne 0 -or $_.max_attempts -ne 1 -or @($_.fallback_on | Where-Object { $null -ne $_ -and $_ -ne '' }).Count -ne 0 -or $_.max_model_switches -ne 0 }).Count -ne 0) { throw 'optimization-audit failed: code worker received an automatic model fallback' }
$executionEvidence = 'schema_version,worker_id,requested_model,session_key,host_dispatch_id,write_boundary,attempts,effective_model,model_switch_count,result_handle,stable_prefix_bytes,stable_prefix_sha256,source_execution_bytes,context_handle_ids,context_bytes_sent,context_payload_sha256,task_contract_bytes,task_contract_sha256,delegation_gate_reason,input_tokens,cached_input_tokens,output_tokens,retry_count,accepted_by_aji,attempt_failure_kinds,cache_domain,billing_unit,total_cost_microunits,aji_baseline_microunits,savings_microunits'
if (@($searchRoute.workers + $codeRoute.workers | Where-Object { -not $_.execution_evidence_required -or ($_.execution_evidence_fields -join ',') -ne $executionEvidence }).Count -ne 0) { throw 'optimization-audit failed: worker execution evidence contract is incomplete' }
$officerRoute = Invoke-WujiJson @('route', '--query', 'white-hat review this architecture')
$officerWorker = @($officerRoute.officer_workers)[0]
if (@($officerRoute.officers).Count -ne 1 -or @($officerRoute.officer_workers).Count -ne 1 -or $officerWorker.stage -ne 'officer' -or $officerWorker.writes -or -not $officerWorker.session_key -or -not $officerWorker.execution_evidence_required -or ($officerWorker.execution_evidence_fields -join ',') -ne $executionEvidence) {
  throw 'optimization-audit failed: explicit white-hat is not a receipt-bound read-only worker'
}
if ($codeRoute.delegation_policy.cross_model_cache_assumed -or $codeRoute.delegation_policy.cache_scope -ne 'model-local stable-prefix only' -or $codeRoute.delegation_policy.max_task_contract_bytes -ne 2048 -or $codeRoute.delegation_policy.max_shared_context_bytes -ne 4096 -or $codeRoute.delegation_policy.max_total_replay_bytes -ne 8192 -or $codeRoute.delegation_policy.min_context_coverage_basis_points -ne 6000 -or -not $codeRoute.delegation_policy.require_code_excerpt -or -not $codeRoute.delegation_policy.require_content_anchor -or $codeRoute.delegation_policy.fallback_only_on_availability_error) { throw 'optimization-audit failed: cross-model cost policy is incomplete' }
$codeWorker = @($codeRoute.workers)[0]
if (-not $codeRoute.delegation_decision.allowed -or $codeRoute.delegation_decision.context_handle -ne $codeContext.context_handle -or $codeRoute.delegation_decision.context_coverage_basis_points -lt 6000 -or $codeRoute.delegation_decision.code_excerpt_count -lt 1 -or $codeRoute.delegation_decision.content_anchor_count -lt 1 -or $codeWorker.allocated_context_bytes -ne $codeContext.payload_bytes -or $codeWorker.context_payload_sha256 -ne $codeContext.payload_sha256 -or -not $codeWorker.context_payload -or $codeWorker.allocated_task_contract_bytes -ne ([Text.Encoding]::UTF8.GetByteCount([string]$codeWorker.task_contract)) -or -not $codeWorker.task_contract_sha256 -or -not $codeWorker.stable_capability_prefix -or $codeWorker.stable_prefix_bytes -ne ([Text.Encoding]::UTF8.GetByteCount([string]$codeWorker.stable_capability_prefix)) -or -not $codeWorker.stable_prefix_sha256 -or $codeRoute.delegation_decision.estimated_replay_bytes -ne ($codeWorker.stable_prefix_bytes + $codeWorker.allocated_context_bytes + $codeWorker.allocated_task_contract_bytes) -or ($codeWorker.prompt_order -join ',') -ne 'stable_capability_prefix,context_payload,task_contract' -or $codeWorker.max_attempts -ne 1 -or @($codeWorker.fallback_on | Where-Object { $null -ne $_ -and $_ -ne '' }).Count -ne 0 -or @($codeWorker.fallback_models | Where-Object { $null -ne $_ -and $_ -ne '' }).Count -ne 0 -or $codeWorker.max_model_switches -ne 0) { throw 'optimization-audit failed: verified context handoff is incomplete' }
if ($codeRoute.model_policy.main_model -ne 'gpt-5.6-terra' -or $codeRoute.model_policy.class_models.sol -ne 'gpt-5.6-sol' -or $codeRoute.model_policy.class_models.luna -ne 'gpt-5.6-luna' -or $null -ne $codeRoute.model_policy.class_models.terra) { throw 'optimization-audit failed: executable model policy is incomplete' }

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
  verified_office_tools = @($sourceLock.office_tools).Count
  nested_skill_count = @($nestedSkillFiles).Count
  manifest_bytes = $manifestBytes
} | ConvertTo-Json
