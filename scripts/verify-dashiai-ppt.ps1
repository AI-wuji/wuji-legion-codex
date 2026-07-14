param([string]$Root = $env:WUJI_ROOT)

$ErrorActionPreference = 'Stop'
if (-not $Root) { $Root = Split-Path $PSScriptRoot -Parent }

$skillRoot = Join-Path $env:USERPROFILE '.codex\skills\dashiai-ppt-skill'
$project = Join-Path $skillRoot 'project'
foreach ($path in @('SKILL.md', 'project\package.json', 'project\scripts\layout-query.mjs', 'project\scripts\render-goal-deck.jsx', 'project\scripts\validate-swiss-deck.mjs', 'project\scripts\validate-goal-copy.mjs')) {
  if (-not (Test-Path -LiteralPath (Join-Path $skillRoot $path) -PathType Leaf)) {
    throw "DashiAI local runtime is incomplete: $path"
  }
}

$evidenceDir = $env:WUJI_PROBE_EVIDENCE_DIR
if (-not $evidenceDir -or -not (Test-Path -LiteralPath $evidenceDir -PathType Container)) {
  throw 'WUJI_PROBE_EVIDENCE_DIR is required for a behavior probe'
}

$work = Join-Path ([IO.Path]::GetTempPath()) ("wuji-dashiai-probe-" + [Guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $work -Force | Out-Null
try {
  Push-Location $work
  try {
    & npm --prefix $project run --silent layout:query -- --theme theme01 --role cover --limit 1 --seed wuji-dashiai-probe
    if ($LASTEXITCODE -ne 0) { throw 'DashiAI layout query failed' }

    $goal = [ordered]@{
      title = 'Wuji Probe'
      goal = 'Validate an offline HTML deck.'
      themePack = 'theme01'
      pageCount = 2
      slides = @(
        [ordered]@{
          layout = 'theme01_page004'
          props = [ordered]@{
            kicker = 'WUJI'; issue = 'PROBE'; overline = 'AUTOMATION'
            titleTop = 'Wuji'; titleBottom = 'HTML Deck'
            subtitle = 'Validate an offline deck.'; sticker = 'READY'
            stats = @(
              [ordered]@{ value = '2'; label = 'slides' },
              [ordered]@{ value = '1'; label = 'workflow' },
              [ordered]@{ value = '100%'; label = 'local' }
            )
            statCount = 3; footnote = 'Offline HTML output'
          }
        },
        [ordered]@{
          layout = 'theme01_page047'
          props = [ordered]@{
            kicker = 'RESULT'; quote = 'Validate every deck.'
            contrastWord = 'offline'; highlightWord = 'HTML'
            attribution = 'Wuji Probe'; caption = 'The generated deck passed validation.'
            notes = @([ordered]@{ label = 'rendered' }, [ordered]@{ label = 'checked' })
            noteCount = 2
          }
        }
      )
    }
    [IO.File]::WriteAllText((Join-Path $work 'goal.json'), ($goal | ConvertTo-Json -Depth 8), [Text.UTF8Encoding]::new($false))

    & npm --prefix $project run --silent validate:goal-spec -- 'goal.json'
    if ($LASTEXITCODE -ne 0) { throw 'DashiAI goal-spec validation failed' }
    & npm --prefix $project run --silent render:goal -- 'goal.json' 'ppt\index.html'
    if ($LASTEXITCODE -ne 0) { throw 'DashiAI deck render failed' }
    & npm --prefix $project run --silent validate:swiss -- 'ppt\index.html'
    if ($LASTEXITCODE -ne 0) { throw 'DashiAI Swiss output validation failed' }
    & npm --prefix $project run --silent validate:goal-copy -- 'goal.json' 'ppt\index.html'
    if ($LASTEXITCODE -ne 0) { throw 'DashiAI goal-copy validation failed' }
  } finally {
    Pop-Location
  }

  $html = Join-Path $work 'ppt\index.html'
  if (-not (Test-Path -LiteralPath $html -PathType Leaf) -or (Get-Item -LiteralPath $html).Length -lt 1024) {
    throw 'DashiAI render produced an empty deck'
  }
  $reportPath = Join-Path $evidenceDir 'dashiai-ppt-assertions.json'
  $report = [ordered]@{
    fixture = 'dashiai-ppt-local-v1'
    source = $skillRoot
    layout_query = 'passed'
    goal_spec = 'passed'
    render = 'passed'
    swiss_validation = 'passed'
    goal_copy_validation = 'passed'
    html_bytes = (Get-Item -LiteralPath $html).Length
    html_sha256 = (Get-FileHash -Algorithm SHA256 -LiteralPath $html).Hash.ToLowerInvariant()
  }
  [IO.File]::WriteAllText($reportPath, ($report | ConvertTo-Json -Compress), [Text.UTF8Encoding]::new($false))
  $reportHash = (Get-FileHash -Algorithm SHA256 -LiteralPath $reportPath).Hash.ToLowerInvariant()
  Write-Output (@{
    wuji_probe = 'behavior'
    fixture = 'dashiai-ppt-local-v1'
    passed = $true
    evidence = @(@{ id = 'assertions'; path = 'dashiai-ppt-assertions.json'; sha256 = $reportHash })
    signature = 'dashiai-ppt-local-v1'
  } | ConvertTo-Json -Compress -Depth 5)
} finally {
  if (Test-Path -LiteralPath $work) { Remove-Item -LiteralPath $work -Recurse -Force }
}
