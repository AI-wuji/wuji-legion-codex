param(
  [Parameter(Mandatory = $true)][ValidateNotNullOrEmpty()][string]$Prompt,
  [string]$OutputDir = 'outputs/image/xiaobai-image2',
  [string]$Model = 'gpt-image-2',
  [ValidatePattern('^\d{2,5}x\d{2,5}$')][string]$Size = '1024x1024',
  [ValidateSet('low','medium','high','auto')][string]$Quality = 'low',
  [string]$BaseUrl = '',
  [ValidateRange(10,60000)][int]$PollIntervalMilliseconds = 1000,
  [ValidateRange(1,1800)][int]$TimeoutSeconds = 180,
  [ValidateRange(1024,104857600)][int]$MaxImageBytes = 26214400,
  [string]$ReportPath = '',
  [switch]$AllowInsecureLoopback
)
$ErrorActionPreference = 'Stop'

$apiKey = [string]$env:XIAOBAI_API_KEY
if ([string]::IsNullOrWhiteSpace($apiKey)) {
  throw 'XIAOBAI_API_KEY is required in the process environment; credentials are never read from repository files.'
}
if ([string]::IsNullOrWhiteSpace($BaseUrl)) {
  $BaseUrl = if ($env:XIAOBAI_BASE_URL) { [string]$env:XIAOBAI_BASE_URL } else { 'https://new.xiaobaiapi.cc' }
}
$baseUri = $null
if (-not [Uri]::TryCreate($BaseUrl.TrimEnd('/'), [UriKind]::Absolute, [ref]$baseUri)) { throw 'BaseUrl must be an absolute URI.' }
if ($baseUri.Scheme -ne 'https') {
  if (-not $AllowInsecureLoopback -or $baseUri.Scheme -ne 'http' -or -not $baseUri.IsLoopback) {
    throw 'Image provider endpoints must use HTTPS; HTTP is allowed only for an explicit loopback behavior probe.'
  }
}
$resolvedBase = $baseUri.AbsoluteUri.TrimEnd('/')
$headers = @{ Authorization = "Bearer $($apiKey.Trim())" }

function Invoke-ProviderJson([string]$Method, [string]$Uri, $Body = $null) {
  $params = @{
    Method = $Method
    Uri = $Uri
    Headers = $headers
    TimeoutSec = [Math]::Min(120, $TimeoutSeconds)
  }
  if ($null -ne $Body) {
    $params.ContentType = 'application/json; charset=utf-8'
    $params.Body = $Body | ConvertTo-Json -Depth 8 -Compress
  }
  return Invoke-RestMethod @params
}

function Get-ImageItem($Response) {
  if ($null -eq $Response) { return $null }
  foreach ($candidate in @($Response.data, $Response.result.data, $Response.output)) {
    $items = @($candidate)
    if ($items.Count -gt 0 -and $null -ne $items[0]) { return $items[0] }
  }
  return $null
}

function Get-ImageExtension([byte[]]$Bytes) {
  if ($Bytes.Length -ge 8 -and $Bytes[0] -eq 0x89 -and $Bytes[1] -eq 0x50 -and $Bytes[2] -eq 0x4e -and $Bytes[3] -eq 0x47) { return '.png' }
  if ($Bytes.Length -ge 3 -and $Bytes[0] -eq 0xff -and $Bytes[1] -eq 0xd8 -and $Bytes[2] -eq 0xff) { return '.jpg' }
  if ($Bytes.Length -ge 12 -and [Text.Encoding]::ASCII.GetString($Bytes,0,4) -eq 'RIFF' -and [Text.Encoding]::ASCII.GetString($Bytes,8,4) -eq 'WEBP') { return '.webp' }
  throw 'Provider result is not a supported PNG, JPEG, or WebP image.'
}

$request = [ordered]@{ model = $Model; prompt = $Prompt; n = 1; size = $Size; quality = $Quality }
$created = Invoke-ProviderJson -Method Post -Uri "$resolvedBase/v1/images/generations" -Body $request
$result = $created
$item = Get-ImageItem $result
$taskId = if ($created.task_id) { [string]$created.task_id } elseif ($created.taskId) { [string]$created.taskId } else { '' }

if ($null -eq $item) {
  if ([string]::IsNullOrWhiteSpace($taskId)) { throw 'Xiaobai Image2 returned neither image data nor a task id.' }
  $deadline = [DateTime]::UtcNow.AddSeconds($TimeoutSeconds)
  do {
    $result = Invoke-ProviderJson -Method Get -Uri "$resolvedBase/v1/images/generations/$([Uri]::EscapeDataString($taskId))"
    $status = ([string]$result.status).ToUpperInvariant()
    if ($status -in @('FAILED','FAILURE','ERROR','CANCELLED')) { throw "Xiaobai Image2 task failed with status $status." }
    $item = Get-ImageItem $result
    if ($status -eq 'SUCCESS' -and $null -ne $item) { break }
    if ([DateTime]::UtcNow -ge $deadline) { throw "Xiaobai Image2 task timed out: $taskId" }
    Start-Sleep -Milliseconds $PollIntervalMilliseconds
  } while ($true)
}

$bytes = $null
if ($item.b64_json) {
  try { $bytes = [Convert]::FromBase64String([string]$item.b64_json) }
  catch { throw 'Provider returned invalid base64 image data.' }
} elseif ($item.url) {
  $imageUri = $null
  if (-not [Uri]::TryCreate([string]$item.url, [UriKind]::Absolute, [ref]$imageUri)) { throw 'Provider returned an invalid image URL.' }
  if ($imageUri.Scheme -ne 'https' -and -not ($AllowInsecureLoopback -and $imageUri.Scheme -eq 'http' -and $imageUri.IsLoopback)) {
    throw 'Provider image URL must use HTTPS.'
  }
  $temporary = [IO.Path]::GetTempFileName()
  try {
    Invoke-WebRequest -Uri $imageUri -OutFile $temporary -TimeoutSec ([Math]::Min(120, $TimeoutSeconds)) | Out-Null
    if ((Get-Item -LiteralPath $temporary).Length -gt $MaxImageBytes) { throw 'Provider image exceeds MaxImageBytes.' }
    $bytes = [IO.File]::ReadAllBytes($temporary)
  } finally { Remove-Item -LiteralPath $temporary -Force -ErrorAction SilentlyContinue }
} else {
  throw 'Provider result contains neither b64_json nor url.'
}
if ($bytes.Length -gt $MaxImageBytes) { throw 'Provider image exceeds MaxImageBytes.' }
$extension = Get-ImageExtension $bytes
$targetDir = [IO.Path]::GetFullPath($OutputDir)
New-Item -ItemType Directory -Path $targetDir -Force | Out-Null
$name = 'xiaobai-image2-{0}-{1}{2}' -f (Get-Date -Format 'yyyyMMdd-HHmmss'),([guid]::NewGuid().ToString('N').Substring(0,8)),$extension
$imagePath = Join-Path $targetDir $name
[IO.File]::WriteAllBytes($imagePath, $bytes)

$summary = [ordered]@{
  ok = $true
  provider = 'xiaobai-image2'
  model = $Model
  task_id = $taskId
  size = $Size
  quality = $Quality
  bytes = $bytes.Length
  image_path = $imagePath
}
if ($ReportPath) {
  $report = [IO.Path]::GetFullPath($ReportPath)
  New-Item -ItemType Directory -Path (Split-Path $report -Parent) -Force | Out-Null
  [IO.File]::WriteAllText($report, ($summary | ConvertTo-Json -Depth 5), [Text.UTF8Encoding]::new($false))
}
$summary | ConvertTo-Json -Depth 5
