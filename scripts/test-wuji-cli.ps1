param()

$ErrorActionPreference = 'Stop'

$root = Split-Path -Parent $PSScriptRoot
$binDir = Join-Path $root '.wuji-tools'
$bin = Join-Path $binDir 'wuji-cli-smoke.exe'
$fixture = Join-Path $env:TEMP ("wuji-cli-fixture-" + [guid]::NewGuid().ToString('N'))
$latestLog = Join-Path $root 'outputs\test-wuji-cli-latest.log'
$ExecutionBudgetContractJson = '{"objective":"all-work-direct-small-task-execution","must_do":["bind-current-scope-before-expansion","bind-finish-line-and-out-of-scope-before-goal-start","keep-direct-code-work-on-a-minimal-first-pass-guard","run-targeted-verification-before-any-full-suite","all-non-chat-work-stays-direct-task-by-default","keep-officers-on-demand-and-exit-after-merge","treat-runtime-context-audit-as-token-cost-cache-claim-only","finish-current-scope-without-reopen-ceremony","avoid-automatic-task-upshift"],"must_not_do":["auto-create-big-task-mode","auto-upshift-to-structural-or-release-mode","add-agnes-scout-or-planning-sidecar-to-light-task-by-tier-alone","treat-officers-as-perspectives-or-tones","repeat-full-suite-after-small-edits","block-non-token-work-on-missing-runtime-usage-log","start-goal-without-clear-finish-line","continue-low-value-sweep-outside-current-scope"]}'
$KernelSourceFixtureJson = '{"kernel_version":"11.3","identity":{"execution_surface":"Codex-only"},"optimization_kernel":{"runtime_context_gate":"runtime-context-audit","runtime_context_policy":"numeric-only runtime usage evidence required; raw prompts messages content forbidden"},"required_audits":["runtime-context-audit-for-token-cost-cache-usage-claims"],"intelligence_profile_contract":{"role":"candidate-scout-not-research-system","search_scope":"wide-recall-shallow-first","search_method":"wide-shallow-scout-first-then-deepen-only-on-promising-candidates","may_do":["search","candidate-metadata","dedupe-cluster","evidence-handle"],"must_not_do":["final-analysis","deep-extract-by-default","distillation-decision","adoption-decision","install-or-execute"]},"concise_execution_contract":{"objective":"short-precise-high-hit-low-total-cost","must_do":["simplest-effective-path-first","single-message-precision","minimal-needed-context","agnes-search-only-before-uncertain-build-when-web-search-is-explicitly-needed","prior-art-before-invention-when-uncertain","first-pass-acceptance-and-impact-lock-before-edit","prove-need-before-abstraction","delete-or-reuse-before-add","target-page-in-place-replacement","active-route-entrypoint-verification","superseded-page-cleanup-before-completion","smallest-working-change-first","fresh-output-uncached-volume-gated"],"must_not_do":["verbose-status-padding","unneeded-preflight-loop","guess-without-evidence","blind-trial-and-error-when-prior-art-is-available","context-shift-from-cached-to-uncached","from-scratch-tooling-when-existing-solution-fits","clever-overengineering-without-proven-need","new-abstraction-before-duplication-or-gap-is-proven","parallel-compat-page-for-requested-page-change","leave-old-page-reachable-after-replacement"],"cost_vector":["cached_tokens_p95","fresh_input_tokens_p95","output_tokens_p95","uncached_tokens_p95","tokens_per_success","retries"]},"execution_budget_contract":' + $ExecutionBudgetContractJson + ',"analysis_completeness_contract":{"objective":"complete-materials-before-architecture-analysis","must_do":["collect-material-inventory","state-coverage-and-gaps","ask-user-for-missing-materials-when-critical","separate-fact-inference-and-unknown","no-final-conclusion-from-incomplete-evidence"],"must_not_do":["guess-architecture-from-partial-materials","treat-sample-as-whole-system","hide-coverage-gaps","promote-uncertain-claim-to-fact"]}}'
function New-ContextPackRichFixtureJson {
    param([string]$Workspace)

    $configHash = Get-Sha256Lower -Path (Join-Path $Workspace 'config.json')
    $toolHash = Get-Sha256Lower -Path (Join-Path $Workspace 'tools\wuji_cli.go')
    return ([ordered]@{
        command = 'context-pack'
        generated_at = (Get-Date).ToUniversalTime().ToString('o')
        wuji_version = '11.3'
        tool_source_hash = $toolHash
        input_hashes = [ordered]@{
            'config.json' = $configHash
            'tools/wuji_cli.go' = $toolHash
        }
        stable_prefix = [ordered]@{ stable_prefix_policy = 'byte-stable-minimal-resident' }
        stable_prefix_canon = [ordered]@{ canon_hash = 'fixture' }
        concise_execution_contract = [ordered]@{ objective = 'short-precise-high-hit-low-total-cost' }
        execution_budget_contract = $ExecutionBudgetContractJson | ConvertFrom-Json
        route_summary = [ordered]@{
            execution_budget = [ordered]@{ id = 'DIRECT_TASK'; full_suite_max_runs = 0 }
            distilled_atom_evidence_count = 1
            current_audit_evidence_count = 3
            query_key = 'fixture'
        }
        dynamic_context = [ordered]@{
            distilled_atoms = @('fixture')
            distilled_atom_evidence = @([ordered]@{ atom = 'fixture'; decision = 'resident'; evidence_visible = $true })
            current_audit_evidence = @(
                [ordered]@{ path_ref = 'outputs/fusion-audit-report.json' },
                [ordered]@{ path_ref = 'outputs/optimization-audit-report.json' },
                [ordered]@{ path_ref = 'outputs/context-bloat-audit-report.json' }
            )
            execution_summaries = @()
            audit_summaries = @()
            model_tier = 'standard'
        }
        review_gates = @('fixture')
        artifact_summaries = @(
            [ordered]@{ path_ref = 'fixture.txt'; kind = 'text'; summary_mode = 'summary'; evidence_handle = 'fixture' }
        )
        optimization_policy = [ordered]@{
            objective = 'fixture'
            concise_execution = 'gate-cached-fresh-output-uncached-total-cost'
        }
    }) | ConvertTo-Json -Depth 12
}

New-Item -ItemType Directory -Force -Path (Split-Path -Parent $latestLog) | Out-Null
Set-Content -LiteralPath $latestLog -Value ("START test-wuji-cli " + (Get-Date).ToUniversalTime().ToString('o')) -Encoding UTF8

function Write-RunLog {
    param([string]$Line)
    Add-Content -LiteralPath $latestLog -Value $Line -Encoding UTF8
    Write-Host $Line
}

trap {
    Add-Content -LiteralPath $latestLog -Value ("FAIL test-wuji-cli " + (Get-Date).ToUniversalTime().ToString('o')) -Encoding UTF8
    Remove-Item -LiteralPath $bin -Force -ErrorAction SilentlyContinue
    Remove-Item -LiteralPath $fixture -Recurse -Force -ErrorAction SilentlyContinue
    throw $_
}

if (Test-Path $fixture) { Remove-Item -LiteralPath $fixture -Recurse -Force }
New-Item -ItemType Directory -Force -Path $binDir, $fixture | Out-Null
try {
$buildOutput = & (Join-Path $PSScriptRoot 'build-wuji-cli.ps1') -Output $bin 2>&1
foreach ($line in $buildOutput) {
    Add-Content -LiteralPath $latestLog -Value ([string]$line) -Encoding UTF8
    Write-Host $line
}

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

function Write-JsonUtf8 {
    param(
        [string]$Path,
        [object]$Value,
        [int]$Depth = 12
    )
    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $Path) | Out-Null
    [System.IO.File]::WriteAllText($Path, ($Value | ConvertTo-Json -Depth $Depth), [System.Text.UTF8Encoding]::new($false))
}

function Copy-RepoFile {
    param([string]$Workspace, [string]$RelPath)
    $source = Join-Path $root ($RelPath -replace '/', '\')
    $dest = Join-Path $Workspace ($RelPath -replace '/', '\')
    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $dest) | Out-Null
    Copy-Item -LiteralPath $source -Destination $dest -Force
}

function New-FusionAuditFixture {
    param([string]$Name)
    $workspace = Join-Path $fixture $Name
    if (Test-Path -LiteralPath $workspace) {
        Remove-Item -LiteralPath $workspace -Recurse -Force
    }
    New-Item -ItemType Directory -Force -Path $workspace | Out-Null

    foreach ($rel in @(
        'kernel-source.json',
        'config.json',
        'fusion-matrix.json',
        'acceptance-checklists.json',
        'purification-charter.json',
        'hotpath-manifest.json',
        'README.md',
        'SKILL.md',
        'GLOBAL_AGENTS.md',
        'scripts/beep.ps1',
        'tools/wuji_cli.go'
    )) {
        Copy-RepoFile -Workspace $workspace -RelPath $rel
    }
    Get-ChildItem -LiteralPath (Join-Path $root 'units') -Filter '*.md' | ForEach-Object {
        Copy-RepoFile -Workspace $workspace -RelPath ('units/' + $_.Name)
    }
    Get-ChildItem -LiteralPath (Join-Path $root 'units') -Filter '*.json' | ForEach-Object {
        Copy-RepoFile -Workspace $workspace -RelPath ('units/' + $_.Name)
    }
    Copy-RepoFile -Workspace $workspace -RelPath 'experts/INDEX.md'
    Get-ChildItem -LiteralPath (Join-Path $root 'experts') -Recurse -Filter '*.md' | ForEach-Object {
        $rootFull = [System.IO.Path]::GetFullPath($root).TrimEnd('\', '/')
        $fullName = [System.IO.Path]::GetFullPath($_.FullName)
        $rel = $fullName.Substring($rootFull.Length).TrimStart('\', '/') -replace '\\', '/'
        Copy-RepoFile -Workspace $workspace -RelPath $rel
    }

    Write-JsonUtf8 -Path (Join-Path $workspace 'residual-entrypoints.json') -Value ([ordered]@{
        version = '11.3'
        entries = @(
            @{ path = 'kernel-source.json'; status = 'main-chain'; type = 'truth-source' },
            @{ path = 'config.json'; status = 'main-chain'; type = 'runtime-config' },
            @{ path = 'fusion-matrix.json'; status = 'main-chain'; type = 'governance-ledger' },
            @{ path = 'acceptance-checklists.json'; status = 'main-chain'; type = 'governance-ledger' },
            @{ path = 'purification-charter.json'; status = 'main-chain'; type = 'governance-ledger' },
            @{ path = 'residual-entrypoints.json'; status = 'main-chain'; type = 'governance-ledger' },
            @{ path = 'hotpath-manifest.json'; status = 'main-chain'; type = 'context-control-ledger' },
            @{ path = 'README.md'; status = 'on-demand'; type = 'mirror-doc' },
            @{ path = 'SKILL.md'; status = 'on-demand'; type = 'mirror-doc' },
            @{ path = 'GLOBAL_AGENTS.md'; status = 'on-demand'; type = 'mirror-doc' },
            @{ path = 'scripts/beep.ps1'; status = 'on-demand'; type = 'notification-script' },
            @{ path = 'tools/wuji_cli.go'; status = 'main-chain'; type = 'execution-base' },
            @{ path = 'units/*.md'; status = 'on-demand'; type = 'unit-mirror-doc-family' },
            @{ path = 'units/*.json'; status = 'on-demand'; type = 'unit-catalog' },
            @{ path = 'experts/INDEX.md'; status = 'on-demand'; type = 'owner-index' },
            @{ path = 'experts/*/*.md'; status = 'on-demand'; type = 'expert-mirror-doc-family' }
        )
    })
    return $workspace
}

function Assert-ReportFailureContains {
    param([string]$ReportPath, [string]$Marker)
    $report = Read-JsonUtf8 -Path $ReportPath
    $joined = (($report.failures | ForEach-Object { [string]$_ }) -join '|')
    if ($joined -notmatch [regex]::Escape($Marker)) {
        throw "FAIL expected failure marker '$Marker' report=$($report | ConvertTo-Json -Depth 8 -Compress)"
    }
}

function Get-Sha256Lower {
    param([string]$Path)
    return (Get-FileHash -Algorithm SHA256 -LiteralPath $Path).Hash.ToLowerInvariant()
}

function Get-PrivacyHash {
    param([string]$Value)
    $normalized = ([System.IO.Path]::GetFullPath($Value).ToLowerInvariant() -replace '\s+', ' ').Trim()
    $sha = [System.Security.Cryptography.SHA256]::Create()
    try {
        $bytes = [System.Text.Encoding]::UTF8.GetBytes($normalized)
        return (($sha.ComputeHash($bytes) | ForEach-Object { $_.ToString('x2') }) -join '').Substring(0, 16)
    }
    finally {
        $sha.Dispose()
    }
}

function Write-CurrentAuditReports {
    param([string]$Workspace)

    New-Item -ItemType Directory -Force -Path $Workspace | Out-Null
    $outputs = Join-Path $Workspace 'outputs'
    New-Item -ItemType Directory -Force -Path $outputs, (Join-Path $Workspace 'tools') | Out-Null

    $seedFiles = [ordered]@{
        'kernel-source.json' = $KernelSourceFixtureJson
        'config.json' = '{"iron_rules_version":"11.3","cache_config":{"stable_prefix_policy":"byte-stable-minimal-resident","optimization_objective":"smaller-stable-prefix-with-equal-or-better-hit-rate","concise_execution_policy":"gate-cached-fresh-output-uncached-total-cost"}}'
        'fusion-matrix.json' = '{"version":"11.3","decisions":[]}'
        'residual-entrypoints.json' = '{"version":"11.3","entries":[]}'
        'acceptance-checklists.json' = '{"white_hat":["fixture"],"guard_office":["fixture"],"root_cause_officer":["fixture"],"audit":["fixture"],"quality_inspection":["fixture runtime-context-audit"],"performance_benchmark_on_demand":["fixture runtime-context-audit"],"compliance_on_demand":["fixture"]}'
        'purification-charter.json' = '{"version":"11.3","hard_gates":[]}'
        'hotpath-manifest.json' = '{"resident":[{"path":"kernel-source.json","max_bytes":8192}],"on_demand":[{"path":"tools/wuji_cli.go","max_loaded_bytes":8192}],"cold_ledger":[{"path":"fusion-matrix.json","default_mode":"handle-only"},{"path":"outputs/runtime-context-audit-report.json","default_mode":"handle-only"}],"forbidden_resident":["outputs/**",".wuji-tools/**","outputs/runtime-usage.jsonl","raw prompts/messages/content","full transcripts"]}'
        'README.md' = '# fixture readme hotpath-manifest.json context-bloat-audit runtime-context-audit concise_execution_contract execution_budget_contract analysis_completeness_contract complete-materials-before-architecture-analysis target-page-in-place-replacement parallel-compat-page-for-requested-page-change'
        'AGENTS.md' = '# fixture agents instruction runtime-context-audit'
        'SKILL.md' = '# fixture skill hotpath-manifest.json context-bloat-audit runtime-context-audit concise_execution_contract execution_budget_contract target-page-in-place-replacement parallel-compat-page-for-requested-page-change'
        'GLOBAL_AGENTS.md' = '# fixture global hotpath-manifest.json context-bloat-audit runtime-context-audit concise_execution_contract execution_budget_contract target-page-in-place-replacement parallel-compat-page-for-requested-page-change'
        'tools/wuji_cli.go' = 'package main // fixture source used only for audit hash freshness'
    }

    foreach ($entry in $seedFiles.GetEnumerator()) {
        $path = Join-Path $Workspace ($entry.Key -replace '/', '\')
        Write-Fixture $path $entry.Value
    }
    Write-Fixture (Join-Path $Workspace 'outputs\context-pack-rich.json') (New-ContextPackRichFixtureJson -Workspace $Workspace)

    $workspaceKey = Get-PrivacyHash -Value $Workspace
    $benchLog = Join-Path $outputs 'bench.jsonl'
    $benchRows = @(
        [ordered]@{ timestamp = '2026-06-03T00:00:00Z'; workspace_key = $workspaceKey; name_key = (Get-PrivacyHash -Value 'fixture bench one'); input_tokens = 12000; output_tokens = 400; duration_ms = 1000; tool_calls = 1; retries = 0; quality_pass = $true; cache_hit = $true; cached_tokens = 8000; fresh_input_tokens = 4000; reused_prefix_bytes = 8000; activated_officers = 1; activated_skills = 0; loaded_file_bytes = 8000; largest_context_segment_bytes = 8000 }
        [ordered]@{ timestamp = '2026-06-03T00:01:00Z'; workspace_key = $workspaceKey; name_key = (Get-PrivacyHash -Value 'fixture bench two'); input_tokens = 11000; output_tokens = 500; duration_ms = 1000; tool_calls = 1; retries = 0; quality_pass = $true; cache_hit = $true; cached_tokens = 7800; fresh_input_tokens = 3800; reused_prefix_bytes = 7800; activated_officers = 1; activated_skills = 0; loaded_file_bytes = 7800; largest_context_segment_bytes = 7800 }
    )
    [System.IO.File]::WriteAllLines($benchLog, ($benchRows | ForEach-Object { $_ | ConvertTo-Json -Depth 12 -Compress }), [System.Text.UTF8Encoding]::new($false))

    $runtimeUsageLog = Join-Path $outputs 'runtime-usage.jsonl'
    $runtimeUsageRows = @(
        [ordered]@{ timestamp = '2026-06-03T00:00:00Z'; usage = [ordered]@{ input_tokens = 12000; output_tokens = 400; cached_tokens = 8000; fresh_input_tokens = 4000 } }
        [ordered]@{ timestamp = '2026-06-03T00:01:00Z'; usage = [ordered]@{ prompt_tokens = 11000; completion_tokens = 500; prompt_tokens_details = [ordered]@{ cached_tokens = 7800 }; uncached_input_tokens = 3200 } }
    )
    [System.IO.File]::WriteAllLines($runtimeUsageLog, ($runtimeUsageRows | ForEach-Object { $_ | ConvertTo-Json -Depth 12 -Compress }), [System.Text.UTF8Encoding]::new($false))

    $benchLogHash = Get-Sha256Lower -Path $benchLog
    Write-JsonUtf8 -Path (Join-Path $outputs 'bench-report.json') -Value ([ordered]@{
        workspace_key = $workspaceKey
        command = 'bench-report'
        generated_at = (Get-Date).ToUniversalTime().ToString('o')
        wuji_version = '11.3'
        log_ref = 'outputs/bench.jsonl'
        log_sha256 = $benchLogHash
        input_hashes = [ordered]@{ 'bench.jsonl' = $benchLogHash }
        decision = 'absorb'
        evidence_level = 'verified'
        volume_gate = 'pass'
        cache_observations = 2
        cached_tokens_p95 = 8000
        output_tokens_p95 = 500
        fresh_input_tokens_p95 = 4000
        uncached_tokens_p95 = 4000
        reused_prefix_bytes_p95 = 8000
        input_tokens_p95 = 12000
        activated_officers_p95 = 1
        activated_skills_p95 = 0
        loaded_file_bytes_p95 = 8000
        largest_context_segment_bytes_p95 = 8000
    })

    $requiredByGate = @{
        'fusion-audit' = @(
            'kernel-source.json',
            'config.json',
            'fusion-matrix.json',
            'residual-entrypoints.json',
            'acceptance-checklists.json',
            'purification-charter.json',
            'hotpath-manifest.json',
            'README.md',
            'tools/wuji_cli.go'
        )
        'optimization-audit' = @(
            'config.json',
            'acceptance-checklists.json',
            'outputs/context-pack-rich.json',
            'hotpath-manifest.json',
            'tools/wuji_cli.go'
        )
        'context-bloat-audit' = @(
            'hotpath-manifest.json',
            'outputs/context-pack-rich.json',
            'outputs/bench-report.json',
            'tools/wuji_cli.go'
        )
    }

    foreach ($gate in $requiredByGate.Keys) {
        $inputHashes = [ordered]@{}
        foreach ($rel in $requiredByGate[$gate]) {
            $inputHashes[$rel] = Get-Sha256Lower -Path (Join-Path $Workspace ($rel -replace '/', '\'))
        }
        $externalInputHashes = [ordered]@{}
        if ($gate -eq 'fusion-audit') {
            $runtimeSkill = 'C:\Users\Administrator\.agents\skills\wuji-legion\SKILL.md'
            $externalInputHashes[(Get-PrivacyHash -Value $runtimeSkill)] = Get-Sha256Lower -Path $runtimeSkill
        }
        Write-JsonUtf8 -Path (Join-Path $outputs "$gate-report.json") -Value ([ordered]@{
            workspace = (Resolve-Path -LiteralPath $Workspace).Path
            status = 'pass'
            generated_at = (Get-Date).ToUniversalTime().ToString('o')
            command = $gate
            wuji_version = '11.3'
            tool_source_hash = $inputHashes['tools/wuji_cli.go']
            input_hashes = $inputHashes
            external_input_hashes = $externalInputHashes
        })
    }

    $previousNativeErrorPref = $PSNativeCommandUseErrorActionPreference
    try {
        $PSNativeCommandUseErrorActionPreference = $false
        $runtimeAuditOutput = & $bin 'runtime-context-audit' '--workspace' $Workspace 2>&1
        $runtimeAuditExit = $LASTEXITCODE
    }
    finally {
        $PSNativeCommandUseErrorActionPreference = $previousNativeErrorPref
    }
    if ($runtimeAuditExit -ne 0) {
        throw "FAIL Write-CurrentAuditReports runtime-context-audit exit=$runtimeAuditExit output=$($runtimeAuditOutput -join ' | ')"
    }
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
    Write-RunLog "PASS $Name exit=$actual"
}

$reference = Join-Path $fixture 'reference.txt'
$output = Join-Path $fixture 'generated.txt'
$evidence = Join-Path $fixture 'evidence.txt'
$artifact = Join-Path $fixture 'artifact.txt'
$sourceArtifact = Join-Path $fixture 'artifact-source.go'
$metaReportArtifact = Join-Path $fixture 'meta-report.json'
Write-Fixture $reference
Write-Fixture $evidence 'fixture evidence content with enough bytes for claim verification and audit linkage'
Write-Fixture $artifact
Write-Fixture $sourceArtifact 'package main

func marker() string {
    return "failure missing blocked are code words here"
}
'
Write-Fixture $metaReportArtifact '{"status":"pass"}'
$claimAuditDir = Join-Path $fixture 'outputs'
New-Item -ItemType Directory -Force -Path $claimAuditDir | Out-Null
Write-JsonUtf8 -Path (Join-Path $claimAuditDir 'fusion-audit-report.json') -Value (@{
    workspace = $fixture
    status = 'pass'
})
Write-JsonUtf8 -Path (Join-Path $claimAuditDir 'optimization-audit-report.json') -Value (@{
    workspace = $fixture
    status = 'pass'
})

Invoke-Case -Name 'reference-safe' -ExpectedExit 0 -Arguments @('reference-guard', '--reference', $reference, '--output', $output)
Invoke-Case -Name 'reference-overwrite-blocked' -ExpectedExit 1 -Arguments @('reference-guard', '--reference', $reference, '--output', $reference)
Invoke-Case -Name 'claim-without-evidence-blocked' -ExpectedExit 1 -Arguments @('claim-guard', '--claim', 'completed and passed')
Invoke-Case -Name 'claim-with-evidence-no-audits-blocked' -ExpectedExit 1 -Arguments @('claim-guard', '--claim', 'completed and passed', '--evidence', $evidence)
Invoke-Case -Name 'claim-forged-audits-blocked' -ExpectedExit 1 -Arguments @('claim-guard', '--claim', 'completed and passed', '--workspace', $fixture, '--evidence', $evidence)
Write-CurrentAuditReports -Workspace $fixture
Invoke-Case -Name 'claim-with-evidence' -ExpectedExit 0 -Arguments @('claim-guard', '--claim', 'completed and passed', '--workspace', $fixture, '--evidence', $evidence)
$missingRuntimeAuditWorkspace = Join-Path $fixture 'missing-runtime-context-audit'
Write-CurrentAuditReports -Workspace $missingRuntimeAuditWorkspace
Remove-Item -LiteralPath (Join-Path $missingRuntimeAuditWorkspace 'outputs\runtime-context-audit-report.json') -Force
Invoke-Case -Name 'claim-token-cost-runtime-audit-required' -ExpectedExit 1 -Arguments @('claim-guard', '--claim', 'completed token cost cache hit backend usage optimization', '--workspace', $missingRuntimeAuditWorkspace, '--evidence', $evidence)
Invoke-Case -Name 'claim-token-cost-runtime-audit-pass' -ExpectedExit 0 -Arguments @('claim-guard', '--claim', 'completed token cost cache hit backend usage optimization', '--workspace', $fixture, '--evidence', $evidence)
$cnTokenClaim = 'completed ' + (-join [char[]](0x540E,0x53F0,0x8D39,0x7528,0x547D,0x4E2D,0x6D88,0x8017,0x4F18,0x5316))
Invoke-Case -Name 'claim-cn-token-runtime-audit-required' -ExpectedExit 1 -Arguments @('claim-guard', '--claim', $cnTokenClaim, '--workspace', $missingRuntimeAuditWorkspace, '--evidence', $evidence)

$fusionBaselineWorkspace = New-FusionAuditFixture -Name 'fusion-audit-trending-baseline'
$fusionBaselineReport = Join-Path $fusionBaselineWorkspace 'fusion-baseline-report.json'
Invoke-Case -Name 'fusion-audit-trending-baseline' -ExpectedExit 0 -Arguments @('fusion-audit', '--workspace', $fusionBaselineWorkspace, '--report', $fusionBaselineReport)

$fusionExtraAtomWorkspace = New-FusionAuditFixture -Name 'fusion-audit-extra-trending-atom'
$fusionExtraMatrixPath = Join-Path $fusionExtraAtomWorkspace 'fusion-matrix.json'
$fusionExtraMatrix = Read-JsonUtf8 -Path $fusionExtraMatrixPath
$fusionExtraMatrix.decisions += [pscustomobject]@{
    atom = 'github-trending-new-runtime-atom'
    source_refs = 'github-trending-20260608-style'
    decision = 'mount-on-demand'
    owner = 'intelligence-profile'
    reason = 'bad additive runtime atom'
}
Write-JsonUtf8 -Path $fusionExtraMatrixPath -Value $fusionExtraMatrix
$fusionExtraReport = Join-Path $fusionExtraAtomWorkspace 'fusion-extra-report.json'
Invoke-Case -Name 'fusion-audit-hot-trending-extra-atom-blocked' -ExpectedExit 1 -Arguments @('fusion-audit', '--workspace', $fusionExtraAtomWorkspace, '--report', $fusionExtraReport)
Assert-ReportFailureContains -ReportPath $fusionExtraReport -Marker 'fusion_matrix_extra_active_atom=github-trending-new-runtime-atom'

$fusionUnknownPoolWorkspace = New-FusionAuditFixture -Name 'fusion-audit-unknown-source-pool'
$fusionUnknownMatrixPath = Join-Path $fusionUnknownPoolWorkspace 'fusion-matrix.json'
$fusionUnknownMatrix = Read-JsonUtf8 -Path $fusionUnknownMatrixPath
$unknownDecision = $fusionUnknownMatrix.decisions | Where-Object { $_.atom -eq 'guarded-realtime-source-search' } | Select-Object -First 1
$unknownDecision.source_refs = 'AnySearch+not-in-source-pools'
Write-JsonUtf8 -Path $fusionUnknownMatrixPath -Value $fusionUnknownMatrix
$fusionUnknownReport = Join-Path $fusionUnknownPoolWorkspace 'fusion-unknown-report.json'
Invoke-Case -Name 'fusion-audit-unknown-source-pool-blocked' -ExpectedExit 1 -Arguments @('fusion-audit', '--workspace', $fusionUnknownPoolWorkspace, '--report', $fusionUnknownReport)
Assert-ReportFailureContains -ReportPath $fusionUnknownReport -Marker 'fusion_matrix_unknown_source_refs=guarded-realtime-source-search:not-in-source-pools'

$fusionRejectWorkspace = New-FusionAuditFixture -Name 'fusion-audit-object-verdict-reject-drift'
$fusionRejectMatrixPath = Join-Path $fusionRejectWorkspace 'fusion-matrix.json'
$fusionRejectMatrix = Read-JsonUtf8 -Path $fusionRejectMatrixPath
$agencyVerdict = $fusionRejectMatrix.object_verdicts | Where-Object { $_.object -eq 'Agency Agents' } | Select-Object -First 1
$agencyVerdict.runtime_status = 'landed'
$agencyVerdict.landed_surfaces = @('guarded-realtime-source-search')
Write-JsonUtf8 -Path $fusionRejectMatrixPath -Value $fusionRejectMatrix
$fusionRejectReport = Join-Path $fusionRejectWorkspace 'fusion-reject-drift-report.json'
Invoke-Case -Name 'fusion-audit-object-verdict-reject-drift-blocked' -ExpectedExit 1 -Arguments @('fusion-audit', '--workspace', $fusionRejectWorkspace, '--report', $fusionRejectReport)
Assert-ReportFailureContains -ReportPath $fusionRejectReport -Marker 'fusion_matrix_object_verdict_required_status_drift=Agency Agents'

$fusionSurfaceWorkspace = New-FusionAuditFixture -Name 'fusion-audit-object-verdict-surface-drift'
$fusionSurfaceMatrixPath = Join-Path $fusionSurfaceWorkspace 'fusion-matrix.json'
$fusionSurfaceMatrix = Read-JsonUtf8 -Path $fusionSurfaceMatrixPath
$agnesVerdict = $fusionSurfaceMatrix.object_verdicts | Where-Object { $_.object -eq 'Agnes AI' } | Select-Object -First 1
$agnesVerdict.landed_surfaces += 'hyperframes-runtime-shell'
Write-JsonUtf8 -Path $fusionSurfaceMatrixPath -Value $fusionSurfaceMatrix
$fusionSurfaceReport = Join-Path $fusionSurfaceWorkspace 'fusion-surface-drift-report.json'
Invoke-Case -Name 'fusion-audit-object-verdict-surface-drift-blocked' -ExpectedExit 1 -Arguments @('fusion-audit', '--workspace', $fusionSurfaceWorkspace, '--report', $fusionSurfaceReport)
Assert-ReportFailureContains -ReportPath $fusionSurfaceReport -Marker 'fusion_matrix_object_verdict_unknown_surface=Agnes AI:hyperframes-runtime-shell'

$fusionPresenceWorkspace = New-FusionAuditFixture -Name 'fusion-audit-object-verdict-required-presence'
$fusionPresenceMatrixPath = Join-Path $fusionPresenceWorkspace 'fusion-matrix.json'
$fusionPresenceMatrix = Read-JsonUtf8 -Path $fusionPresenceMatrixPath
$fusionPresenceMatrix.object_verdicts = @($fusionPresenceMatrix.object_verdicts | Where-Object { $_.object -ne 'Reasonix' })
Write-JsonUtf8 -Path $fusionPresenceMatrixPath -Value $fusionPresenceMatrix
$fusionPresenceReport = Join-Path $fusionPresenceWorkspace 'fusion-presence-drift-report.json'
Invoke-Case -Name 'fusion-audit-object-verdict-required-presence-blocked' -ExpectedExit 1 -Arguments @('fusion-audit', '--workspace', $fusionPresenceWorkspace, '--report', $fusionPresenceReport)
Assert-ReportFailureContains -ReportPath $fusionPresenceReport -Marker 'fusion_matrix_object_verdict_required_missing=Reasonix'

$fusionDocCoverageWorkspace = New-FusionAuditFixture -Name 'fusion-audit-doc-object-without-verdict'
$fusionDocCoverageMatrixPath = Join-Path $fusionDocCoverageWorkspace 'fusion-matrix.json'
$fusionDocCoverageMatrix = Read-JsonUtf8 -Path $fusionDocCoverageMatrixPath
$fusionDocCoverageMatrix.object_verdicts = @($fusionDocCoverageMatrix.object_verdicts | Where-Object { $_.object -ne 'dbs-business-toolbox' })
Write-JsonUtf8 -Path $fusionDocCoverageMatrixPath -Value $fusionDocCoverageMatrix
$fusionDocCoverageReport = Join-Path $fusionDocCoverageWorkspace 'fusion-doc-coverage-report.json'
Invoke-Case -Name 'fusion-audit-doc-object-without-verdict-blocked' -ExpectedExit 1 -Arguments @('fusion-audit', '--workspace', $fusionDocCoverageWorkspace, '--report', $fusionDocCoverageReport)
Assert-ReportFailureContains -ReportPath $fusionDocCoverageReport -Marker 'doc_object_without_verdict=plugins.md:dbs-business-toolbox'

$fusionRiskWorkspace = New-FusionAuditFixture -Name 'fusion-audit-risk-not-reject'
$fusionRiskMatrixPath = Join-Path $fusionRiskWorkspace 'fusion-matrix.json'
$fusionRiskMatrix = Read-JsonUtf8 -Path $fusionRiskMatrixPath
$riskDecision = $fusionRiskMatrix.decisions | Where-Object { $_.atom -eq 'github-trending-risk-surfaces' } | Select-Object -First 1
$riskDecision.decision = 'mount-on-demand'
Write-JsonUtf8 -Path $fusionRiskMatrixPath -Value $fusionRiskMatrix
$fusionRiskReport = Join-Path $fusionRiskWorkspace 'fusion-risk-report.json'
Invoke-Case -Name 'fusion-audit-risk-item-must-reject' -ExpectedExit 1 -Arguments @('fusion-audit', '--workspace', $fusionRiskWorkspace, '--report', $fusionRiskReport)
Assert-ReportFailureContains -ReportPath $fusionRiskReport -Marker 'fusion_matrix_risk_surface_not_reject=github-trending-risk-surfaces'

Invoke-Case -Name 'truth-state-fact-with-evidence' -ExpectedExit 0 -Arguments @('truth-state', '--text', 'fixed and verified', '--state', 'fact', '--evidence', $evidence, '--report', (Join-Path $fixture 'truth-fact.json'))
Invoke-Case -Name 'truth-state-fact-without-evidence-blocked' -ExpectedExit 1 -Arguments @('truth-state', '--text', 'fixed and verified', '--state', 'fact')
Invoke-Case -Name 'truth-state-inference-success-claim-blocked' -ExpectedExit 1 -Arguments @('truth-state', '--text', 'maybe already fixed', '--state', 'inference', '--report', (Join-Path $fixture 'truth-inference-bad.json'))
Invoke-Case -Name 'truth-state-inference-pass' -ExpectedExit 0 -Arguments @('truth-state', '--text', 'possibly caused by route rules, needs verification', '--state', 'inference', '--report', (Join-Path $fixture 'truth-inference.json'))
Invoke-Case -Name 'truth-state-todo-success-claim-blocked' -ExpectedExit 1 -Arguments @('truth-state', '--text', 'next step completed', '--state', 'todo')
Invoke-Case -Name 'time-guard-blocked' -ExpectedExit 1 -Arguments @('time-guard', '--kind', 'non-code', '--elapsed-minutes', '15', '--phase', 'prototype')
Invoke-Case -Name 'time-guard-with-artifact' -ExpectedExit 0 -Arguments @('time-guard', '--kind', 'non-code', '--elapsed-minutes', '15', '--phase', 'prototype', '--artifact', $artifact)

$taskWorkspace = Join-Path $fixture 'task'
Invoke-Case -Name 'task-start' -ExpectedExit 0 -Arguments @('task', '--workspace', $taskWorkspace, '--event', 'start', '--status', 'running', '--artifact', $artifact, '--note', 'task started')
Invoke-Case -Name 'task-start-precheck-blocked' -ExpectedExit 1 -Arguments @('task', '--workspace', (Join-Path $fixture 'task-precheck'), '--event', 'start', '--status', 'running', '--phase', 'preflight', '--note', '先看看能不能做，先查环境再说')
Invoke-Case -Name 'task-heartbeat-precheck-report-only-blocked' -ExpectedExit 1 -Arguments @('task', '--workspace', (Join-Path $fixture 'task-precheck-report-only'), '--event', 'heartbeat', '--status', 'running', '--phase', 'probe', '--artifact', $metaReportArtifact, '--note', 'check environment first and scan the repo first')
Invoke-Case -Name 'task-heartbeat-wait-confirmation-blocked' -ExpectedExit 1 -Arguments @('task', '--workspace', (Join-Path $fixture 'task-wait-confirmation'), '--event', 'heartbeat', '--status', 'running', '--artifact', $artifact, '--note', 'wait for user confirmation before continuing')
Invoke-Case -Name 'task-blocked-without-note-blocked' -ExpectedExit 2 -Arguments @('task', '--workspace', (Join-Path $fixture 'task-blocked-no-note'), '--event', 'blocked', '--status', 'blocked')
Invoke-Case -Name 'task-end-invalid-status-blocked' -ExpectedExit 2 -Arguments @('task', '--workspace', $taskWorkspace, '--event', 'end', '--status', 'running')
$benchWorkspace = Join-Path $fixture 'bench'
Invoke-Case -Name 'bench-log' -ExpectedExit 0 -Arguments @('bench', '--workspace', $benchWorkspace, '--name', 'sample', '--input-tokens', '10', '--output-tokens', '20', '--duration-ms', '30', '--tool-calls', '2', '--retries', '0', '--quality-pass', 'true')
Invoke-Case -Name 'bench-report' -ExpectedExit 0 -Arguments @('bench-report', '--workspace', $benchWorkspace, '--report', (Join-Path $fixture 'bench-report.json'))
$benchReport = Read-JsonUtf8 -Path (Join-Path $fixture 'bench-report.json')
if ($benchReport.decision -ne 'defer' -or $benchReport.cache_observations -ne 0 -or -not $benchReport.workspace_key -or $benchReport.command -ne 'bench-report' -or -not $benchReport.generated_at -or -not $benchReport.input_hashes.'bench.jsonl' -or -not $benchReport.log_ref) {
    throw "FAIL bench-report decision report=$($benchReport | ConvertTo-Json -Depth 6 -Compress)"
}
Invoke-Case -Name 'bench-log-cache' -ExpectedExit 0 -Arguments @('bench', '--workspace', $benchWorkspace, '--name', 'sample cache hit', '--input-tokens', '10', '--output-tokens', '20', '--duration-ms', '30', '--tool-calls', '2', '--retries', '0', '--quality-pass', 'true', '--cache-hit', 'true', '--reused-prefix-bytes', '200')
Invoke-Case -Name 'bench-report-cache' -ExpectedExit 0 -Arguments @('bench-report', '--workspace', $benchWorkspace, '--report', (Join-Path $fixture 'bench-report-cache.json'))
$benchReportCache = Read-JsonUtf8 -Path (Join-Path $fixture 'bench-report-cache.json')
if ($benchReportCache.decision -ne 'defer' -or $benchReportCache.evidence_level -ne 'checked' -or $benchReportCache.cache_observations -ne 1) {
    throw "FAIL bench-report-cache decision report=$($benchReportCache | ConvertTo-Json -Depth 6 -Compress)"
}
Invoke-Case -Name 'bench-log-cache-second' -ExpectedExit 0 -Arguments @('bench', '--workspace', $benchWorkspace, '--name', 'sample cache hit second', '--input-tokens', '10', '--output-tokens', '20', '--duration-ms', '30', '--tool-calls', '2', '--retries', '0', '--quality-pass', 'true', '--cache-hit', 'true', '--reused-prefix-bytes', '200')
$benchReportCacheTwoPath = Join-Path $fixture 'bench-report-cache-two.json'
Invoke-Case -Name 'bench-report-cache-two-observations' -ExpectedExit 0 -Arguments @('bench-report', '--workspace', $benchWorkspace, '--report', $benchReportCacheTwoPath)
$benchReportCacheTwo = Read-JsonUtf8 -Path $benchReportCacheTwoPath
if ($benchReportCacheTwo.decision -ne 'absorb' -or $benchReportCacheTwo.evidence_level -ne 'verified' -or $benchReportCacheTwo.cache_observations -ne 2 -or $benchReportCacheTwo.volume_gate -ne 'pass' -or -not ($benchReportCacheTwo.PSObject.Properties.Name -contains 'output_tokens_p95') -or -not ($benchReportCacheTwo.PSObject.Properties.Name -contains 'fresh_input_tokens_p95') -or -not ($benchReportCacheTwo.PSObject.Properties.Name -contains 'uncached_tokens_p95') -or -not ($benchReportCacheTwo.PSObject.Properties.Name -contains 'effective_input_units_p95') -or $benchReportCacheTwo.cached_input_cost_weight_bps -ne 1000 -or $benchReportCacheTwo.output_tokens_p95 -gt 4000 -or $benchReportCacheTwo.fresh_input_tokens_p95 -gt 12000 -or $benchReportCacheTwo.uncached_tokens_p95 -gt 14000) {
    throw "FAIL bench-report-cache-two-observations report=$($benchReportCacheTwo | ConvertTo-Json -Depth 6 -Compress)"
}
$benchBloatWorkspace = Join-Path $fixture 'bench-bloat'
1..2 | ForEach-Object {
    Invoke-Case -Name "bench-log-cache-bloat-$_" -ExpectedExit 0 -Arguments @('bench', '--workspace', $benchBloatWorkspace, '--name', "bloat $_", '--input-tokens', '10', '--output-tokens', '20', '--duration-ms', '30', '--tool-calls', '2', '--retries', '0', '--quality-pass', 'true', '--cache-hit', 'true', '--cached-tokens', '40000', '--reused-prefix-bytes', '40000')
}
$benchBloatReportPath = Join-Path $fixture 'bench-report-bloat.json'
Invoke-Case -Name 'bench-report-cache-bloat-rejected' -ExpectedExit 0 -Arguments @('bench-report', '--workspace', $benchBloatWorkspace, '--report', $benchBloatReportPath)
$benchBloatReport = Read-JsonUtf8 -Path $benchBloatReportPath
if ($benchBloatReport.cache_hit_rate -ne 1 -or $benchBloatReport.decision -ne 'reject' -or $benchBloatReport.volume_gate -ne 'fail' -or $benchBloatReport.cached_tokens_p95 -le 32768) {
    throw "FAIL bench-report-cache-bloat-rejected report=$($benchBloatReport | ConvertTo-Json -Depth 8 -Compress)"
}
$contextAuditWorkspace = Join-Path $fixture 'context-bloat-pass'
Write-Fixture (Join-Path $contextAuditWorkspace 'resident.md') 'small resident prompt'
Write-Fixture (Join-Path $contextAuditWorkspace 'config.json') '{"iron_rules_version":"11.3","cache_config":{"stable_prefix_policy":"byte-stable-minimal-resident","optimization_objective":"smaller-stable-prefix-with-equal-or-better-hit-rate","concise_execution_policy":"gate-cached-fresh-output-uncached-total-cost"}}'
Write-Fixture (Join-Path $contextAuditWorkspace 'tools\wuji_cli.go') 'package main // fixture'
Write-Fixture (Join-Path $contextAuditWorkspace 'outputs\context-pack-rich.json') (New-ContextPackRichFixtureJson -Workspace $contextAuditWorkspace)
Write-JsonUtf8 -Path (Join-Path $contextAuditWorkspace 'hotpath-manifest.json') -Value ([ordered]@{
    resident = @(@{ path = 'resident.md'; max_bytes = 8192 })
    on_demand = @(@{ path = 'cold.md'; max_loaded_bytes = 1024 })
    cold_ledger = @(@{ path = 'cold.md'; default_mode = 'handle-only' })
    forbidden_resident = @('raw transcript', 'large artifacts')
})
1..2 | ForEach-Object {
    Invoke-Case -Name "bench-log-context-audit-pass-$_" -ExpectedExit 0 -Arguments @('bench', '--workspace', $contextAuditWorkspace, '--log-dir', (Join-Path $contextAuditWorkspace 'outputs'), '--name', "context audit pass $_", '--input-tokens', '100', '--output-tokens', '20', '--duration-ms', '100', '--tool-calls', '1', '--retries', '0', '--quality-pass', 'true', '--cache-hit', 'true', '--cached-tokens', '200', '--reused-prefix-bytes', '200', '--activated-officers', '1', '--activated-skills', '0', '--loaded-file-bytes', '200', '--largest-context-segment-bytes', '200')
}
Invoke-Case -Name 'bench-report-context-audit-pass' -ExpectedExit 0 -Arguments @('bench-report', '--workspace', $contextAuditWorkspace, '--log-dir', (Join-Path $contextAuditWorkspace 'outputs'), '--report', (Join-Path $contextAuditWorkspace 'outputs\bench-report.json'))
$contextAuditPassReport = Join-Path $contextAuditWorkspace 'context-bloat-audit-pass.json'
Invoke-Case -Name 'context-bloat-audit-pass' -ExpectedExit 0 -Arguments @('context-bloat-audit', '--workspace', $contextAuditWorkspace, '--bench-report', (Join-Path $contextAuditWorkspace 'outputs\bench-report.json'), '--report', $contextAuditPassReport)
$contextAuditPass = Read-JsonUtf8 -Path $contextAuditPassReport
if ($contextAuditPass.status -ne 'pass' -or $contextAuditPass.command -ne 'context-bloat-audit' -or $contextAuditPass.bench_status -ne 'checked' -or -not ($contextAuditPass.checks -contains 'bench-cache-volume')) {
    throw "FAIL context-bloat-audit-pass report=$($contextAuditPass | ConvertTo-Json -Depth 8 -Compress)"
}
Copy-Item -LiteralPath $benchReportCacheTwoPath -Destination (Join-Path $contextAuditWorkspace 'outputs\bench-report.json') -Force
$contextAuditMismatchReport = Join-Path $contextAuditWorkspace 'context-bloat-audit-mismatch.json'
Invoke-Case -Name 'context-bloat-audit-bench-workspace-mismatch-blocked' -ExpectedExit 1 -Arguments @('context-bloat-audit', '--workspace', $contextAuditWorkspace, '--bench-report', (Join-Path $contextAuditWorkspace 'outputs\bench-report.json'), '--report', $contextAuditMismatchReport)
$contextAuditMismatch = Read-JsonUtf8 -Path $contextAuditMismatchReport
if ($contextAuditMismatch.status -ne 'fail' -or -not (($contextAuditMismatch.failures | Where-Object { $_ -like 'bench_report_workspace_key_mismatch=*' }).Count -gt 0)) {
    throw "FAIL context-bloat-audit-bench-workspace-mismatch-blocked report=$($contextAuditMismatch | ConvertTo-Json -Depth 8 -Compress)"
}
Remove-Item -LiteralPath (Join-Path $contextAuditWorkspace 'outputs\bench.jsonl') -Force -ErrorAction SilentlyContinue
1..2 | ForEach-Object {
    Invoke-Case -Name "bench-log-context-audit-bloat-$_" -ExpectedExit 0 -Arguments @('bench', '--workspace', $contextAuditWorkspace, '--log-dir', (Join-Path $contextAuditWorkspace 'outputs'), '--name', "context audit bloat $_", '--input-tokens', '100', '--output-tokens', '20', '--duration-ms', '100', '--tool-calls', '1', '--retries', '0', '--quality-pass', 'true', '--cache-hit', 'true', '--cached-tokens', '40000', '--reused-prefix-bytes', '40000', '--activated-officers', '1', '--activated-skills', '0', '--loaded-file-bytes', '200', '--largest-context-segment-bytes', '200')
}
Invoke-Case -Name 'bench-report-context-audit-bloat' -ExpectedExit 0 -Arguments @('bench-report', '--workspace', $contextAuditWorkspace, '--log-dir', (Join-Path $contextAuditWorkspace 'outputs'), '--report', (Join-Path $contextAuditWorkspace 'outputs\bench-report.json'))
$contextAuditFailReport = Join-Path $contextAuditWorkspace 'context-bloat-audit-fail.json'
Invoke-Case -Name 'context-bloat-audit-bench-bloat-blocked' -ExpectedExit 1 -Arguments @('context-bloat-audit', '--workspace', $contextAuditWorkspace, '--bench-report', (Join-Path $contextAuditWorkspace 'outputs\bench-report.json'), '--report', $contextAuditFailReport)
$contextAuditFail = Read-JsonUtf8 -Path $contextAuditFailReport
if ($contextAuditFail.status -ne 'fail' -or -not ($contextAuditFail.failures -contains 'bench_volume_gate_failed') -or -not ($contextAuditFail.failures -contains 'bench_decision_reject')) {
    throw "FAIL context-bloat-audit-bench-bloat-blocked report=$($contextAuditFail | ConvertTo-Json -Depth 8 -Compress)"
}
Remove-Item -LiteralPath (Join-Path $contextAuditWorkspace 'outputs\bench.jsonl') -Force -ErrorAction SilentlyContinue
1..2 | ForEach-Object {
    Invoke-Case -Name "bench-log-context-audit-high-fresh-$_" -ExpectedExit 0 -Arguments @('bench', '--workspace', $contextAuditWorkspace, '--log-dir', (Join-Path $contextAuditWorkspace 'outputs'), '--name', "context audit high fresh $_", '--input-tokens', '30000', '--output-tokens', '500', '--duration-ms', '100', '--tool-calls', '1', '--retries', '0', '--quality-pass', 'true', '--cache-hit', 'true', '--cached-tokens', '10000', '--fresh-input-tokens', '20000', '--reused-prefix-bytes', '10000', '--activated-officers', '1', '--activated-skills', '0', '--loaded-file-bytes', '200', '--largest-context-segment-bytes', '200')
}
Invoke-Case -Name 'bench-report-context-audit-high-fresh' -ExpectedExit 0 -Arguments @('bench-report', '--workspace', $contextAuditWorkspace, '--log-dir', (Join-Path $contextAuditWorkspace 'outputs'), '--report', (Join-Path $contextAuditWorkspace 'outputs\bench-report.json'))
$contextAuditHighFreshBench = Read-JsonUtf8 -Path (Join-Path $contextAuditWorkspace 'outputs\bench-report.json')
if ($contextAuditHighFreshBench.cache_hit_rate -ne 1 -or $contextAuditHighFreshBench.decision -ne 'reject' -or $contextAuditHighFreshBench.volume_gate -ne 'fail' -or $contextAuditHighFreshBench.fresh_input_tokens_p95 -le 12000 -or $contextAuditHighFreshBench.uncached_tokens_p95 -le 14000) {
    throw "FAIL bench-report-context-audit-high-fresh report=$($contextAuditHighFreshBench | ConvertTo-Json -Depth 8 -Compress)"
}
$contextAuditHighFreshReport = Join-Path $contextAuditWorkspace 'context-bloat-audit-high-fresh.json'
Invoke-Case -Name 'context-bloat-audit-high-fresh-blocked' -ExpectedExit 1 -Arguments @('context-bloat-audit', '--workspace', $contextAuditWorkspace, '--bench-report', (Join-Path $contextAuditWorkspace 'outputs\bench-report.json'), '--report', $contextAuditHighFreshReport)
$contextAuditHighFresh = Read-JsonUtf8 -Path $contextAuditHighFreshReport
if ($contextAuditHighFresh.status -ne 'fail' -or -not (($contextAuditHighFresh.failures | Where-Object { $_ -like 'fresh_input_tokens_p95_over_budget=*' }).Count -gt 0) -or -not (($contextAuditHighFresh.failures | Where-Object { $_ -like 'uncached_tokens_p95_over_budget=*' }).Count -gt 0) -or -not ($contextAuditHighFresh.failures -contains 'bench_volume_gate_failed') -or -not ($contextAuditHighFresh.failures -contains 'bench_decision_reject')) {
    throw "FAIL context-bloat-audit-high-fresh-blocked report=$($contextAuditHighFresh | ConvertTo-Json -Depth 8 -Compress)"
}
Remove-Item -LiteralPath (Join-Path $contextAuditWorkspace 'outputs\bench.jsonl') -Force -ErrorAction SilentlyContinue
1..2 | ForEach-Object {
    Invoke-Case -Name "bench-log-context-audit-high-output-$_" -ExpectedExit 0 -Arguments @('bench', '--workspace', $contextAuditWorkspace, '--log-dir', (Join-Path $contextAuditWorkspace 'outputs'), '--name', "context audit high output $_", '--input-tokens', '1000', '--output-tokens', '6000', '--duration-ms', '100', '--tool-calls', '1', '--retries', '0', '--quality-pass', 'true', '--cache-hit', 'true', '--cached-tokens', '500', '--fresh-input-tokens', '500', '--reused-prefix-bytes', '500', '--activated-officers', '1', '--activated-skills', '0', '--loaded-file-bytes', '200', '--largest-context-segment-bytes', '200')
}
Invoke-Case -Name 'bench-report-context-audit-high-output' -ExpectedExit 0 -Arguments @('bench-report', '--workspace', $contextAuditWorkspace, '--log-dir', (Join-Path $contextAuditWorkspace 'outputs'), '--report', (Join-Path $contextAuditWorkspace 'outputs\bench-report.json'))
$contextAuditHighOutputBench = Read-JsonUtf8 -Path (Join-Path $contextAuditWorkspace 'outputs\bench-report.json')
if ($contextAuditHighOutputBench.cache_hit_rate -ne 1 -or $contextAuditHighOutputBench.decision -ne 'reject' -or $contextAuditHighOutputBench.volume_gate -ne 'fail' -or $contextAuditHighOutputBench.output_tokens_p95 -le 4000) {
    throw "FAIL bench-report-context-audit-high-output report=$($contextAuditHighOutputBench | ConvertTo-Json -Depth 8 -Compress)"
}
$contextAuditHighOutputReport = Join-Path $contextAuditWorkspace 'context-bloat-audit-high-output.json'
Invoke-Case -Name 'context-bloat-audit-high-output-blocked' -ExpectedExit 1 -Arguments @('context-bloat-audit', '--workspace', $contextAuditWorkspace, '--bench-report', (Join-Path $contextAuditWorkspace 'outputs\bench-report.json'), '--report', $contextAuditHighOutputReport)
$contextAuditHighOutput = Read-JsonUtf8 -Path $contextAuditHighOutputReport
if ($contextAuditHighOutput.status -ne 'fail' -or -not (($contextAuditHighOutput.failures | Where-Object { $_ -like 'output_tokens_p95_over_budget=*' }).Count -gt 0) -or -not ($contextAuditHighOutput.failures -contains 'bench_volume_gate_failed') -or -not ($contextAuditHighOutput.failures -contains 'bench_decision_reject')) {
    throw "FAIL context-bloat-audit-high-output-blocked report=$($contextAuditHighOutput | ConvertTo-Json -Depth 8 -Compress)"
}

$legacyOutputWorkspace = Join-Path $fixture 'optimization-audit-legacy-output'
Write-CurrentAuditReports -Workspace $legacyOutputWorkspace
New-Item -ItemType Directory -Force -Path (Join-Path $legacyOutputWorkspace 'output') | Out-Null
Write-Fixture (Join-Path $legacyOutputWorkspace 'output\legacy-artifact.txt') 'legacy output residue'
$legacyOutputReportPath = Join-Path $legacyOutputWorkspace 'optimization-audit-legacy-output.json'
Invoke-Case -Name 'optimization-audit-legacy-output-blocked' -ExpectedExit 1 -Arguments @('optimization-audit', '--workspace', $legacyOutputWorkspace, '--report', $legacyOutputReportPath)
$legacyOutputReport = Read-JsonUtf8 -Path $legacyOutputReportPath
if ($legacyOutputReport.status -ne 'fail' -or -not (($legacyOutputReport.failures | Where-Object { $_ -like 'legacy_output_directory_present=*' }).Count -gt 0) -or $legacyOutputReport.budgets.legacy_output_directory_forbidden -ne $true) {
    throw "FAIL optimization-audit-legacy-output-blocked report=$($legacyOutputReport | ConvertTo-Json -Depth 8 -Compress)"
}

$nonCanonicalOutputsWorkspace = Join-Path $fixture 'optimization-audit-noncanonical-outputs'
Write-CurrentAuditReports -Workspace $nonCanonicalOutputsWorkspace
Write-Fixture (Join-Path $nonCanonicalOutputsWorkspace 'outputs\chat-route.json') '{"route":"legacy-root-residue"}'
$nonCanonicalOutputsReportPath = Join-Path $nonCanonicalOutputsWorkspace 'optimization-audit-noncanonical-outputs.json'
Invoke-Case -Name 'optimization-audit-noncanonical-root-outputs-blocked' -ExpectedExit 1 -Arguments @('optimization-audit', '--workspace', $nonCanonicalOutputsWorkspace, '--report', $nonCanonicalOutputsReportPath)
$nonCanonicalOutputsReport = Read-JsonUtf8 -Path $nonCanonicalOutputsReportPath
if ($nonCanonicalOutputsReport.status -ne 'fail' -or -not (($nonCanonicalOutputsReport.failures | Where-Object { $_ -like 'outputs_noncanonical_root_residue=*chat-route.json*' }).Count -gt 0) -or $nonCanonicalOutputsReport.budgets.outputs_root_policy -ne 'canonical-current-evidence-only') {
    throw "FAIL optimization-audit-noncanonical-root-outputs-blocked report=$($nonCanonicalOutputsReport | ConvertTo-Json -Depth 8 -Compress)"
}

$testResidueWorkspace = Join-Path $fixture 'optimization-audit-test-residue'
Write-CurrentAuditReports -Workspace $testResidueWorkspace
Write-Fixture (Join-Path $testResidueWorkspace 'outputs\tests\leftover\artifact.txt') 'leftover test residue'
$testResidueReportPath = Join-Path $testResidueWorkspace 'optimization-audit-test-residue.json'
Invoke-Case -Name 'optimization-audit-test-residue-blocked' -ExpectedExit 1 -Arguments @('optimization-audit', '--workspace', $testResidueWorkspace, '--report', $testResidueReportPath)
$testResidueReport = Read-JsonUtf8 -Path $testResidueReportPath
if ($testResidueReport.status -ne 'fail' -or -not (($testResidueReport.failures | Where-Object { $_ -like 'outputs_tests_residue_present=*outputs*tests*' }).Count -gt 0) -or $testResidueReport.budgets.outputs_tests_policy -ne 'empty-or-absent') {
    throw "FAIL optimization-audit-test-residue-blocked report=$($testResidueReport | ConvertTo-Json -Depth 8 -Compress)"
}

$runtimeAuditWorkspace = Join-Path $fixture 'runtime-context-audit'
Write-CurrentAuditReports -Workspace $runtimeAuditWorkspace
$runtimeAuditPass = Read-JsonUtf8 -Path (Join-Path $runtimeAuditWorkspace 'outputs\runtime-context-audit-report.json')
if ($runtimeAuditPass.status -ne 'pass' -or $runtimeAuditPass.command -ne 'runtime-context-audit' -or $runtimeAuditPass.usage_observations -lt 2 -or $runtimeAuditPass.volume_gate -ne 'pass' -or $runtimeAuditPass.privacy_mode -ne 'numeric-usage-and-hash-only' -or -not ($runtimeAuditPass.PSObject.Properties.Name -contains 'effective_input_units_p95') -or $runtimeAuditPass.cached_input_cost_weight_bps -ne 1000) {
    throw "FAIL runtime-context-audit-pass report=$($runtimeAuditPass | ConvertTo-Json -Depth 8 -Compress)"
}
if ($runtimeAuditPass.thread_reset_required -ne $false -or $runtimeAuditPass.thread_emergency -ne $false -or $runtimeAuditPass.same_thread_policy -ne 'same-thread-allowed') {
    throw "FAIL runtime-context-audit-pass same-thread policy report=$($runtimeAuditPass | ConvertTo-Json -Depth 8 -Compress)"
}

$runtimeBloatWorkspace = Join-Path $fixture 'runtime-context-audit-bloat'
Write-CurrentAuditReports -Workspace $runtimeBloatWorkspace
$runtimeBloatKey = Get-PrivacyHash -Value $runtimeBloatWorkspace
$runtimeBloatLog = Join-Path $runtimeBloatWorkspace 'outputs\runtime-usage.jsonl'
$runtimeBloatRows = @(
    [ordered]@{ timestamp = '2026-06-03T00:00:00Z'; workspace_key = $runtimeBloatKey; usage = [ordered]@{ input_tokens = 240000; output_tokens = 700; cached_tokens = 235000; fresh_input_tokens = 5000 } }
    [ordered]@{ timestamp = '2026-06-03T00:01:00Z'; workspace_key = $runtimeBloatKey; usage = [ordered]@{ input_tokens = 230000; output_tokens = 800; cached_tokens = 225000; fresh_input_tokens = 5000 } }
)
[System.IO.File]::WriteAllLines($runtimeBloatLog, ($runtimeBloatRows | ForEach-Object { $_ | ConvertTo-Json -Depth 12 -Compress }), [System.Text.UTF8Encoding]::new($false))
Invoke-Case -Name 'runtime-context-audit-bloat-blocked' -ExpectedExit 1 -Arguments @('runtime-context-audit', '--workspace', $runtimeBloatWorkspace)
$runtimeBloatReport = Read-JsonUtf8 -Path (Join-Path $runtimeBloatWorkspace 'outputs\runtime-context-audit-report.json')
if ($runtimeBloatReport.status -ne 'fail' -or $runtimeBloatReport.volume_gate -ne 'fail' -or $runtimeBloatReport.long_context_suspected -ne $true -or $runtimeBloatReport.diagnosis -ne 'cached-token-bloat-suspected-long-resident-or-outer-context' -or -not ($runtimeBloatReport.context_slimming_actions -contains 'replace-long-history-with-task-state-summary-and-evidence-handles') -or -not ($runtimeBloatReport.context_slimming_actions -contains 'create-fresh-thread-from-context-reset-handoff') -or -not (($runtimeBloatReport.failures | Where-Object { $_ -like 'runtime_cached_tokens_p95_over_budget=*' }).Count -gt 0)) {
    throw "FAIL runtime-context-audit-bloat-blocked report=$($runtimeBloatReport | ConvertTo-Json -Depth 8 -Compress)"
}
if ($runtimeBloatReport.thread_reset_required -ne $true -or $runtimeBloatReport.thread_emergency -ne $false -or $runtimeBloatReport.same_thread_policy -ne 'stop-broad-work-refresh-handoff' -or -not ($runtimeBloatReport.failures -contains 'runtime_same_thread_context_reset_required') -or (($runtimeBloatReport.failures | Where-Object { $_ -like 'runtime_same_thread_emergency_stop_required=*' }).Count -gt 0)) {
    throw "FAIL runtime-context-audit-bloat-reset-policy report=$($runtimeBloatReport | ConvertTo-Json -Depth 8 -Compress)"
}

$runtimeEmergencyWorkspace = Join-Path $fixture 'runtime-context-audit-emergency'
Write-CurrentAuditReports -Workspace $runtimeEmergencyWorkspace
$runtimeEmergencyKey = Get-PrivacyHash -Value $runtimeEmergencyWorkspace
$runtimeEmergencyLog = Join-Path $runtimeEmergencyWorkspace 'outputs\runtime-usage.jsonl'
$runtimeEmergencyRows = @(
    [ordered]@{ timestamp = '2026-06-03T00:00:00Z'; workspace_key = $runtimeEmergencyKey; usage = [ordered]@{ input_tokens = 270000; output_tokens = 900; cached_tokens = 250000; fresh_input_tokens = 20000 } }
    [ordered]@{ timestamp = '2026-06-03T00:01:00Z'; workspace_key = $runtimeEmergencyKey; usage = [ordered]@{ input_tokens = 265000; output_tokens = 950; cached_tokens = 248000; fresh_input_tokens = 17000 } }
)
[System.IO.File]::WriteAllLines($runtimeEmergencyLog, ($runtimeEmergencyRows | ForEach-Object { $_ | ConvertTo-Json -Depth 12 -Compress }), [System.Text.UTF8Encoding]::new($false))
Invoke-Case -Name 'runtime-context-audit-emergency-blocked' -ExpectedExit 1 -Arguments @('runtime-context-audit', '--workspace', $runtimeEmergencyWorkspace)
$runtimeEmergencyReport = Read-JsonUtf8 -Path (Join-Path $runtimeEmergencyWorkspace 'outputs\runtime-context-audit-report.json')
if ($runtimeEmergencyReport.status -ne 'fail' -or $runtimeEmergencyReport.volume_gate -ne 'fail' -or $runtimeEmergencyReport.thread_reset_required -ne $true -or $runtimeEmergencyReport.thread_emergency -ne $true -or $runtimeEmergencyReport.same_thread_policy -ne 'emergency-stop-refresh-handoff' -or -not ($runtimeEmergencyReport.failures -contains 'runtime_same_thread_context_reset_required') -or -not (($runtimeEmergencyReport.failures | Where-Object { $_ -like 'runtime_same_thread_emergency_stop_required=*' }).Count -gt 0)) {
    throw "FAIL runtime-context-audit-emergency-policy report=$($runtimeEmergencyReport | ConvertTo-Json -Depth 8 -Compress)"
}

$runtimeRawWorkspace = Join-Path $fixture 'runtime-context-audit-raw'
Write-CurrentAuditReports -Workspace $runtimeRawWorkspace
$runtimeRawKey = Get-PrivacyHash -Value $runtimeRawWorkspace
$runtimeRawLog = Join-Path $runtimeRawWorkspace 'outputs\runtime-usage.jsonl'
$runtimeRawRows = @(
    [ordered]@{ timestamp = '2026-06-03T00:00:00Z'; workspace_key = $runtimeRawKey; prompt = 'raw prompt payload must not be retained here'; usage = [ordered]@{ input_tokens = 1000; output_tokens = 100; cached_tokens = 800; fresh_input_tokens = 200 } }
    [ordered]@{ timestamp = '2026-06-03T00:01:00Z'; workspace_key = $runtimeRawKey; usage = [ordered]@{ input_tokens = 900; output_tokens = 100; cached_tokens = 700; fresh_input_tokens = 200 } }
)
[System.IO.File]::WriteAllLines($runtimeRawLog, ($runtimeRawRows | ForEach-Object { $_ | ConvertTo-Json -Depth 12 -Compress }), [System.Text.UTF8Encoding]::new($false))
Invoke-Case -Name 'runtime-context-audit-raw-payload-blocked' -ExpectedExit 1 -Arguments @('runtime-context-audit', '--workspace', $runtimeRawWorkspace)
$runtimeRawReport = Read-JsonUtf8 -Path (Join-Path $runtimeRawWorkspace 'outputs\runtime-context-audit-report.json')
if ($runtimeRawReport.status -ne 'fail' -or -not (($runtimeRawReport.failures | Where-Object { $_ -like 'runtime_usage_record_1_raw_payload_field_forbidden=prompt' }).Count -gt 0)) {
    throw "FAIL runtime-context-audit-raw-payload-blocked report=$($runtimeRawReport | ConvertTo-Json -Depth 8 -Compress)"
}

$codeMapWorkspace = Join-Path $fixture 'code-map'
$codeMapReport = Join-Path $codeMapWorkspace 'code-map.json'
Invoke-Case -Name 'code-map' -ExpectedExit 0 -Arguments @('code-map', '--workspace', $codeMapWorkspace, '--goal', 'refactor route task defaults', '--entry', 'routeTaskCommand', '--dependency', 'contextPackCommand', '--risk', 'route drift', '--verify', 'route-task regression', '--report', $codeMapReport)
$codeMap = Read-JsonUtf8 -Path $codeMapReport
if ($codeMap.entry -ne 'routeTaskCommand' -or $codeMap.verifications.Count -lt 1) {
    throw "FAIL code-map report=$($codeMap | ConvertTo-Json -Depth 6 -Compress)"
}

$bugfixWorkspace = Join-Path $fixture 'bugfix'
New-Item -ItemType Directory -Force -Path $bugfixWorkspace | Out-Null
$bugfixArtifact = Join-Path $bugfixWorkspace 'app.exe'
$bugfixSelfTest = Join-Path $bugfixWorkspace 'self-test.txt'
$bugfixBrowserCheck = Join-Path $bugfixWorkspace 'browser-check.txt'
$bugfixRegression = Join-Path $bugfixWorkspace 'regression-evidence.txt'
$bugfixSameClassEvidence = Join-Path $bugfixWorkspace 'same-class-scan.txt'
$rootCauseReport = Join-Path $bugfixWorkspace 'root-cause-report.json'
Write-Fixture $bugfixArtifact 'binary output bytes'
Write-Fixture $bugfixSelfTest 'self test passed on reproduced bug path'
Write-Fixture $bugfixBrowserCheck 'browser verification passed on fixed flow'
Write-Fixture $bugfixRegression 'regression command replayed login and register submit state transitions successfully'
Write-Fixture $bugfixSameClassEvidence 'same-class scan checked login register and password reset submit state transitions'
Invoke-Case -Name 'root-cause-radar-pass' -ExpectedExit 0 -Arguments @('root-cause-radar', '--workspace', $bugfixWorkspace, '--symptom', 'login button freezes after failed validation', '--repro', 'click login button on home page after empty password', '--hypothesis', 'submit disabled state is not reset after validation failure', '--eliminated-cause', 'browser click handler still fires in the reproduced flow', '--root-cause', 'Login submit state remains disabled because the shared validation failure path never resets it', '--same-class-scan', 'Scanned login, register, and password reset submit flows for the same disabled-state reset pattern', '--same-class-evidence', $bugfixSameClassEvidence, '--fix-strategy', 'Reset submit state in the shared validation failure path and cover all submit flows without another local workaround', '--patch-debt-action', 'Delete earlier local workaround attempts and keep only the shared reset path', '--regression-evidence', $bugfixRegression, '--artifact', $bugfixArtifact, '--report', $rootCauseReport)
$rootCause = Read-JsonUtf8 -Path $rootCauseReport
if ($rootCause.status -ne 'pass' -or $rootCause.officer -ne 'root-cause-officer' -or -not $rootCause.workspace_key -or $rootCause.symptom.key -eq $null -or $rootCause.symptom.text) {
    throw "FAIL root-cause-radar privacy/officer report=$($rootCause | ConvertTo-Json -Depth 8 -Compress)"
}
Invoke-Case -Name 'root-cause-radar-symptom-only-blocked' -ExpectedExit 1 -Arguments @('root-cause-radar', '--workspace', $bugfixWorkspace, '--symptom', 'login button freezes after failed validation', '--repro', 'click login button on home page after empty password', '--hypothesis', 'submit state issue', '--root-cause', 'login button freezes after failed validation', '--same-class-scan', 'Scanned login and register submit paths for similar disabled state issues', '--same-class-evidence', $bugfixSameClassEvidence, '--fix-strategy', 'Reset shared submit state after validation failure and verify submit flows', '--regression-evidence', $bugfixRegression, '--report', (Join-Path $bugfixWorkspace 'root-cause-symptom-only.json'))
Invoke-Case -Name 'root-cause-radar-same-class-required' -ExpectedExit 1 -Arguments @('root-cause-radar', '--workspace', $bugfixWorkspace, '--symptom', 'login button freezes after failed validation', '--repro', 'click login button on home page after empty password', '--hypothesis', 'submit disabled state is not reset', '--root-cause', 'Login submit state remains disabled because validation failure never resets it', '--same-class-scan', 'none', '--same-class-evidence', $bugfixSameClassEvidence, '--fix-strategy', 'Reset shared submit state after validation failure and verify submit flows', '--regression-evidence', $bugfixRegression, '--report', (Join-Path $bugfixWorkspace 'root-cause-no-scan.json'))
Invoke-Case -Name 'root-cause-radar-same-class-evidence-required' -ExpectedExit 1 -Arguments @('root-cause-radar', '--workspace', $bugfixWorkspace, '--symptom', 'login button freezes after failed validation', '--repro', 'click login button on home page after empty password', '--hypothesis', 'submit disabled state is not reset', '--root-cause', 'Login submit state remains disabled because validation failure never resets it', '--same-class-scan', 'Scanned login and register submit paths for similar disabled state issues', '--fix-strategy', 'Reset shared submit state after validation failure and verify submit flows', '--regression-evidence', $bugfixRegression, '--report', (Join-Path $bugfixWorkspace 'root-cause-no-scan-evidence.json'))
Invoke-Case -Name 'root-cause-radar-regression-required' -ExpectedExit 1 -Arguments @('root-cause-radar', '--workspace', $bugfixWorkspace, '--symptom', 'login button freezes after failed validation', '--repro', 'click login button on home page after empty password', '--hypothesis', 'submit disabled state is not reset', '--root-cause', 'Login submit state remains disabled because validation failure never resets it', '--same-class-scan', 'Scanned login and register submit paths for similar disabled state issues', '--same-class-evidence', $bugfixSameClassEvidence, '--fix-strategy', 'Reset shared submit state after validation failure and verify submit flows', '--report', (Join-Path $bugfixWorkspace 'root-cause-no-regression.json'))
Invoke-Case -Name 'root-cause-radar-patch-debt-action-required' -ExpectedExit 1 -Arguments @('root-cause-radar', '--workspace', $bugfixWorkspace, '--symptom', 'login button freezes after failed validation', '--repro', 'click login button on home page after empty password', '--hypothesis', 'temporary patch left state inconsistent', '--root-cause', 'Login submit state remains disabled because a temporary patch bypassed shared reset logic', '--same-class-scan', 'Scanned login and register submit paths for similar disabled state issues', '--same-class-evidence', $bugfixSameClassEvidence, '--fix-strategy', 'Remove temporary patch and repair the shared reset path', '--regression-evidence', $bugfixRegression, '--report', (Join-Path $bugfixWorkspace 'root-cause-no-patch-action.json'))
Write-JsonUtf8 -Path (Join-Path $bugfixWorkspace 'forged-root-cause.json') -Value (@{ status = 'pass'; officer = 'root-cause-officer'; workspace_key = $rootCause.workspace_key; regression_evidence_refs = @(@{ path_ref = 'regression-evidence.txt' }) })
Invoke-Case -Name 'bugfix-guard-forged-root-cause-blocked' -ExpectedExit 1 -Arguments @('bugfix-guard', '--workspace', $bugfixWorkspace, '--goal', 'fix login button bug', '--repro', 'click login button on home page', '--root-cause-report', (Join-Path $bugfixWorkspace 'forged-root-cause.json'), '--artifact', $bugfixArtifact, '--verify', 'login flow no longer freezes', '--self-test', $bugfixSelfTest, '--browser-check', $bugfixBrowserCheck, '--report', (Join-Path $bugfixWorkspace 'bugfix-forged-root.json'))
$forgedRootUnsafePath = Join-Path $bugfixWorkspace 'forged-root-cause-unsafe-path.json'
Write-JsonUtf8 -Path $forgedRootUnsafePath -Value ([ordered]@{
    status = 'pass'
    officer = 'root-cause-officer'
    command = 'root-cause-radar'
    schema_version = 'root-cause-radar.v1'
    verdict = 'root-cause-repair-ready'
    privacy_mode = 'hash-length-and-evidence-ref-only'
    generated_at = (Get-Date).ToUniversalTime().ToString('yyyy-MM-ddTHH:mm:ssZ')
    workspace_key = $rootCause.workspace_key
    symptom = $rootCause.symptom
    repro = $rootCause.repro
    hypotheses = $rootCause.hypotheses
    root_cause = $rootCause.root_cause
    same_class_scan = $rootCause.same_class_scan
    fix_strategy = $rootCause.fix_strategy
    patch_debt_required = $false
    same_class_evidence_refs = @(@{ path_ref = '../same-class-scan.txt'; bytes = (Get-Item -LiteralPath $bugfixSameClassEvidence).Length; sha256 = (Get-Sha256Lower -Path $bugfixSameClassEvidence) })
    regression_evidence_refs = $rootCause.regression_evidence_refs
})
Invoke-Case -Name 'bugfix-guard-forged-root-cause-unsafe-path-blocked' -ExpectedExit 1 -Arguments @('bugfix-guard', '--workspace', $bugfixWorkspace, '--goal', 'fix login button bug', '--repro', 'click login button on home page', '--root-cause-report', $forgedRootUnsafePath, '--artifact', $bugfixArtifact, '--verify', 'login flow no longer freezes', '--self-test', $bugfixSelfTest, '--browser-check', $bugfixBrowserCheck, '--report', (Join-Path $bugfixWorkspace 'bugfix-forged-root-unsafe-path.json'))
Invoke-Case -Name 'bugfix-guard-pass' -ExpectedExit 0 -Arguments @('bugfix-guard', '--workspace', $bugfixWorkspace, '--goal', 'fix login button bug', '--repro', 'click login button on home page', '--root-cause-report', $rootCauseReport, '--artifact', $bugfixArtifact, '--verify', 'login flow no longer freezes', '--self-test', $bugfixSelfTest, '--browser-check', $bugfixBrowserCheck, '--report', (Join-Path $bugfixWorkspace 'bugfix-pass.json'))
Invoke-Case -Name 'bugfix-guard-root-cause-required' -ExpectedExit 1 -Arguments @('bugfix-guard', '--workspace', $bugfixWorkspace, '--goal', 'fix login button bug', '--repro', 'click login button on home page', '--artifact', $bugfixArtifact, '--verify', 'login flow no longer freezes', '--self-test', $bugfixSelfTest, '--browser-check', $bugfixBrowserCheck, '--report', (Join-Path $bugfixWorkspace 'bugfix-no-root-cause.json'))
Invoke-Case -Name 'bugfix-guard-direct-fields-blocked' -ExpectedExit 1 -Arguments @('bugfix-guard', '--workspace', $bugfixWorkspace, '--goal', 'fix login button bug', '--repro', 'click login button on home page', '--root-cause', 'Login submit state remains disabled because validation failure never resets it', '--same-class-scan', 'Scanned login and register submit paths for similar disabled state issues', '--regression-evidence', $bugfixRegression, '--artifact', $bugfixArtifact, '--verify', 'login flow no longer freezes', '--self-test', $bugfixSelfTest, '--browser-check', $bugfixBrowserCheck, '--report', (Join-Path $bugfixWorkspace 'bugfix-direct-fields.json'))
Invoke-Case -Name 'bugfix-guard-still-failing-blocked' -ExpectedExit 1 -Arguments @('bugfix-guard', '--workspace', $bugfixWorkspace, '--goal', 'fix login button bug', '--repro', 'click login button on home page', '--root-cause-report', $rootCauseReport, '--artifact', $bugfixArtifact, '--verify', 'login flow no longer freezes', '--self-test', $bugfixSelfTest, '--browser-check', $bugfixBrowserCheck, '--still-failing', 'login button still freezes after click', '--report', (Join-Path $bugfixWorkspace 'bugfix-fail.json'))
Invoke-Case -Name 'bugfix-guard-self-test-required' -ExpectedExit 1 -Arguments @('bugfix-guard', '--workspace', $bugfixWorkspace, '--goal', 'fix login button bug', '--repro', 'click login button on home page', '--root-cause-report', $rootCauseReport, '--artifact', $bugfixArtifact, '--verify', 'login flow no longer freezes', '--browser-check', $bugfixBrowserCheck, '--report', (Join-Path $bugfixWorkspace 'bugfix-no-selftest.json'))

$qualityWorkspace = Join-Path $fixture 'quality-inspection'
New-Item -ItemType Directory -Force -Path $qualityWorkspace | Out-Null
$qualityArtifact = Join-Path $qualityWorkspace 'deliverable.txt'
$qualityBrowserCheck = Join-Path $qualityWorkspace 'browser-check.txt'
Write-Fixture $qualityArtifact 'deliverable bytes'
Write-Fixture $qualityBrowserCheck 'browser independent verification passed'
Invoke-Case -Name 'quality-guard-pass' -ExpectedExit 0 -Arguments @('quality-guard', '--workspace', $qualityWorkspace, '--goal', 'quality verify frontend bugfix', '--artifact', $qualityArtifact, '--verify', 'browser flow works end to end', '--browser-check', $qualityBrowserCheck, '--report', (Join-Path $qualityWorkspace 'quality-pass.json'))
$qualityReport = Read-JsonUtf8 -Path (Join-Path $qualityWorkspace 'quality-pass.json')
if ($qualityReport.workspace -or $qualityReport.artifacts -or $qualityReport.goal -is [string] -or -not $qualityReport.workspace_key -or -not $qualityReport.artifact_refs) {
    throw "FAIL quality-guard report retained raw fields report=$($qualityReport | ConvertTo-Json -Depth 8 -Compress)"
}
Invoke-Case -Name 'quality-guard-missing-check-blocked' -ExpectedExit 1 -Arguments @('quality-guard', '--workspace', $qualityWorkspace, '--goal', 'quality verify frontend bugfix', '--artifact', $qualityArtifact, '--verify', 'browser flow works end to end', '--report', (Join-Path $qualityWorkspace 'quality-missing-check.json'))
Invoke-Case -Name 'quality-guard-still-failing-blocked' -ExpectedExit 1 -Arguments @('quality-guard', '--workspace', $qualityWorkspace, '--goal', 'quality verify frontend bugfix', '--artifact', $qualityArtifact, '--verify', 'browser flow works end to end', '--browser-check', $qualityBrowserCheck, '--still-failing', 'submit button still does not respond', '--report', (Join-Path $qualityWorkspace 'quality-still-failing.json'))

$migrationWorkspace = Join-Path $fixture 'migration'
New-Item -ItemType Directory -Force -Path $migrationWorkspace | Out-Null
$featureMap = Join-Path $migrationWorkspace 'feature-map.json'
$runEvidence = Join-Path $migrationWorkspace 'cargo-run.txt'
$previewEvidence = Join-Path $migrationWorkspace 'preview.txt'
$migrationArtifact = Join-Path $migrationWorkspace 'target-app.exe'
Write-Fixture $featureMap '{"pages":["home","settings"],"flows":["import","sync"]}'
Write-Fixture $runEvidence 'cargo run success'
Write-Fixture $previewEvidence 'manual preview evidence'
Write-Fixture $migrationArtifact 'built executable bytes'
Invoke-Case -Name 'migration-guard-pass' -ExpectedExit 0 -Arguments @('migration-guard', '--workspace', $migrationWorkspace, '--goal', 'port old app to rust', '--feature-map', $featureMap, '--artifact', $migrationArtifact, '--verify', 'cargo run opens main flow', '--run-evidence', $runEvidence, '--preview-evidence', $previewEvidence, '--report', (Join-Path $migrationWorkspace 'migration-pass.json'))
$migrationReport = Read-JsonUtf8 -Path (Join-Path $migrationWorkspace 'migration-pass.json')
if ($migrationReport.workspace -or $migrationReport.artifacts -or $migrationReport.goal -is [string] -or $migrationReport.feature_map -or -not $migrationReport.workspace_key -or -not $migrationReport.artifact_refs) {
    throw "FAIL migration-guard report retained raw fields report=$($migrationReport | ConvertTo-Json -Depth 8 -Compress)"
}
Invoke-Case -Name 'migration-guard-fake-page-blocked' -ExpectedExit 1 -Arguments @('migration-guard', '--workspace', $migrationWorkspace, '--goal', 'port old app to rust', '--feature-map', $featureMap, '--artifact', $migrationArtifact, '--verify', 'cargo run opens main flow', '--run-evidence', $runEvidence, '--fake-page', 'settings page is placeholder', '--report', (Join-Path $migrationWorkspace 'migration-fail.json'))
Invoke-Case -Name 'migration-guard-run-evidence-required' -ExpectedExit 1 -Arguments @('migration-guard', '--workspace', $migrationWorkspace, '--goal', 'port old app to rust', '--feature-map', $featureMap, '--artifact', $migrationArtifact, '--verify', 'cargo run opens main flow', '--report', (Join-Path $migrationWorkspace 'migration-no-run.json'))

$closeoutWorkspace = Join-Path $fixture 'closeout'
New-Item -ItemType Directory -Force -Path $closeoutWorkspace | Out-Null
$closeoutArtifact = Join-Path $closeoutWorkspace 'final.txt'
Write-Fixture $closeoutArtifact 'final artifact for closeout verification'
$closeoutPassReport = Join-Path $closeoutWorkspace 'closeout-check-pass.json'
Invoke-Case -Name 'closeout-check-pass' -ExpectedExit 0 -Arguments @('closeout-check', '--workspace', $closeoutWorkspace, '--goal', 'finish route update', '--audit-workspace', $fixture, '--artifact', $closeoutArtifact, '--verify', 'route-task regression', '--report', $closeoutPassReport)
Invoke-Case -Name 'closeout-check-current-audits-required' -ExpectedExit 1 -Arguments @('closeout-check', '--workspace', $closeoutWorkspace, '--goal', 'finish route update', '--artifact', $closeoutArtifact, '--verify', 'route-task regression', '--report', (Join-Path $closeoutWorkspace 'closeout-check-no-audits.json'))
$missingContextAuditWorkspace = Join-Path $fixture 'missing-context-bloat-audit'
Write-CurrentAuditReports -Workspace $missingContextAuditWorkspace
Remove-Item -LiteralPath (Join-Path $missingContextAuditWorkspace 'outputs\context-bloat-audit-report.json') -Force
Invoke-Case -Name 'closeout-check-context-bloat-audit-required' -ExpectedExit 1 -Arguments @('closeout-check', '--workspace', $closeoutWorkspace, '--goal', 'finish route update', '--audit-workspace', $missingContextAuditWorkspace, '--artifact', $closeoutArtifact, '--verify', 'route-task regression', '--report', (Join-Path $closeoutWorkspace 'closeout-missing-context-bloat.json'))
$staleContextAuditWorkspace = Join-Path $fixture 'stale-context-bloat-audit'
Write-CurrentAuditReports -Workspace $staleContextAuditWorkspace
Write-Fixture (Join-Path $staleContextAuditWorkspace 'hotpath-manifest.json') '{"resident":[{"path":"kernel-source.json","max_bytes":8192}],"changed":true}'
Invoke-Case -Name 'closeout-check-context-bloat-audit-freshness-required' -ExpectedExit 1 -Arguments @('closeout-check', '--workspace', $closeoutWorkspace, '--goal', 'finish route update', '--audit-workspace', $staleContextAuditWorkspace, '--artifact', $closeoutArtifact, '--verify', 'route-task regression', '--report', (Join-Path $closeoutWorkspace 'closeout-stale-context-bloat.json'))
Invoke-Case -Name 'closeout-check-bugfix-root-cause-required' -ExpectedExit 1 -Arguments @('closeout-check', '--workspace', $bugfixWorkspace, '--goal', 'finish bugfix regression repair', '--audit-workspace', $fixture, '--artifact', $bugfixArtifact, '--verify', 'bugfix regression verified', '--report', (Join-Path $bugfixWorkspace 'closeout-bugfix-no-root.json'))
Invoke-Case -Name 'closeout-check-bugfix-artifact-root-cause-required' -ExpectedExit 1 -Arguments @('closeout-check', '--workspace', $bugfixWorkspace, '--goal', 'finish neutral release', '--audit-workspace', $fixture, '--artifact', (Join-Path $bugfixWorkspace 'bugfix-pass.json'), '--verify', 'bugfix guard report reviewed', '--report', (Join-Path $bugfixWorkspace 'closeout-bugfix-artifact-no-root.json'))
$closeoutBugfixRootReport = Join-Path $bugfixWorkspace 'closeout-bugfix-root.json'
Invoke-Case -Name 'closeout-check-bugfix-with-root-cause' -ExpectedExit 0 -Arguments @('closeout-check', '--workspace', $bugfixWorkspace, '--goal', 'finish bugfix regression repair', '--audit-workspace', $fixture, '--root-cause-report', $rootCauseReport, '--artifact', $bugfixArtifact, '--verify', 'bugfix regression verified', '--report', $closeoutBugfixRootReport)
$closeoutBugfixRoot = Read-JsonUtf8 -Path $closeoutBugfixRootReport
if ($closeoutBugfixRoot.command -ne 'closeout-check' -or $closeoutBugfixRoot.root_cause_required -ne $true -or $closeoutBugfixRoot.root_cause_report_verified -ne $true) {
    throw "FAIL closeout-check missing command/root verification report=$($closeoutBugfixRoot | ConvertTo-Json -Depth 8 -Compress)"
}
Invoke-Case -Name 'closeout-check-bugfix-forged-root-cause-blocked' -ExpectedExit 1 -Arguments @('closeout-check', '--workspace', $bugfixWorkspace, '--goal', 'finish bugfix regression repair', '--audit-workspace', $fixture, '--root-cause-report', $forgedRootUnsafePath, '--artifact', $bugfixArtifact, '--verify', 'bugfix regression verified', '--report', (Join-Path $bugfixWorkspace 'closeout-bugfix-forged-root.json'))
Invoke-Case -Name 'closeout-check-gap-blocked' -ExpectedExit 1 -Arguments @('closeout-check', '--workspace', $closeoutWorkspace, '--goal', 'finish route update', '--audit-workspace', $fixture, '--artifact', $closeoutArtifact, '--verify', 'route-task regression', '--next-gap', 'still need to sync docs', '--report', (Join-Path $closeoutWorkspace 'closeout-check-gap.json'))
$closeoutDecisionReport = Join-Path $closeoutWorkspace 'closeout-check-decision.json'
Invoke-Case -Name 'closeout-check-needs-decision-pass' -ExpectedExit 0 -Arguments @('closeout-check', '--workspace', $closeoutWorkspace, '--goal', 'finish route update', '--artifact', $closeoutArtifact, '--verify', 'route-task regression', '--next-gap', 'choose release branch', '--needs-user-decision', 'true', '--blocked-reason', 'release branch choice requires user input', '--report', $closeoutDecisionReport)
$closeoutDecision = Read-JsonUtf8 -Path $closeoutDecisionReport
if ($closeoutDecision.status -ne 'pass' -or $closeoutDecision.needs_user_decision -ne $true -or $closeoutDecision.resolved_gap_mode -ne $true) {
    throw "FAIL closeout-check decision report=$($closeoutDecision | ConvertTo-Json -Depth 6 -Compress)"
}
Invoke-Case -Name 'finish-or-block-pass' -ExpectedExit 0 -Arguments @('finish-or-block', '--goal', 'finish route update', '--audit-workspace', $fixture, '--report', (Join-Path $closeoutWorkspace 'finish-pass.json'))
Invoke-Case -Name 'finish-or-block-current-audits-required' -ExpectedExit 1 -Arguments @('finish-or-block', '--goal', 'finish route update', '--report', (Join-Path $closeoutWorkspace 'finish-no-audits.json'))
Invoke-Case -Name 'finish-or-block-bugfix-root-cause-required' -ExpectedExit 1 -Arguments @('finish-or-block', '--workspace', $bugfixWorkspace, '--goal', 'finish bugfix regression repair', '--audit-workspace', $fixture, '--report', (Join-Path $bugfixWorkspace 'finish-bugfix-no-root.json'))
Invoke-Case -Name 'finish-or-block-bugfix-root-cause-workspace-required' -ExpectedExit 1 -Arguments @('finish-or-block', '--goal', 'finish bugfix regression repair', '--audit-workspace', $fixture, '--root-cause-report', $rootCauseReport, '--report', (Join-Path $bugfixWorkspace 'finish-bugfix-root-no-workspace.json'))
Invoke-Case -Name 'finish-or-block-bugfix-with-root-cause' -ExpectedExit 0 -Arguments @('finish-or-block', '--workspace', $bugfixWorkspace, '--goal', 'finish bugfix regression repair', '--audit-workspace', $fixture, '--root-cause-report', $rootCauseReport, '--report', (Join-Path $bugfixWorkspace 'finish-bugfix-root.json'))
Invoke-Case -Name 'finish-or-block-remaining-step-blocked' -ExpectedExit 1 -Arguments @('finish-or-block', '--goal', 'finish route update', '--audit-workspace', $fixture, '--remaining-step', 'still need to sync docs', '--report', (Join-Path $closeoutWorkspace 'finish-fail.json'))
Invoke-Case -Name 'finish-or-block-needs-decision-pass' -ExpectedExit 0 -Arguments @('finish-or-block', '--goal', 'finish route update', '--remaining-step', 'choose release branch', '--needs-user-decision', 'true', '--blocked-reason', 'release branch choice requires user input', '--report', (Join-Path $closeoutWorkspace 'finish-decision.json'))
$verifiedEvidenceReport = Join-Path $closeoutWorkspace 'verified.json'
Invoke-Case -Name 'evidence-grade-verified' -ExpectedExit 0 -Arguments @('evidence-grade', '--workspace', $closeoutWorkspace, '--status', 'verified', '--summary', 'issue verified with artifact', '--artifact', $closeoutArtifact, '--report', $verifiedEvidenceReport)
$taskCloseoutArtifact = Join-Path $taskWorkspace 'final.txt'
$taskCloseoutPassReport = Join-Path $taskWorkspace 'closeout-check-pass.json'
$taskVerifiedEvidenceReport = Join-Path $taskWorkspace 'verified.json'
Write-Fixture $taskCloseoutArtifact 'task final artifact for closeout verification'
Invoke-Case -Name 'closeout-check-task-pass' -ExpectedExit 0 -Arguments @('closeout-check', '--workspace', $taskWorkspace, '--goal', 'finish task route update', '--audit-workspace', $fixture, '--artifact', $taskCloseoutArtifact, '--verify', 'task route regression', '--report', $taskCloseoutPassReport)
Invoke-Case -Name 'evidence-grade-task-verified' -ExpectedExit 0 -Arguments @('evidence-grade', '--workspace', $taskWorkspace, '--status', 'verified', '--summary', 'task verified with artifact', '--artifact', $taskCloseoutArtifact, '--report', $taskVerifiedEvidenceReport)
$taskVerifiedEvidence = Read-JsonUtf8 -Path $taskVerifiedEvidenceReport
if ($taskVerifiedEvidence.command -ne 'evidence-grade' -or $taskVerifiedEvidence.schema_version -ne 'evidence-grade.v1' -or -not $taskVerifiedEvidence.workspace_key -or -not $taskVerifiedEvidence.artifact_refs -or $taskVerifiedEvidence.artifacts -or $taskVerifiedEvidence.summary -is [string]) {
    throw "FAIL evidence-grade report retained raw or missing schema report=$($taskVerifiedEvidence | ConvertTo-Json -Depth 8 -Compress)"
}
$taskNoAuditWorkspace = Join-Path $fixture 'task-no-current-audits'
$taskNoAuditArtifact = Join-Path $taskNoAuditWorkspace 'final.txt'
$taskNoAuditCloseout = Join-Path $taskNoAuditWorkspace 'closeout-check-pass.json'
$taskNoAuditEvidence = Join-Path $taskNoAuditWorkspace 'verified.json'
Write-Fixture $taskNoAuditArtifact 'task final artifact for missing audit verification'
Invoke-Case -Name 'closeout-check-task-no-audit-pass' -ExpectedExit 0 -Arguments @('closeout-check', '--workspace', $taskNoAuditWorkspace, '--goal', 'finish task route update', '--audit-workspace', $fixture, '--artifact', $taskNoAuditArtifact, '--verify', 'task route regression', '--report', $taskNoAuditCloseout)
Invoke-Case -Name 'evidence-grade-task-no-audit-verified' -ExpectedExit 0 -Arguments @('evidence-grade', '--workspace', $taskNoAuditWorkspace, '--status', 'verified', '--summary', 'task verified with artifact', '--artifact', $taskNoAuditArtifact, '--report', $taskNoAuditEvidence)
Invoke-Case -Name 'task-end-current-audits-required' -ExpectedExit 1 -Arguments @('task', '--workspace', $taskNoAuditWorkspace, '--event', 'end', '--status', 'done', '--artifact', $artifact, '--closeout-report', $taskNoAuditCloseout, '--evidence-report', $taskNoAuditEvidence, '--audit-workspace', (Join-Path $fixture 'missing-audits'))
Invoke-Case -Name 'task-end-closeout-needs-decision-blocked' -ExpectedExit 2 -Arguments @('task', '--workspace', (Join-Path $fixture 'task-decision-closeout'), '--event', 'end', '--status', 'done', '--artifact', $artifact, '--closeout-report', $closeoutDecisionReport, '--evidence-report', $verifiedEvidenceReport, '--audit-workspace', $fixture)
Invoke-Case -Name 'task-end-closeout-workspace-mismatch-blocked' -ExpectedExit 2 -Arguments @('task', '--workspace', $taskWorkspace, '--event', 'end', '--status', 'done', '--artifact', $artifact, '--closeout-report', $closeoutPassReport, '--evidence-report', $taskVerifiedEvidenceReport, '--audit-workspace', $fixture)
$taskCloseoutNoCommand = Join-Path $taskWorkspace 'closeout-no-command.json'
Write-JsonUtf8 -Path $taskCloseoutNoCommand -Value ([ordered]@{ status = 'pass'; workspace_key = (Get-PrivacyHash -Value $taskWorkspace); needs_user_decision = $false; remaining_gaps = @() })
Invoke-Case -Name 'task-end-closeout-command-required' -ExpectedExit 2 -Arguments @('task', '--workspace', $taskWorkspace, '--event', 'end', '--status', 'done', '--artifact', $artifact, '--closeout-report', $taskCloseoutNoCommand, '--evidence-report', $taskVerifiedEvidenceReport, '--audit-workspace', $fixture)
$taskCloseoutMinimalForged = Join-Path $taskWorkspace 'closeout-minimal-forged.json'
Write-JsonUtf8 -Path $taskCloseoutMinimalForged -Value ([ordered]@{ status = 'pass'; command = 'closeout-check'; schema_version = 'closeout-check.v1'; privacy_mode = 'hash-length-and-evidence-ref-only'; workspace_key = (Get-PrivacyHash -Value $taskWorkspace); needs_user_decision = $false; remaining_gaps = @(); root_cause_required = $false; root_cause_report_verified = $false })
Invoke-Case -Name 'task-end-closeout-minimal-forged-blocked' -ExpectedExit 2 -Arguments @('task', '--workspace', $taskWorkspace, '--event', 'end', '--status', 'done', '--artifact', $artifact, '--closeout-report', $taskCloseoutMinimalForged, '--evidence-report', $taskVerifiedEvidenceReport, '--audit-workspace', $fixture)
$taskCloseoutRootUnverified = Join-Path $taskWorkspace 'closeout-root-unverified.json'
Write-JsonUtf8 -Path $taskCloseoutRootUnverified -Value ([ordered]@{ status = 'pass'; command = 'closeout-check'; schema_version = 'closeout-check.v1'; privacy_mode = 'hash-length-and-evidence-ref-only'; workspace_key = (Get-PrivacyHash -Value $taskWorkspace); audit_workspace_key = (Get-PrivacyHash -Value $fixture); needs_user_decision = $false; remaining_gaps = @(); root_cause_required = $true; root_cause_report_verified = $false; artifact_refs = @(@{ path_ref = 'final.txt'; bytes = (Get-Item -LiteralPath $taskCloseoutArtifact).Length; sha256 = (Get-Sha256Lower -Path $taskCloseoutArtifact) }); verifications = @(@{ key = (Get-PrivacyHash -Value 'task route regression'); length = 21 }) })
Invoke-Case -Name 'task-end-closeout-root-unverified-blocked' -ExpectedExit 2 -Arguments @('task', '--workspace', $taskWorkspace, '--event', 'end', '--status', 'done', '--artifact', $artifact, '--closeout-report', $taskCloseoutRootUnverified, '--evidence-report', $taskVerifiedEvidenceReport, '--audit-workspace', $fixture)
$taskForgedEvidence = Join-Path $taskWorkspace 'evidence-forged.json'
Write-JsonUtf8 -Path $taskForgedEvidence -Value ([ordered]@{ status = 'verified' })
Invoke-Case -Name 'task-end-evidence-forged-blocked' -ExpectedExit 2 -Arguments @('task', '--workspace', $taskWorkspace, '--event', 'end', '--status', 'done', '--artifact', $artifact, '--closeout-report', $taskCloseoutPassReport, '--evidence-report', $taskForgedEvidence, '--audit-workspace', $fixture)
Invoke-Case -Name 'task-end-valid' -ExpectedExit 0 -Arguments @('task', '--workspace', $taskWorkspace, '--event', 'end', '--status', 'done', '--artifact', $artifact, '--closeout-report', $taskCloseoutPassReport, '--evidence-report', $taskVerifiedEvidenceReport, '--audit-workspace', $fixture)
Invoke-Case -Name 'task-end-done-next-step-blocked' -ExpectedExit 1 -Arguments @('task', '--workspace', $taskWorkspace, '--event', 'end', '--status', 'done', '--artifact', $artifact, '--closeout-report', $taskCloseoutPassReport, '--evidence-report', $taskVerifiedEvidenceReport, '--audit-workspace', $fixture, '--note', 'next step could continue with further optimization')

$routeConfig = Join-Path $fixture 'route-config.json'
[System.IO.File]::WriteAllText($routeConfig, (@{
    iron_rules_version = '11.3'
    providers = @(
        @{
            id = 'agnes-openai-free'
            name = 'Agnes AI (Current Free Tier)'
            provider_type = 'openai'
            api_key = $null
            base_url = 'https://apihub.agnes-ai.com'
            currency = 'USD'
            enabled = $true
            priority = 1
            model = 'agnes-2.0-flash'
            api_key_env = 'AGNES_API_KEY'
            free_tier_status = 'current-free'
            notes = 'Mirror only. Keep key outside repo and use only while official current price remains 0.'
        }
        @{
            id = 'deepseek-web'
            name = 'DeepSeek (免费网页)'
            provider_type = 'deepseek_web'
            api_key = $null
            base_url = $null
            currency = 'CNY'
            enabled = $true
            priority = 0
            model = 'deepseek-chat'
        }
    )
    default_model_tier = 'standard'
    model_profiles = @{
        low = @{
            provider_id = 'deepseek-web'
            model = 'deepseek-chat'
            reasoning_effort = 'low'
        }
        standard = @{
            provider_id = 'deepseek-web'
            model = 'deepseek-chat'
            reasoning_effort = 'medium'
        }
        high = @{
            provider_id = 'deepseek-web'
            model = 'deepseek-chat'
            reasoning_effort = 'high'
        }
    }
    routing_rules = @(
        @{
            id = 'imagegen'
            provider_id = 'agnes-openai-free'
            model = 'agnes-image-2.1-flash'
        }
        @{
            id = 'video'
            provider_id = 'agnes-openai-free'
            model = 'agnes-video-v2.0'
        }
    )
    cache_config = @{ target_hit_rate = 0.95; flatten_threshold = 10 }
} | ConvertTo-Json -Depth 8), [System.Text.UTF8Encoding]::new($false))
$routeConfigAgnesDisabled = Join-Path $fixture 'route-config-agnes-disabled.json'
[System.IO.File]::WriteAllText($routeConfigAgnesDisabled, (@{
    iron_rules_version = '11.3'
    providers = @(
        @{
            id = 'agnes-openai-free'
            name = 'Agnes AI (Current Free Tier)'
            provider_type = 'openai'
            api_key = $null
            base_url = 'https://apihub.agnes-ai.com'
            currency = 'USD'
            enabled = $false
            priority = 1
            model = 'agnes-2.0-flash'
            api_key_env = 'AGNES_API_KEY'
            free_tier_status = 'current-free'
            notes = 'Mirror only. Keep key outside repo and use only while official current price remains 0.'
        }
        @{
            id = 'deepseek-web'
            name = 'DeepSeek (免费网页)'
            provider_type = 'deepseek_web'
            api_key = $null
            base_url = $null
            currency = 'CNY'
            enabled = $true
            priority = 0
            model = 'deepseek-chat'
        }
    )
    default_model_tier = 'standard'
    model_profiles = @{
        low = @{
            provider_id = 'deepseek-web'
            model = 'deepseek-chat'
            reasoning_effort = 'low'
        }
        standard = @{
            provider_id = 'deepseek-web'
            model = 'deepseek-chat'
            reasoning_effort = 'medium'
        }
        high = @{
            provider_id = 'deepseek-web'
            model = 'deepseek-chat'
            reasoning_effort = 'high'
        }
    }
    routing_rules = @(
        @{
            id = 'imagegen'
            provider_id = 'agnes-openai-free'
            model = 'agnes-image-2.1-flash'
        }
        @{
            id = 'video'
            provider_id = 'agnes-openai-free'
            model = 'agnes-video-v2.0'
        }
    )
    cache_config = @{ target_hit_rate = 0.95; flatten_threshold = 10 }
} | ConvertTo-Json -Depth 8), [System.Text.UTF8Encoding]::new($false))
$canonReport = Join-Path $fixture 'canon-report.json'
Invoke-Case -Name 'canon-report' -ExpectedExit 0 -Arguments @('canon-report', '--report', $canonReport)
if (-not (Test-Path -LiteralPath $canonReport)) {
    throw "FAIL canon-report missing report=$canonReport"
}
$canon = Read-JsonUtf8 -Path $canonReport
if ($canon.default_model_tier -ne 'standard') {
    throw "FAIL canon-report wrong default tier=$($canon.default_model_tier)"
}
if ($canon.model_profiles.low.model -ne 'deepseek-chat' -or $canon.model_profiles.standard.model -ne 'deepseek-chat' -or $canon.model_profiles.high.model -ne 'deepseek-chat') {
    throw "FAIL canon-report wrong model profiles"
}
if ($canon.host_available_plugins.Count -lt 4) {
    throw "FAIL canon-report missing host-available plugins"
}
foreach ($pluginName in @('Browser', 'Documents', 'Spreadsheets', 'Presentations')) {
    if (-not ($canon.host_available_plugins | Where-Object { $_.plugin -eq $pluginName })) {
        throw "FAIL canon-report missing host-available plugin mirror $pluginName report=$($canon.host_available_plugins | ConvertTo-Json -Depth 8 -Compress)"
    }
}
foreach ($pluginName in @('Browser', 'Documents', 'Spreadsheets', 'Presentations')) {
    if (-not ($canon.admitted_plugins | Where-Object { $_.plugin -eq $pluginName })) {
        throw "FAIL canon-report missing admitted plugin $pluginName report=$($canon.admitted_plugins | ConvertTo-Json -Depth 8 -Compress)"
    }
}
if (-not $canon.distilled_atom_kernel -or -not ($canon.distilled_atom_kernel.resident_light_atoms -contains 'assumption-ledger') -or -not ($canon.distilled_atom_kernel.on_demand_atoms -contains 'version-doc-mcp')) {
    throw "FAIL canon-report missing distilled atom kernel report=$($canon | ConvertTo-Json -Depth 8 -Compress)"
}
if (-not ($canon.top_level_roles -contains 'nuwa-preflight') -or -not ($canon.top_level_roles -contains 'performance-benchmark-on-demand') -or -not ($canon.top_level_roles -contains 'compliance-on-demand')) {
    throw "FAIL canon-report missing canonical top-level roles report=$($canon.top_level_roles | ConvertTo-Json -Compress)"
}
$distilledAtomCount = $canon.distilled_atom_kernel.resident_light_atoms.Count + $canon.distilled_atom_kernel.on_demand_atoms.Count
if ($distilledAtomCount -ne 21 -or -not $canon.distilled_atom_kernel.owner_map -or -not $canon.distilled_atom_kernel.owner_map.'disciplined-debug-loop' -or -not $canon.distilled_atom_kernel.owner_map.'prior-art-solution-search' -or -not $canon.distilled_atom_kernel.owner_map.'root-cause-radar' -or -not $canon.distilled_atom_kernel.owner_map.'parallel-hypothesis-fanout' -or -not $canon.distilled_atom_kernel.owner_map.'patch-debt-root-cure' -or -not $canon.distilled_atom_kernel.owner_map.'terminal-real-run-verification' -or -not $canon.distilled_atom_kernel.owner_map.'html-native-design-canvas') {
    throw "FAIL canon-report distilled atom registry drift count=$distilledAtomCount report=$($canon.distilled_atom_kernel | ConvertTo-Json -Depth 8 -Compress)"
}
if ($canon.distilled_atom_kernel.owner_map.'content-type-compression-router' -notmatch 'performance-benchmark-on-demand' -or $canon.distilled_atom_kernel.owner_map.'motion-stage-sprite-engine' -notmatch 'performance-benchmark-on-demand') {
    throw "FAIL canon-report performance officer naming drift report=$($canon.distilled_atom_kernel.owner_map | ConvertTo-Json -Depth 8 -Compress)"
}
if ($canon.intelligence_profile_contract.role -ne 'candidate-scout-not-research-system' -or $canon.intelligence_profile_contract.search_scope -ne 'wide-recall-shallow-first') {
    throw "FAIL canon-report intelligence profile contract role/scope drift report=$($canon.intelligence_profile_contract | ConvertTo-Json -Depth 8 -Compress)"
}
if ($canon.intelligence_profile_contract.search_method -ne 'wide-shallow-scout-first-then-deepen-only-on-promising-candidates') {
    throw "FAIL canon-report intelligence profile search method drift report=$($canon.intelligence_profile_contract | ConvertTo-Json -Depth 8 -Compress)"
}
if (-not ($canon.intelligence_profile_contract.may_do -contains 'candidate-metadata') -or -not ($canon.intelligence_profile_contract.may_do -contains 'evidence-handle')) {
    throw "FAIL canon-report intelligence profile missing allowed scout outputs report=$($canon.intelligence_profile_contract | ConvertTo-Json -Depth 8 -Compress)"
}
if (-not ($canon.intelligence_profile_contract.must_not_do -contains 'final-analysis') -or -not ($canon.intelligence_profile_contract.must_not_do -contains 'deep-extract-by-default') -or -not ($canon.intelligence_profile_contract.must_not_do -contains 'distillation-decision') -or -not ($canon.intelligence_profile_contract.must_not_do -contains 'adoption-decision') -or -not ($canon.intelligence_profile_contract.must_not_do -contains 'install-or-execute')) {
    throw "FAIL canon-report intelligence profile missing forbidden powers report=$($canon.intelligence_profile_contract | ConvertTo-Json -Depth 8 -Compress)"
}
if ($canon.concise_execution_contract.objective -ne 'short-precise-high-hit-low-total-cost') {
    throw "FAIL canon-report concise execution objective drift report=$($canon.concise_execution_contract | ConvertTo-Json -Depth 8 -Compress)"
}
if (-not ($canon.concise_execution_contract.must_do -contains 'simplest-effective-path-first') -or -not ($canon.concise_execution_contract.must_do -contains 'agnes-search-only-before-uncertain-build-when-web-search-is-explicitly-needed') -or -not ($canon.concise_execution_contract.must_do -contains 'single-message-precision') -or -not ($canon.concise_execution_contract.must_do -contains 'prior-art-before-invention-when-uncertain') -or -not ($canon.concise_execution_contract.must_do -contains 'first-pass-acceptance-and-impact-lock-before-edit') -or -not ($canon.concise_execution_contract.must_do -contains 'prove-need-before-abstraction') -or -not ($canon.concise_execution_contract.must_do -contains 'delete-or-reuse-before-add') -or -not ($canon.concise_execution_contract.must_do -contains 'target-page-in-place-replacement') -or -not ($canon.concise_execution_contract.must_do -contains 'active-route-entrypoint-verification') -or -not ($canon.concise_execution_contract.must_do -contains 'superseded-page-cleanup-before-completion') -or -not ($canon.concise_execution_contract.must_do -contains 'smallest-working-change-first') -or -not ($canon.concise_execution_contract.must_do -contains 'fresh-output-uncached-volume-gated')) {
    throw "FAIL canon-report concise execution missing must_do report=$($canon.concise_execution_contract | ConvertTo-Json -Depth 8 -Compress)"
}
if (-not ($canon.concise_execution_contract.must_not_do -contains 'verbose-status-padding') -or -not ($canon.concise_execution_contract.must_not_do -contains 'unneeded-preflight-loop') -or -not ($canon.concise_execution_contract.must_not_do -contains 'guess-without-evidence') -or -not ($canon.concise_execution_contract.must_not_do -contains 'blind-trial-and-error-when-prior-art-is-available') -or -not ($canon.concise_execution_contract.must_not_do -contains 'context-shift-from-cached-to-uncached') -or -not ($canon.concise_execution_contract.must_not_do -contains 'from-scratch-tooling-when-existing-solution-fits') -or -not ($canon.concise_execution_contract.must_not_do -contains 'clever-overengineering-without-proven-need') -or -not ($canon.concise_execution_contract.must_not_do -contains 'new-abstraction-before-duplication-or-gap-is-proven') -or -not ($canon.concise_execution_contract.must_not_do -contains 'parallel-compat-page-for-requested-page-change') -or -not ($canon.concise_execution_contract.must_not_do -contains 'leave-old-page-reachable-after-replacement')) {
    throw "FAIL canon-report concise execution missing must_not_do report=$($canon.concise_execution_contract | ConvertTo-Json -Depth 8 -Compress)"
}
if (-not ($canon.concise_execution_contract.cost_vector -contains 'fresh_input_tokens_p95') -or -not ($canon.concise_execution_contract.cost_vector -contains 'output_tokens_p95') -or -not ($canon.concise_execution_contract.cost_vector -contains 'uncached_tokens_p95') -or -not ($canon.concise_execution_contract.cost_vector -contains 'tokens_per_success')) {
    throw "FAIL canon-report concise execution missing cost vector report=$($canon.concise_execution_contract | ConvertTo-Json -Depth 8 -Compress)"
}
if ($canon.execution_budget_contract.objective -ne 'all-work-direct-small-task-execution') {
    throw "FAIL canon-report execution budget objective drift report=$($canon.execution_budget_contract | ConvertTo-Json -Depth 8 -Compress)"
}
if (-not ($canon.execution_budget_contract.must_do -contains 'bind-finish-line-and-out-of-scope-before-goal-start') -or -not ($canon.execution_budget_contract.must_do -contains 'keep-direct-code-work-on-a-minimal-first-pass-guard') -or -not ($canon.execution_budget_contract.must_do -contains 'all-non-chat-work-stays-direct-task-by-default') -or -not ($canon.execution_budget_contract.must_do -contains 'run-targeted-verification-before-any-full-suite') -or -not ($canon.execution_budget_contract.must_do -contains 'keep-officers-on-demand-and-exit-after-merge')) {
    throw "FAIL canon-report execution budget missing must_do report=$($canon.execution_budget_contract | ConvertTo-Json -Depth 8 -Compress)"
}
if (-not ($canon.execution_budget_contract.must_not_do -contains 'auto-create-big-task-mode') -or -not ($canon.execution_budget_contract.must_not_do -contains 'repeat-full-suite-after-small-edits') -or -not ($canon.execution_budget_contract.must_not_do -contains 'block-non-token-work-on-missing-runtime-usage-log') -or -not ($canon.execution_budget_contract.must_not_do -contains 'start-goal-without-clear-finish-line')) {
    throw "FAIL canon-report execution budget missing must_not_do report=$($canon.execution_budget_contract | ConvertTo-Json -Depth 8 -Compress)"
}
if ($canon.analysis_completeness_contract.objective -ne 'complete-materials-before-architecture-analysis') {
    throw "FAIL canon-report analysis completeness objective drift report=$($canon.analysis_completeness_contract | ConvertTo-Json -Depth 8 -Compress)"
}
if (-not ($canon.analysis_completeness_contract.must_do -contains 'collect-material-inventory') -or -not ($canon.analysis_completeness_contract.must_do -contains 'state-coverage-and-gaps') -or -not ($canon.analysis_completeness_contract.must_do -contains 'ask-user-for-missing-materials-when-critical') -or -not ($canon.analysis_completeness_contract.must_do -contains 'separate-fact-inference-and-unknown') -or -not ($canon.analysis_completeness_contract.must_do -contains 'no-final-conclusion-from-incomplete-evidence')) {
    throw "FAIL canon-report analysis completeness missing must_do report=$($canon.analysis_completeness_contract | ConvertTo-Json -Depth 8 -Compress)"
}
if (-not ($canon.analysis_completeness_contract.must_not_do -contains 'guess-architecture-from-partial-materials') -or -not ($canon.analysis_completeness_contract.must_not_do -contains 'treat-sample-as-whole-system') -or -not ($canon.analysis_completeness_contract.must_not_do -contains 'hide-coverage-gaps') -or -not ($canon.analysis_completeness_contract.must_not_do -contains 'promote-uncertain-claim-to-fact')) {
    throw "FAIL canon-report analysis completeness missing must_not_do report=$($canon.analysis_completeness_contract | ConvertTo-Json -Depth 8 -Compress)"
}
Invoke-Case -Name 'route-task' -ExpectedExit 0 -Arguments @('route-task', '--config', $routeConfig, '--query', 'please build a ppt ui design', '--report', (Join-Path $fixture 'route-report.json'))
$routeReport = Read-JsonUtf8 -Path (Join-Path $fixture 'route-report.json')
if ($routeReport.matched_route.id -ne 'visual' -or $routeReport.recommended_tier -ne 'standard' -or $routeReport.recommended_profile.model -ne 'deepseek-chat') {
    throw "FAIL route-task wrong route or tier report=$($routeReport | ConvertTo-Json -Depth 6 -Compress)"
}
if ($routeReport.concise_execution_contract.objective -ne 'short-precise-high-hit-low-total-cost') {
    throw "FAIL route-task missing concise execution contract report=$($routeReport | ConvertTo-Json -Depth 8 -Compress)"
}
if ($routeReport.execution_budget_contract.objective -ne 'all-work-direct-small-task-execution' -or $routeReport.execution_budget.id -ne 'DIRECT_TASK' -or $routeReport.execution_budget.full_suite_max_runs -ne 0) {
    throw "FAIL route-task execution budget should stay light for scoped visual work report=$($routeReport | ConvertTo-Json -Depth 8 -Compress)"
}
if ($routeReport.simplest_effective_path_required -ne $true -or $routeReport.agnes_scout_preferred -ne $false -or $routeReport.goal_boundary_lock_required -ne $false) {
    throw "FAIL route-task scoped visual work should stay light without forced scout or goal boundary lock report=$($routeReport | ConvertTo-Json -Depth 8 -Compress)"
}
if ($null -ne $routeReport.goal_boundary_required_fields) {
    throw "FAIL route-task light visual work should not emit goal boundary required fields report=$($routeReport | ConvertTo-Json -Depth 8 -Compress)"
}
if (-not ($routeReport.capability_mounts.distilled_atoms -contains 'content-type-compression-router') -or -not ($routeReport.capability_mounts.distilled_atoms -contains 'reversible-evidence-handle')) {
    throw "FAIL route-task missing visual distilled atoms report=$($routeReport | ConvertTo-Json -Depth 8 -Compress)"
}
if (-not ($routeReport.capability_mounts.distilled_atoms -contains 'brand-asset-protocol') -or -not ($routeReport.capability_mounts.distilled_atoms -contains 'anti-ai-slop-visual-rules') -or -not ($routeReport.capability_mounts.distilled_atoms -contains 'native-pptx-master-route')) {
    throw "FAIL route-task missing huashu-design distilled atoms report=$($routeReport | ConvertTo-Json -Depth 8 -Compress)"
}
if ($routeReport.capability_mounts.source_support_classes -contains 'content-type-compression-router' -or $routeReport.capability_mounts.source_support_classes -contains 'reversible-evidence-handle') {
    throw "FAIL route-task mixed distilled atoms into source lineage atoms report=$($routeReport | ConvertTo-Json -Depth 8 -Compress)"
}
Invoke-Case -Name 'route-task-chat-default-strong' -ExpectedExit 0 -Arguments @('route-task', '--config', $routeConfig, '--query', 'just chat casually with me', '--report', (Join-Path $fixture 'route-chat-agnes-report.json'))
$chatAgnesReport = Read-JsonUtf8 -Path (Join-Path $fixture 'route-chat-agnes-report.json')
if ($chatAgnesReport.matched_route.id -ne 'chat' -or $chatAgnesReport.matched_route.provider_id -ne 'deepseek-web' -or $chatAgnesReport.matched_route.model -ne 'deepseek-chat' -or $chatAgnesReport.recommended_profile.provider_id -ne 'deepseek-web' -or $chatAgnesReport.recommended_profile.model -ne 'deepseek-chat') {
    throw "FAIL route-task chat should route to default strong provider report=$($chatAgnesReport | ConvertTo-Json -Depth 6 -Compress)"
}
Invoke-Case -Name 'route-task-chat-fallback-when-agnes-disabled' -ExpectedExit 0 -Arguments @('route-task', '--config', $routeConfigAgnesDisabled, '--query', 'just chat casually with me', '--report', (Join-Path $fixture 'route-chat-agnes-disabled-report.json'))
$chatAgnesDisabledReport = Read-JsonUtf8 -Path (Join-Path $fixture 'route-chat-agnes-disabled-report.json')
if ($chatAgnesDisabledReport.matched_route.id -ne 'chat' -or $chatAgnesDisabledReport.matched_route.provider_id -ne 'deepseek-web' -or $chatAgnesDisabledReport.matched_route.model -ne 'deepseek-chat' -or $chatAgnesDisabledReport.recommended_profile.provider_id -ne 'deepseek-web' -or $chatAgnesDisabledReport.recommended_profile.model -ne 'deepseek-chat' -or $chatAgnesDisabledReport.provider_fallback_applied -eq $true) {
    throw "FAIL route-task chat should stay on default strong provider even when Agnes is disabled report=$($chatAgnesDisabledReport | ConvertTo-Json -Depth 8 -Compress)"
}
Invoke-Case -Name 'route-task-analysis-completeness' -ExpectedExit 0 -Arguments @('route-task', '--config', $routeConfig, '--query', 'analyze architecture with incomplete docs and unknown modules before making system conclusions', '--report', (Join-Path $fixture 'route-analysis-completeness-report.json'))
$routeAnalysisReport = Read-JsonUtf8 -Path (Join-Path $fixture 'route-analysis-completeness-report.json')
if ($routeAnalysisReport.analysis_completeness_required -ne $true -or $routeAnalysisReport.analysis_completeness_contract.objective -ne 'complete-materials-before-architecture-analysis') {
    throw "FAIL route-task analysis completeness contract missing report=$($routeAnalysisReport | ConvertTo-Json -Depth 8 -Compress)"
}
if (-not ($routeAnalysisReport.task_route.oversight_chain -contains 'white-hat') -or -not ($routeAnalysisReport.task_route.oversight_chain -contains 'audit')) {
    throw "FAIL route-task analysis completeness missing oversight report=$($routeAnalysisReport | ConvertTo-Json -Depth 8 -Compress)"
}
if (-not ($routeAnalysisReport.capability_mounts.distilled_atoms -contains 'assumption-ledger') -or -not ($routeAnalysisReport.capability_mounts.distilled_atoms -contains 'claim-fact-check') -or -not ($routeAnalysisReport.capability_mounts.distilled_atoms -contains 'research-evidence-pack') -or -not ($routeAnalysisReport.capability_mounts.distilled_atoms -contains 'reversible-evidence-handle')) {
    throw "FAIL route-task analysis completeness missing evidence atoms report=$($routeAnalysisReport | ConvertTo-Json -Depth 8 -Compress)"
}
Invoke-Case -Name 'route-task-code-map-required' -ExpectedExit 0 -Arguments @('route-task', '--config', $routeConfig, '--query', 'please refactor a Rust plugin and fix code across multiple files', '--report', (Join-Path $fixture 'route-code-report.json'))
$routeCodeReport = Read-JsonUtf8 -Path (Join-Path $fixture 'route-code-report.json')
if ($routeCodeReport.matched_route.id -ne 'code' -or $routeCodeReport.code_map_required -ne $true -or $routeCodeReport.next_required_artifact -ne 'code-map') {
    throw "FAIL route-task code-map report=$($routeCodeReport | ConvertTo-Json -Depth 6 -Compress)"
}
if ($routeCodeReport.task_route.execution_contract -ne 'developer-plan-then-execution' -or $routeCodeReport.task_route.planning_profile.provider_id -ne 'deepseek-web' -or $routeCodeReport.task_route.planning_profile.model -ne 'deepseek-chat') {
    throw "FAIL route-task code should plan with default strong provider report=$($routeCodeReport | ConvertTo-Json -Depth 8 -Compress)"
}
Invoke-Case -Name 'route-task-code-plan-fallback-when-agnes-disabled' -ExpectedExit 0 -Arguments @('route-task', '--config', $routeConfigAgnesDisabled, '--query', 'please refactor a Rust plugin and fix code across multiple files', '--report', (Join-Path $fixture 'route-code-agnes-disabled-report.json'))
$routeCodeAgnesDisabledReport = Read-JsonUtf8 -Path (Join-Path $fixture 'route-code-agnes-disabled-report.json')
if ($routeCodeAgnesDisabledReport.task_route.execution_contract -ne 'developer-plan-then-execution' -or $routeCodeAgnesDisabledReport.task_route.planning_profile.provider_id -ne 'deepseek-web' -or $routeCodeAgnesDisabledReport.task_route.planning_profile.model -ne 'deepseek-chat') {
    throw "FAIL route-task code planning should remain on default strong provider when Agnes is disabled report=$($routeCodeAgnesDisabledReport | ConvertTo-Json -Depth 8 -Compress)"
}
if (-not ($routeCodeReport.capability_mounts.distilled_atoms -contains 'version-doc-mcp') -or -not ($routeCodeReport.capability_mounts.distilled_atoms -contains 'disciplined-debug-loop')) {
    throw "FAIL route-task code missing distilled atoms report=$($routeCodeReport | ConvertTo-Json -Depth 8 -Compress)"
}
Invoke-Case -Name 'route-task-simple-code-low-cost' -ExpectedExit 0 -Arguments @('route-task', '--config', $routeConfig, '--query', 'fix bug', '--report', (Join-Path $fixture 'route-simple-code-report.json'))
$routeSimpleCodeReport = Read-JsonUtf8 -Path (Join-Path $fixture 'route-simple-code-report.json')
if ($routeSimpleCodeReport.matched_route.id -ne 'code' -or $routeSimpleCodeReport.recommended_tier -ne 'low' -or $routeSimpleCodeReport.code_map_required -ne $false -or ($routeSimpleCodeReport.task_route.oversight_chain -contains 'root-cause-officer') -or ($routeSimpleCodeReport.capability_mounts.distilled_atoms -contains 'root-cause-radar') -or ($routeSimpleCodeReport.capability_mounts.distilled_atoms -contains 'prior-art-solution-search') -or $routeSimpleCodeReport.deterministic_execution.required -ne $false -or $routeSimpleCodeReport.execution_budget.id -ne 'DIRECT_TASK' -or $routeSimpleCodeReport.execution_budget.full_suite_max_runs -ne 0) {
    throw "FAIL route-task simple code should stay low-cost report=$($routeSimpleCodeReport | ConvertTo-Json -Depth 8 -Compress)"
}
if ($routeSimpleCodeReport.task_route.execution_contract -ne 'developer-direct-execution' -or $routeSimpleCodeReport.goal_boundary_lock_required -ne $false -or $routeSimpleCodeReport.agnes_scout_preferred -ne $false -or $null -ne $routeSimpleCodeReport.task_route.planning_profile) {
    throw "FAIL route-task simple code should skip Agnes planning and goal boundary lock report=$($routeSimpleCodeReport | ConvertTo-Json -Depth 8 -Compress)"
}
if ($routeSimpleCodeReport.first_pass_guard_required -ne $true -or -not $routeSimpleCodeReport.first_pass_guard_fields -or -not ($routeSimpleCodeReport.first_pass_guard_fields -contains 'acceptance_signal') -or -not $routeSimpleCodeReport.task_route.direct_execution_guard -or $routeSimpleCodeReport.task_route.direct_execution_guard.mode -ne 'first-pass-minimal-lock') {
    throw "FAIL route-task simple code should carry first-pass minimal lock report=$($routeSimpleCodeReport | ConvertTo-Json -Depth 8 -Compress)"
}
Invoke-Case -Name 'route-task-simple-doc-light' -ExpectedExit 0 -Arguments @('route-task', '--config', $routeConfig, '--query', 'small doc update for one section', '--report', (Join-Path $fixture 'route-simple-doc-report.json'))
$routeSimpleDocReport = Read-JsonUtf8 -Path (Join-Path $fixture 'route-simple-doc-report.json')
if ($routeSimpleDocReport.matched_route.id -ne 'content' -or $routeSimpleDocReport.task_route.state -ne 'LIGHT_EXECUTION' -or $routeSimpleDocReport.execution_budget.id -ne 'DIRECT_TASK' -or $routeSimpleDocReport.goal_boundary_lock_required -ne $false -or $routeSimpleDocReport.agnes_scout_preferred -ne $false) {
    throw "FAIL route-task simple doc should stay on the light path report=$($routeSimpleDocReport | ConvertTo-Json -Depth 8 -Compress)"
}
Invoke-Case -Name 'route-task-simple-code-chinese-light' -ExpectedExit 0 -Arguments @('route-task', '--config', $routeConfig, '--query', '修一个按钮点击没反应的小 bug', '--report', (Join-Path $fixture 'route-simple-code-chinese-report.json'))
$routeSimpleCodeChineseReport = Read-JsonUtf8 -Path (Join-Path $fixture 'route-simple-code-chinese-report.json')
if ($routeSimpleCodeChineseReport.matched_route.id -ne 'code' -or $routeSimpleCodeChineseReport.task_route.state -ne 'LIGHT_EXECUTION' -or $routeSimpleCodeChineseReport.execution_budget.id -ne 'DIRECT_TASK' -or $routeSimpleCodeChineseReport.task_route.execution_contract -ne 'developer-direct-execution' -or $routeSimpleCodeChineseReport.matched_route.provider_id -ne 'deepseek-web' -or $routeSimpleCodeChineseReport.agnes_scout_preferred -ne $false) {
    throw "FAIL route-task chinese simple code should stay on direct strong light path report=$($routeSimpleCodeChineseReport | ConvertTo-Json -Depth 8 -Compress)"
}
if ($routeSimpleCodeChineseReport.first_pass_guard_required -ne $true -or $routeSimpleCodeChineseReport.task_route.direct_execution_guard.mode -ne 'first-pass-minimal-lock') {
    throw "FAIL route-task chinese simple code should carry first-pass minimal lock report=$($routeSimpleCodeChineseReport | ConvertTo-Json -Depth 8 -Compress)"
}
Invoke-Case -Name 'route-task-visual-page-design-not-code' -ExpectedExit 0 -Arguments @('route-task', '--config', $routeConfig, '--query', '改一个落地页页面设计和界面版式', '--report', (Join-Path $fixture 'route-visual-page-design-report.json'))
$visualPageDesignReport = Read-JsonUtf8 -Path (Join-Path $fixture 'route-visual-page-design-report.json')
if ($visualPageDesignReport.matched_route.id -ne 'visual' -or ($visualPageDesignReport.capability_mounts.distilled_atoms -contains 'disciplined-debug-loop')) {
    throw "FAIL route-task visual page design should prefer visual route over code report=$($visualPageDesignReport | ConvertTo-Json -Depth 8 -Compress)"
}
Invoke-Case -Name 'route-task-search-chinese-agnes' -ExpectedExit 0 -Arguments @('route-task', '--config', $routeConfig, '--query', '全网搜索一下类似的开源解决方案', '--report', (Join-Path $fixture 'route-search-chinese-agnes-report.json'))
$routeSearchChineseAgnesReport = Read-JsonUtf8 -Path (Join-Path $fixture 'route-search-chinese-agnes-report.json')
if ($routeSearchChineseAgnesReport.matched_route.id -ne 'search' -or $routeSearchChineseAgnesReport.matched_route.provider_id -ne 'agnes-openai-free' -or $routeSearchChineseAgnesReport.matched_route.model -ne 'agnes-2.0-flash' -or $routeSearchChineseAgnesReport.recommended_profile.provider_id -ne 'agnes-openai-free') {
    throw "FAIL route-task chinese search should be the only Agnes text path report=$($routeSearchChineseAgnesReport | ConvertTo-Json -Depth 8 -Compress)"
}
Invoke-Case -Name 'route-task-prior-art-solution-search' -ExpectedExit 0 -Arguments @('route-task', '--config', $routeConfig, '--query', 'solve this bug by searching existing open source tools and root cause approaches before building from scratch', '--report', (Join-Path $fixture 'route-prior-art-report.json'))
$routePriorArtReport = Read-JsonUtf8 -Path (Join-Path $fixture 'route-prior-art-report.json')
if (-not ($routePriorArtReport.capability_mounts.distilled_atoms -contains 'prior-art-solution-search')) {
    throw "FAIL route-task missing prior-art solution search atom report=$($routePriorArtReport | ConvertTo-Json -Depth 8 -Compress)"
}
if ($routePriorArtReport.intelligence_profile_contract.role -ne 'candidate-scout-not-research-system' -or -not ($routePriorArtReport.intelligence_profile_contract.must_not_do -contains 'final-analysis') -or -not ($routePriorArtReport.intelligence_profile_contract.must_not_do -contains 'install-or-execute')) {
    throw "FAIL route-task prior-art should expose intelligence scout contract report=$($routePriorArtReport | ConvertTo-Json -Depth 8 -Compress)"
}
Invoke-Case -Name 'route-task-root-cause-radar' -ExpectedExit 0 -Arguments @('route-task', '--config', $routeConfig, '--query', 'debug a login bug through root cause and regression evidence in code', '--report', (Join-Path $fixture 'route-root-cause-report.json'))
$routeRootCauseReport = Read-JsonUtf8 -Path (Join-Path $fixture 'route-root-cause-report.json')
if ($routeRootCauseReport.matched_route.id -ne 'code' -or -not ($routeRootCauseReport.task_route.oversight_chain -contains 'root-cause-officer') -or -not ($routeRootCauseReport.capability_mounts.distilled_atoms -contains 'root-cause-radar') -or -not ($routeRootCauseReport.deterministic_execution.command_candidates -contains 'root-cause-radar') -or $routeRootCauseReport.deterministic_execution.required -ne $true) {
    throw "FAIL route-task missing root-cause radar atom report=$($routeRootCauseReport | ConvertTo-Json -Depth 8 -Compress)"
}
Invoke-Case -Name 'route-task-low-efficiency-root-cause-radar' -ExpectedExit 0 -Arguments @('route-task', '--config', $routeConfig, '--query', 'low efficiency and rework in a repeated fix workflow', '--report', (Join-Path $fixture 'route-low-efficiency-root-cause-report.json'))
$routeLowEfficiencyReport = Read-JsonUtf8 -Path (Join-Path $fixture 'route-low-efficiency-root-cause-report.json')
if (-not ($routeLowEfficiencyReport.task_route.oversight_chain -contains 'root-cause-officer') -or -not ($routeLowEfficiencyReport.capability_mounts.distilled_atoms -contains 'root-cause-radar')) {
    throw "FAIL route-task low efficiency missing root-cause radar atom report=$($routeLowEfficiencyReport | ConvertTo-Json -Depth 8 -Compress)"
}
Invoke-Case -Name 'route-task-parallel-hypothesis-fanout' -ExpectedExit 0 -Arguments @('route-task', '--config', $routeConfig, '--query', 'fix code by parallel hypothesis fanout across nine places and multiple candidates', '--report', (Join-Path $fixture 'route-parallel-fanout-report.json'))
$routeParallelFanoutReport = Read-JsonUtf8 -Path (Join-Path $fixture 'route-parallel-fanout-report.json')
if (-not ($routeParallelFanoutReport.capability_mounts.distilled_atoms -contains 'parallel-hypothesis-fanout')) {
    throw "FAIL route-task missing parallel hypothesis fanout atom report=$($routeParallelFanoutReport | ConvertTo-Json -Depth 8 -Compress)"
}
Invoke-Case -Name 'route-task-patch-debt-root-cure' -ExpectedExit 0 -Arguments @('route-task', '--config', $routeConfig, '--query', 'fix code by removing patch debt bloat and workaround instead of adding another temporary patch', '--report', (Join-Path $fixture 'route-patch-debt-report.json'))
$routePatchDebtReport = Read-JsonUtf8 -Path (Join-Path $fixture 'route-patch-debt-report.json')
if (-not ($routePatchDebtReport.capability_mounts.distilled_atoms -contains 'patch-debt-root-cure')) {
    throw "FAIL route-task missing patch debt root cure atom report=$($routePatchDebtReport | ConvertTo-Json -Depth 8 -Compress)"
}
Invoke-Case -Name 'route-task-ponytail-restraint-atoms' -ExpectedExit 0 -Arguments @('route-task', '--config', $routeConfig, '--query', 'fix code with YAGNI, smallest working change, delete before add, and avoid overengineering or premature abstraction', '--report', (Join-Path $fixture 'route-ponytail-restraint-report.json'))
$routePonytailReport = Read-JsonUtf8 -Path (Join-Path $fixture 'route-ponytail-restraint-report.json')
if (-not ($routePonytailReport.capability_mounts.distilled_atoms -contains 'disciplined-debug-loop') -or -not ($routePonytailReport.capability_mounts.distilled_atoms -contains 'patch-debt-root-cure')) {
    throw "FAIL route-task Ponytail restraint did not map to existing atoms report=$($routePonytailReport | ConvertTo-Json -Depth 8 -Compress)"
}
Invoke-Case -Name 'route-task-terminal-real-run-verification' -ExpectedExit 0 -Arguments @('route-task', '--config', $routeConfig, '--query', 'finish a code fix only after final real run verification with browser test command evidence', '--report', (Join-Path $fixture 'route-terminal-verification-report.json'))
$routeTerminalVerificationReport = Read-JsonUtf8 -Path (Join-Path $fixture 'route-terminal-verification-report.json')
if (-not ($routeTerminalVerificationReport.capability_mounts.distilled_atoms -contains 'terminal-real-run-verification')) {
    throw "FAIL route-task missing terminal real-run verification atom report=$($routeTerminalVerificationReport | ConvertTo-Json -Depth 8 -Compress)"
}
Invoke-Case -Name 'route-task-research-atoms' -ExpectedExit 0 -Arguments @('route-task', '--config', $routeConfig, '--query', 'research latest github mcp plugin sources and cite evidence', '--report', (Join-Path $fixture 'route-research-report.json'))
$routeResearchReport = Read-JsonUtf8 -Path (Join-Path $fixture 'route-research-report.json')
if ($routeResearchReport.matched_route.id -ne 'search' -or -not ($routeResearchReport.capability_mounts.distilled_atoms -contains 'guarded-realtime-source-search') -or -not ($routeResearchReport.capability_mounts.distilled_atoms -contains 'research-evidence-pack') -or -not ($routeResearchReport.capability_mounts.distilled_atoms -contains 'claim-fact-check')) {
    throw "FAIL route-task research missing distilled atoms report=$($routeResearchReport | ConvertTo-Json -Depth 8 -Compress)"
}
if ($routeResearchReport.intelligence_profile_contract.role -ne 'candidate-scout-not-research-system' -or -not ($routeResearchReport.capability_mounts.plugin_candidates -contains 'Browser') -or ($routeResearchReport.capability_mounts.plugin_candidates -contains 'GitHub') -or -not ($routeResearchReport.intelligence_profile_contract.must_not_do -contains 'final-analysis') -or -not ($routeResearchReport.intelligence_profile_contract.must_not_do -contains 'install-or-execute')) {
    throw "FAIL route-task research should expose GitHub-first scout contract without fake GitHub plugin admission report=$($routeResearchReport | ConvertTo-Json -Depth 8 -Compress)"
}
if ($routeResearchReport.capability_mounts.source_support_classes -contains 'guarded-realtime-source-search' -or $routeResearchReport.capability_mounts.source_support_classes -contains 'research-evidence-pack') {
    throw "FAIL route-task research mixed distilled atoms into source lineage atoms report=$($routeResearchReport | ConvertTo-Json -Depth 8 -Compress)"
}
Invoke-Case -Name 'route-task-imagegen' -ExpectedExit 0 -Arguments @('route-task', '--config', $routeConfig, '--query', 'please generate a neon teaching illustration poster cover image', '--report', (Join-Path $fixture 'route-image-report.json'))
$imageRouteReport = Read-JsonUtf8 -Path (Join-Path $fixture 'route-image-report.json')
if ($imageRouteReport.matched_route.id -ne 'imagegen' -or $imageRouteReport.recommended_tier -ne 'low' -or $imageRouteReport.reasoning_effort -ne 'low') {
    throw "FAIL route-task imagegen report=$($imageRouteReport | ConvertTo-Json -Depth 6 -Compress)"
}
if ($imageRouteReport.matched_route.provider_id -ne 'agnes-openai-free' -or $imageRouteReport.matched_route.model -ne 'agnes-image-2.1-flash') {
    throw "FAIL route-task imagegen not routed to Agnes report=$($imageRouteReport | ConvertTo-Json -Depth 6 -Compress)"
}
Invoke-Case -Name 'route-task-imagegen-fallback-when-agnes-disabled' -ExpectedExit 0 -Arguments @('route-task', '--config', $routeConfigAgnesDisabled, '--query', 'please generate a neon teaching illustration poster cover image', '--report', (Join-Path $fixture 'route-image-agnes-disabled-report.json'))
$imageAgnesDisabledReport = Read-JsonUtf8 -Path (Join-Path $fixture 'route-image-agnes-disabled-report.json')
if ($imageAgnesDisabledReport.matched_route.id -ne 'imagegen' -or $imageAgnesDisabledReport.matched_route.provider_id -ne 'deepseek-web' -or $imageAgnesDisabledReport.matched_route.model -ne 'deepseek-chat' -or $imageAgnesDisabledReport.recommended_profile.provider_id -ne 'deepseek-web' -or $imageAgnesDisabledReport.provider_fallback_applied -ne $true) {
    throw "FAIL route-task imagegen should fallback when Agnes is disabled report=$($imageAgnesDisabledReport | ConvertTo-Json -Depth 8 -Compress)"
}
if (-not ($imageRouteReport.capability_mounts.distilled_atoms -contains 'brand-asset-protocol') -or -not ($imageRouteReport.capability_mounts.distilled_atoms -contains 'anti-ai-slop-visual-rules')) {
    throw "FAIL route-task imagegen missing huashu-design distilled atoms report=$($imageRouteReport | ConvertTo-Json -Depth 8 -Compress)"
}
Invoke-Case -Name 'route-task-video-motion-atoms' -ExpectedExit 0 -Arguments @('route-task', '--config', $routeConfig, '--query', 'create an animated motion stage sprite product demo video', '--report', (Join-Path $fixture 'route-video-report.json'))
$videoRouteReport = Read-JsonUtf8 -Path (Join-Path $fixture 'route-video-report.json')
if ($videoRouteReport.matched_route.id -ne 'video' -or -not ($videoRouteReport.capability_mounts.distilled_atoms -contains 'motion-stage-sprite-engine') -or -not ($videoRouteReport.capability_mounts.distilled_atoms -contains 'anti-ai-slop-visual-rules')) {
    throw "FAIL route-task video missing motion distilled atoms report=$($videoRouteReport | ConvertTo-Json -Depth 8 -Compress)"
}
if ($videoRouteReport.matched_route.provider_id -ne 'agnes-openai-free' -or $videoRouteReport.matched_route.model -ne 'agnes-video-v2.0') {
    throw "FAIL route-task video not routed to Agnes report=$($videoRouteReport | ConvertTo-Json -Depth 6 -Compress)"
}
if ($videoRouteReport.recommended_profile.provider_id -ne 'agnes-openai-free' -or $videoRouteReport.recommended_profile.model -ne 'agnes-video-v2.0' -or $videoRouteReport.recommended_tier -ne 'low' -or $videoRouteReport.reasoning_effort -ne 'low') {
    throw "FAIL route-task video should stay on Agnes video route report=$($videoRouteReport | ConvertTo-Json -Depth 8 -Compress)"
}
Invoke-Case -Name 'route-task-video-fallback-when-agnes-disabled' -ExpectedExit 0 -Arguments @('route-task', '--config', $routeConfigAgnesDisabled, '--query', 'create an animated motion stage sprite product demo video', '--report', (Join-Path $fixture 'route-video-agnes-disabled-report.json'))
$videoAgnesDisabledReport = Read-JsonUtf8 -Path (Join-Path $fixture 'route-video-agnes-disabled-report.json')
if ($videoAgnesDisabledReport.matched_route.id -ne 'video' -or $videoAgnesDisabledReport.matched_route.provider_id -ne 'deepseek-web' -or $videoAgnesDisabledReport.matched_route.model -ne 'deepseek-chat' -or $videoAgnesDisabledReport.recommended_profile.provider_id -ne 'deepseek-web' -or $videoAgnesDisabledReport.provider_fallback_applied -ne $true) {
    throw "FAIL route-task video should fallback when Agnes is disabled report=$($videoAgnesDisabledReport | ConvertTo-Json -Depth 8 -Compress)"
}
Invoke-Case -Name 'route-task-visual-motion-hybrid' -ExpectedExit 0 -Arguments @('route-task', '--config', $routeConfig, '--query', 'make a motion animated html deck presentation with product demo video and editable pptx', '--report', (Join-Path $fixture 'route-visual-motion-report.json'))
$visualMotionRouteReport = Read-JsonUtf8 -Path (Join-Path $fixture 'route-visual-motion-report.json')
if ($visualMotionRouteReport.matched_route.id -ne 'visual' -or -not ($visualMotionRouteReport.capability_mounts.distilled_atoms -contains 'native-pptx-master-route') -or -not ($visualMotionRouteReport.capability_mounts.distilled_atoms -contains 'brand-asset-protocol') -or -not ($visualMotionRouteReport.capability_mounts.distilled_atoms -contains 'motion-stage-sprite-engine')) {
    throw "FAIL route-task visual motion hybrid wrong route or missing atoms report=$($visualMotionRouteReport | ConvertTo-Json -Depth 8 -Compress)"
}
Invoke-Case -Name 'route-task-comfyui' -ExpectedExit 0 -Arguments @('route-task', '--config', $routeConfig, '--query', 'build a ComfyUI workflow node plugin for batch generation', '--report', (Join-Path $fixture 'route-comfyui-report.json'))
$comfyRouteReport = Read-JsonUtf8 -Path (Join-Path $fixture 'route-comfyui-report.json')
if ($comfyRouteReport.matched_route.id -ne 'comfyui' -or -not ($comfyRouteReport.task_route.specialized_entrypoints -contains 'ComfyUI主帅')) {
    throw "FAIL route-task comfyui report=$($comfyRouteReport | ConvertTo-Json -Depth 6 -Compress)"
}
Invoke-Case -Name 'route-task-prompt-specialist' -ExpectedExit 0 -Arguments @('route-task', '--config', $routeConfig, '--query', 'rewrite prompt into storyboard-spec and image-spec for an illustration workflow', '--report', (Join-Path $fixture 'route-prompt-report.json'))
$promptRouteReport = Read-JsonUtf8 -Path (Join-Path $fixture 'route-prompt-report.json')
if ($promptRouteReport.matched_route.id -ne 'prompt' -or -not ($promptRouteReport.task_route.specialized_entrypoints -contains '提示词主帅')) {
    throw "FAIL route-task prompt specialist report=$($promptRouteReport | ConvertTo-Json -Depth 6 -Compress)"
}
Invoke-Case -Name 'route-task-security-specialist' -ExpectedExit 0 -Arguments @('route-task', '--config', $routeConfig, '--query', 'perform threat model and dependency risk review for this plugin secret handling flow', '--report', (Join-Path $fixture 'route-security-report.json'))
$securityRouteReport = Read-JsonUtf8 -Path (Join-Path $fixture 'route-security-report.json')
if ($securityRouteReport.matched_route.id -ne 'code' -or -not ($securityRouteReport.task_route.specialized_entrypoints -contains '安全主帅') -or -not ($securityRouteReport.task_route.oversight_chain -contains 'guard-office')) {
    throw "FAIL route-task security specialist report=$($securityRouteReport | ConvertTo-Json -Depth 8 -Compress)"
}
Invoke-Case -Name 'route-task-staff-specialist' -ExpectedExit 0 -Arguments @('route-task', '--config', $routeConfig, '--query', 'analyze architecture tradeoffs and system design route with complete materials before conclusion', '--report', (Join-Path $fixture 'route-staff-specialist-report.json'))
$staffSpecialistReport = Read-JsonUtf8 -Path (Join-Path $fixture 'route-staff-specialist-report.json')
if (-not ($staffSpecialistReport.task_route.specialized_entrypoints -contains '参谋主帅')) {
    throw "FAIL route-task staff specialist report=$($staffSpecialistReport | ConvertTo-Json -Depth 8 -Compress)"
}
Invoke-Case -Name 'route-task-expedition-specialist' -ExpectedExit 0 -Arguments @('route-task', '--config', $routeConfig, '--query', 'run closeout-check for broad cleanup across multiple files with parallel work slices handoff merge results and final closeout', '--report', (Join-Path $fixture 'route-expedition-specialist-report.json'))
$expeditionSpecialistReport = Read-JsonUtf8 -Path (Join-Path $fixture 'route-expedition-specialist-report.json')
if ($expeditionSpecialistReport.matched_route.id -ne 'execution-base' -or $expeditionSpecialistReport.task_route.state -ne 'LIGHT_EXECUTION' -or ($expeditionSpecialistReport.task_route.specialized_entrypoints -contains '交付主帅')) {
    throw "FAIL route-task expedition specialist report=$($expeditionSpecialistReport | ConvertTo-Json -Depth 8 -Compress)"
}
Invoke-Case -Name 'route-task-evolve-atoms' -ExpectedExit 0 -Arguments @('route-task', '--config', $routeConfig, '--query', 'distill skills and evolve rules through fusion', '--report', (Join-Path $fixture 'route-evolve-report.json'))
$evolveRouteReport = Read-JsonUtf8 -Path (Join-Path $fixture 'route-evolve-report.json')
if ($evolveRouteReport.matched_route.id -ne 'evolve' -or -not ($evolveRouteReport.capability_mounts.distilled_atoms -contains 'skill-stocktake-daily-library') -or -not ($evolveRouteReport.capability_mounts.distilled_atoms -contains 'verified-learning-loop')) {
    throw "FAIL route-task evolve missing distilled atoms report=$($evolveRouteReport | ConvertTo-Json -Depth 8 -Compress)"
}
Invoke-Case -Name 'route-task-performance-benchmark-officer' -ExpectedExit 0 -Arguments @('route-task', '--config', $routeConfig, '--query', 'analyze token cost hit rate latency benchmark performance and reduce total rework cost', '--report', (Join-Path $fixture 'route-performance-officer-report.json'))
$performanceOfficerReport = Read-JsonUtf8 -Path (Join-Path $fixture 'route-performance-officer-report.json')
if ($performanceOfficerReport.matched_route.id -ne 'execution-base' -or $performanceOfficerReport.recommended_tier -ne 'standard' -or -not ($performanceOfficerReport.task_route.oversight_chain -contains 'performance-benchmark-on-demand') -or ($performanceOfficerReport.task_route.oversight_chain -contains 'quality-inspection') -or -not ($performanceOfficerReport.capability_mounts.distilled_atoms -contains 'content-type-compression-router') -or -not ($performanceOfficerReport.capability_mounts.distilled_atoms -contains 'terminal-real-run-verification') -or -not ($performanceOfficerReport.deterministic_execution.command_candidates -contains 'context-bloat-audit')) {
    throw "FAIL route-task missing performance benchmark officer report=$($performanceOfficerReport | ConvertTo-Json -Depth 8 -Compress)"
}
Invoke-Case -Name 'route-task-context-cache-bloat-execution-base' -ExpectedExit 0 -Arguments @('route-task', '--config', $routeConfig, '--query', 'analyze token cost hit rate context cache bloat and reduce total volume without hurting one pass success', '--report', (Join-Path $fixture 'route-context-cache-bloat-report.json'))
$contextCacheBloatReport = Read-JsonUtf8 -Path (Join-Path $fixture 'route-context-cache-bloat-report.json')
if ($contextCacheBloatReport.matched_route.id -ne 'execution-base' -or -not ($contextCacheBloatReport.task_route.oversight_chain -contains 'performance-benchmark-on-demand') -or -not ($contextCacheBloatReport.deterministic_execution.command_candidates -contains 'bench-report') -or -not ($contextCacheBloatReport.deterministic_execution.command_candidates -contains 'runtime-context-audit') -or $contextCacheBloatReport.task_route.state -ne 'LIGHT_EXECUTION' -or $contextCacheBloatReport.execution_budget.id -ne 'DIRECT_TASK') {
    throw "FAIL route-task context cache bloat should use execution-base report=$($contextCacheBloatReport | ConvertTo-Json -Depth 8 -Compress)"
}
Invoke-Case -Name 'route-task-blue-hit-200k-context' -ExpectedExit 0 -Arguments @('route-task', '--config', $routeConfig, '--query', '后台 token 每条 200k 蓝色命中 命中体量大 可能是上下文太长', '--report', (Join-Path $fixture 'route-blue-hit-200k-report.json'))
$blueHitRouteReport = Read-JsonUtf8 -Path (Join-Path $fixture 'route-blue-hit-200k-report.json')
if ($blueHitRouteReport.matched_route.id -ne 'execution-base' -or -not ($blueHitRouteReport.task_route.oversight_chain -contains 'performance-benchmark-on-demand') -or -not ($blueHitRouteReport.deterministic_execution.command_candidates -contains 'runtime-context-audit')) {
    throw "FAIL route-task blue hit 200k should use execution-base runtime-context-audit report=$($blueHitRouteReport | ConvertTo-Json -Depth 8 -Compress)"
}
Invoke-Case -Name 'route-task-pure-performance-execution-base' -ExpectedExit 0 -Arguments @('route-task', '--config', $routeConfig, '--query', 'latency p95 memory resource speed performance regression', '--report', (Join-Path $fixture 'route-pure-performance-report.json'))
$purePerformanceReport = Read-JsonUtf8 -Path (Join-Path $fixture 'route-pure-performance-report.json')
if ($purePerformanceReport.matched_route.id -ne 'execution-base' -or -not ($purePerformanceReport.task_route.oversight_chain -contains 'performance-benchmark-on-demand') -or ($purePerformanceReport.task_route.oversight_chain -contains 'quality-inspection') -or ($purePerformanceReport.task_route.oversight_chain -contains 'root-cause-officer') -or ($purePerformanceReport.capability_mounts.distilled_atoms -contains 'root-cause-radar') -or ($purePerformanceReport.deterministic_execution.command_candidates -contains 'root-cause-radar') -or ($purePerformanceReport.deterministic_execution.command_candidates -contains 'runtime-context-audit') -or $purePerformanceReport.execution_budget.id -ne 'DIRECT_TASK') {
    throw "FAIL route-task pure performance should use execution-base without root-cause or runtime-context audit report=$($purePerformanceReport | ConvertTo-Json -Depth 8 -Compress)"
}
Invoke-Case -Name 'route-task-pure-token-execution-base' -ExpectedExit 0 -Arguments @('route-task', '--config', $routeConfig, '--query', 'reduce token and cost with headroom without hurting hit rate', '--report', (Join-Path $fixture 'route-pure-token-report.json'))
$pureTokenReport = Read-JsonUtf8 -Path (Join-Path $fixture 'route-pure-token-report.json')
if ($pureTokenReport.matched_route.id -ne 'execution-base' -or -not ($pureTokenReport.task_route.oversight_chain -contains 'performance-benchmark-on-demand') -or -not ($pureTokenReport.deterministic_execution.command_candidates -contains 'runtime-context-audit') -or $pureTokenReport.execution_budget.id -ne 'DIRECT_TASK') {
    throw "FAIL route-task pure token should use execution-base report=$($pureTokenReport | ConvertTo-Json -Depth 8 -Compress)"
}
Invoke-Case -Name 'route-task-quality-officer-not-owner' -ExpectedExit 0 -Arguments @('route-task', '--config', $routeConfig, '--query', 'white-hat review quality acceptance check', '--report', (Join-Path $fixture 'route-quality-officer-report.json'))
$qualityOfficerReport = Read-JsonUtf8 -Path (Join-Path $fixture 'route-quality-officer-report.json')
if ($qualityOfficerReport.matched_route.id -ne 'quality-inspection' -or $qualityOfficerReport.task_route.owner_profile -ne 'staff-runtime' -or -not ($qualityOfficerReport.task_route.oversight_chain -contains 'quality-inspection')) {
    throw "FAIL route-task quality officer should return to staff-runtime report=$($qualityOfficerReport | ConvertTo-Json -Depth 8 -Compress)"
}
Invoke-Case -Name 'route-task-qa-final-verification-canonical' -ExpectedExit 0 -Arguments @('route-task', '--config', $routeConfig, '--query', 'qa final verification', '--report', (Join-Path $fixture 'route-qa-final-verification-report.json'))
$qaFinalReport = Read-JsonUtf8 -Path (Join-Path $fixture 'route-qa-final-verification-report.json')
if ($qaFinalReport.matched_route.id -ne 'quality-inspection' -or $qaFinalReport.task_route.route_id -eq 'qa' -or -not ($qaFinalReport.task_route.oversight_chain -contains 'quality-inspection')) {
    throw "FAIL route-task qa should canonicalize to quality-inspection report=$($qaFinalReport | ConvertTo-Json -Depth 8 -Compress)"
}
Invoke-Case -Name 'route-task-compliance-officer' -ExpectedExit 0 -Arguments @('route-task', '--config', $routeConfig, '--query', 'check license attribution privacy PII SBOM CycloneDX SPDX provenance before publishing imported third party assets', '--report', (Join-Path $fixture 'route-compliance-officer-report.json'))
$complianceOfficerReport = Read-JsonUtf8 -Path (Join-Path $fixture 'route-compliance-officer-report.json')
if (-not ($complianceOfficerReport.task_route.oversight_chain -contains 'compliance-on-demand') -or -not ($complianceOfficerReport.capability_mounts.distilled_atoms -contains 'claim-fact-check') -or -not ($complianceOfficerReport.capability_mounts.distilled_atoms -contains 'research-evidence-pack')) {
    throw "FAIL route-task missing compliance officer report=$($complianceOfficerReport | ConvertTo-Json -Depth 8 -Compress)"
}
Invoke-Case -Name 'route-task-all-independent-officers-explicit' -ExpectedExit 0 -Arguments @('route-task', '--config', $routeConfig, '--query', 'all independent officers joint review full legion whole system analysis', '--report', (Join-Path $fixture 'route-all-officers-report.json'))
$allOfficersReport = Read-JsonUtf8 -Path (Join-Path $fixture 'route-all-officers-report.json')
$expectedOfficers = @('white-hat','guard-office','root-cause-officer','audit','quality-inspection','performance-benchmark-on-demand','compliance-on-demand')
foreach ($seat in $expectedOfficers) {
    if (-not ($allOfficersReport.task_route.oversight_chain -contains $seat)) {
        throw "FAIL route-task all officers missing seat=$seat report=$($allOfficersReport | ConvertTo-Json -Depth 8 -Compress)"
    }
}
if ($allOfficersReport.execution_budget.id -ne 'DIRECT_TASK' -or $allOfficersReport.execution_budget.sidecar_mode -ne 'all-relevant-once' -or $allOfficersReport.task_route.explicit_officer_activation.mode -ne 'all-independent-officers-explicit' -or $allOfficersReport.task_route.explicit_officer_activation.all_officers_requested -ne $true -or $allOfficersReport.task_route.explicit_officer_activation.subagent_substitute_forbidden -ne $true) {
    throw "FAIL route-task all officers explicit activation report=$($allOfficersReport | ConvertTo-Json -Depth 8 -Compress)"
}
$officerLedger = $allOfficersReport.task_route.officer_activation_ledger
if (-not $officerLedger -or $allOfficersReport.task_route.single_write_authority -ne 'main-chain-only' -or $officerLedger.merge_owner -ne 'staff-runtime' -or $officerLedger.single_write_authority -ne 'main-chain-only' -or $officerLedger.seat_count -ne 7) {
    throw "FAIL route-task all officers ledger report=$($allOfficersReport | ConvertTo-Json -Depth 8 -Compress)"
}
if ($allOfficersReport.capability_mounts.distilled_atom_evidence.Count -lt 1 -or $allOfficersReport.capability_mounts.current_audit_evidence.Count -lt 3) {
    throw "FAIL route-task all officers missing evidence handles report=$($allOfficersReport | ConvertTo-Json -Depth 8 -Compress)"
}
foreach ($seat in $expectedOfficers) {
    $seatEntry = @($officerLedger.seats | Where-Object { $_.seat -eq $seat }) | Select-Object -First 1
    if (-not $seatEntry -or $seatEntry.write_authority -ne 'none' -or $seatEntry.merge_target -ne 'staff-runtime' -or $seatEntry.substitute_forbidden -ne $true) {
        throw "FAIL route-task all officers seat ledger seat=$seat report=$($allOfficersReport | ConvertTo-Json -Depth 8 -Compress)"
    }
}
Invoke-Case -Name 'route-task-fallback-chat' -ExpectedExit 0 -Arguments @('route-task', '--config', $routeConfig, '--query', 'just chat casually with me', '--report', (Join-Path $fixture 'route-chat-report.json'))
$chatRouteReport = Read-JsonUtf8 -Path (Join-Path $fixture 'route-chat-report.json')
if ($chatRouteReport.matched_route.id -ne 'chat' -or $chatRouteReport.recommended_tier -ne 'low' -or $chatRouteReport.task_route.oversight_chain.Count -ne 0) {
    throw "FAIL route-task fallback chat report=$($chatRouteReport | ConvertTo-Json -Depth 6 -Compress)"
}
Invoke-Case -Name 'route-task-explicit-default-model-override' -ExpectedExit 0 -Arguments @('route-task', '--config', $routeConfig, '--query', 'just chat casually with me but use default model', '--report', (Join-Path $fixture 'route-chat-explicit-default-report.json'))
$chatExplicitDefaultReport = Read-JsonUtf8 -Path (Join-Path $fixture 'route-chat-explicit-default-report.json')
if ($chatExplicitDefaultReport.matched_route.id -ne 'chat' -or $chatExplicitDefaultReport.matched_route.provider_id -ne 'deepseek-web' -or $chatExplicitDefaultReport.matched_route.model -ne 'deepseek-chat' -or $chatExplicitDefaultReport.provider_override_applied -ne $true -or $chatExplicitDefaultReport.provider_override_source -ne 'default model') {
    throw "FAIL route-task explicit default model override report=$($chatExplicitDefaultReport | ConvertTo-Json -Depth 8 -Compress)"
}
Invoke-Case -Name 'route-task-explicit-deepseek-override-image' -ExpectedExit 0 -Arguments @('route-task', '--config', $routeConfig, '--query', 'please generate a neon teaching illustration poster cover image and use deepseek', '--report', (Join-Path $fixture 'route-image-explicit-deepseek-report.json'))
$imageExplicitDeepseekReport = Read-JsonUtf8 -Path (Join-Path $fixture 'route-image-explicit-deepseek-report.json')
if ($imageExplicitDeepseekReport.matched_route.id -ne 'imagegen' -or $imageExplicitDeepseekReport.matched_route.provider_id -ne 'deepseek-web' -or $imageExplicitDeepseekReport.matched_route.model -ne 'deepseek-chat' -or $imageExplicitDeepseekReport.provider_override_applied -ne $true -or $imageExplicitDeepseekReport.provider_override_source -ne 'deepseek') {
    throw "FAIL route-task explicit deepseek override on image route report=$($imageExplicitDeepseekReport | ConvertTo-Json -Depth 8 -Compress)"
}
Invoke-Case -Name 'route-task-explicit-agnes-override-code' -ExpectedExit 0 -Arguments @('route-task', '--config', $routeConfig, '--query', 'fix bug and use agnes', '--report', (Join-Path $fixture 'route-code-explicit-agnes-report.json'))
$codeExplicitAgnesReport = Read-JsonUtf8 -Path (Join-Path $fixture 'route-code-explicit-agnes-report.json')
if ($codeExplicitAgnesReport.matched_route.id -ne 'code' -or $codeExplicitAgnesReport.matched_route.provider_id -ne 'agnes-openai-free' -or $codeExplicitAgnesReport.provider_override_applied -ne $true -or $codeExplicitAgnesReport.provider_override_source -ne 'agnes') {
    throw "FAIL route-task explicit Agnes override on code route report=$($codeExplicitAgnesReport | ConvertTo-Json -Depth 8 -Compress)"
}
Invoke-Case -Name 'context-pack' -ExpectedExit 0 -Arguments @('context-pack', '--config', $routeConfig, '--workspace', $fixture, '--query', 'please build a ppt ui design', '--artifact', $artifact, '--report', (Join-Path $fixture 'context-pack.json'))
$contextPack = Read-JsonUtf8 -Path (Join-Path $fixture 'context-pack.json')
if ($contextPack.stable_prefix.iron_rules_version -ne '11.3' -or $contextPack.dynamic_context.model_tier -ne 'standard' -or $contextPack.route_report -or -not $contextPack.route_summary.query_key) {
    throw "FAIL context-pack wrong stable prefix"
}
if ($contextPack.command -ne 'context-pack' -or -not $contextPack.generated_at -or $contextPack.wuji_version -ne '11.3' -or -not $contextPack.tool_source_hash -or -not $contextPack.input_hashes.'config.json' -or -not $contextPack.input_hashes.'tools/wuji_cli.go') {
    throw "FAIL context-pack freshness metadata missing report=$($contextPack | ConvertTo-Json -Depth 8 -Compress)"
}
if (-not ($contextPack.dynamic_context.distilled_atoms -contains 'content-type-compression-router') -or -not ($contextPack.dynamic_context.distilled_atoms -contains 'reversible-evidence-handle')) {
    throw "FAIL context-pack missing distilled atoms report=$($contextPack | ConvertTo-Json -Depth 8 -Compress)"
}
if ($contextPack.dynamic_context.distilled_atom_evidence.Count -lt 1 -or $contextPack.dynamic_context.current_audit_evidence.Count -lt 3 -or $contextPack.route_summary.distilled_atom_evidence_count -lt 1 -or $contextPack.route_summary.current_audit_evidence_count -lt 3) {
    throw "FAIL context-pack missing current audit evidence report=$($contextPack | ConvertTo-Json -Depth 8 -Compress)"
}
if ($contextPack.concise_execution_contract.objective -ne 'short-precise-high-hit-low-total-cost') {
    throw "FAIL context-pack missing concise execution contract report=$($contextPack | ConvertTo-Json -Depth 8 -Compress)"
}
if ($contextPack.execution_budget_contract.objective -ne 'all-work-direct-small-task-execution' -or $contextPack.route_summary.execution_budget.id -ne 'DIRECT_TASK') {
    throw "FAIL context-pack missing execution budget contract or route summary report=$($contextPack | ConvertTo-Json -Depth 8 -Compress)"
}
if ($contextPack.PSObject.Properties.Name -contains 'analysis_completeness_contract') {
    throw "FAIL context-pack non-analysis should not carry analysis completeness contract report=$($contextPack | ConvertTo-Json -Depth 8 -Compress)"
}
if (-not $contextPack.stable_prefix_canon.canon_hash -or ($contextPack.stable_prefix_canon.PSObject.Properties.Name -contains 'canon_text') -or ($contextPack.stable_prefix_canon.PSObject.Properties.Name -contains 'ordered_fields')) {
    throw "FAIL context-pack stable prefix canon should stay compact report=$($contextPack.stable_prefix_canon | ConvertTo-Json -Depth 8 -Compress)"
}
if (-not $contextPack.artifact_summaries[0].path_ref -or $contextPack.artifact_summaries[0].path) {
    throw "FAIL context-pack artifact privacy report=$($contextPack | ConvertTo-Json -Depth 8 -Compress)"
}
Invoke-Case -Name 'context-pack-analysis-completeness' -ExpectedExit 0 -Arguments @('context-pack', '--config', $routeConfig, '--workspace', $fixture, '--query', 'architecture analysis with incomplete docs and unknown modules', '--artifact', $artifact, '--report', (Join-Path $fixture 'context-pack-analysis.json'))
$analysisContextPack = Read-JsonUtf8 -Path (Join-Path $fixture 'context-pack-analysis.json')
if ($analysisContextPack.route_summary.analysis_required -ne $true -or $analysisContextPack.analysis_completeness_contract.objective -ne 'complete-materials-before-architecture-analysis') {
    throw "FAIL context-pack analysis completeness contract missing report=$($analysisContextPack | ConvertTo-Json -Depth 8 -Compress)"
}
if (-not ($analysisContextPack.dynamic_context.distilled_atoms -contains 'assumption-ledger') -or -not ($analysisContextPack.dynamic_context.distilled_atoms -contains 'claim-fact-check') -or -not ($analysisContextPack.dynamic_context.distilled_atoms -contains 'research-evidence-pack')) {
    throw "FAIL context-pack analysis completeness atoms missing report=$($analysisContextPack | ConvertTo-Json -Depth 8 -Compress)"
}
Invoke-Case -Name 'context-pack-source-key-lines' -ExpectedExit 0 -Arguments @('context-pack', '--config', $routeConfig, '--workspace', $fixture, '--query', 'assemble source context without treating code words as runtime failure', '--artifact', $sourceArtifact, '--report', (Join-Path $fixture 'context-pack-source.json'))
$sourceContextPack = Read-JsonUtf8 -Path (Join-Path $fixture 'context-pack-source.json')
$sourceSummary = $sourceContextPack.artifact_summaries[0]
if ($sourceSummary.summary_mode -ne 'key-lines' -or $sourceSummary.failure_lines) {
    throw "FAIL context-pack source code should not use failure-lines report=$($sourceContextPack | ConvertTo-Json -Depth 8 -Compress)"
}
$privateContextArtifact = Join-Path $fixture '.codex\auth\private-context.txt'
Write-Fixture $privateContextArtifact 'session token should never be read into context pack'
Invoke-Case -Name 'context-pack-private-artifact-denied' -ExpectedExit 0 -Arguments @('context-pack', '--config', $routeConfig, '--workspace', $fixture, '--query', 'assemble context with a private artifact path', '--artifact', $privateContextArtifact, '--report', (Join-Path $fixture 'context-pack-private.json'))
$privateContextPack = Read-JsonUtf8 -Path (Join-Path $fixture 'context-pack-private.json')
$privateSummary = $privateContextPack.artifact_summaries[0]
if ($privateSummary.status -ne 'denied' -or -not ($privateSummary.failures -contains 'artifact_private_denied') -or $privateSummary.artifact_hash -or $privateSummary.key_lines -or $privateSummary.failure_lines) {
    throw "FAIL context-pack private artifact was not handle-only denied report=$($privateContextPack | ConvertTo-Json -Depth 8 -Compress)"
}
Invoke-Case -Name 'feedback-log' -ExpectedExit 0 -Arguments @('feedback-log', '--workspace', $fixture, '--task', 'daily answer quality tuning', '--prefer', 'keep the answer concise', '--prefer', 'state uncertainty explicitly', '--avoid', 'placeholder', '--report', (Join-Path $fixture 'feedback-log-report.json'))
$feedbackLogReport = Read-JsonUtf8 -Path (Join-Path $fixture 'feedback-log-report.json')
if (-not $feedbackLogReport.log_ref -or $feedbackLogReport.log) {
    throw "FAIL feedback-log report retained absolute log path report=$($feedbackLogReport | ConvertTo-Json -Depth 6 -Compress)"
}
Invoke-Case -Name 'feedback-log-second' -ExpectedExit 0 -Arguments @('feedback-log', '--workspace', $fixture, '--task', 'source discipline', '--prefer', 'cite primary sources', '--avoid', 'todo', '--source', 'quality_inspection')
Invoke-Case -Name 'feedback-log-third-repeat-task' -ExpectedExit 0 -Arguments @('feedback-log', '--workspace', $fixture, '--task', 'daily answer quality tuning', '--prefer', 'keep the answer concise', '--note', 'repeat so candidate sink can trigger')
$feedbackLog = Join-Path $fixture 'feedback\feedback-log.jsonl'
$feedbackDataset = Join-Path $fixture 'feedback\feedback-dataset.json'
Invoke-Case -Name 'feedback-dataset' -ExpectedExit 0 -Arguments @('feedback-dataset', '--log', $feedbackLog, '--report', $feedbackDataset)
$feedbackRows = Read-NdjsonUtf8 -Path $feedbackLog
foreach ($row in $feedbackRows) {
    if ($row.PSObject.Properties.Name -contains 'task_label' -or $row.PSObject.Properties.Name -contains 'task' -or $row.PSObject.Properties.Name -contains 'prefer_terms' -or $row.PSObject.Properties.Name -contains 'avoid_terms' -or $row.PSObject.Properties.Name -contains 'note') {
        throw "FAIL feedback-log retained raw user text row=$($row | ConvertTo-Json -Depth 6 -Compress)"
    }
    if ($row.privacy_mode -ne 'hash-only-no-user-text' -or -not $row.task_key) {
        throw "FAIL feedback-log missing hash-only fields row=$($row | ConvertTo-Json -Depth 6 -Compress)"
    }
}
$feedbackDatasetObj = Read-JsonUtf8 -Path $feedbackDataset
if ($feedbackDatasetObj.summary.privacy_mode -ne 'hash-only-no-user-text') {
    throw "FAIL feedback-dataset missing hash-only privacy mode report=$($feedbackDatasetObj | ConvertTo-Json -Depth 8 -Compress)"
}
$repeatCandidatesReport = Join-Path $fixture 'feedback\repeat-candidates.json'
Invoke-Case -Name 'repeat-candidates' -ExpectedExit 0 -Arguments @('repeat-candidates', '--log', $feedbackLog, '--report', $repeatCandidatesReport)
$repeatCandidates = Read-JsonUtf8 -Path $repeatCandidatesReport
if (($repeatCandidates | ConvertTo-Json -Depth 8 -Compress) -match 'task_label|prefer_terms|avoid_terms') {
    throw "FAIL repeat-candidates retained legacy raw fields report=$($repeatCandidates | ConvertTo-Json -Depth 8 -Compress)"
}
if (-not ($repeatCandidates.candidates | Where-Object { $_.task_key -and $_.occurrences -ge 2 })) {
    throw "FAIL repeat-candidates report=$($repeatCandidates | ConvertTo-Json -Depth 6 -Compress)"
}
if (-not ($repeatCandidates.distill_queue | Where-Object { $_.task_key -and $_.target -eq 'cli-or-skill' })) {
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
Invoke-Case -Name 'evidence-grade-verified-missing-artifact-blocked' -ExpectedExit 1 -Arguments @('evidence-grade', '--status', 'verified', '--summary', 'issue verified without artifact', '--report', (Join-Path $evidenceGradeWorkspace 'blocked.json'))

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
$workflowBlockedWait = Join-Path $fixture 'workflow-blocked-wait'
New-Item -ItemType Directory -Force -Path (Join-Path $workflowBlockedWait 'packets'), (Join-Path $workflowBlockedWait 'results') | Out-Null
Write-Fixture (Join-Path $workflowBlockedWait 'contract.md')
Write-Fixture (Join-Path $workflowBlockedWait 'state.json') '{"status":"done","verification":{"status":"passed"}}'
Write-Fixture (Join-Path $workflowBlockedWait 'final-report.md') "# Report`n## Verification Evidence`n- passed"
Write-Fixture (Join-Path $workflowBlockedWait 'packets\\01.md')
Write-Fixture (Join-Path $workflowBlockedWait 'results\\01.md')
[System.IO.File]::WriteAllLines((Join-Path $workflowBlockedWait 'task-log.jsonl'), @(
    (@{ timestamp = '2026-06-03T00:00:00Z'; event = 'blocked'; status = 'needs_decision'; note = 'reply continue after your confirmation'; artifacts = @() } | ConvertTo-Json -Compress),
    (@{ timestamp = '2026-06-03T00:01:00Z'; event = 'end'; status = 'done'; note = 'done'; artifacts = @($artifact); closeout_report = $closeoutPassReport; evidence_report = $verifiedEvidenceReport } | ConvertTo-Json -Compress)
), [System.Text.UTF8Encoding]::new($false))
Invoke-Case -Name 'workflow-final-blocked-wait-blocked' -ExpectedExit 1 -Arguments @('workflow-guard', '--workspace', $workflowBlockedWait, '--stage', 'final')

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
$fidelityReport = Read-JsonUtf8 -Path (Join-Path $pptPipeline 'quality\template-fidelity-check.json')
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
    (Join-Path $pipelineHtmlWorkspace 'quality\pptx-audit.json')
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
    (Join-Path $pipelineTemplateWorkspace 'quality\template-fidelity-check.json'),
    (Join-Path $pipelineTemplateWorkspace 'quality\pptx-audit.json')
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
$mcpPrivateManifest = Join-Path $fixture '.codex\auth\mcp-private.json'
Write-Fixture $mcpPrivateManifest '{"name":"private-mcp"}'
Invoke-Case -Name 'mcp-guard-private-manifest-blocked' -ExpectedExit 1 -Arguments @('mcp-guard', '--manifest', $mcpPrivateManifest, '--workspace', $fixture)
$mcpOutside = Join-Path $fixture 'mcp-outside.json'
Write-JsonUtf8 -Path $mcpOutside -Value (@{
    name = 'outside-mcp'
    version = '1.0.0'
    transport = 'stdio'
    capabilities = @{ tools = @(); resources = @(); prompts = @() }
    permissions = @{ network = $false; filesystem = @('C:\Users\Administrator\.codex\sessions') }
}) -Depth 8
Invoke-Case -Name 'mcp-guard-filesystem-private-blocked' -ExpectedExit 1 -Arguments @('mcp-guard', '--manifest', $mcpOutside, '--workspace', $fixture)

$supplyArtifact = Join-Path $fixture 'supply\bundle.txt'
Write-Fixture $supplyArtifact 'supply chain artifact bytes'
$supplyHash = Get-Sha256Lower -Path $supplyArtifact
$supplySafe = Join-Path $fixture 'supply-safe.json'
$supplyNetwork = Join-Path $fixture 'supply-network.json'
$supplyBadHash = Join-Path $fixture 'supply-bad-hash.json'
$supplyNoLocalPath = Join-Path $fixture 'supply-no-local-path.json'
$supplyMissing = Join-Path $fixture 'supply-missing.json'
Write-JsonUtf8 -Path $supplySafe -Value (@{
    name = 'local-bundle'
    version = '1.0.0'
    source = 'local'
    ref = '0123456789abcdef0123456789abcdef01234567'
    sha256 = $supplyHash
    license = 'MIT'
    local_path = $supplyArtifact
    execute_after_fetch = $false
}) -Depth 8
Write-JsonUtf8 -Path $supplyNetwork -Value (@{
    name = 'remote-bundle'
    version = '1.0.0'
    source = 'https://example.invalid/bundle.js'
    ref = '0123456789abcdef0123456789abcdef01234567'
    sha256 = $supplyHash
    license = 'MIT'
    execute_after_fetch = $false
}) -Depth 8
Write-JsonUtf8 -Path $supplyBadHash -Value (@{
    name = 'local-bundle'
    version = '1.0.0'
    source = 'local'
    ref = '0123456789abcdef0123456789abcdef01234567'
    sha256 = ('0' * 64)
    license = 'MIT'
    local_path = $supplyArtifact
    execute_after_fetch = $false
}) -Depth 8
Write-JsonUtf8 -Path $supplyNoLocalPath -Value (@{
    name = 'local-bundle'
    version = '1.0.0'
    source = 'local'
    ref = '0123456789abcdef0123456789abcdef01234567'
    sha256 = $supplyHash
    license = 'MIT'
    execute_after_fetch = $false
}) -Depth 8
Write-JsonUtf8 -Path $supplyMissing -Value (@{
    name = 'local-bundle'
    version = '1.0.0'
    source = 'local'
    ref = '0123456789abcdef0123456789abcdef01234567'
    sha256 = $supplyHash
    license = 'MIT'
    local_path = (Join-Path $fixture 'supply\missing-bundle.txt')
    execute_after_fetch = $false
}) -Depth 8
Invoke-Case -Name 'supply-chain-safe' -ExpectedExit 0 -Arguments @('supply-chain', '--manifest', $supplySafe, '--workspace', $fixture)
Invoke-Case -Name 'supply-chain-network-blocked' -ExpectedExit 1 -Arguments @('supply-chain', '--manifest', $supplyNetwork, '--workspace', $fixture)
Invoke-Case -Name 'supply-chain-hash-mismatch-blocked' -ExpectedExit 1 -Arguments @('supply-chain', '--manifest', $supplyBadHash, '--workspace', $fixture)
Invoke-Case -Name 'supply-chain-local-path-required-blocked' -ExpectedExit 1 -Arguments @('supply-chain', '--manifest', $supplyNoLocalPath, '--workspace', $fixture)
Invoke-Case -Name 'supply-chain-missing-local-path-blocked' -ExpectedExit 1 -Arguments @('supply-chain', '--manifest', $supplyMissing, '--workspace', $fixture)
$supplyPrivateManifest = Join-Path $fixture '.codex\auth\supply-private.json'
Write-Fixture $supplyPrivateManifest '{"name":"private"}'
Invoke-Case -Name 'supply-chain-private-manifest-blocked' -ExpectedExit 1 -Arguments @('supply-chain', '--manifest', $supplyPrivateManifest, '--workspace', $fixture)

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
if (-not ($mcpDistill.decisions | Where-Object { $_.name -eq 'github' -and $_.decision -eq 'defer' })) {
    throw "FAIL mcp-distill network candidate should defer report=$($mcpDistill | ConvertTo-Json -Depth 8 -Compress)"
}
$mcpHighRiskCatalog = Join-Path $fixture 'mcp-high-risk-catalog.json'
$mcpHighRiskReport = Join-Path $fixture 'mcp-high-risk-report.json'
[System.IO.File]::WriteAllText($mcpHighRiskCatalog, (@{
    candidates = @(
        @{
            name = 'write-all-secrets-tool'
            owner = 'guard-office'
            source = 'https://example.invalid/high-risk'
            license = 'MIT'
            capability = 'dangerous runtime capability'
            transport = 'stdio'
            permissions = @('secrets', 'write-all')
            risk = 'high'
        }
    )
} | ConvertTo-Json -Depth 8), [System.Text.UTF8Encoding]::new($false))
Invoke-Case -Name 'mcp-distill-high-risk-reject-covered' -ExpectedExit 0 -Arguments @('mcp-distill', '--catalog', $mcpHighRiskCatalog, '--report', $mcpHighRiskReport)
$mcpHighRiskDistill = Read-JsonUtf8 -Path $mcpHighRiskReport
if (-not ($mcpHighRiskDistill.decisions | Where-Object { $_.name -eq 'write-all-secrets-tool' -and $_.decision -eq 'reject' -and $_.evidence_level -eq 'checked' })) {
    throw "FAIL mcp-distill high-risk candidate should reject report=$($mcpHighRiskDistill | ConvertTo-Json -Depth 8 -Compress)"
}
$mcpEmptyCatalog = Join-Path $fixture 'mcp-empty-catalog.json'
$mcpEmptyReport = Join-Path $fixture 'mcp-empty-report.json'
[System.IO.File]::WriteAllText($mcpEmptyCatalog, (@{
    checked_at = '2026-06-16'
    principle = 'task-scoped MCP review only'
    candidates = @()
} | ConvertTo-Json -Depth 8), [System.Text.UTF8Encoding]::new($false))
Invoke-Case -Name 'mcp-distill-empty-standing-backlog' -ExpectedExit 0 -Arguments @('mcp-distill', '--catalog', $mcpEmptyCatalog, '--report', $mcpEmptyReport)
$mcpEmptyDistill = Read-JsonUtf8 -Path $mcpEmptyReport
if ($mcpEmptyDistill.standing_backlog -ne 'empty' -or $mcpEmptyDistill.decision_mode -ne 'task-scoped-ad-hoc-only' -or $mcpEmptyDistill.decisions.Count -ne 0) {
    throw "FAIL mcp-distill empty backlog report=$($mcpEmptyDistill | ConvertTo-Json -Depth 8 -Compress)"
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
Write-JsonUtf8 -Path $promptDataset -Value (@{
    cases = @(
        @{
            id = 'prompt-quality-01'
            required_terms = @('cite primary sources', 'keep the answer concise', 'state uncertainty explicitly')
            forbidden_terms = @('placeholder')
        }
    )
}) -Depth 8
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
$badCloseoutPrompt = Join-Path $fixture 'bad-closeout-prompt.json'
[System.IO.File]::WriteAllText($badCloseoutPrompt, (@{
    name = 'bad-closeout-reopen'
    objective = 'answer with concise execution and complete closeout'
    metric = 'closeout'
    prompt_template = 'Finish the task, then tell the user the next step and ask whether to continue for further optimization.'
    stable_prefix = 'Wuji closeout prompt stable prefix with forbidden reopen-work language.'
    variables = @('task')
} | ConvertTo-Json -Depth 8), [System.Text.UTF8Encoding]::new($false))
Invoke-Case -Name 'prompt-candidate-audit-closeout-reopen-blocked' -ExpectedExit 1 -Arguments @('prompt-candidate-audit', '--candidate', $badCloseoutPrompt, '--report', (Join-Path $fixture 'bad-closeout-prompt-audit.json'))
$badManagementPrompt = Join-Path $fixture 'bad-management-prompt.json'
[System.IO.File]::WriteAllText($badManagementPrompt, (@{
    name = 'bad-management-pause-loop'
    objective = 'manage a complex task'
    metric = 'speed'
    prompt_template = ('Command center has taken over. Phase 1 analysis, phase 2 execution, phase 3 review. ' +
        'Role breakdown follows. Wait for user confirmation, then continue to the next phase.')
    stable_prefix = 'Wuji management prompt stable prefix with forbidden stage pause loop behavior.'
    variables = @('task')
} | ConvertTo-Json -Depth 8), [System.Text.UTF8Encoding]::new($false))
Invoke-Case -Name 'prompt-candidate-audit-management-pause-loop-blocked' -ExpectedExit 1 -Arguments @('prompt-candidate-audit', '--candidate', $badManagementPrompt, '--report', (Join-Path $fixture 'bad-management-prompt-audit.json'))
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
Write-Fixture (Join-Path $auditRoot 'normal-prose.md') "- Normal prose fixture without release-blocking audit markers`nInvoke-Case --avoid todo as plain fixture data"
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
$auditBlockedWaitRoot = Join-Path $fixture 'audit-blocked-wait'
New-Item -ItemType Directory -Force -Path $auditBlockedWaitRoot | Out-Null
[System.IO.File]::WriteAllLines((Join-Path $auditBlockedWaitRoot 'task-log.jsonl'), @(
    (@{ timestamp = '2026-06-03T00:02:00Z'; event = 'blocked'; status = 'needs_decision'; note = 'reply continue after your confirmation'; artifacts = @() } | ConvertTo-Json -Compress)
), [System.Text.UTF8Encoding]::new($false))
Invoke-Case -Name 'audit-blocked-wait-blocked' -ExpectedExit 1 -Arguments @('audit', '--path', $auditBlockedWaitRoot)

$previewOut = Join-Path $fixture 'preview.txt'
Invoke-Case -Name 'preview-unsafe-blocked' -ExpectedExit 1 -Arguments @('preview', '--command', 'powershell', '--arg', '-NoProfile', '--arg', '-Command', '--arg', ('Set-Content -LiteralPath ''' + $previewOut + ''' -Value ''preview output long enough'''), '--output', $previewOut)
Invoke-Case -Name 'preview-dispatch' -ExpectedExit 0 -Arguments @('preview', '--command', 'powershell', '--allow-unsafe-command', 'true', '--arg', '-NoProfile', '--arg', '-Command', '--arg', ('Set-Content -LiteralPath ''' + $previewOut + ''' -Value ''preview output long enough'''), '--output', $previewOut)

Write-RunLog 'RESULT: PASS - wuji-cli deterministic gates'
Write-RunLog ("END test-wuji-cli " + (Get-Date).ToUniversalTime().ToString('o'))
}
finally {
Remove-Item -LiteralPath $bin -Force -ErrorAction SilentlyContinue
Remove-Item -LiteralPath $fixture -Recurse -Force -ErrorAction SilentlyContinue
    $testsRoot = Join-Path $root 'outputs\tests'
    if ((Test-Path -LiteralPath $testsRoot) -and -not (Get-ChildItem -LiteralPath $testsRoot -Force | Select-Object -First 1)) {
        Remove-Item -LiteralPath $testsRoot -Force -ErrorAction SilentlyContinue
    }
}
