param(
  [Parameter(Mandatory)][ValidateSet('editable-pptx','web-deck')][string]$Engine,
  [Parameter(Mandatory)][string]$AssetId,
  [Parameter(Mandatory)][string]$AssetPath,
  [Parameter(Mandatory)][string]$AssetSHA256,
  [Parameter(Mandatory)][Int64]$AssetBytes,
  [string]$Output = ''
)
$ErrorActionPreference = 'Stop'
. (Join-Path $PSScriptRoot '..\..\..\scripts\sha256.ps1')
$root = Split-Path $PSScriptRoot -Parent
$projectRoot = [IO.Path]::GetFullPath((Join-Path $root '..\..'))
function Get-AssetDigest([string]$Path) {
  $item = Get-Item -LiteralPath $Path
  if (-not $item.PSIsContainer) {
    return [pscustomobject]@{
      sha256 = (Get-FileHash -Algorithm SHA256 -LiteralPath $Path).Hash.ToLowerInvariant()
      bytes = [Int64]$item.Length
      kind = 'file'
    }
  }

  $base = $item.FullName.TrimEnd('\','/')
  $files = @(Get-ChildItem -LiteralPath $base -Recurse -File | Sort-Object {
    $_.FullName.Substring($base.Length).TrimStart('\','/').Replace('\','/')
  })
  $totalBytes = [Int64]0
  $tree = [Text.StringBuilder]::new()
  foreach ($file in $files) {
    $relative = $file.FullName.Substring($base.Length).TrimStart('\','/').Replace('\','/')
    $fileHash = (Get-FileHash -Algorithm SHA256 -LiteralPath $file.FullName).Hash.ToLowerInvariant()
    $totalBytes += [Int64]$file.Length
    [void]$tree.Append($relative).Append("`t").Append($fileHash).Append("`t").Append($file.Length).Append("`n")
  }
  $algorithm = [Security.Cryptography.SHA256]::Create()
  try {
    $treeHash = ([BitConverter]::ToString($algorithm.ComputeHash([Text.UTF8Encoding]::new($false).GetBytes($tree.ToString())))).Replace('-','').ToLowerInvariant()
  } finally {
    $algorithm.Dispose()
  }
  return [pscustomobject]@{ sha256 = $treeHash; bytes = $totalBytes; kind = 'directory' }
}
if ([IO.Path]::IsPathRooted($AssetPath) -or $AssetPath -match '(^|[\\/])\.\.([\\/]|$)') { throw 'asset path must remain inside presentation capability' }
$full = [IO.Path]::GetFullPath((Join-Path $root ($AssetPath -replace '/', '\\')))
if (-not $full.StartsWith($root + '\', [StringComparison]::OrdinalIgnoreCase) -or -not (Test-Path -LiteralPath $full -PathType Leaf)) { throw 'contract asset is unavailable' }
$info = Get-Item -LiteralPath $full
$hash = (Get-FileHash -Algorithm SHA256 -LiteralPath $full).Hash.ToLowerInvariant()
if ($hash -ne $AssetSHA256.ToLowerInvariant() -or $info.Length -ne $AssetBytes) { throw 'contract asset hash or byte count does not match' }
$catalogPath = Join-Path $root 'assets\template-catalog.json'
$catalog = Get-Content -Raw -Encoding UTF8 -LiteralPath $catalogPath | ConvertFrom-Json
$entry = @($catalog.entries | Where-Object scenario -eq $Engine | Sort-Object category,id | Select-Object -First 1)
if ($entry.Count -ne 1) { throw "catalog has no $Engine asset" }
$selected = $entry[0].preferred
$selectedPath = & (Join-Path $projectRoot 'scripts\expand-wuji-path.ps1') -PathValue $selected.path -Root $projectRoot
if (-not (Test-Path -LiteralPath $selectedPath)) { throw "selected catalog asset is unavailable: $($selected.path)" }
$selectedDigest = Get-AssetDigest $selectedPath
$record = [ordered]@{
  invocation = 'presentation-fusion-runtime'
  engine = $Engine
  contract_asset_id = $AssetId
  contract_asset_path = $AssetPath
  contract_asset_sha256 = $hash
  contract_asset_bytes = [Int64]$info.Length
  selected_catalog_id = $entry[0].id
  selected_catalog_source = $selected.source
  selected_catalog_path = $selected.path
  selected_catalog_kind = $selectedDigest.kind
  selected_catalog_sha256 = $selectedDigest.sha256
  selected_catalog_bytes = $selectedDigest.bytes
}
$json = $record | ConvertTo-Json -Depth 4
if ($Output) { [IO.File]::WriteAllText($Output, $json, [Text.UTF8Encoding]::new($false)) }
$json
