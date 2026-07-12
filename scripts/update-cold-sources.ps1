param(
  [string[]]$SourceId = @(),
  [switch]$Apply,
  [ValidateSet('Auto','Git','Archive','TreeDelta')][string]$Transport = 'Auto'
)
$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'
$root = Split-Path $PSScriptRoot -Parent
$SourceId = @($SourceId | ForEach-Object { $_ -split ',' } | ForEach-Object { $_.Trim() } | Where-Object { $_ } | Select-Object -Unique)
$projects = if ($env:WUJI_PROJECTS) { [IO.Path]::GetFullPath($env:WUJI_PROJECTS) } else { [IO.Path]::GetFullPath((Join-Path $root '..')) }
$lockPath = Join-Path $root 'sources.lock.json'
$lock = Get-Content -Raw -Encoding UTF8 -LiteralPath $lockPath | ConvertFrom-Json
$headers = @{ 'User-Agent' = 'Wuji-Legion-Source-Audit'; 'Accept' = 'application/vnd.github+json' }

function Get-SourceTreeHash([string]$Path) {
  $full = [IO.Path]::GetFullPath($Path).TrimEnd('\','/')
  $lines = @(Get-ChildItem -LiteralPath $full -Recurse -File -Force | ForEach-Object {
    $relative = $_.FullName.Substring($full.Length).TrimStart('\','/') -replace '\\','/'
    $hash = (Get-FileHash -Algorithm SHA256 -LiteralPath $_.FullName).Hash.ToLowerInvariant()
    "$relative`t$hash"
  } | Sort-Object)
  $payload = ($lines -join "`n") + "`n"
  $sha = [Security.Cryptography.SHA256]::Create()
  try { return ([BitConverter]::ToString($sha.ComputeHash([Text.Encoding]::UTF8.GetBytes($payload))) -replace '-','').ToLowerInvariant() }
  finally { $sha.Dispose() }
}

$rows = @()
foreach ($source in @($lock.sources)) {
  if ($SourceId.Count -gt 0 -and $source.id -notin $SourceId) { continue }
  if ([string]$source.repo -notmatch '^https://github\.com/(?<owner>[^/]+)/(?<repo>[^/]+?)(?:\.git)?$') { continue }
  $owner = $Matches.owner
  $repo = $Matches.repo
  $remote = Invoke-RestMethod -Headers $headers -Uri "https://api.github.com/repos/$owner/$repo/commits?per_page=1"
  $head = [string]$remote[0].sha
  if ($head -notmatch '^[a-f0-9]{40}$') { throw "Invalid remote HEAD for $($source.id)" }
  $sourcePath = & (Join-Path $PSScriptRoot 'expand-wuji-path.ps1') -PathValue $source.path -Root $root
  $rows += [pscustomobject]@{
    id = [string]$source.id
    path = $sourcePath
    locked = [string]$source.commit
    head = $head
    update_available = ([string]$source.commit -ne $head)
    owner = $owner
    repo_name = $repo
    source = $source
  }
}
if ($SourceId.Count -gt 0) {
  $missing = @($SourceId | Where-Object { $_ -notin $rows.id })
  if ($missing.Count -gt 0) { throw "Unknown or non-GitHub source ids: $($missing -join ', ')" }
}

$pending = @($rows | Where-Object update_available)
if ($Apply -and $pending.Count -gt 0) {
  $scratch = Join-Path $env:TEMP ('wuji-source-update-' + [guid]::NewGuid().ToString('N'))
  New-Item -ItemType Directory -Path $scratch | Out-Null
  $plans = @()
  $switched = @()
  $installed = @()
  try {
    foreach ($item in $pending) {
      $mode = $Transport
      if ($mode -eq 'TreeDelta') {
        $staged = Join-Path $scratch "$($item.id)-tree"
        $materialized = & (Join-Path $PSScriptRoot 'materialize-github-tree.ps1') -Owner $item.owner -Repo $item.repo_name -Commit $item.head -BasePath $item.path -OutputPath $staged | ConvertFrom-Json
        if (-not $materialized.ok -or $materialized.target_commit -ne $item.head) { throw "Tree materialization failed: $($item.id)" }
        $target = Join-Path $projects "wuji-capability-sources\upstream-snapshots\$($item.id)"
        $plans += [pscustomobject]@{ item = $item; mode = 'TreeDelta'; staged = $staged; target = $target; tree = (Get-SourceTreeHash $staged) }
        continue
      }
      if ($mode -ne 'Archive') {
        if (-not (Test-Path -LiteralPath (Join-Path $item.path '.git'))) {
          if ($mode -eq 'Git') { throw "Git metadata missing for $($item.id): $($item.path)" }
          $mode = 'Archive'
        } else {
          $dirty = @(& git -C $item.path status --porcelain)
          if ($LASTEXITCODE -ne 0 -or $dirty.Count -gt 0) { throw "Cold source must be a clean Git worktree: $($item.id)" }
          & git -c http.lowSpeedLimit=1 -c http.lowSpeedTime=30 -C $item.path fetch --depth=1 origin $item.head
          if ($LASTEXITCODE -ne 0) {
            if ($mode -eq 'Git') { throw "Fetch failed; no lock entries were changed: $($item.id)" }
            $mode = 'Archive'
          } else { $mode = 'Git' }
        }
      }
      if ($mode -eq 'Archive') {
        $archive = Join-Path $scratch "$($item.id).zip"
        $extract = Join-Path $scratch "$($item.id)-extract"
        Invoke-WebRequest -UseBasicParsing -Headers $headers -Uri "https://codeload.github.com/$($item.owner)/$($item.repo_name)/zip/$($item.head)" -OutFile $archive -TimeoutSec 180
        Expand-Archive -LiteralPath $archive -DestinationPath $extract -Force
        $roots = @(Get-ChildItem -LiteralPath $extract -Directory)
        if ($roots.Count -ne 1) { throw "Unexpected archive layout for $($item.id)" }
        $target = Join-Path $projects "wuji-capability-sources\upstream-snapshots\$($item.id)"
        $plans += [pscustomobject]@{ item = $item; mode = 'Archive'; staged = $roots[0].FullName; target = $target; tree = (Get-SourceTreeHash $roots[0].FullName) }
      } else {
        $plans += [pscustomobject]@{ item = $item; mode = 'Git'; staged = ''; target = $item.path; tree = '' }
      }
    }

    foreach ($plan in $plans) {
      $item = $plan.item
      if ($plan.mode -eq 'Git') {
        & git -C $item.path switch --detach $item.head
        if ($LASTEXITCODE -ne 0) { throw "Checkout failed: $($item.id)" }
        $switched += $item
        continue
      }
      $snapshotRoot = [IO.Path]::GetFullPath((Join-Path $projects 'wuji-capability-sources\upstream-snapshots')).TrimEnd('\')
      $target = [IO.Path]::GetFullPath($plan.target)
      if (-not $target.StartsWith($snapshotRoot + '\', [StringComparison]::OrdinalIgnoreCase)) { throw "Unsafe snapshot target: $target" }
      New-Item -ItemType Directory -Path $snapshotRoot -Force | Out-Null
      $backup = ''
      if (Test-Path -LiteralPath $target) {
        $backup = "$target.backup-$([guid]::NewGuid().ToString('N'))"
        Move-Item -LiteralPath $target -Destination $backup
      }
      $installed += [pscustomobject]@{ target = $target; backup = $backup }
      Move-Item -LiteralPath $plan.staged -Destination $target
    }
  } catch {
    foreach ($entry in @($installed | Select-Object -Reverse)) {
      if (Test-Path -LiteralPath $entry.target) { Remove-Item -LiteralPath $entry.target -Recurse -Force }
      if ($entry.backup -and (Test-Path -LiteralPath $entry.backup)) { Move-Item -LiteralPath $entry.backup -Destination $entry.target }
    }
    foreach ($item in @($switched | Select-Object -Reverse)) {
      & git -C $item.path switch --detach $item.locked 2>$null | Out-Null
    }
    throw
  } finally {
    Remove-Item -LiteralPath $scratch -Recurse -Force -ErrorAction SilentlyContinue
  }

  foreach ($plan in $plans) {
    $item = $plan.item
    $item.source.commit = $item.head
    if ($plan.mode -in @('Archive','TreeDelta')) {
      $item.source.path = '${WUJI_PROJECTS}/wuji-capability-sources/upstream-snapshots/' + $item.id
      if ($item.source.PSObject.Properties.Name -contains 'tree_sha256') { $item.source.tree_sha256 = $plan.tree }
      else { $item.source | Add-Member -NotePropertyName tree_sha256 -NotePropertyValue $plan.tree }
    }
  }
  $json = $lock | ConvertTo-Json -Depth 12
  $temp = "$lockPath.tmp"
  [IO.File]::WriteAllText($temp, $json, [Text.UTF8Encoding]::new($false))
  Move-Item -LiteralPath $temp -Destination $lockPath -Force
  foreach ($entry in $installed) {
    if ($entry.backup -and (Test-Path -LiteralPath $entry.backup)) { Remove-Item -LiteralPath $entry.backup -Recurse -Force }
  }
}

$report = [ordered]@{
  checked_at = (Get-Date).ToUniversalTime().ToString('o')
  applied = [bool]$Apply
  transport = $Transport
  checked = $rows.Count
  updates = $pending.Count
  sources = @($rows | ForEach-Object {
    [ordered]@{
      id = $_.id
      locked = $_.locked
      head = $_.head
      update_available = $_.update_available
      path = $_.path
    }
  })
}
$report | ConvertTo-Json -Depth 6
