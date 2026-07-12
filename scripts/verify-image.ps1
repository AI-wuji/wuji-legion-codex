$ErrorActionPreference = 'Stop'
$root = Split-Path $PSScriptRoot -Parent
$client = Join-Path $root 'capabilities\image\providers\xiaobai-image2\invoke.ps1'
$scratch = Join-Path $env:TEMP ('wuji-image-probe-' + [guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $scratch | Out-Null
$portProbe = [Net.Sockets.TcpListener]::new([Net.IPAddress]::Loopback, 0)
$portProbe.Start()
$port = ([Net.IPEndPoint]$portProbe.LocalEndpoint).Port
$portProbe.Stop()
$ready = Join-Path $scratch 'ready.txt'
$png = 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9Wl2nXQAAAAASUVORK5CYII='

$job = Start-Job -ArgumentList $port,$ready,$png -ScriptBlock {
  param($Port,$Ready,$Png)
  $listener = [Net.Sockets.TcpListener]::new([Net.IPAddress]::Loopback, $Port)
  $listener.Start()
  [IO.File]::WriteAllText($Ready, 'ready')
  try {
    for ($requestNumber = 0; $requestNumber -lt 2; $requestNumber++) {
      $tcp = $listener.AcceptTcpClient()
      try {
        $stream = $tcp.GetStream()
        $reader = [IO.StreamReader]::new($stream, [Text.Encoding]::UTF8, $false, 4096, $true)
        $requestLine = $reader.ReadLine()
        $contentLength = 0
        $authorization = ''
        while ($true) {
          $line = $reader.ReadLine()
          if ([string]::IsNullOrEmpty($line)) { break }
          if ($line -match '^Content-Length:\s*(\d+)$') { $contentLength = [int]$Matches[1] }
          if ($line -match '^Authorization:\s*(.+)$') { $authorization = $Matches[1].Trim() }
        }
        $body = ''
        if ($contentLength -gt 0) {
          $buffer = [char[]]::new($contentLength)
          $read = 0
          while ($read -lt $contentLength) { $read += $reader.Read($buffer, $read, $contentLength - $read) }
          $body = -join $buffer
        }
        if ($authorization -ne 'Bearer fixture-secret') { throw 'authorization header mismatch' }
        if ($requestNumber -eq 0) {
          if ($requestLine -notmatch '^POST /v1/images/generations ') { throw "unexpected create request: $requestLine" }
          $payload = $body | ConvertFrom-Json
          if ($payload.prompt -ne 'fixture prompt' -or $payload.model -ne 'gpt-image-2') { throw 'create payload mismatch' }
          $response = '{"task_id":"fixture-task","status":"PENDING"}'
        } else {
          if ($requestLine -notmatch '^GET /v1/images/generations/fixture-task ') { throw "unexpected poll request: $requestLine" }
          $response = '{"status":"SUCCESS","result":{"data":[{"b64_json":"' + $Png + '"}]}}'
        }
        $bytes = [Text.Encoding]::UTF8.GetBytes($response)
        $header = [Text.Encoding]::ASCII.GetBytes("HTTP/1.1 200 OK`r`nContent-Type: application/json`r`nContent-Length: $($bytes.Length)`r`nConnection: close`r`n`r`n")
        $stream.Write($header, 0, $header.Length)
        $stream.Write($bytes, 0, $bytes.Length)
        $stream.Flush()
      } finally { $tcp.Dispose() }
    }
  } finally { $listener.Stop() }
}

$previousKey = $env:XIAOBAI_API_KEY
try {
  for ($i = 0; $i -lt 100 -and -not (Test-Path -LiteralPath $ready); $i++) { Start-Sleep -Milliseconds 50 }
  if (-not (Test-Path -LiteralPath $ready)) { throw 'Image provider fixture did not start' }
  $env:XIAOBAI_API_KEY = 'fixture-secret'
  $report = Join-Path $scratch 'report.json'
  $output = & $client -Prompt 'fixture prompt' -OutputDir (Join-Path $scratch 'images') -BaseUrl "http://127.0.0.1:$port" -AllowInsecureLoopback -PollIntervalMilliseconds 10 -TimeoutSeconds 10 -ReportPath $report | ConvertFrom-Json
  if ($null -eq $output -or -not $output.ok -or $output.provider -ne 'xiaobai-image2') { throw 'Direct provider client failed' }
  if (-not (Test-Path -LiteralPath $output.image_path -PathType Leaf)) { throw 'Direct provider image was not materialized' }
  $bytes = [IO.File]::ReadAllBytes($output.image_path)
  if ($bytes.Length -lt 8 -or $bytes[0] -ne 0x89 -or $bytes[1] -ne 0x50) { throw 'Direct provider image signature is invalid' }
  $reportText = Get-Content -Raw -Encoding UTF8 -LiteralPath $report
  if ($reportText -match 'fixture-secret|Authorization|api.?key') { throw 'Provider credential leaked into report' }
  Wait-Job -Job $job -Timeout 10 | Out-Null
  if ($job.State -ne 'Completed') { throw "Image provider fixture failed: $($job.ChildJobs[0].JobStateInfo.Reason)" }
  Receive-Job -Job $job -ErrorAction Stop | Out-Null
  Write-Output 'xiaobai-direct-client-behavior-ok async=poll image=png credential=ephemeral'
} finally {
  $env:XIAOBAI_API_KEY = $previousKey
  Stop-Job -Job $job -ErrorAction SilentlyContinue
  Remove-Job -Job $job -Force -ErrorAction SilentlyContinue
  Remove-Item -LiteralPath $scratch -Recurse -Force -ErrorAction SilentlyContinue
}
