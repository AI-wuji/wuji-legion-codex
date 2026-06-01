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

function Invoke-Case {
    param([string]$Name, [int]$ExpectedExit, [string[]]$Arguments)
    $output = & $bin @Arguments 2>&1
    $actual = $LASTEXITCODE
    if ($actual -ne $ExpectedExit) {
        throw "FAIL $Name expected=$ExpectedExit actual=$actual output=$($output -join ' | ')"
    }
    Write-Host "PASS $Name exit=$actual"
}

$reference = Join-Path $fixture 'reference.txt'
$output = Join-Path $fixture 'generated.txt'
$evidence = Join-Path $fixture 'evidence.txt'
$artifact = Join-Path $fixture 'artifact.txt'
Write-Fixture $reference
Write-Fixture $evidence
Write-Fixture $artifact

Invoke-Case -Name 'reference-safe' -ExpectedExit 0 -Arguments @('reference-guard', '--reference', $reference, '--output', $output)
Invoke-Case -Name 'reference-overwrite-blocked' -ExpectedExit 1 -Arguments @('reference-guard', '--reference', $reference, '--output', $reference)
Invoke-Case -Name 'claim-without-evidence-blocked' -ExpectedExit 1 -Arguments @('claim-guard', '--claim', 'completed and passed')
Invoke-Case -Name 'claim-with-evidence' -ExpectedExit 0 -Arguments @('claim-guard', '--claim', 'completed and passed', '--evidence', $evidence)
Invoke-Case -Name 'time-guard-blocked' -ExpectedExit 1 -Arguments @('time-guard', '--kind', 'non-code', '--elapsed-minutes', '15', '--phase', 'prototype')
Invoke-Case -Name 'time-guard-with-artifact' -ExpectedExit 0 -Arguments @('time-guard', '--kind', 'non-code', '--elapsed-minutes', '15', '--phase', 'prototype', '--artifact', $artifact)

$taskWorkspace = Join-Path $fixture 'task'
Invoke-Case -Name 'task-start' -ExpectedExit 0 -Arguments @('task', '--workspace', $taskWorkspace, '--event', 'start', '--status', 'running', '--artifact', $artifact, '--note', 'task started')
$benchWorkspace = Join-Path $fixture 'bench'
Invoke-Case -Name 'bench-log' -ExpectedExit 0 -Arguments @('bench', '--workspace', $benchWorkspace, '--name', 'sample', '--input-tokens', '10', '--output-tokens', '20', '--duration-ms', '30', '--tool-calls', '2', '--retries', '0', '--qa-pass', 'true')
Invoke-Case -Name 'bench-report' -ExpectedExit 0 -Arguments @('bench-report', '--workspace', $benchWorkspace, '--report', (Join-Path $fixture 'bench-report.json'))

$routeConfig = Join-Path $fixture 'route-config.json'
[System.IO.File]::WriteAllText($routeConfig, (@{
    iron_rules_version = '10.6'
    default_model_tier = 'low'
    model_profiles = @{
        low = @{ provider_id = 'openai-api'; model = 'gpt-5.4-mini'; reasoning_effort = 'low' }
        standard = @{ provider_id = 'openai-api'; model = 'gpt-5.4'; reasoning_effort = 'medium' }
        high = @{ provider_id = 'openai-api'; model = 'gpt-5.5'; reasoning_effort = 'high' }
    }
    cache_config = @{ target_hit_rate = 0.95; flatten_threshold = 10 }
    routing_rules = @(
        @{ id = 'visual'; name = 'visual'; provider_id = 'deepseek-web'; model = ''; priority = 60; keywords = @('ppt', 'ui', 'design') },
        @{ id = 'code'; name = 'code'; provider_id = 'deepseek-web'; model = ''; priority = 80; keywords = @('go', 'code', 'compile') },
        @{ id = 'chat'; name = 'chat'; provider_id = 'deepseek-web'; model = ''; priority = 0; keywords = @() }
    )
} | ConvertTo-Json -Depth 8), [System.Text.UTF8Encoding]::new($false))
Invoke-Case -Name 'route-task' -ExpectedExit 0 -Arguments @('route-task', '--config', $routeConfig, '--query', 'please build a ppt ui design', '--report', (Join-Path $fixture 'route-report.json'))
Invoke-Case -Name 'context-pack' -ExpectedExit 0 -Arguments @('context-pack', '--config', $routeConfig, '--workspace', $fixture, '--query', 'please build a ppt ui design', '--artifact', $artifact, '--report', (Join-Path $fixture 'context-pack.json'))
Invoke-Case -Name 'feedback-log' -ExpectedExit 0 -Arguments @('feedback-log', '--workspace', $fixture, '--task', 'daily answer quality tuning', '--prefer', 'keep the answer concise', '--prefer', 'state uncertainty explicitly', '--avoid', 'placeholder', '--report', (Join-Path $fixture 'feedback-log-report.json'))
Invoke-Case -Name 'feedback-log-second' -ExpectedExit 0 -Arguments @('feedback-log', '--workspace', $fixture, '--task', 'source discipline', '--prefer', 'cite primary sources', '--avoid', 'todo', '--source', 'qa')
$feedbackLog = Join-Path $fixture 'feedback\feedback-log.jsonl'
$feedbackDataset = Join-Path $fixture 'feedback\feedback-dataset.json'
Invoke-Case -Name 'feedback-dataset' -ExpectedExit 0 -Arguments @('feedback-dataset', '--log', $feedbackLog, '--report', $feedbackDataset)

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
Invoke-Case -Name 'workflow-final' -ExpectedExit 0 -Arguments @('workflow-guard', '--workspace', $workflow, '--stage', 'final')

$pptx = Join-Path $fixture 'pptx'
New-Item -ItemType Directory -Force -Path $pptx,(Join-Path $pptx 'ppt\slides'),(Join-Path $pptx 'ppt\media'),(Join-Path $pptx 'ppt\slideLayouts'),(Join-Path $pptx 'ppt\theme') | Out-Null
Set-Content -LiteralPath (Join-Path $pptx 'ppt\slides\slide1.xml') -Value '<p:sld><p:cSld><p:spTree><p:sp/><p:pic/><a:t>Hello</a:t></p:spTree></p:cSld></p:sld>' -Encoding UTF8
Set-Content -LiteralPath (Join-Path $pptx 'ppt\slideLayouts\slideLayout1.xml') -Value '<layout />' -Encoding UTF8
Set-Content -LiteralPath (Join-Path $pptx 'ppt\theme\theme1.xml') -Value '<theme />' -Encoding UTF8
Set-Content -LiteralPath (Join-Path $pptx 'ppt\media\image1.png') -Value 'png bytes long enough for test' -Encoding UTF8
$pptxFile = Join-Path $fixture 'sample.pptx'
if (Test-Path $pptxFile) { Remove-Item -LiteralPath $pptxFile -Force }
$zipFile = Join-Path $fixture 'sample.zip'
if (Test-Path $zipFile) { Remove-Item -LiteralPath $zipFile -Force }
Compress-Archive -Path (Join-Path $pptx '*') -DestinationPath $zipFile
Move-Item -LiteralPath $zipFile -Destination $pptxFile

$assetWorkspace = Join-Path $fixture 'asset-map'
Invoke-Case -Name 'asset-map' -ExpectedExit 0 -Arguments @('asset-map', '--pptx', $pptxFile, '--workspace', $assetWorkspace)
Write-Fixture (Join-Path $assetWorkspace 'pilot-page.pptx')
Write-Fixture (Join-Path $assetWorkspace 'pilot-preview.png')
Write-Fixture (Join-Path $assetWorkspace 'pilot-score.md')
Invoke-Case -Name 'pptx-audit' -ExpectedExit 0 -Arguments @('pptx-audit', '--pptx', $pptxFile, '--report', (Join-Path $fixture 'pptx-audit.json'))
Invoke-Case -Name 'pptx-preflight' -ExpectedExit 0 -Arguments @('pptx-preflight', '--workspace', $assetWorkspace)
Invoke-Case -Name 'pptx-batch-gate' -ExpectedExit 0 -Arguments @('pptx-batch-gate', '--workspace', $assetWorkspace)

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
Invoke-Case -Name 'prompt-eval' -ExpectedExit 0 -Arguments @('prompt-eval', '--candidate', $candidatePrompt, '--dataset', $promptDataset, '--report', $promptReport)
Invoke-Case -Name 'prompt-distill' -ExpectedExit 0 -Arguments @('prompt-distill', '--baseline', $baselinePrompt, '--candidate', $candidatePrompt, '--dataset', $promptDataset, '--report', $promptDistillReport)
if (-not (Test-Path -LiteralPath $promptDistillReport)) {
    throw "FAIL prompt-distill missing report=$promptDistillReport"
}

$auditRoot = Join-Path $fixture 'audit'
New-Item -ItemType Directory -Force -Path $auditRoot | Out-Null
Write-Fixture (Join-Path $auditRoot 'clean.md') 'all clean content for audit'
$auditSarif = Join-Path $fixture 'audit-report.sarif'
Invoke-Case -Name 'audit-clean' -ExpectedExit 0 -Arguments @('audit', '--path', $auditRoot, '--report', (Join-Path $fixture 'audit-report.json'), '--sarif', $auditSarif)
if (-not (Test-Path -LiteralPath $auditSarif)) {
    throw "FAIL audit-sarif missing=$auditSarif"
}

$previewOut = Join-Path $fixture 'preview.txt'
Invoke-Case -Name 'preview-dispatch' -ExpectedExit 0 -Arguments @('preview', '--command', 'powershell', '--arg', '-NoProfile', '--arg', '-Command', '--arg', ('Set-Content -LiteralPath ''' + $previewOut + ''' -Value ''preview output long enough'''), '--output', $previewOut)

Write-Host 'RESULT: PASS - wuji-cli deterministic gates'
Remove-Item -LiteralPath $bin -Force -ErrorAction SilentlyContinue
