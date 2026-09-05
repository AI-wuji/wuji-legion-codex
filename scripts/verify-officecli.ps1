param(
  [string]$Root = (Split-Path $PSScriptRoot -Parent),
  [string]$EvidenceDir = $env:WUJI_PROBE_EVIDENCE_DIR
)

$ErrorActionPreference = 'Stop'
. (Join-Path $PSScriptRoot 'sha256.ps1')
$dotnetRoot = Join-Path $env:LOCALAPPDATA 'WujiLegion\dotnet'
if (Test-Path -LiteralPath (Join-Path $dotnetRoot 'dotnet.exe') -PathType Leaf) {
  $env:DOTNET_ROOT = $dotnetRoot
  $env:Path = $dotnetRoot + [IO.Path]::PathSeparator + $env:Path
}
$evidence = if ($EvidenceDir) { [IO.Path]::GetFullPath($EvidenceDir) } else { Join-Path $env:TEMP ('wuji-officecli-evidence-' + [guid]::NewGuid().ToString('N')) }
New-Item -ItemType Directory -Force -Path $evidence | Out-Null
$adapter = Join-Path $Root 'capabilities\documents\adapters\officecli\invoke-officecli.ps1'
$cli = Join-Path $env:LOCALAPPDATA 'OfficeCLI\officecli.exe'
if (-not (Test-Path -LiteralPath $cli -PathType Leaf)) { throw "OfficeCLI is not installed; run scripts/install-officecli.ps1 first: $cli" }
$scratch = Join-Path $env:TEMP ('wuji-officecli-probe-' + [guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Force -Path $scratch | Out-Null
function Quote-OfficeCliArgument([string]$Value) {
  if ($Value -notmatch '[\s"]') { return $Value }
  return '"' + ($Value -replace '(\\*)"', '$1$1\\"' -replace '(\\*)$', '$1$1') + '"'
}

function Invoke-OfficeCliChecked([string[]]$Arguments, [string]$Description) {
  $psi = [Diagnostics.ProcessStartInfo]::new()
  $psi.FileName = $cli
  $psi.Arguments = (($Arguments | ForEach-Object { Quote-OfficeCliArgument ([string]$_) }) -join ' ')
  $psi.UseShellExecute = $false
  $psi.RedirectStandardOutput = $true
  $psi.RedirectStandardError = $true
  $process = [Diagnostics.Process]::new()
  $process.StartInfo = $psi
  try {
    if (-not $process.Start()) { throw "OfficeCLI $Description did not start" }
    $stdoutTask = $process.StandardOutput.ReadToEndAsync()
    $stderrTask = $process.StandardError.ReadToEndAsync()
    if (-not $process.WaitForExit(45000)) {
      & taskkill /PID $process.Id /T /F | Out-Null
      throw "OfficeCLI $Description timed out after 45 seconds"
    }
    $stdout = $stdoutTask.GetAwaiter().GetResult()
    $stderr = $stderrTask.GetAwaiter().GetResult()
    if ($process.ExitCode -ne 0) { throw "OfficeCLI $Description failed exit=$($process.ExitCode): $stderr$stdout" }
  } finally {
    $process.Dispose()
  }
}
try {
  $env:OFFICECLI_NO_AUTO_RESIDENT = '1'
  $beforeProcesses = @((Get-Process -Name officecli -ErrorAction SilentlyContinue)).Count
  $checks = @()
  foreach ($case in @(
    @{ Format = 'docx'; Sentinel = 'WUJI_OFFICECLI_DOCX_SENTINEL'; File = 'officecli-probe.docx' },
    @{ Format = 'xlsx'; Sentinel = 'WUJI_OFFICECLI_XLSX_SENTINEL'; File = 'officecli-probe.xlsx' }
  )) {
    $input = Join-Path $scratch $case.File
    Invoke-OfficeCliChecked @('create', $input) "create for $($case.Format)"
    if ($case.Format -eq 'docx') { Invoke-OfficeCliChecked @('add', $input, '/body', '--type', 'paragraph', '--prop', "text=$($case.Sentinel)") "sentinel write for $($case.Format)" } else { Invoke-OfficeCliChecked @('set', $input, '/Sheet1/A1', '--prop', "value=$($case.Sentinel)") "sentinel write for $($case.Format)" }
    $html = Join-Path $evidence "officecli-probe.$($case.Format).html"
    $json = Join-Path $evidence "officecli-probe.$($case.Format).json"
    $htmlResult = & $adapter -Operation ViewHtml -InputPath $input -OutputPath $html -OfficeCliPath $cli | ConvertFrom-Json
    $dumpResult = & $adapter -Operation DumpJson -InputPath $input -OutputPath $json -OfficeCliPath $cli | ConvertFrom-Json
    $htmlText = Get-Content -Raw -Encoding UTF8 -LiteralPath $html
    $dumpText = Get-Content -Raw -Encoding UTF8 -LiteralPath $json
    $null = $dumpText | ConvertFrom-Json
    if (-not $htmlText.Contains($case.Sentinel) -or -not $dumpText.Contains($case.Sentinel)) { throw "OfficeCLI probe lost sentinel for $($case.Format)" }
    Invoke-OfficeCliChecked @('validate', $input, '--json') "OpenXML validation for $($case.Format)"
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
