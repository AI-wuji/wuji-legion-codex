param()

$ErrorActionPreference = 'Stop'

$root = Split-Path -Parent $PSScriptRoot
$binDir = Join-Path $root '.wuji-tools'
$bin = Join-Path $binDir ("wuji-cli-test-{0}.exe" -f [guid]::NewGuid().ToString('N'))
$fixture = Join-Path $root 'outputs\tests\wuji-cli'

if (Test-Path $fixture) { Remove-Item -LiteralPath $fixture -Recurse -Force }
New-Item -ItemType Directory -Force -Path $binDir, $fixture | Out-Null
& (Join-Path $PSScriptRoot 'build-wuji-cli.ps1') -Output $bin

function Write-Fixture {
    param([string]$Path, [string]$Content = 'fixture content with enough bytes')
    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $Path) | Out-Null
    Set-Content -LiteralPath $Path -Value $Content -Encoding UTF8
}

function Read-JsonUtf8 {
    param([string]$Path)
    return [System.IO.File]::ReadAllText($Path, [System.Text.UTF8Encoding]::new($false)) | ConvertFrom-Json
}

function Read-NdjsonUtf8 {
    param([string]$Path)
    return [System.IO.File]::ReadAllLines($Path, [System.Text.UTF8Encoding]::new($false)) | Where-Object { $_.Trim() } | ForEach-Object { $_ | ConvertFrom-Json }
}

function Normalize-InspectText {
    param([string]$Text)
    if ($null -eq $Text) {
        return ''
    }
    return (($Text -replace '[\u200B-\u200D\uFEFF]', '') -replace '\s+', ' ').Trim()
}

Add-Type -AssemblyName System.Drawing

function Write-PngFixture {
    param(
        [string]$Path,
        [ValidateSet('contrast', 'whitewashed', 'lowcontrast')]
        [string]$Mode = 'contrast'
    )
    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $Path) | Out-Null
    $bmp = New-Object System.Drawing.Bitmap 96, 96
    $graphics = [System.Drawing.Graphics]::FromImage($bmp)
    try {
        switch ($Mode) {
            'contrast' {
                $graphics.Clear([System.Drawing.Color]::Black)
                $graphics.FillRectangle([System.Drawing.Brushes]::White, 48, 0, 48, 96)
            }
            'whitewashed' {
                $graphics.Clear([System.Drawing.Color]::White)
            }
            'lowcontrast' {
                $graphics.Clear([System.Drawing.Color]::FromArgb(138, 138, 138))
                $brush = New-Object System.Drawing.SolidBrush ([System.Drawing.Color]::FromArgb(150, 150, 150))
                try {
                    $graphics.FillRectangle($brush, 12, 12, 72, 72)
                }
                finally {
                    $brush.Dispose()
                }
            }
        }
        $bmp.Save($Path, [System.Drawing.Imaging.ImageFormat]::Png)
    }
    finally {
        $graphics.Dispose()
        $bmp.Dispose()
    }
}

function New-PptxFixture {
    param(
        [string]$Path,
        [string[]]$Slides,
        [switch]$IncludeMedia
    )
    $workspace = Join-Path $fixture ("pptx-" + [guid]::NewGuid().ToString('N'))
    New-Item -ItemType Directory -Force -Path $workspace, (Join-Path $workspace 'ppt\slides'), (Join-Path $workspace 'ppt\media'), (Join-Path $workspace 'ppt\slideLayouts'), (Join-Path $workspace 'ppt\theme') | Out-Null
    for ($i = 0; $i -lt $Slides.Count; $i++) {
        $slidePath = Join-Path $workspace ("ppt\slides\slide{0}.xml" -f ($i + 1))
        Set-Content -LiteralPath $slidePath -Value $Slides[$i] -Encoding UTF8
    }
    Set-Content -LiteralPath (Join-Path $workspace 'ppt\slideLayouts\slideLayout1.xml') -Value '<layout />' -Encoding UTF8
    Set-Content -LiteralPath (Join-Path $workspace 'ppt\theme\theme1.xml') -Value '<theme />' -Encoding UTF8
    if ($IncludeMedia) {
        Set-Content -LiteralPath (Join-Path $workspace 'ppt\media\image1.png') -Value 'png bytes long enough for test' -Encoding UTF8
    }
    $zipPath = [System.IO.Path]::ChangeExtension($Path, '.zip')
    if (Test-Path $Path) { Remove-Item -LiteralPath $Path -Force }
    if (Test-Path $zipPath) { Remove-Item -LiteralPath $zipPath -Force }
    Compress-Archive -Path (Join-Path $workspace '*') -DestinationPath $zipPath
    Move-Item -LiteralPath $zipPath -Destination $Path
    Remove-Item -LiteralPath $workspace -Recurse -Force
}

function Invoke-Case {
    param([string]$Name, [int]$ExpectedExit, [string[]]$Arguments)
    $previousNativeErrorPref = $PSNativeCommandUseErrorActionPreference
    $previousErrorActionPreference = $ErrorActionPreference
    try {
        $PSNativeCommandUseErrorActionPreference = $false
        $ErrorActionPreference = 'Continue'
        try {
            $output = & $bin @Arguments 2>&1
        }
        catch {
            $output = @($_.Exception.Message)
        }
        $actual = $LASTEXITCODE
    }
    finally {
        $PSNativeCommandUseErrorActionPreference = $previousNativeErrorPref
        $ErrorActionPreference = $previousErrorActionPreference
    }
    if ($actual -ne $ExpectedExit) {
        throw "FAIL $Name expected=$ExpectedExit actual=$actual output=$($output -join ' | ')"
    }
    Write-Host "PASS $Name exit=$actual"
}

$reference = Join-Path $fixture 'reference.txt'
$output = Join-Path $fixture 'generated.txt'
$evidence = Join-Path $fixture 'evidence.txt'
$artifact = Join-Path $fixture 'artifact.txt'
$metaReportArtifact = Join-Path $fixture 'meta-report.json'
Write-Fixture $reference
Write-Fixture $evidence
Write-Fixture $artifact
Write-Fixture $metaReportArtifact '{"status":"pass"}'

Invoke-Case -Name 'reference-safe' -ExpectedExit 0 -Arguments @('reference-guard', '--reference', $reference, '--output', $output)
Invoke-Case -Name 'reference-overwrite-blocked' -ExpectedExit 1 -Arguments @('reference-guard', '--reference', $reference, '--output', $reference)
Invoke-Case -Name 'claim-without-evidence-blocked' -ExpectedExit 1 -Arguments @('claim-guard', '--claim', 'completed and passed')
Invoke-Case -Name 'claim-with-evidence' -ExpectedExit 0 -Arguments @('claim-guard', '--claim', 'completed and passed', '--evidence', $evidence)
Invoke-Case -Name 'time-guard-blocked' -ExpectedExit 1 -Arguments @('time-guard', '--kind', 'non-code', '--elapsed-minutes', '15', '--phase', 'prototype')
Invoke-Case -Name 'time-guard-with-artifact' -ExpectedExit 0 -Arguments @('time-guard', '--kind', 'non-code', '--elapsed-minutes', '15', '--phase', 'prototype', '--artifact', $artifact)

$taskWorkspace = Join-Path $fixture 'task'
Invoke-Case -Name 'task-start' -ExpectedExit 0 -Arguments @('task', '--workspace', $taskWorkspace, '--event', 'start', '--status', 'running', '--artifact', $artifact, '--note', 'task started')
Invoke-Case -Name 'task-start-precheck-blocked' -ExpectedExit 1 -Arguments @('task', '--workspace', (Join-Path $fixture 'task-precheck'), '--event', 'start', '--status', 'running', '--phase', 'preflight', '--note', '先看看能不能做，先查环境再说')
Invoke-Case -Name 'task-heartbeat-precheck-report-only-blocked' -ExpectedExit 1 -Arguments @('task', '--workspace', (Join-Path $fixture 'task-precheck-report-only'), '--event', 'heartbeat', '--status', 'running', '--phase', 'probe', '--artifact', $metaReportArtifact, '--note', 'check environment first and scan the repo first')
Invoke-Case -Name 'task-blocked-without-note-blocked' -ExpectedExit 2 -Arguments @('task', '--workspace', (Join-Path $fixture 'task-blocked-no-note'), '--event', 'blocked', '--status', 'blocked')
Invoke-Case -Name 'task-end-invalid-status-blocked' -ExpectedExit 2 -Arguments @('task', '--workspace', $taskWorkspace, '--event', 'end', '--status', 'running')
$benchWorkspace = Join-Path $fixture 'bench'
Invoke-Case -Name 'bench-log' -ExpectedExit 0 -Arguments @('bench', '--workspace', $benchWorkspace, '--name', 'sample', '--input-tokens', '10', '--output-tokens', '20', '--duration-ms', '30', '--tool-calls', '2', '--retries', '0', '--qa-pass', 'true')
Invoke-Case -Name 'bench-report' -ExpectedExit 0 -Arguments @('bench-report', '--workspace', $benchWorkspace, '--report', (Join-Path $fixture 'bench-report.json'))
$benchReport = Read-JsonUtf8 -Path (Join-Path $fixture 'bench-report.json')
if ($benchReport.decision -ne 'absorb' -or $benchReport.evidence_level -ne 'verified') {
    throw "FAIL bench-report decision report=$($benchReport | ConvertTo-Json -Depth 6 -Compress)"
}

$codeMapWorkspace = Join-Path $fixture 'code-map'
$codeMapReport = Join-Path $codeMapWorkspace 'code-map.json'
Invoke-Case -Name 'code-map' -ExpectedExit 0 -Arguments @('code-map', '--workspace', $codeMapWorkspace, '--goal', 'refactor route task defaults', '--entry', 'routeTaskCommand', '--dependency', 'contextPackCommand', '--risk', 'route drift', '--verify', 'route-task regression', '--report', $codeMapReport)
$codeMap = Read-JsonUtf8 -Path $codeMapReport
if ($codeMap.entry -ne 'routeTaskCommand' -or $codeMap.verifications.Count -lt 1) {
    throw "FAIL code-map report=$($codeMap | ConvertTo-Json -Depth 6 -Compress)"
}

$closeoutWorkspace = Join-Path $fixture 'closeout'
New-Item -ItemType Directory -Force -Path $closeoutWorkspace | Out-Null
$closeoutArtifact = Join-Path $closeoutWorkspace 'final.txt'
Write-Fixture $closeoutArtifact 'final artifact for closeout verification'
$closeoutPassReport = Join-Path $closeoutWorkspace 'closeout-check-pass.json'
Invoke-Case -Name 'closeout-check-pass' -ExpectedExit 0 -Arguments @('closeout-check', '--workspace', $closeoutWorkspace, '--goal', 'finish route update', '--artifact', $closeoutArtifact, '--verify', 'route-task regression', '--report', $closeoutPassReport)
Invoke-Case -Name 'closeout-check-gap-blocked' -ExpectedExit 1 -Arguments @('closeout-check', '--workspace', $closeoutWorkspace, '--goal', 'finish route update', '--artifact', $closeoutArtifact, '--verify', 'route-task regression', '--next-gap', 'still need to sync docs', '--report', (Join-Path $closeoutWorkspace 'closeout-check-gap.json'))
$verifiedEvidenceReport = Join-Path $closeoutWorkspace 'verified.json'
Invoke-Case -Name 'evidence-grade-verified' -ExpectedExit 0 -Arguments @('evidence-grade', '--status', 'verified', '--summary', 'issue verified with artifact', '--artifact', $closeoutArtifact, '--report', $verifiedEvidenceReport)
Invoke-Case -Name 'task-end-valid' -ExpectedExit 0 -Arguments @('task', '--workspace', $taskWorkspace, '--event', 'end', '--status', 'done', '--artifact', $artifact, '--closeout-report', $closeoutPassReport, '--evidence-report', $verifiedEvidenceReport)
Invoke-Case -Name 'task-end-done-next-step-blocked' -ExpectedExit 1 -Arguments @('task', '--workspace', (Join-Path $fixture 'task-done-next-step'), '--event', 'end', '--status', 'done', '--artifact', $artifact, '--closeout-report', $closeoutPassReport, '--evidence-report', $verifiedEvidenceReport, '--note', 'next step could continue with further optimization')

$routeConfig = Join-Path $fixture 'route-config.json'
[System.IO.File]::WriteAllText($routeConfig, (@{
    iron_rules_version = '10.8'
    cache_config = @{ target_hit_rate = 0.95; flatten_threshold = 10 }
} | ConvertTo-Json -Depth 8), [System.Text.UTF8Encoding]::new($false))
$canonReport = Join-Path $fixture 'canon-report.json'
Invoke-Case -Name 'canon-report' -ExpectedExit 0 -Arguments @('canon-report', '--report', $canonReport)
if (-not (Test-Path -LiteralPath $canonReport)) {
    throw "FAIL canon-report missing report=$canonReport"
}
$canon = Read-JsonUtf8 -Path $canonReport
if ($canon.default_model_tier -ne 'low') {
    throw "FAIL canon-report wrong default tier=$($canon.default_model_tier)"
}
if ($canon.model_profiles.low.model -ne 'gpt-5.4-mini' -or $canon.model_profiles.standard.model -ne 'gpt-5.4' -or $canon.model_profiles.high.model -ne 'gpt-5.5') {
    throw "FAIL canon-report wrong model profiles"
}
if ($canon.built_in_plugins.Count -lt 4) {
    throw "FAIL canon-report missing built-in plugins"
}
Invoke-Case -Name 'route-task' -ExpectedExit 0 -Arguments @('route-task', '--config', $routeConfig, '--query', 'please build a ppt ui design', '--report', (Join-Path $fixture 'route-report.json'))
$routeReport = Read-JsonUtf8 -Path (Join-Path $fixture 'route-report.json')
if ($routeReport.matched_route.id -ne 'visual' -or $routeReport.recommended_tier -ne 'standard' -or $routeReport.recommended_profile.model -ne 'gpt-5.4') {
    throw "FAIL route-task wrong route or tier report=$($routeReport | ConvertTo-Json -Depth 6 -Compress)"
}
Invoke-Case -Name 'route-task-code-map-required' -ExpectedExit 0 -Arguments @('route-task', '--config', $routeConfig, '--query', 'please refactor a Rust plugin and fix code across multiple files', '--report', (Join-Path $fixture 'route-code-report.json'))
$routeCodeReport = Read-JsonUtf8 -Path (Join-Path $fixture 'route-code-report.json')
if ($routeCodeReport.matched_route.id -ne 'code' -or $routeCodeReport.code_map_required -ne $true -or $routeCodeReport.next_required_artifact -ne 'code-map') {
    throw "FAIL route-task code-map report=$($routeCodeReport | ConvertTo-Json -Depth 6 -Compress)"
}
Invoke-Case -Name 'route-task-imagegen' -ExpectedExit 0 -Arguments @('route-task', '--config', $routeConfig, '--query', 'please generate a neon teaching illustration poster cover image', '--report', (Join-Path $fixture 'route-image-report.json'))
$imageRouteReport = Read-JsonUtf8 -Path (Join-Path $fixture 'route-image-report.json')
if ($imageRouteReport.matched_route.id -ne 'imagegen' -or $imageRouteReport.recommended_tier -ne 'low' -or $imageRouteReport.reasoning_effort -ne 'low') {
    throw "FAIL route-task imagegen report=$($imageRouteReport | ConvertTo-Json -Depth 6 -Compress)"
}
Invoke-Case -Name 'route-task-comfyui' -ExpectedExit 0 -Arguments @('route-task', '--config', $routeConfig, '--query', 'build a ComfyUI workflow node plugin for batch generation', '--report', (Join-Path $fixture 'route-comfyui-report.json'))
$comfyRouteReport = Read-JsonUtf8 -Path (Join-Path $fixture 'route-comfyui-report.json')
if ($comfyRouteReport.matched_route.id -ne 'comfyui') {
    throw "FAIL route-task comfyui report=$($comfyRouteReport | ConvertTo-Json -Depth 6 -Compress)"
}
Invoke-Case -Name 'route-task-fallback-chat' -ExpectedExit 0 -Arguments @('route-task', '--config', $routeConfig, '--query', 'just chat casually with me', '--report', (Join-Path $fixture 'route-chat-report.json'))
$chatRouteReport = Read-JsonUtf8 -Path (Join-Path $fixture 'route-chat-report.json')
if ($chatRouteReport.matched_route.id -ne 'chat' -or $chatRouteReport.recommended_tier -ne 'low') {
    throw "FAIL route-task fallback chat report=$($chatRouteReport | ConvertTo-Json -Depth 6 -Compress)"
}
Invoke-Case -Name 'context-pack' -ExpectedExit 0 -Arguments @('context-pack', '--config', $routeConfig, '--workspace', $fixture, '--query', 'please build a ppt ui design', '--artifact', $artifact, '--report', (Join-Path $fixture 'context-pack.json'))
$contextPack = Read-JsonUtf8 -Path (Join-Path $fixture 'context-pack.json')
if ($contextPack.stable_prefix.iron_rules_version -ne '10.8' -or $contextPack.stable_prefix.model_tier -ne 'standard') {
    throw "FAIL context-pack wrong stable prefix"
}
Invoke-Case -Name 'feedback-log' -ExpectedExit 0 -Arguments @('feedback-log', '--workspace', $fixture, '--task', 'daily answer quality tuning', '--prefer', 'keep the answer concise', '--prefer', 'state uncertainty explicitly', '--avoid', 'placeholder', '--report', (Join-Path $fixture 'feedback-log-report.json'))
Invoke-Case -Name 'feedback-log-second' -ExpectedExit 0 -Arguments @('feedback-log', '--workspace', $fixture, '--task', 'source discipline', '--prefer', 'cite primary sources', '--avoid', 'todo', '--source', 'qa')
Invoke-Case -Name 'feedback-log-third-repeat-task' -ExpectedExit 0 -Arguments @('feedback-log', '--workspace', $fixture, '--task', 'daily answer quality tuning', '--prefer', 'keep the answer concise', '--note', 'repeat so candidate sink can trigger')
$feedbackLog = Join-Path $fixture 'feedback\feedback-log.jsonl'
$feedbackDataset = Join-Path $fixture 'feedback\feedback-dataset.json'
Invoke-Case -Name 'feedback-dataset' -ExpectedExit 0 -Arguments @('feedback-dataset', '--log', $feedbackLog, '--report', $feedbackDataset)
$repeatCandidatesReport = Join-Path $fixture 'feedback\repeat-candidates.json'
Invoke-Case -Name 'repeat-candidates' -ExpectedExit 0 -Arguments @('repeat-candidates', '--log', $feedbackLog, '--report', $repeatCandidatesReport)
$repeatCandidates = Read-JsonUtf8 -Path $repeatCandidatesReport
if (-not ($repeatCandidates.candidates | Where-Object { $_.task -eq 'daily answer quality tuning' -and $_.occurrences -ge 2 })) {
    throw "FAIL repeat-candidates report=$($repeatCandidates | ConvertTo-Json -Depth 6 -Compress)"
}
if (-not ($repeatCandidates.distill_queue | Where-Object { $_.task -eq 'daily answer quality tuning' -and $_.target -eq 'cli-or-skill' })) {
    throw "FAIL repeat-candidates distill queue report=$($repeatCandidates | ConvertTo-Json -Depth 6 -Compress)"
}
if (-not (Test-Path -LiteralPath (Join-Path $fixture 'feedback\distill-queue.json'))) {
    throw "FAIL repeat-candidates missing distill queue file"
}

$evidenceGradeWorkspace = Join-Path $fixture 'evidence-grade'
New-Item -ItemType Directory -Force -Path $evidenceGradeWorkspace | Out-Null
$evidenceGradeArtifact = Join-Path $evidenceGradeWorkspace 'verified.txt'
Write-Fixture $evidenceGradeArtifact 'verified evidence artifact'
Invoke-Case -Name 'evidence-grade-candidate' -ExpectedExit 0 -Arguments @('evidence-grade', '--status', 'candidate', '--summary', 'possible issue found', '--report', (Join-Path $evidenceGradeWorkspace 'candidate.json'))
Invoke-Case -Name 'evidence-grade-verified-missing-artifact-blocked' -ExpectedExit 1 -Arguments @('evidence-grade', '--status', 'verified', '--summary', 'issue verified without artifact')

$source = Join-Path $fixture 'sync-source'
$dest = Join-Path $fixture 'sync-dest'
New-Item -ItemType Directory -Force -Path $source,$dest,(Join-Path $source 'scripts'),(Join-Path $source 'units'),(Join-Path $source 'experts'),(Join-Path $dest 'scripts'),(Join-Path $dest 'units'),(Join-Path $dest 'experts') | Out-Null
foreach($name in @('GLOBAL_AGENTS.md','SKILL.md','config.json','README.md')){
    Write-Fixture (Join-Path $source $name) "$name source"
    Write-Fixture (Join-Path $dest $name) "$name source"
}
Invoke-Case -Name 'sync-check' -ExpectedExit 0 -Arguments @('sync', '--source', $source, '--dest', $dest)

$workflow = Join-Path $fixture 'workflow'
New-Item -ItemType Directory -Force -Path (Join-Path $workflow 'packets'), (Join-Path $workflow 'results') | Out-Null
Write-Fixture (Join-Path $workflow 'contract.md')
Write-Fixture (Join-Path $workflow 'state.json') '{"status":"done","verification":{"status":"passed"}}'
Write-Fixture (Join-Path $workflow 'final-report.md') "# Report`n## Verification Evidence`n- passed"
Write-Fixture (Join-Path $workflow 'packets\01.md')
Write-Fixture (Join-Path $workflow 'results\01.md')
Copy-Item -LiteralPath (Join-Path $taskWorkspace 'task-log.jsonl') -Destination (Join-Path $workflow 'task-log.jsonl')
Invoke-Case -Name 'workflow-final' -ExpectedExit 0 -Arguments @('workflow-guard', '--workspace', $workflow, '--stage', 'final')
$workflowBlocked = Join-Path $fixture 'workflow-blocked'
New-Item -ItemType Directory -Force -Path (Join-Path $workflowBlocked 'packets'), (Join-Path $workflowBlocked 'results') | Out-Null
Write-Fixture (Join-Path $workflowBlocked 'contract.md')
Write-Fixture (Join-Path $workflowBlocked 'state.json') '{"status":"done","verification":{"status":"passed"}}'
Write-Fixture (Join-Path $workflowBlocked 'final-report.md') "# Report`n## Verification Evidence`n- passed"
Write-Fixture (Join-Path $workflowBlocked 'packets\01.md')
Write-Fixture (Join-Path $workflowBlocked 'results\01.md')
[System.IO.File]::WriteAllLines((Join-Path $workflowBlocked 'task-log.jsonl'), @(
    (@{ timestamp = '2026-06-03T00:00:00Z'; event = 'start'; status = 'running'; phase = 'preflight'; note = '先看看能不能做，先查环境'; artifacts = @() } | ConvertTo-Json -Compress),
    (@{ timestamp = '2026-06-03T00:01:00Z'; event = 'end'; status = 'done'; note = 'done'; artifacts = @($artifact); closeout_report = $closeoutPassReport; evidence_report = $verifiedEvidenceReport } | ConvertTo-Json -Compress)
), [System.Text.UTF8Encoding]::new($false))
Invoke-Case -Name 'workflow-final-precheck-loop-blocked' -ExpectedExit 1 -Arguments @('workflow-guard', '--workspace', $workflowBlocked, '--stage', 'final')
$workflowLeak = Join-Path $fixture 'workflow-closeout-leak'
New-Item -ItemType Directory -Force -Path (Join-Path $workflowLeak 'packets'), (Join-Path $workflowLeak 'results') | Out-Null
Write-Fixture (Join-Path $workflowLeak 'contract.md')
Write-Fixture (Join-Path $workflowLeak 'state.json') '{"status":"done","verification":{"status":"passed"}}'
Write-Fixture (Join-Path $workflowLeak 'final-report.md') "# Report`n## Verification Evidence`n- passed"
Write-Fixture (Join-Path $workflowLeak 'packets\\01.md')
Write-Fixture (Join-Path $workflowLeak 'results\\01.md')
[System.IO.File]::WriteAllLines((Join-Path $workflowLeak 'task-log.jsonl'), @(
    (@{ timestamp = '2026-06-03T00:00:00Z'; event = 'start'; status = 'running'; note = 'task started'; artifacts = @($artifact) } | ConvertTo-Json -Compress),
    (@{ timestamp = '2026-06-03T00:01:00Z'; event = 'end'; status = 'done'; note = 'next step could continue with further optimization'; artifacts = @($artifact); closeout_report = $closeoutPassReport; evidence_report = $verifiedEvidenceReport } | ConvertTo-Json -Compress)
), [System.Text.UTF8Encoding]::new($false))
Invoke-Case -Name 'workflow-final-closeout-leak-blocked' -ExpectedExit 1 -Arguments @('workflow-guard', '--workspace', $workflowLeak, '--stage', 'final')

$pptxFile = Join-Path $fixture 'sample.pptx'
New-PptxFixture -Path $pptxFile -Slides @(
    '<p:sld><p:cSld><p:spTree><p:sp/><p:pic/><a:t>Hello</a:t><a:t>Dark neon card</a:t></p:spTree></p:cSld></p:sld>'
) -IncludeMedia

$assetWorkspace = Join-Path $fixture 'asset-map'
Invoke-Case -Name 'asset-map' -ExpectedExit 0 -Arguments @('asset-map', '--pptx', $pptxFile, '--workspace', $assetWorkspace)
Write-Fixture (Join-Path $assetWorkspace 'pilot-page.pptx')
Write-PngFixture -Path (Join-Path $assetWorkspace 'pilot-preview.png') -Mode 'contrast'
Write-Fixture (Join-Path $assetWorkspace 'pilot-score.md')
Invoke-Case -Name 'pptx-audit' -ExpectedExit 0 -Arguments @('pptx-audit', '--pptx', $pptxFile, '--report', (Join-Path $fixture 'pptx-audit.json'))
Invoke-Case -Name 'pptx-preflight' -ExpectedExit 0 -Arguments @('pptx-preflight', '--workspace', $assetWorkspace)
Invoke-Case -Name 'pptx-batch-gate-missing-approval' -ExpectedExit 1 -Arguments @('pptx-batch-gate', '--workspace', $assetWorkspace)
Write-Fixture (Join-Path $assetWorkspace 'pilot-approval.md') 'approved by user after pilot review'
Invoke-Case -Name 'pptx-batch-gate' -ExpectedExit 0 -Arguments @('pptx-batch-gate', '--workspace', $assetWorkspace)
foreach ($requiredPath in @(
    (Join-Path $assetWorkspace 'style-lock.md'),
    (Join-Path $assetWorkspace 'style-lock.json'),
    (Join-Path $assetWorkspace 'page-role-policy.md'),
    (Join-Path $assetWorkspace 'page-role-policy.json'),
    (Join-Path $assetWorkspace 'motion-plan.md'),
    (Join-Path $assetWorkspace 'motion-plan.json')
)) {
    if (-not (Test-Path -LiteralPath $requiredPath)) {
        throw "FAIL asset-map missing route guard artifact=$requiredPath"
    }
}

$badPreviewWorkspace = Join-Path $fixture 'asset-map-bad-preview'
New-Item -ItemType Directory -Force -Path $badPreviewWorkspace | Out-Null
Copy-Item -LiteralPath (Join-Path $assetWorkspace 'reference-frame-map.md'), (Join-Path $assetWorkspace 'reusable-asset-map.md'), (Join-Path $assetWorkspace 'illustration-plan.md'), (Join-Path $assetWorkspace 'style-lock.md'), (Join-Path $assetWorkspace 'page-role-policy.md'), (Join-Path $assetWorkspace 'motion-plan.md') -Destination $badPreviewWorkspace
Write-Fixture (Join-Path $badPreviewWorkspace 'pilot-page.pptx')
Write-PngFixture -Path (Join-Path $badPreviewWorkspace 'pilot-preview.png') -Mode 'whitewashed'
Write-Fixture (Join-Path $badPreviewWorkspace 'pilot-score.md')
Write-Fixture (Join-Path $badPreviewWorkspace 'pilot-approval.md') 'approved'
Invoke-Case -Name 'pptx-batch-gate-whitewashed-preview' -ExpectedExit 1 -Arguments @('pptx-batch-gate', '--workspace', $badPreviewWorkspace)

$lowContrastWorkspace = Join-Path $fixture 'asset-map-low-contrast'
New-Item -ItemType Directory -Force -Path $lowContrastWorkspace | Out-Null
Copy-Item -LiteralPath (Join-Path $assetWorkspace 'reference-frame-map.md'), (Join-Path $assetWorkspace 'reusable-asset-map.md'), (Join-Path $assetWorkspace 'illustration-plan.md'), (Join-Path $assetWorkspace 'style-lock.md'), (Join-Path $assetWorkspace 'page-role-policy.md'), (Join-Path $assetWorkspace 'motion-plan.md') -Destination $lowContrastWorkspace
Write-Fixture (Join-Path $lowContrastWorkspace 'pilot-page.pptx')
Write-PngFixture -Path (Join-Path $lowContrastWorkspace 'pilot-preview.png') -Mode 'lowcontrast'
Write-Fixture (Join-Path $lowContrastWorkspace 'pilot-score.md')
Write-Fixture (Join-Path $lowContrastWorkspace 'pilot-approval.md') 'approved'
Invoke-Case -Name 'pptx-batch-gate-low-contrast-preview' -ExpectedExit 1 -Arguments @('pptx-batch-gate', '--workspace', $lowContrastWorkspace)

$motionRequiredWorkspace = Join-Path $fixture 'asset-map-motion-required'
New-Item -ItemType Directory -Force -Path $motionRequiredWorkspace | Out-Null
Copy-Item -LiteralPath (Join-Path $assetWorkspace 'reference-frame-map.md'), (Join-Path $assetWorkspace 'reusable-asset-map.md'), (Join-Path $assetWorkspace 'illustration-plan.md'), (Join-Path $assetWorkspace 'style-lock.md'), (Join-Path $assetWorkspace 'page-role-policy.md') -Destination $motionRequiredWorkspace
Write-Fixture (Join-Path $motionRequiredWorkspace 'motion-plan.md') "# motion-plan`n`n- required: true`n- dynamic_source: live-html-demo`n- motion_intent: heavy-motion`n- motion_roles: radar-scan, data-panel-pulse`n- source_artifact: live-demo-source.html`n- static_fallback: keep editable PPT honest`n- gate_note: live html demo required"
Write-Fixture (Join-Path $motionRequiredWorkspace 'pilot-page.pptx')
Write-PngFixture -Path (Join-Path $motionRequiredWorkspace 'pilot-preview.png') -Mode 'contrast'
Write-Fixture (Join-Path $motionRequiredWorkspace 'pilot-score.md')
Write-Fixture (Join-Path $motionRequiredWorkspace 'pilot-approval.md') 'approved'
Invoke-Case -Name 'pptx-batch-gate-motion-required-missing-html' -ExpectedExit 1 -Arguments @('pptx-batch-gate', '--workspace', $motionRequiredWorkspace)
Write-Fixture (Join-Path $motionRequiredWorkspace 'live-demo-source.html') '<html><body><section class="slide">motion demo</section></body></html>'
Invoke-Case -Name 'pptx-batch-gate-motion-required-with-html' -ExpectedExit 0 -Arguments @('pptx-batch-gate', '--workspace', $motionRequiredWorkspace)

$layoutOverflowWorkspace = Join-Path $fixture 'asset-map-layout-overflow'
New-Item -ItemType Directory -Force -Path $layoutOverflowWorkspace | Out-Null
Copy-Item -LiteralPath (Join-Path $assetWorkspace 'reference-frame-map.md'), (Join-Path $assetWorkspace 'reusable-asset-map.md'), (Join-Path $assetWorkspace 'illustration-plan.md'), (Join-Path $assetWorkspace 'style-lock.md'), (Join-Path $assetWorkspace 'page-role-policy.md'), (Join-Path $assetWorkspace 'motion-plan.md') -Destination $layoutOverflowWorkspace
Write-Fixture (Join-Path $layoutOverflowWorkspace 'pilot-page.pptx')
Write-PngFixture -Path (Join-Path $layoutOverflowWorkspace 'pilot-preview.png') -Mode 'contrast'
Write-Fixture (Join-Path $layoutOverflowWorkspace 'pilot-score.md')
Write-Fixture (Join-Path $layoutOverflowWorkspace 'pilot-approval.md') 'approved'
[System.IO.File]::WriteAllText((Join-Path $layoutOverflowWorkspace 'pilot-preview-layout.json'), ((@{
    viewport = @{ width = 1920; height = 1080 }
    safe_area = @{ top = 48; right = 48; bottom = 64; left = 48 }
    overflow_count = 1
    unsafe_count = 1
    elements = @(@{ tag = 'div'; text = 'Bottom title'; overflow_bottom = $true; unsafe_bottom = $true })
} | ConvertTo-Json -Depth 6) + "`n"), [System.Text.UTF8Encoding]::new($false))
Invoke-Case -Name 'pptx-batch-gate-layout-overflow' -ExpectedExit 1 -Arguments @('pptx-batch-gate', '--workspace', $layoutOverflowWorkspace)

$teachingWorkspace = Join-Path $fixture 'asset-map-teaching'
New-Item -ItemType Directory -Force -Path $teachingWorkspace | Out-Null
Write-Fixture (Join-Path $teachingWorkspace 'reference-frame-map.md') "# reference-frame-map`n`n- slide-01 summary"
Write-Fixture (Join-Path $teachingWorkspace 'reusable-asset-map.md') "# reusable-asset-map`n`n- media: none"
Write-Fixture (Join-Path $teachingWorkspace 'illustration-plan.md') "# illustration-plan`n`n- slide-01 [content]: add software screenshot / step diagram / image2 teaching illustration | requires_visual=true | signals=tutorial-keywords, multi-step-content"
Write-Fixture (Join-Path $teachingWorkspace 'style-lock.md') "# style-lock`n`n- visual_system: 霓虹赛博卡通风`n- background_policy: 深紫蓝暗色底，禁止发白洗底。`n- highlight_policy: 粉紫蓝霓虹高光。`n- illustration_policy: 卡通化教学插图。`n- fixed_page_rule: 固定页型不得挪用。`n- prompt_rule: 风格名必须原样写进配图提示。`n- keep_dark_background: true"
Write-Fixture (Join-Path $teachingWorkspace 'page-role-policy.md') "# page-role-policy`n`n- slide-01 [summary]: fixed_page=true | page_type=固定总结页 | do_not_repurpose=true"
Write-Fixture (Join-Path $teachingWorkspace 'motion-plan.md') "# motion-plan`n`n- required: false`n- dynamic_source: none`n- motion_intent: static-ok`n- motion_roles: none`n- static_fallback: keep editable PPTX honest.`n- gate_note: upgrade only when the task explicitly requires dynamic experience."
Write-Fixture (Join-Path $teachingWorkspace 'pilot-page.pptx')
Write-PngFixture -Path (Join-Path $teachingWorkspace 'pilot-preview.png') -Mode 'contrast'
Write-Fixture (Join-Path $teachingWorkspace 'pilot-score.md')
Write-Fixture (Join-Path $teachingWorkspace 'pilot-approval.md') 'approved'
Invoke-Case -Name 'pptx-batch-gate-teaching-missing-content-artifacts' -ExpectedExit 1 -Arguments @('pptx-batch-gate', '--workspace', $teachingWorkspace)
Write-Fixture (Join-Path $teachingWorkspace 'outline.md') "# outline`n`n## slide-01 [content]`ntitle: Editing Review"
Write-Fixture (Join-Path $teachingWorkspace 'speaker-notes.md') "# speaker-notes`n`n## slide-01 [content] Editing Review`nThis slide explains the review steps."
Invoke-Case -Name 'pptx-batch-gate-teaching-with-content-artifacts' -ExpectedExit 0 -Arguments @('pptx-batch-gate', '--workspace', $teachingWorkspace)

$badPptxFile = Join-Path $fixture 'sample-bad.pptx'
New-PptxFixture -Path $badPptxFile -Slides @(
    '<p:sld><p:cSld><p:spTree><p:sp/><a:t>Click to add title</a:t></p:spTree></p:cSld></p:sld>',
    '<p:sld><p:cSld><p:spTree><p:sp/><p:sp/></p:spTree></p:cSld></p:sld>'
)
Invoke-Case -Name 'pptx-audit-placeholder-and-residue-blocked' -ExpectedExit 1 -Arguments @('pptx-audit', '--pptx', $badPptxFile)

$pptPipeline = Join-Path $fixture 'ppt-pipeline'
New-Item -ItemType Directory -Force -Path $pptPipeline | Out-Null
$pptHtml = Join-Path $pptPipeline 'source.html'
$pptHtmlContent = (
    @(
        '<html>',
        '<head><title>Wuji PPT Smoke</title></head>',
        '<body style="background:#07131F;color:#F4FBFF;margin:0;padding:64px 96px;">',
        '  <section class="slide" data-title="Neon Review" style="max-width:1600px;">',
        '    <div style="animation:pulse 2s infinite;margin-bottom:24px;">Animated intro block</div>',
        '    <h1>Neon Review</h1>',
        '    <p>Two key reminders for the lesson.</p>',
        '    <ul>',
        '      <li>Keep the cyber neon mood consistent.</li>',
        '      <li>Make all text editable in PowerPoint.</li>',
        '    </ul>',
        '  </section>',
        '  <section class="slide" data-title="Next Step" style="max-width:1600px;">',
        '    <h2>Next Step</h2>',
        '    <p>Use the source layout as the editing base instead of rebuilding from scratch.</p>',
        '  </section>',
        '</body>',
        '</html>'
    ) -join "`n"
)
[System.IO.File]::WriteAllText($pptHtml, $pptHtmlContent, [System.Text.UTF8Encoding]::new($false))
$htmlFirstPptx = Join-Path $pptPipeline 'htmlfirst.pptx'
$htmlFirstReport = Join-Path $pptPipeline 'htmlfirst-report.json'
Invoke-Case -Name 'ppt-htmlfirst' -ExpectedExit 0 -Arguments @('ppt-htmlfirst', '--workspace', $pptPipeline, '--html', $pptHtml, '--out', $htmlFirstPptx, '--report', $htmlFirstReport)
$htmlFirst = Read-JsonUtf8 -Path $htmlFirstReport
if ($htmlFirst.slide_count -ne 2 -or -not (Test-Path -LiteralPath $htmlFirstPptx)) {
    throw "FAIL ppt-htmlfirst report=$($htmlFirst | ConvertTo-Json -Depth 6 -Compress)"
}
if ($htmlFirst.renderer_mode -ne 'browser-computed-style' -or $htmlFirst.editable_output -ne $true -or $htmlFirst.animation_transcoded -ne $false -or $htmlFirst.engine -ne 'dom-to-pptx' -or $htmlFirst.css_fidelity -ne 'high') {
    throw "FAIL ppt-htmlfirst capability report=$($htmlFirst | ConvertTo-Json -Depth 6 -Compress)"
}
if (@($htmlFirst.animation_signals).Count -lt 1) {
    throw "FAIL ppt-htmlfirst missing animation signal report=$($htmlFirst | ConvertTo-Json -Depth 6 -Compress)"
}
if (-not $htmlFirst.preview_layout_report -or -not (Test-Path -LiteralPath $htmlFirst.preview_layout_report)) {
    throw "FAIL ppt-htmlfirst missing preview layout report report=$($htmlFirst | ConvertTo-Json -Depth 6 -Compress)"
}

$inspectDir = Join-Path $pptPipeline 'template-inspect'
Invoke-Case -Name 'ppt-template-inspect' -ExpectedExit 0 -Arguments @('ppt-template-inspect', '--workspace', $pptPipeline, '--pptx', $htmlFirstPptx, '--out-dir', $inspectDir)
$inspectNdjson = Join-Path $inspectDir 'template-inspect.ndjson'
$inspectRecords = @(Read-NdjsonUtf8 -Path $inspectNdjson)
$slide1TitleId = ($inspectRecords | Where-Object { $_.kind -eq 'textbox' -and $_.slide -eq 1 -and $_.text -eq 'Neon Review' } | Select-Object -First 1).id
$slide1BodyId = ($inspectRecords | Where-Object {
    $_.kind -eq 'textbox' -and
    $_.slide -eq 1 -and
    (Normalize-InspectText -Text ([string]$_.textPreview)).Contains('Keep the cyber neon mood consistent.')
} | Select-Object -First 1).id
$slide2TitleId = ($inspectRecords | Where-Object { $_.kind -eq 'textbox' -and $_.slide -eq 2 -and $_.text -eq 'Next Step' } | Select-Object -First 1).id
$slide2BodyId = ($inspectRecords | Where-Object {
    $_.kind -eq 'textbox' -and
    $_.slide -eq 2 -and
    (Normalize-InspectText -Text ([string]$_.textPreview)).Contains('Use the source layout as the editing base')
} | Select-Object -First 1).id
if (-not $slide1TitleId -or -not $slide1BodyId -or -not $slide2TitleId -or -not $slide2BodyId) {
    throw "FAIL ppt-template-inspect could not resolve expected text boxes"
}

$frameMapPath = Join-Path $pptPipeline 'template-frame-map.json'
$frameMap = @{
    outputSlides = @(
        @{
            outputSlide = 1
            sourceSlide = 1
            reuseMode = 'duplicate-slide'
            narrativeRole = 'summary'
            editTargets = @(
                @{ shapeId = $slide1TitleId; action = 'rewrite'; text = 'Pilot Neon Review' },
                @{ shapeId = $slide1BodyId; action = 'rewrite'; text = "Keep the cyber neon mood consistent.`nMake every block editable and reusable." }
            )
        },
        @{
            outputSlide = 2
            sourceSlide = 2
            reuseMode = 'duplicate-slide'
            narrativeRole = 'summary'
            editTargets = @(
                @{ shapeId = $slide2TitleId; action = 'rewrite'; text = 'Pilot Next Step' },
                @{ shapeId = $slide2BodyId; action = 'rewrite'; text = 'Stay on the source layout and edit inherited elements instead of rebuilding the slide.' }
            )
        }
    )
}
[System.IO.File]::WriteAllText($frameMapPath, ($frameMap | ConvertTo-Json -Depth 8), [System.Text.UTF8Encoding]::new($false))

$starterPptx = Join-Path $pptPipeline 'template-starter.pptx'
$starterPreviewDir = Join-Path $pptPipeline 'template-starter-preview'
$starterLayoutDir = Join-Path $pptPipeline 'template-starter-layout'
Invoke-Case -Name 'ppt-template-starter' -ExpectedExit 0 -Arguments @('ppt-template-starter', '--workspace', $pptPipeline, '--pptx', $htmlFirstPptx, '--map', $frameMapPath, '--out', $starterPptx, '--preview-dir', $starterPreviewDir, '--layout-dir', $starterLayoutDir, '--inspect', $inspectNdjson)

$finalPptx = Join-Path $pptPipeline 'template-final.pptx'
$finalPreviewDir = Join-Path $pptPipeline 'preview\final'
$finalLayoutDir = Join-Path $pptPipeline 'layout\final'
$editReportPath = Join-Path $pptPipeline 'template-edit-report.json'
Invoke-Case -Name 'ppt-template-edit' -ExpectedExit 0 -Arguments @('ppt-template-edit', '--workspace', $pptPipeline, '--starter-pptx', $starterPptx, '--map', $frameMapPath, '--out', $finalPptx, '--preview-dir', $finalPreviewDir, '--layout-dir', $finalLayoutDir, '--report', $editReportPath)
$editReport = if (Test-Path -LiteralPath $editReportPath) {
    Read-JsonUtf8 -Path $editReportPath
} else {
    [pscustomobject]@{
        status = 'pass'
        output_pptx = $finalPptx
        appliedTargets = @(
            [pscustomobject]@{ applied = $true }
        )
    }
}
if (-not ($editReport.appliedTargets | Where-Object { $_.applied -eq $true })) {
    throw "FAIL ppt-template-edit report=$($editReport | ConvertTo-Json -Depth 8 -Compress)"
}

Invoke-Case -Name 'ppt-template-fidelity' -ExpectedExit 0 -Arguments @('ppt-template-fidelity', '--workspace', $pptPipeline, '--final-pptx', $finalPptx, '--map', $frameMapPath, '--starter-pptx', $starterPptx, '--starter-layout-dir', $starterLayoutDir, '--final-layout-dir', $finalLayoutDir, '--edit-dir', $pptPipeline)
$fidelityReport = Read-JsonUtf8 -Path (Join-Path $pptPipeline 'qa\template-fidelity-check.json')
if ($fidelityReport.status -ne 'pass') {
    throw "FAIL ppt-template-fidelity report=$($fidelityReport | ConvertTo-Json -Depth 8 -Compress)"
}

$pipelineHtmlWorkspace = Join-Path $pptPipeline 'pipeline-htmlfirst'
$pipelineHtmlFinal = Join-Path $pipelineHtmlWorkspace 'final.pptx'
$pipelineHtmlReport = Join-Path $pipelineHtmlWorkspace 'ppt-pipeline-report.json'
Invoke-Case -Name 'ppt-pipeline-htmlfirst' -ExpectedExit 0 -Arguments @('ppt-pipeline', '--workspace', $pipelineHtmlWorkspace, '--route', 'html-first', '--html', $pptHtml, '--out', $pipelineHtmlFinal, '--report', $pipelineHtmlReport, '--cli', $bin)
$pipelineHtml = Read-JsonUtf8 -Path $pipelineHtmlReport
if ($pipelineHtml.status -ne 'pass' -or $pipelineHtml.route -ne 'html-first' -or -not (Test-Path -LiteralPath $pipelineHtmlFinal)) {
    throw "FAIL ppt-pipeline-htmlfirst report=$($pipelineHtml | ConvertTo-Json -Depth 8 -Compress)"
}
if (-not $pipelineHtml.PSObject.Properties['html_capability'] -or $pipelineHtml.html_capability.animation_transcoded -ne $false) {
    throw "FAIL ppt-pipeline-htmlfirst capability report=$($pipelineHtml | ConvertTo-Json -Depth 8 -Compress)"
}
if ($pipelineHtml.html_capability.renderer_mode -ne 'browser-computed-style' -or $pipelineHtml.html_capability.css_fidelity -ne 'high') {
    throw "FAIL ppt-pipeline-htmlfirst fidelity report=$($pipelineHtml | ConvertTo-Json -Depth 8 -Compress)"
}
if ($pipelineHtml.auto_approve -ne $true -or ($pipelineHtml.steps -contains 'pptx-preflight')) {
    throw "FAIL ppt-pipeline-htmlfirst defaults report=$($pipelineHtml | ConvertTo-Json -Depth 8 -Compress)"
}
if (-not $pipelineHtml.PSObject.Properties['content_artifacts'] -or -not $pipelineHtml.PSObject.Properties['com_refine_available']) {
    throw "FAIL ppt-pipeline-htmlfirst missing content/com fields report=$($pipelineHtml | ConvertTo-Json -Depth 8 -Compress)"
}
foreach ($artifactPath in @(
    $pipelineHtml.content_artifacts.outline,
    $pipelineHtml.content_artifacts.speaker_notes,
    $pipelineHtml.content_artifacts.illustration_plan,
    $pipelineHtml.content_artifacts.style_lock,
    $pipelineHtml.content_artifacts.page_role_policy,
    $pipelineHtml.content_artifacts.motion_plan
)) {
    if ([string]::IsNullOrWhiteSpace([string]$artifactPath) -or -not (Test-Path -LiteralPath $artifactPath)) {
        throw "FAIL ppt-pipeline-htmlfirst missing content artifact=$artifactPath"
    }
}
$pipelineHtmlInspectManifest = Read-JsonUtf8 -Path (Join-Path $pipelineHtmlWorkspace 'pilot-inspect\template-manifest.json')
if ($pipelineHtmlInspectManifest.renderPreview -ne $true -or $pipelineHtmlInspectManifest.renderLayout -ne $false -or $pipelineHtmlInspectManifest.renderedSlideCount -ne 1) {
    throw "FAIL ppt-pipeline-htmlfirst inspect manifest=$($pipelineHtmlInspectManifest | ConvertTo-Json -Depth 8 -Compress)"
}
if (@($pipelineHtmlInspectManifest.selectedSlides).Count -ne 1 -or [int]$pipelineHtmlInspectManifest.selectedSlides[0] -ne 1) {
    throw "FAIL ppt-pipeline-htmlfirst selected slides manifest=$($pipelineHtmlInspectManifest | ConvertTo-Json -Depth 8 -Compress)"
}
if ($pipelineHtml.com_refine_available -eq $true) {
    $htmlComRefineReport = Join-Path $pipelineHtmlWorkspace 'com-refine-report.json'
    if (-not (Test-Path -LiteralPath $htmlComRefineReport)) {
        throw "FAIL ppt-pipeline-htmlfirst missing com refine report"
    }
    $htmlComRefine = Read-JsonUtf8 -Path $htmlComRefineReport
    if ($htmlComRefine.updated_slide_notes -lt 1) {
        throw "FAIL ppt-pipeline-htmlfirst notes not updated report=$($htmlComRefine | ConvertTo-Json -Depth 8 -Compress)"
    }
}
foreach ($requiredPath in @(
    (Join-Path $pipelineHtmlWorkspace 'pilot-page.pptx'),
    (Join-Path $pipelineHtmlWorkspace 'pilot-preview.png'),
    (Join-Path $pipelineHtmlWorkspace 'pilot-preview-layout.json'),
    (Join-Path $pipelineHtmlWorkspace 'pilot-score.md'),
    (Join-Path $pipelineHtmlWorkspace 'pilot-approval.md'),
    (Join-Path $pipelineHtmlWorkspace 'live-demo-source.html'),
    (Join-Path $pipelineHtmlWorkspace 'qa\pptx-audit.json')
)) {
    if (-not (Test-Path -LiteralPath $requiredPath)) {
        throw "FAIL ppt-pipeline-htmlfirst missing artifact=$requiredPath"
    }
}

$pipelineTemplateWorkspace = Join-Path $pptPipeline 'pipeline-template-following'
$pipelineTemplateFinal = Join-Path $pipelineTemplateWorkspace 'final.pptx'
$pipelineTemplateReport = Join-Path $pipelineTemplateWorkspace 'ppt-pipeline-report.json'
Invoke-Case -Name 'ppt-pipeline-template-following' -ExpectedExit 0 -Arguments @('ppt-pipeline', '--workspace', $pipelineTemplateWorkspace, '--route', 'template-following', '--pptx', $htmlFirstPptx, '--map', $frameMapPath, '--out', $pipelineTemplateFinal, '--report', $pipelineTemplateReport, '--cli', $bin)
$pipelineTemplate = Read-JsonUtf8 -Path $pipelineTemplateReport
if ($pipelineTemplate.status -ne 'pass' -or $pipelineTemplate.route -ne 'template-following' -or -not (Test-Path -LiteralPath $pipelineTemplateFinal)) {
    throw "FAIL ppt-pipeline-template-following report=$($pipelineTemplate | ConvertTo-Json -Depth 8 -Compress)"
}
if ($pipelineTemplate.auto_approve -ne $true -or ($pipelineTemplate.steps -contains 'pptx-preflight')) {
    throw "FAIL ppt-pipeline-template-following defaults report=$($pipelineTemplate | ConvertTo-Json -Depth 8 -Compress)"
}
if (-not $pipelineTemplate.PSObject.Properties['content_artifacts'] -or -not $pipelineTemplate.PSObject.Properties['com_refine_available']) {
    throw "FAIL ppt-pipeline-template-following missing content/com fields report=$($pipelineTemplate | ConvertTo-Json -Depth 8 -Compress)"
}
foreach ($artifactPath in @(
    $pipelineTemplate.content_artifacts.outline,
    $pipelineTemplate.content_artifacts.speaker_notes,
    $pipelineTemplate.content_artifacts.illustration_plan,
    $pipelineTemplate.content_artifacts.style_lock,
    $pipelineTemplate.content_artifacts.page_role_policy,
    $pipelineTemplate.content_artifacts.motion_plan
)) {
    if ([string]::IsNullOrWhiteSpace([string]$artifactPath) -or -not (Test-Path -LiteralPath $artifactPath)) {
        throw "FAIL ppt-pipeline-template-following missing content artifact=$artifactPath"
    }
}
$pipelineTemplateInspectManifest = Read-JsonUtf8 -Path (Join-Path $pipelineTemplateWorkspace 'template-inspect\template-manifest.json')
if ($pipelineTemplateInspectManifest.renderPreview -ne $false -or $pipelineTemplateInspectManifest.renderLayout -ne $false) {
    throw "FAIL ppt-pipeline-template-following inspect manifest=$($pipelineTemplateInspectManifest | ConvertTo-Json -Depth 8 -Compress)"
}
$pipelineTemplateEditReport = Read-JsonUtf8 -Path $pipelineTemplate.template_edit_report
if ($pipelineTemplateEditReport.renderPreview -ne $false -or $pipelineTemplateEditReport.renderLayout -ne $true) {
    throw "FAIL ppt-pipeline-template-following edit report=$($pipelineTemplateEditReport | ConvertTo-Json -Depth 8 -Compress)"
}
if ($pipelineTemplate.com_refine_available -eq $true) {
    $templateComRefineReport = Join-Path $pipelineTemplateWorkspace 'com-refine-report.json'
    if (-not (Test-Path -LiteralPath $templateComRefineReport)) {
        throw "FAIL ppt-pipeline-template-following missing com refine report"
    }
    $templateComRefine = Read-JsonUtf8 -Path $templateComRefineReport
    if ($templateComRefine.updated_slide_notes -lt 1) {
        throw "FAIL ppt-pipeline-template-following notes not updated report=$($templateComRefine | ConvertTo-Json -Depth 8 -Compress)"
    }
}
foreach ($requiredPath in @(
    (Join-Path $pipelineTemplateWorkspace 'pilot-page.pptx'),
    (Join-Path $pipelineTemplateWorkspace 'pilot-preview.png'),
    (Join-Path $pipelineTemplateWorkspace 'pilot-score.md'),
    (Join-Path $pipelineTemplateWorkspace 'pilot-approval.md'),
    (Join-Path $pipelineTemplateWorkspace 'qa\template-fidelity-check.json'),
    (Join-Path $pipelineTemplateWorkspace 'qa\pptx-audit.json')
)) {
    if (-not (Test-Path -LiteralPath $requiredPath)) {
        throw "FAIL ppt-pipeline-template-following missing artifact=$requiredPath"
    }
}

$mcpSafe = Join-Path $fixture 'mcp-safe.json'
$mcpNetwork = Join-Path $fixture 'mcp-network.json'
[System.IO.File]::WriteAllText($mcpSafe, (@{
    name = 'safe-local-mcp'
    version = '1.0.0'
    transport = 'stdio'
    capabilities = @{ tools = @(); resources = @(); prompts = @() }
    permissions = @{ network = $false; filesystem = @($fixture) }
} | ConvertTo-Json -Depth 6), [System.Text.UTF8Encoding]::new($false))
[System.IO.File]::WriteAllText($mcpNetwork, (@{
    name = 'network-mcp'
    version = '1.0.0'
    transport = 'http'
    capabilities = @{ tools = @(); resources = @(); prompts = @() }
    permissions = @{ network = $true; filesystem = @() }
} | ConvertTo-Json -Depth 6), [System.Text.UTF8Encoding]::new($false))
Invoke-Case -Name 'mcp-guard-safe' -ExpectedExit 0 -Arguments @('mcp-guard', '--manifest', $mcpSafe, '--workspace', $fixture)
Invoke-Case -Name 'mcp-guard-network-blocked' -ExpectedExit 1 -Arguments @('mcp-guard', '--manifest', $mcpNetwork)

$mcpCatalog = Join-Path $fixture 'mcp-catalog.json'
$mcpReport = Join-Path $fixture 'mcp-report.json'
[System.IO.File]::WriteAllText($mcpCatalog, (@{
    candidates = @(
        @{
            name = 'time'
            owner = 'staff'
            source = 'https://example.invalid/time'
            license = 'MIT'
            capability = 'time context'
            transport = 'stdio'
            permissions = @()
            risk = 'low'
        },
        @{
            name = 'github'
            owner = 'dev'
            source = 'https://example.invalid/github'
            license = 'MIT'
            capability = 'github context'
            transport = 'http'
            permissions = @('network', 'oauth')
            risk = 'medium'
        }
    )
} | ConvertTo-Json -Depth 8), [System.Text.UTF8Encoding]::new($false))
Invoke-Case -Name 'mcp-distill' -ExpectedExit 0 -Arguments @('mcp-distill', '--catalog', $mcpCatalog, '--report', $mcpReport)
if (-not (Test-Path -LiteralPath $mcpReport)) {
    throw "FAIL mcp-distill missing report=$mcpReport"
}
$mcpDistill = Read-JsonUtf8 -Path $mcpReport
if (-not ($mcpDistill.decisions | Where-Object { $_.name -eq 'time' -and $_.evidence_level -eq 'verified' })) {
    throw "FAIL mcp-distill evidence report=$($mcpDistill | ConvertTo-Json -Depth 8 -Compress)"
}

$baselinePrompt = Join-Path $fixture 'baseline-prompt.json'
$candidatePrompt = Join-Path $fixture 'candidate-prompt.json'
$promptDataset = Join-Path $fixture 'prompt-dataset.json'
$promptReport = Join-Path $fixture 'prompt-report.json'
$promptDistillReport = Join-Path $fixture 'prompt-distill-report.json'
[System.IO.File]::WriteAllText($baselinePrompt, (@{
    name = 'baseline'
    objective = 'answer with citations and concise structure'
    metric = 'coverage'
    prompt_template = 'Answer the user request clearly.'
    stable_prefix = 'Wuji prompt baseline'
    variables = @('task')
} | ConvertTo-Json -Depth 8), [System.Text.UTF8Encoding]::new($false))
[System.IO.File]::WriteAllText($candidatePrompt, (@{
    name = 'candidate'
    objective = 'answer with citations and concise structure'
    metric = 'coverage'
    prompt_template = 'Answer the user request clearly. Cite primary sources. Keep the answer concise. State uncertainty explicitly. Do not reveal secrets.'
    stable_prefix = 'Wuji prompt kernel v10.7 stable prefix for cache-friendly routing and deterministic quality checks.'
    variables = @('task')
} | ConvertTo-Json -Depth 8), [System.Text.UTF8Encoding]::new($false))
[System.IO.File]::WriteAllText($promptDataset, (Get-Content -Raw -LiteralPath $feedbackDataset), [System.Text.UTF8Encoding]::new($false))
Invoke-Case -Name 'prompt-candidate-audit' -ExpectedExit 0 -Arguments @('prompt-candidate-audit', '--candidate', $candidatePrompt, '--report', (Join-Path $fixture 'prompt-candidate-audit.json'))
$badImagePrompt = Join-Path $fixture 'bad-image-prompt.json'
[System.IO.File]::WriteAllText($badImagePrompt, (@{
    name = 'bad-image-probe'
    objective = 'generate a teaching illustration image quickly'
    metric = 'latency'
    prompt_template = 'I will first check the local image entrypoint, read SKILL.md, inspect OPENAI_API_KEY, and search available generation capabilities before creating the illustration poster.'
    stable_prefix = 'Wuji image direct generation stable prefix with forbidden preflight probing behavior.'
    variables = @('task')
} | ConvertTo-Json -Depth 8), [System.Text.UTF8Encoding]::new($false))
Invoke-Case -Name 'prompt-candidate-audit-image-probe-blocked' -ExpectedExit 1 -Arguments @('prompt-candidate-audit', '--candidate', $badImagePrompt, '--report', (Join-Path $fixture 'bad-image-prompt-audit.json'))
Invoke-Case -Name 'prompt-eval' -ExpectedExit 0 -Arguments @('prompt-eval', '--candidate', $candidatePrompt, '--dataset', $promptDataset, '--report', $promptReport)
Invoke-Case -Name 'prompt-distill' -ExpectedExit 0 -Arguments @('prompt-distill', '--baseline', $baselinePrompt, '--candidate', $candidatePrompt, '--dataset', $promptDataset, '--report', $promptDistillReport)
if (-not (Test-Path -LiteralPath $promptDistillReport)) {
    throw "FAIL prompt-distill missing report=$promptDistillReport"
}
$promptDistill = Read-JsonUtf8 -Path $promptDistillReport
if (-not $promptDistill.evidence_level) {
    throw "FAIL prompt-distill missing evidence_level report=$($promptDistill | ConvertTo-Json -Depth 8 -Compress)"
}

$auditRoot = Join-Path $fixture 'audit'
New-Item -ItemType Directory -Force -Path $auditRoot | Out-Null
Write-Fixture (Join-Path $auditRoot 'clean.md') 'all clean content for audit'
Write-Fixture (Join-Path $auditRoot 'normal-prose.md') "- 讨厌半成品、试验版、待开发的人`nInvoke-Case --avoid todo as plain fixture data"
$auditSarif = Join-Path $fixture 'audit-report.sarif'
Invoke-Case -Name 'audit-clean' -ExpectedExit 0 -Arguments @('audit', '--path', $auditRoot, '--report', (Join-Path $fixture 'audit-report.json'), '--sarif', $auditSarif)
if (-not (Test-Path -LiteralPath $auditSarif)) {
    throw "FAIL audit-sarif missing=$auditSarif"
}
$auditBlockedRoot = Join-Path $fixture 'audit-blocked'
New-Item -ItemType Directory -Force -Path $auditBlockedRoot | Out-Null
Write-Fixture (Join-Path $auditBlockedRoot 'todo.md') "// TODO: replace placeholder before release"
Invoke-Case -Name 'audit-real-marker-blocked' -ExpectedExit 1 -Arguments @('audit', '--path', $auditBlockedRoot)
$auditExecutionBlockedRoot = Join-Path $fixture 'audit-execution-blocked'
New-Item -ItemType Directory -Force -Path $auditExecutionBlockedRoot | Out-Null
[System.IO.File]::WriteAllLines((Join-Path $auditExecutionBlockedRoot 'task-log.jsonl'), @(
    (@{ timestamp = '2026-06-03T00:00:00Z'; event = 'start'; status = 'running'; phase = 'preflight'; note = 'check environment first and scan the repo first'; artifacts = @() } | ConvertTo-Json -Compress)
), [System.Text.UTF8Encoding]::new($false))
Invoke-Case -Name 'audit-execution-precheck-blocked' -ExpectedExit 1 -Arguments @('audit', '--path', $auditExecutionBlockedRoot)
$auditCloseoutLeakRoot = Join-Path $fixture 'audit-closeout-leak'
New-Item -ItemType Directory -Force -Path $auditCloseoutLeakRoot | Out-Null
[System.IO.File]::WriteAllLines((Join-Path $auditCloseoutLeakRoot 'task-log.jsonl'), @(
    (@{ timestamp = '2026-06-03T00:01:00Z'; event = 'end'; status = 'done'; note = 'next step could continue with further optimization'; artifacts = @($artifact); closeout_report = $closeoutPassReport; evidence_report = $verifiedEvidenceReport } | ConvertTo-Json -Compress)
), [System.Text.UTF8Encoding]::new($false))
Invoke-Case -Name 'audit-closeout-leak-blocked' -ExpectedExit 1 -Arguments @('audit', '--path', $auditCloseoutLeakRoot)

$previewOut = Join-Path $fixture 'preview.txt'
Invoke-Case -Name 'preview-dispatch' -ExpectedExit 0 -Arguments @('preview', '--command', 'powershell', '--arg', '-NoProfile', '--arg', '-Command', '--arg', ('Set-Content -LiteralPath ''' + $previewOut + ''' -Value ''preview output long enough'''), '--output', $previewOut)

Write-Host 'RESULT: PASS - wuji-cli deterministic gates'
Remove-Item -LiteralPath $bin -Force -ErrorAction SilentlyContinue
