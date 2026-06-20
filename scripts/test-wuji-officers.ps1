param()

$ErrorActionPreference = 'Stop'

$root = Split-Path -Parent $PSScriptRoot
$binDir = Join-Path $root '.wuji-tools'
$bin = Join-Path $binDir 'wuji-cli-officer-test.exe'
$fixture = Join-Path $root 'outputs\tests\wuji-officers'
$config = Join-Path $root 'config.json'

function Read-JsonUtf8 {
    param([string]$Path)
    return [System.IO.File]::ReadAllText($Path, [System.Text.UTF8Encoding]::new($false)) | ConvertFrom-Json
}

function Assert-True {
    param(
        [bool]$Condition,
        [string]$Message
    )
    if (-not $Condition) {
        throw $Message
    }
}

if (Test-Path -LiteralPath $fixture) {
    Remove-Item -LiteralPath $fixture -Recurse -Force
}
New-Item -ItemType Directory -Force -Path $binDir, $fixture | Out-Null

try {
& (Join-Path $PSScriptRoot 'build-wuji-cli.ps1') -Output $bin

$allOfficersReportPath = Join-Path $fixture 'route-all-officers.json'
& $bin 'route-task' '--config' $config '--query' 'all independent officers joint review full legion whole system analysis' '--report' $allOfficersReportPath
if ($LASTEXITCODE -ne 0) {
    throw "route-task all officers failed exit=$LASTEXITCODE"
}
$allOfficersReport = Read-JsonUtf8 -Path $allOfficersReportPath
$expectedOfficers = @('white-hat','guard-office','root-cause-officer','audit','quality-inspection','performance-benchmark-on-demand','compliance-on-demand')

Assert-True ($allOfficersReport.matched_route.id -eq 'execution-base') 'all-officers route must force execution-base'
Assert-True ($allOfficersReport.execution_budget.id -eq 'DIRECT_TASK') 'all-officers route must stay in DIRECT_TASK with explicit final review'
Assert-True ($allOfficersReport.task_route.single_write_authority -eq 'main-chain-only') 'all-officers route must keep single write authority'
Assert-True ($allOfficersReport.task_route.explicit_officer_activation.mode -eq 'all-independent-officers-explicit') 'all-officers activation mode drift'
Assert-True ($allOfficersReport.task_route.explicit_officer_activation.subagent_substitute_forbidden -eq $true) 'all-officers generic substitute must be forbidden'
Assert-True ($allOfficersReport.task_route.officer_activation_ledger.merge_owner -eq 'staff-runtime') 'all-officers ledger merge owner drift'
Assert-True ($allOfficersReport.task_route.officer_activation_ledger.single_write_authority -eq 'main-chain-only') 'all-officers ledger write authority drift'
Assert-True ($allOfficersReport.task_route.officer_activation_ledger.seat_count -eq 7) 'all-officers ledger seat count drift'
Assert-True ($allOfficersReport.capability_mounts.distilled_atom_evidence.Count -ge 1) 'all-officers route missing distilled atom evidence'
Assert-True ((@($allOfficersReport.capability_mounts.distilled_atom_evidence | Where-Object { $_.decision_surface -eq 'fusion-matrix' })).Count -ge 1) 'all-officers route missing fusion-matrix-backed atom evidence'
Assert-True ($allOfficersReport.capability_mounts.current_audit_evidence.Count -ge 3) 'all-officers route missing current audit evidence handles'

$ledgerSeats = @($allOfficersReport.task_route.officer_activation_ledger.seats)
foreach ($seat in $expectedOfficers) {
    Assert-True (($allOfficersReport.task_route.oversight_chain -contains $seat)) "all-officers oversight chain missing seat=$seat"
    $entry = @($ledgerSeats | Where-Object { $_.seat -eq $seat }) | Select-Object -First 1
    Assert-True ($null -ne $entry) "all-officers ledger missing seat=$seat"
    Assert-True ($entry.write_authority -eq 'none') "all-officers seat write authority drift seat=$seat"
    Assert-True ($entry.merge_target -eq 'staff-runtime') "all-officers merge target drift seat=$seat"
    Assert-True ($entry.substitute_forbidden -eq $true) "all-officers seat substitute policy drift seat=$seat"
}

$narrowReportPath = Join-Path $fixture 'route-narrow-officers.json'
& $bin 'route-task' '--config' $config '--query' 'white-hat audit quality acceptance check' '--report' $narrowReportPath
if ($LASTEXITCODE -ne 0) {
    throw "route-task narrow officers failed exit=$LASTEXITCODE"
}
$narrowReport = Read-JsonUtf8 -Path $narrowReportPath
Assert-True ($narrowReport.matched_route.id -eq 'execution-base') 'narrow-officers route should stay on execution-base'
Assert-True ($narrowReport.task_route.explicit_officer_activation.mode -eq 'triggered-officers-only') 'narrow-officers mode drift'
Assert-True ($narrowReport.task_route.officer_activation_ledger.single_write_authority -eq 'main-chain-only') 'narrow-officers ledger write authority drift'
Assert-True ($narrowReport.task_route.oversight_chain.Count -lt 7) 'narrow-officers route should not inflate to full legion'
Assert-True ($narrowReport.task_route.oversight_chain -contains 'white-hat') 'narrow-officers route missing white-hat'
Assert-True ($narrowReport.task_route.oversight_chain -contains 'audit') 'narrow-officers route missing audit'

$artifact = Join-Path $fixture 'artifact.txt'
[System.IO.File]::WriteAllText($artifact, 'fixture artifact for officer context pack verification', [System.Text.UTF8Encoding]::new($false))
$contextPackPath = Join-Path $fixture 'context-pack-all-officers.json'
& $bin 'context-pack' '--config' $config '--workspace' $fixture '--query' 'all independent officers joint review full legion whole system analysis' '--artifact' $artifact '--report' $contextPackPath
if ($LASTEXITCODE -ne 0) {
    throw "context-pack all officers failed exit=$LASTEXITCODE"
}
$contextPack = Read-JsonUtf8 -Path $contextPackPath
Assert-True ($contextPack.route_summary.single_write_authority -eq 'main-chain-only') 'context-pack route summary write authority drift'
Assert-True ($contextPack.route_summary.explicit_officer_activation.mode -eq 'all-independent-officers-explicit') 'context-pack explicit activation drift'
Assert-True ($contextPack.route_summary.officer_activation_ledger.merge_owner -eq 'staff-runtime') 'context-pack ledger merge owner drift'
Assert-True ($contextPack.route_summary.officer_activation_ledger.seat_count -eq 7) 'context-pack ledger seat count drift'
Assert-True (-not $contextPack.route_summary.command_candidates) 'context-pack route summary should stay compact'
Assert-True ($contextPack.dynamic_context.distilled_atom_evidence.Count -ge 1) 'context-pack missing distilled atom evidence'
Assert-True ($contextPack.dynamic_context.current_audit_evidence.Count -ge 3) 'context-pack missing current audit evidence'
Assert-True ($contextPack.route_summary.current_audit_evidence_count -ge 3) 'context-pack route summary missing current audit evidence count'

$officerMergePath = Join-Path $fixture 'officer-merge.json'
& $bin 'officer-run' '--workspace' $fixture '--route-report' $allOfficersReportPath '--report' $officerMergePath
if ($LASTEXITCODE -ne 0) {
    throw "officer-run failed exit=$LASTEXITCODE"
}
$officerMerge = Read-JsonUtf8 -Path $officerMergePath
Assert-True ($officerMerge.main_chain_merge_decision -eq 'merged-into-staff-runtime') 'officer-run merge decision drift'
Assert-True ($officerMerge.officer_runtime_entrypoint -eq 'officer-run') 'officer-run entrypoint marker drift'
Assert-True ($officerMerge.seat_count -eq 7) 'officer-run seat count drift'
Assert-True ($officerMerge.distilled_atom_evidence.Count -ge 1) 'officer-run missing distilled atom evidence'
Assert-True ($officerMerge.current_audit_evidence.Count -ge 3) 'officer-run missing current audit evidence'
foreach ($seatVerdict in @($officerMerge.seat_verdicts)) {
    Assert-True ($seatVerdict.required_atom_evidence.Count -ge 1) "officer-run seat verdict missing atom evidence seat=$($seatVerdict.seat)"
    Assert-True ($seatVerdict.current_audit_evidence.Count -ge 3) "officer-run seat verdict missing current audit evidence seat=$($seatVerdict.seat)"
}

Write-Host 'PASS test-wuji-officers'
}
finally {
Remove-Item -LiteralPath $bin -Force -ErrorAction SilentlyContinue
Remove-Item -LiteralPath $fixture -Recurse -Force -ErrorAction SilentlyContinue
    $testsRoot = Join-Path $root 'outputs\tests'
    if ((Test-Path -LiteralPath $testsRoot) -and -not (Get-ChildItem -LiteralPath $testsRoot -Force | Select-Object -First 1)) {
        Remove-Item -LiteralPath $testsRoot -Force -ErrorAction SilentlyContinue
    }
}
