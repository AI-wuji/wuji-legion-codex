param(
  [string]$Root = (Split-Path $PSScriptRoot -Parent),
  [string]$EvidenceDir = $env:WUJI_PROBE_EVIDENCE_DIR
)

$ErrorActionPreference = 'Stop'
$evidence = if ($EvidenceDir) { [IO.Path]::GetFullPath($EvidenceDir) } else { Join-Path $env:TEMP ('wuji-officecli-evidence-' + [guid]::NewGuid().ToString('N')) }
New-Item -ItemType Directory -Force -Path $evidence | Out-Null
$adapter = Join-Path $Root 'capabilities\documents\adapters\officecli\invoke-officecli.ps1'
$cli = Join-Path $env:LOCALAPPDATA 'OfficeCLI\officecli.exe'
if (-not (Test-Path -LiteralPath $cli -PathType Leaf)) { throw "OfficeCLI is not installed; run scripts/install-officecli.ps1 first: $cli" }
$scratch = Join-Path $env:TEMP ('wuji-officecli-probe-' + [guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Force -Path $scratch | Out-Null
try {
  $env:OFFICECLI_NO_AUTO_RESIDENT = '1'
  $beforeProcesses = @((Get-Process -Name officecli -ErrorAction SilentlyContinue)).Count
  $checks = @()
  foreach ($case in @(
    @{ Format = 'docx'; Sentinel = 'WUJI_OFFICECLI_DOCX_SENTINEL'; File = 'officecli-probe.docx' },
    @{ Format = 'xlsx'; Sentinel = 'WUJI_OFFICECLI_XLSX_SENTINEL'; File = 'officecli-probe.xlsx' }
  )) {
    $input = Join-Path $scratch $case.File
    & $cli create $input | Out-Null
    if ($LASTEXITCODE -ne 0) { throw "OfficeCLI create failed for $($case.Format)" }
    if ($case.Format -eq 'docx') { & $cli add $input /body --type paragraph --prop "text=$($case.Sentinel)" | Out-Null } else { & $cli set $input /Sheet1/A1 --prop "value=$($case.Sentinel)" | Out-Null }
    if ($LASTEXITCODE -ne 0) { throw "OfficeCLI sentinel write failed for $($case.Format)" }
    $html = Join-Path $evidence "officecli-probe.$($case.Format).html"
    $json = Join-Path $evidence "officecli-probe.$($case.Format).json"
    $htmlResult = & $adapter -Operation ViewHtml -InputPath $input -OutputPath $html -OfficeCliPath $cli | ConvertFrom-Json
    $dumpResult = & $adapter -Operation DumpJson -InputPath $input -OutputPath $json -OfficeCliPath $cli | ConvertFrom-Json
    $htmlText = Get-Content -Raw -Encoding UTF8 -LiteralPath $html
    $dumpText = Get-Content -Raw -Encoding UTF8 -LiteralPath $json
    $null = $dumpText | ConvertFrom-Json
    if (-not $htmlText.Contains($case.Sentinel) -or -not $dumpText.Contains($case.Sentinel)) { throw "OfficeCLI probe lost sentinel for $($case.Format)" }
    & $cli validate $input --json | Out-Null
    if ($LASTEXITCODE -ne 0) { throw "OfficeCLI OpenXML validation failed for $($case.Format)" }
    $checks += [ordered]@{ format = $case.Format; sentinel = $case.Sentinel; html_sha256 = $htmlResult.output_sha256; dump_sha256 = $dumpResult.output_sha256; html_contains_sentinel = $true; dump_contains_sentinel = $true }
  }
  Start-Sleep -Milliseconds 500
  $afterProcesses = @((Get-Process -Name officecli -ErrorAction SilentlyContinue)).Count
  if ($afterProcesses -ne $beforeProcesses) { throw "OfficeCLI resident process count changed: before=$beforeProcesses after=$afterProcesses" }
  $assertionsPath = Join-Path $evidence 'officecli-assertions.json'
  $assertions = [ordered]@{ fixture = 'officecli-stateless-v1'; passed = $true; auto_resident_disabled = $true; officecli_processes_before = $beforeProcesses; officecli_processes_after = $afterProcesses; checks = $checks }
  [IO.File]::WriteAllText($assertionsPath, ($assertions | ConvertTo-Json -Compress -Depth 6), [Text.UTF8Encoding]::new($false))
  $files = @('officecli-assertions.json','officecli-probe.docx.html','officecli-probe.docx.json','officecli-probe.xlsx.html','officecli-probe.xlsx.json')
  $artifacts = @($files | ForEach-Object { [ordered]@{ path = $_; sha256 = (Get-FileHash -Algorithm SHA256 -LiteralPath (Join-Path $evidence $_)).Hash.ToLowerInvariant() } })
  [pscustomobject]@{ wuji_probe = 'behavior'; fixture = 'officecli-stateless-v1'; passed = $true; evidence_dir = $evidence; evidence = $artifacts } | ConvertTo-Json -Compress -Depth 6
} finally {
  Remove-Item Env:OFFICECLI_NO_AUTO_RESIDENT -ErrorAction SilentlyContinue
  Remove-Item -LiteralPath $scratch -Recurse -Force -ErrorAction SilentlyContinue
}
