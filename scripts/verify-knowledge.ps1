$ErrorActionPreference = 'Stop'
. (Join-Path $PSScriptRoot 'sha256.ps1')
$root = if ($env:WUJI_ROOT) { $env:WUJI_ROOT } else { Split-Path $PSScriptRoot -Parent }
$wuji = Join-Path $root 'bin\wuji.exe'
& (Join-Path $root 'scripts\build.ps1')

$probeRoot = Join-Path ([IO.Path]::GetTempPath()) ('wuji-knowledge-probe-' + [guid]::NewGuid().ToString('N'))
$workspace = Join-Path $probeRoot 'workspace'
$store = Join-Path $probeRoot 'knowledge'
New-Item -ItemType Directory -Force $workspace | Out-Null
try {
  $solution = Join-Path $workspace 'browser-timeout-solution.md'
  $verification = Join-Path $workspace 'browser-timeout-verification.json'
  $otherOne = Join-Path $workspace 'unrelated.go'
  $otherTwo = Join-Path $workspace 'another.go'
  [IO.File]::WriteAllText($solution, "# Browser timeout`nUse the bounded retry adapter.`n", [Text.UTF8Encoding]::new($false))
  $verificationReceipt = [ordered]@{
    schema_version = 1
    type = 'wuji-verification-receipt'
    passed = $true
    verifier = 'knowledge-behavior-probe'
    verified_at = [DateTime]::UtcNow.ToString('o')
  } | ConvertTo-Json -Compress
  [IO.File]::WriteAllText($verification, $verificationReceipt, [Text.UTF8Encoding]::new($false))
  [IO.File]::WriteAllText($otherOne, "package fixture`nfunc UnrelatedAlpha() {}`n", [Text.UTF8Encoding]::new($false))
  [IO.File]::WriteAllText($otherTwo, "package fixture`nfunc UnrelatedBeta() {}`n", [Text.UTF8Encoding]::new($false))

  $recordRaw = & $wuji knowledge-record --store $store --kind failure --key 'browser timeout bounded retry' --workspace $workspace `
    --summary 'Use the verified bounded retry adapter.' --root-cause 'The browser command exceeded its bounded response window.' `
    --location $solution --verification $verification --tags 'browser,timeout,retry' --relations 'solved-by=bounded-retry-adapter'
  if ($LASTEXITCODE -ne 0) { throw 'knowledge-record failed' }
  $record = $recordRaw | ConvertFrom-Json

  $queryRaw = & $wuji knowledge-query --store $store --trigger failure --kind failure --key 'browser timeout bounded retry' --workspace $workspace --limit 3
  if ($LASTEXITCODE -ne 0) { throw 'event-triggered knowledge-query failed' }
  $query = $queryRaw | ConvertFrom-Json
  if (-not $query.exact_match -or $query.full_scan -or $query.index_lookups -ne 1 -or $query.candidate_records -ne 1 -or @($query.matches).Count -ne 1) {
    throw 'knowledge query did not use one exact bounded lookup'
  }
  if ($query.max_index_lookups -ne 12 -or $query.max_candidate_records -ne 128 -or $query.max_results -ne 10 -or $query.max_refs_per_index -ne 256) {
    throw 'knowledge query hard limits are missing'
  }
  if ($query.matches[0].location -notmatch '^object:sha256:[0-9a-f]{64}$' -or $query.matches[0].verification -notmatch '^object:sha256:[0-9a-f]{64}$') {
    throw 'knowledge query exposed raw local paths instead of content-addressed objects'
  }
  $expectedVerificationHash = (Get-FileHash -Algorithm SHA256 -LiteralPath $verification).Hash.ToLowerInvariant()
  if ($query.matches[0].verification_sha256 -ne $expectedVerificationHash) {
    throw 'knowledge query did not preserve verified evidence identity'
  }

  $previousErrorActionPreference = $ErrorActionPreference
  $ErrorActionPreference = 'Continue'
  $normalOutput = & $wuji knowledge-query --store $store --kind failure --key 'browser timeout bounded retry' --workspace $workspace 2>&1
  $normalExitCode = $LASTEXITCODE
  $ErrorActionPreference = $previousErrorActionPreference
  if ($normalExitCode -eq 0 -or ($normalOutput -join "`n") -notmatch 'event trigger') {
    throw 'normal task startup was allowed to query cross-project knowledge'
  }

  $syncRaw = & $wuji graph-sync --workspace $workspace
  if ($LASTEXITCODE -ne 0) { throw 'graph-sync failed' }
  $sync = $syncRaw | ConvertFrom-Json
  if ($sync.max_terms_per_file -ne 512 -or $sync.max_refs_per_term -ne 256 -or $sync.max_lookups -ne 64 -or $sync.max_candidates -ne 128 -or $sync.max_source_bytes -ne 16777216 -or $sync.file_count -lt 3) {
    throw 'workspace graph hard limits are missing'
  }
  $contextRaw = & $wuji context-select --workspace $workspace --query 'bounded retry adapter' --max-bytes 1024
  if ($LASTEXITCODE -ne 0) { throw 'workspace graph context-select failed' }
  $context = $contextRaw | ConvertFrom-Json
  if ($context.retrieval_mode -notlike 'workspace-graph*' -or $context.graph_lookups -lt 1 -or $context.candidate_files -ge $context.indexed_files) {
    throw 'workspace graph did not narrow candidates before source reads'
  }

  $evidenceDir = $env:WUJI_PROBE_EVIDENCE_DIR
  if (-not $evidenceDir -or -not (Test-Path -LiteralPath $evidenceDir -PathType Container)) {
    throw 'WUJI_PROBE_EVIDENCE_DIR is required for a behavior probe'
  }
  $reportPath = Join-Path $evidenceDir 'knowledge-assertions.json'
  $report = [ordered]@{
    fixture = 'bounded-knowledge-retrieval-v1'
    event_trigger = [string]$query.trigger
    exact_match = [bool]$query.exact_match
    full_scan = [bool]$query.full_scan
    index_lookups = [int]$query.index_lookups
    candidate_records = [int]$query.candidate_records
    max_index_lookups = [int]$query.max_index_lookups
    max_candidate_records = [int]$query.max_candidate_records
    max_results = [int]$query.max_results
    max_refs_per_index = [int]$query.max_refs_per_index
    content_addressed_solution = ($query.matches[0].location -match '^object:sha256:[0-9a-f]{64}$')
    capacity_node_count = [int]$query.capacity.node_count
    capacity_max_nodes = [int]$query.capacity.max_nodes
    capacity_store_bytes = [int64]$query.capacity.store_bytes
    capacity_max_store_bytes = [int64]$query.capacity.max_store_bytes
    verification_sha256 = [string]$query.matches[0].verification_sha256
    normal_startup_rejected = $true
    retrieval_mode = [string]$context.retrieval_mode
    indexed_files = [int]$context.indexed_files
    candidate_files = [int]$context.candidate_files
    graph_lookups = [int]$context.graph_lookups
    max_terms_per_file = [int]$sync.max_terms_per_file
    max_refs_per_term = [int]$sync.max_refs_per_term
    max_graph_lookups = [int]$sync.max_lookups
    max_graph_candidates = [int]$sync.max_candidates
    max_graph_source_bytes = [int64]$sync.max_source_bytes
    retrieval_truncated = [bool]$context.retrieval_truncated
    record_revision = [int]$record.revision
  }
  [IO.File]::WriteAllText($reportPath, ($report | ConvertTo-Json -Compress), [Text.UTF8Encoding]::new($false))
  $reportHash = (Get-FileHash -Algorithm SHA256 -LiteralPath $reportPath).Hash.ToLowerInvariant()
  Write-Output (@{
    wuji_probe = 'behavior'
    fixture = 'bounded-knowledge-retrieval-v1'
    passed = $true
    evidence = @(@{ id = 'assertions'; path = 'knowledge-assertions.json'; sha256 = $reportHash })
    signature = 'bounded-knowledge-retrieval-v1'
  } | ConvertTo-Json -Compress -Depth 5)
} finally {
  Remove-Item -LiteralPath $probeRoot -Recurse -Force -ErrorAction SilentlyContinue
}
