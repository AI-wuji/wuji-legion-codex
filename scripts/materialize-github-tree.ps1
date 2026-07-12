param(
  [Parameter(Mandatory = $true)][string]$Owner,
  [Parameter(Mandatory = $true)][string]$Repo,
  [Parameter(Mandatory = $true)][ValidatePattern('^[a-f0-9]{40}$')][string]$Commit,
  [Parameter(Mandatory = $true)][string]$BasePath,
  [Parameter(Mandatory = $true)][string]$OutputPath,
  [ValidateRange(1,16)][int]$Concurrency = 8
)
$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'
$headers = @{ 'User-Agent' = 'Wuji-Legion-Source-Audit'; 'Accept' = 'application/vnd.github+json' }
$base = [IO.Path]::GetFullPath($BasePath).TrimEnd('\')
$output = [IO.Path]::GetFullPath($OutputPath).TrimEnd('\')
if (-not (Test-Path -LiteralPath (Join-Path $base '.git'))) { throw "BasePath must be a Git worktree: $base" }
if (Test-Path -LiteralPath $output) { throw "OutputPath must not exist: $output" }
$dirty = @(& git -C $base status --porcelain --untracked-files=all)
if ($LASTEXITCODE -ne 0 -or $dirty.Count -gt 0) { throw 'BasePath must be a clean Git worktree.' }
$baseCommit = (@(& git -C $base rev-parse HEAD))[0].Trim()
if ($baseCommit -notmatch '^[a-f0-9]{40}$') { throw 'BasePath HEAD is invalid.' }

function Get-GitHubTree([string]$Revision) {
  $commitInfo = Invoke-RestMethod -Headers $headers -Uri "https://api.github.com/repos/$Owner/$Repo/git/commits/$Revision"
  $tree = Invoke-RestMethod -Headers $headers -Uri "https://api.github.com/repos/$Owner/$Repo/git/trees/$($commitInfo.tree.sha)?recursive=1"
  if ($tree.truncated) { throw "GitHub tree response is truncated for $Revision" }
  return $tree
}

$baseTree = Get-GitHubTree $baseCommit
$targetTree = Get-GitHubTree $Commit
$baseBlobs = [Collections.Generic.Dictionary[string,string]]::new([StringComparer]::Ordinal)
foreach ($entry in @($baseTree.tree | Where-Object type -eq 'blob')) { $baseBlobs[$entry.path] = $entry.sha }
$targetBlobs = [Collections.Generic.Dictionary[string,object]]::new([StringComparer]::Ordinal)
foreach ($entry in @($targetTree.tree | Where-Object type -eq 'blob')) {
  if ($targetBlobs.ContainsKey($entry.path)) { throw "Duplicate target path: $($entry.path)" }
  $targetBlobs[$entry.path] = $entry
}

New-Item -ItemType Directory -Path $output | Out-Null
& robocopy $base $output /E /XD (Join-Path $base '.git') /NFL /NDL /NJH /NJS /NP | Out-Null
if ($LASTEXITCODE -ge 8) { throw "Base tree copy failed with robocopy exit code $LASTEXITCODE" }

foreach ($file in @(Get-ChildItem -LiteralPath $output -Recurse -File -Force)) {
  $relative = $file.FullName.Substring($output.Length).TrimStart('\') -replace '\\','/'
  if (-not $targetBlobs.ContainsKey($relative)) { Remove-Item -LiteralPath $file.FullName -Force }
}

$changed = @($targetBlobs.Values | Where-Object {
  $local = Join-Path $output ($_.path -replace '/','\')
  -not $baseBlobs.ContainsKey($_.path) -or $baseBlobs[$_.path] -ne $_.sha -or -not (Test-Path -LiteralPath $local -PathType Leaf)
})

Add-Type -AssemblyName System.Net.Http
$client = [Net.Http.HttpClient]::new()
$client.Timeout = [TimeSpan]::FromSeconds(120)
$client.DefaultRequestHeaders.UserAgent.ParseAdd('Wuji-Legion-Source-Audit')
try {
  for ($offset = 0; $offset -lt $changed.Count; $offset += $Concurrency) {
    $last = [Math]::Min($changed.Count - 1, $offset + $Concurrency - 1)
    $batch = @($changed[$offset..$last])
    $downloads = @()
    foreach ($entry in $batch) {
      $segments = @($entry.path -split '/' | ForEach-Object { [Uri]::EscapeDataString($_) })
      $uri = "https://raw.githubusercontent.com/$Owner/$Repo/$Commit/$($segments -join '/')"
      $downloads += [pscustomobject]@{ entry = $entry; uri = $uri; task = $client.GetByteArrayAsync($uri) }
    }
    try { [Threading.Tasks.Task]::WaitAll([Threading.Tasks.Task[]]@($downloads.task)) }
    catch { }
    foreach ($download in $downloads) {
      $entry = $download.entry
      $target = Join-Path $output ($entry.path -replace '/','\')
      New-Item -ItemType Directory -Path (Split-Path $target -Parent) -Force | Out-Null
      if ($download.task.Status -eq [Threading.Tasks.TaskStatus]::RanToCompletion) {
        [IO.File]::WriteAllBytes($target, [byte[]]$download.task.Result)
      } else {
        $downloaded = $false
        for ($attempt = 1; $attempt -le 3 -and -not $downloaded; $attempt++) {
          try {
            Invoke-WebRequest -UseBasicParsing -Headers $headers -Uri $download.uri -OutFile $target -TimeoutSec 120
            $downloaded = $true
          } catch {
            if ($attempt -eq 3) { throw }
            Start-Sleep -Seconds $attempt
          }
        }
      }
      $actual = (@(& git hash-object --no-filters -- $target))[0].Trim()
      if ($LASTEXITCODE -ne 0 -or $actual -ne $entry.sha) { throw "Git blob verification failed: $($entry.path)" }
    }
  }
} finally { $client.Dispose() }

$materializedFiles = @(Get-ChildItem -LiteralPath $output -Recurse -File -Force)
if ($materializedFiles.Count -ne $targetBlobs.Count) {
  throw "Materialized file count mismatch: expected $($targetBlobs.Count), got $($materializedFiles.Count)"
}
[ordered]@{
  ok = $true
  owner = $Owner
  repo = $Repo
  base_commit = $baseCommit
  target_commit = $Commit
  target_tree = [string]$targetTree.sha
  files = $targetBlobs.Count
  changed_files = $changed.Count
  changed_bytes = ($changed | Measure-Object size -Sum).Sum
  output = $output
} | ConvertTo-Json -Depth 4
