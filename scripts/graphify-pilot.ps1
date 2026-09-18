#requires -Version 7.0
[CmdletBinding()]
param(
  [Parameter(Mandatory)][string]$Workspace,
  [Parameter(Mandatory)][string[]]$File,
  [string]$GraphifyExecutable = 'graphify',
  [string[]]$GraphifyPrefixArgs = @(),
  [ValidateRange(1, 64)][int]$MaxFiles = 32,
  [ValidateRange(1, 16777216)][int]$MaxInputBytes = 1048576,
  [ValidateRange(1, 4096)][int]$MaxNodes = 256,
  [ValidateRange(1, 8192)][int]$MaxEdges = 512,
  [ValidateRange(256, 1048576)][int]$MaxOutputBytes = 16384,
  [ValidateRange(1024, 1048576)][int]$MaxCaptureBytes = 65536,
  [ValidateRange(1, 120)][int]$TimeoutSeconds = 20
)

$ErrorActionPreference = 'Stop'
$root = [IO.Path]::GetFullPath($Workspace).TrimEnd([IO.Path]::DirectorySeparatorChar)
if (-not [IO.Directory]::Exists($root)) { throw "workspace does not exist: $root" }
if ($File.Count -gt $MaxFiles) { throw "file cap exceeded: $($File.Count) > $MaxFiles" }

function Assert-NoReparsePoint([string]$Path) {
  $relative = [IO.Path]::GetRelativePath($root, $Path)
  $cursor = $root
  foreach ($part in @('.') + @($relative -split '[\\/]')) {
    if ($part -ne '.') { $cursor = Join-Path $cursor $part }
    if (([IO.File]::GetAttributes($cursor) -band [IO.FileAttributes]::ReparsePoint) -ne 0) {
      throw "selected path contains a reparse point: $cursor"
    }
  }
}

$selected = @()
$total = 0L
foreach ($item in $File) {
  $path = [IO.Path]::GetFullPath((Join-Path $root $item))
  if (-not ($path.StartsWith($root + [IO.Path]::DirectorySeparatorChar, [StringComparison]::OrdinalIgnoreCase))) {
    throw "file escapes workspace: $item"
  }
  if (-not [IO.File]::Exists($path)) { throw "selected file does not exist: $item" }
  Assert-NoReparsePoint $path
  $info = [IO.FileInfo]$path
  $total += $info.Length
  if ($total -gt $MaxInputBytes) { throw "input byte cap exceeded: $total > $MaxInputBytes" }
  $selected += [pscustomobject]@{
    absolute = $path
    relative = [IO.Path]::GetRelativePath($root, $path).Replace('\', '/')
    bytes = $info.Length
    sha256 = (Get-FileHash -LiteralPath $path -Algorithm SHA256).Hash.ToLowerInvariant()
  }
}

$temp = Join-Path ([IO.Path]::GetTempPath()) ("wuji-graphify-pilot-" + [guid]::NewGuid().ToString('N'))
[IO.Directory]::CreateDirectory($temp) | Out-Null
try {
  foreach ($entry in $selected) {
    $destination = Join-Path $temp $entry.relative
    [IO.Directory]::CreateDirectory([IO.Path]::GetDirectoryName($destination)) | Out-Null
    [IO.File]::Copy($entry.absolute, $destination, $true)
  }
  $out = Join-Path $temp 'out'
  $args = @($GraphifyPrefixArgs) + @('extract', $temp, '--code-only', '--no-cluster', '--max-workers', '1', '--out', $out)
  try {
    $start = [Diagnostics.ProcessStartInfo]::new()
    $start.FileName = $GraphifyExecutable
    $start.UseShellExecute = $false
    $start.CreateNoWindow = $true
    $start.RedirectStandardOutput = $true
    $start.RedirectStandardError = $true
    foreach ($argument in $args) { $start.ArgumentList.Add([string]$argument) }
    $process = [Diagnostics.Process]::Start($start)
  } catch {
    [pscustomobject]@{ schema='wuji.graphify-pilot.v1'; status='unavailable'; reason='runtime-not-found'; detail=$_.Exception.Message; files=$selected } | ConvertTo-Json -Depth 6
    exit 3
  }
  $stdoutBuffer = [char[]]::new(4096)
  $stderrBuffer = [char[]]::new(4096)
  $stdoutBuilder = [Text.StringBuilder]::new()
  $stderrBuilder = [Text.StringBuilder]::new()
  $stdoutBytes = 0
  $stderrBytes = 0
  $stdoutDone = $false
  $stderrDone = $false
  $stdoutTask = $process.StandardOutput.ReadAsync($stdoutBuffer, 0, $stdoutBuffer.Length)
  $stderrTask = $process.StandardError.ReadAsync($stderrBuffer, 0, $stderrBuffer.Length)
  $deadline = [DateTime]::UtcNow.AddSeconds($TimeoutSeconds)
  while (-not $process.HasExited -or -not $stdoutDone -or -not $stderrDone) {
    if (-not $stdoutDone -and $stdoutTask.IsCompleted) {
      $count = $stdoutTask.GetAwaiter().GetResult()
      if ($count -eq 0) { $stdoutDone = $true } else {
        $chunk = [string]::new($stdoutBuffer, 0, $count)
        $stdoutBytes += [Text.Encoding]::UTF8.GetByteCount($chunk)
        if ($stdoutBytes -le $MaxCaptureBytes) { [void]$stdoutBuilder.Append($chunk) }
        $stdoutTask = $process.StandardOutput.ReadAsync($stdoutBuffer, 0, $stdoutBuffer.Length)
      }
    }
    if (-not $stderrDone -and $stderrTask.IsCompleted) {
      $count = $stderrTask.GetAwaiter().GetResult()
      if ($count -eq 0) { $stderrDone = $true } else {
        $chunk = [string]::new($stderrBuffer, 0, $count)
        $stderrBytes += [Text.Encoding]::UTF8.GetByteCount($chunk)
        if ($stderrBytes -le $MaxCaptureBytes) { [void]$stderrBuilder.Append($chunk) }
        $stderrTask = $process.StandardError.ReadAsync($stderrBuffer, 0, $stderrBuffer.Length)
      }
    }
    if ($stdoutBytes -gt $MaxCaptureBytes -or $stderrBytes -gt $MaxCaptureBytes) {
      if (-not $process.HasExited) { $process.Kill($true) }
      [pscustomobject]@{ schema='wuji.graphify-pilot.v1'; status='output-limit'; capture_bytes=$MaxCaptureBytes; files=$selected } | ConvertTo-Json -Depth 6
      exit 5
    }
    if ([DateTime]::UtcNow -ge $deadline) {
      if (-not $process.HasExited) { $process.Kill($true) }
      [pscustomobject]@{ schema='wuji.graphify-pilot.v1'; status='timeout'; timeout_seconds=$TimeoutSeconds; files=$selected } | ConvertTo-Json -Depth 6
      exit 4
    }
    Start-Sleep -Milliseconds 25
  }
  $stdoutText = $stdoutBuilder.ToString()
  $stderrText = $stderrBuilder.ToString()
  if ($process.ExitCode -ne 0) {
    $detail = $stderrText
    if ($detail.Length -gt 2048) { $detail = $detail.Substring(0, 2048) }
    [pscustomobject]@{ schema='wuji.graphify-pilot.v1'; status='unavailable'; reason='upstream-failed'; exit_code=$process.ExitCode; detail=$detail; files=$selected } | ConvertTo-Json -Depth 6
    exit 3
  }
  $graphPath = Join-Path $out 'graphify-out/graph.json'
  if (-not [IO.File]::Exists($graphPath)) { throw 'upstream succeeded without graphify-out/graph.json' }
  if ((Get-Item -LiteralPath $graphPath).Length -gt 16777216) { throw 'raw graph exceeds 16 MiB parse cap' }
  $graph = Get-Content -Raw -LiteralPath $graphPath | ConvertFrom-Json
  $rawNodes = @($graph.nodes)
  $rawEdges = if ($null -ne $graph.links) { @($graph.links) } else { @($graph.edges) }
  $nodes = @($rawNodes | Select-Object -First $MaxNodes | ForEach-Object {
    [ordered]@{ id=$_.id; kind=($_.type ?? $_.kind); name=($_.name ?? $_.label); file=$_.source_file }
  })
  $nodeIds = [Collections.Generic.HashSet[string]]::new([StringComparer]::Ordinal)
  foreach ($node in $nodes) { [void]$nodeIds.Add([string]$node.id) }
  $eligibleEdges = @($rawEdges | Where-Object {
    $nodeIds.Contains([string]$_.source) -and $nodeIds.Contains([string]$_.target)
  })
  $edges = @($eligibleEdges | Select-Object -First $MaxEdges | ForEach-Object {
    [ordered]@{ source=$_.source; target=$_.target; relation=($_.type ?? $_.relation ?? $_.label); authority='EXTRACTED' }
  })
  $result = [ordered]@{
    schema='wuji.graphify-pilot.v1'; status='ok'; upstream='graphifyy'; mode='offline-code-only'
    files=$selected; limits=[ordered]@{ files=$MaxFiles; input_bytes=$MaxInputBytes; nodes=$MaxNodes; edges=$MaxEdges; output_bytes=$MaxOutputBytes; timeout_seconds=$TimeoutSeconds }
    truncated=[ordered]@{ nodes=($rawNodes.Count -gt $nodes.Count); edges=($rawEdges.Count -gt $edges.Count) }
    nodes=$nodes; edges=$edges
  }
  $json = $result | ConvertTo-Json -Depth 8 -Compress
  if ([Text.Encoding]::UTF8.GetByteCount($json) -gt $MaxOutputBytes) { throw 'compact projection exceeds output byte cap' }
  $json
} finally {
  if ([IO.Directory]::Exists($temp)) { Remove-Item -LiteralPath $temp -Recurse -Force }
}
