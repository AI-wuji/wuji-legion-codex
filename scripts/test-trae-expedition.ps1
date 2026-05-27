param(
    [string]$Prompt = "只回答 OK"
)

$ErrorActionPreference = "Stop"

function Test-TraeCli {
    param(
        [string]$Name,
        [string]$Path
    )

    $result = [ordered]@{
        name = $Name
        path = $Path
        exists = Test-Path -LiteralPath $Path
        help = $false
        prompt_exit_code = $null
        prompt_stdout = ""
        usable_for_auto_handoff = $false
    }

    if (-not $result.exists) {
        return [pscustomobject]$result
    }

    $helpOutput = & $Path chat --help 2>&1
    $result.help = ($LASTEXITCODE -eq 0 -and ($helpOutput -join "`n") -match "Usage:")

    $promptOutput = & $Path chat --mode ask $Prompt 2>&1
    $result.prompt_exit_code = $LASTEXITCODE
    $result.prompt_stdout = ($promptOutput -join "`n").Trim()
    $result.usable_for_auto_handoff = (
        $LASTEXITCODE -eq 0 -and
        $result.prompt_stdout.Length -gt 0 -and
        $result.prompt_stdout -notmatch "^Reading from stdin"
    )

    [pscustomobject]$result
}

$candidates = @(
    @{ Name = "Trae"; Path = Join-Path $env:LOCALAPPDATA "Programs\Trae\bin\trae.cmd" },
    @{ Name = "Trae CN"; Path = Join-Path $env:LOCALAPPDATA "Programs\Trae CN\bin\trae-cn.cmd" }
)

$results = foreach ($candidate in $candidates) {
    Test-TraeCli -Name $candidate.Name -Path $candidate.Path
}

$results | Format-Table -AutoSize

if ($results.usable_for_auto_handoff -contains $true) {
    Write-Host "RESULT: PASS - Trae can be used for automatic expedition handoff." -ForegroundColor Green
    exit 0
}

Write-Host "RESULT: FAIL - Trae CLI exists but did not return a machine-readable chat result." -ForegroundColor Yellow
exit 2
