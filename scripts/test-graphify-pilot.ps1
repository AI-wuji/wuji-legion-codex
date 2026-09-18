$ErrorActionPreference = 'Stop'
$repo = Split-Path -Parent $PSScriptRoot
$fixture = Join-Path ([IO.Path]::GetTempPath()) ('wuji graphify test ' + [guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Force $fixture | Out-Null
try {
  @'
def leaf():
    return 1
def root():
    return leaf()
'@ | Set-Content -LiteralPath (Join-Path $fixture 'calls.py') -Encoding utf8NoBOM
  $fake = Join-Path $fixture 'fake.ps1'
  $fakeCmd = Join-Path $fixture 'graphify.cmd'
  @'
param($command,$source,$codeOnly,$noCluster,$workersFlag,$workers,$outFlag,$out)
$graphDir = Join-Path $out 'graphify-out'; New-Item -ItemType Directory -Force $graphDir | Out-Null
$nodes = 1..12 | ForEach-Object { @{ id="n$_"; type='function'; name="f$_"; source_file='calls.py' } }
$links = 1..20 | ForEach-Object { @{ source="n$($_ % 12 + 1)"; target="n$(($_ + 1) % 12 + 1)"; type='CALLS' } }
@{ nodes=$nodes; links=$links } | ConvertTo-Json -Depth 5 | Set-Content -LiteralPath (Join-Path $graphDir 'graph.json') -Encoding utf8NoBOM
'@ | Set-Content -LiteralPath $fake -Encoding utf8NoBOM
  "@`"$((Get-Process -Id $PID).Path)`" -NoProfile -File `"$fake`" %*" | Set-Content -LiteralPath $fakeCmd -Encoding ascii
  $pwsh = (Get-Process -Id $PID).Path
  $raw = & "$PSScriptRoot\graphify-pilot.ps1" -Workspace $fixture -File calls.py -GraphifyExecutable $fakeCmd -MaxNodes 4 -MaxEdges 5 -MaxOutputBytes 8192
  if ($LASTEXITCODE) { throw "bounded fixture failed: $raw" }
  $result = $raw | ConvertFrom-Json
  if ($result.status -ne 'ok' -or @($result.nodes).Count -ne 4 -or @($result.edges).Count -ne 5) { throw 'dense projection caps failed' }
  if (-not $result.truncated.nodes -or -not $result.truncated.edges) { throw 'truncation was not reported' }
  $returnedIds = @{}; @($result.nodes) | ForEach-Object { $returnedIds[[string]$_.id] = $true }
  foreach ($edge in @($result.edges)) {
    if (-not $returnedIds.ContainsKey([string]$edge.source) -or -not $returnedIds.ContainsKey([string]$edge.target)) {
      throw 'projected edge references a truncated node'
    }
  }
  if ($result.edges[0].authority -ne 'EXTRACTED' -or $result.files[0].sha256.Length -ne 64) { throw 'relation authority or stale hash missing' }
  $outside = Join-Path ([IO.Path]::GetTempPath()) ('wuji-graphify-outside-' + [guid]::NewGuid().ToString('N'))
  New-Item -ItemType Directory -Force $outside | Out-Null
  Set-Content -LiteralPath (Join-Path $outside 'escape.py') -Value 'def escaped(): pass' -Encoding utf8NoBOM
  $link = Join-Path $fixture 'linked'
  New-Item -ItemType Junction -Path $link -Target $outside | Out-Null
  $escaped = & $pwsh -NoProfile -File "$PSScriptRoot\graphify-pilot.ps1" -Workspace $fixture -File linked/escape.py -GraphifyExecutable $fakeCmd 2>&1
  if ($LASTEXITCODE -eq 0 -or "$escaped" -notmatch 'reparse point') { throw 'reparse-point escape guard failed' }
  Remove-Item -LiteralPath $link -Force
  Remove-Item -LiteralPath $outside -Recurse -Force
  Remove-Item -LiteralPath (Join-Path $fixture 'calls.py')
  $missing = & $pwsh -NoProfile -File "$PSScriptRoot\graphify-pilot.ps1" -Workspace $fixture -File calls.py -GraphifyExecutable $pwsh 2>&1
  if ($LASTEXITCODE -eq 0 -or "$missing" -notmatch 'does not exist') { throw 'deletion/stale selection guard failed' }
  @'
param($args); Start-Sleep -Seconds 3
'@ | Set-Content -LiteralPath $fake -Encoding utf8NoBOM
  Set-Content -LiteralPath (Join-Path $fixture 'calls.py') -Value 'def x(): pass' -Encoding utf8NoBOM
  $timed = & $pwsh -NoProfile -File "$PSScriptRoot\graphify-pilot.ps1" -Workspace $fixture -File calls.py -GraphifyExecutable $fakeCmd -TimeoutSeconds 1 2>&1
  if ($LASTEXITCODE -ne 4 -or "$timed" -notmatch 'timeout') { throw 'timeout guard failed' }
  @'
param($command,$source,$codeOnly,$noCluster,$workersFlag,$workers,$outFlag,$out)
[Console]::Out.Write(('x' * 8192))
Start-Sleep -Milliseconds 200
'@ | Set-Content -LiteralPath $fake -Encoding utf8NoBOM
  $noisy = & $pwsh -NoProfile -File "$PSScriptRoot\graphify-pilot.ps1" -Workspace $fixture -File calls.py -GraphifyExecutable $fakeCmd -MaxCaptureBytes 1024 2>&1
  if ($LASTEXITCODE -ne 5 -or "$noisy" -notmatch 'output-limit') { throw 'capture output cap failed' }
  'graphify-pilot adapter behavior: PASS'
} finally {
  if (Test-Path $fixture) { Remove-Item -LiteralPath $fixture -Recurse -Force }
}
$global:LASTEXITCODE = 0
