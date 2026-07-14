param(
  [Parameter(Mandatory = $true)][ValidateSet('ViewHtml', 'DumpJson')][string]$Operation,
  [Parameter(Mandatory = $true)][string]$InputPath,
  [Parameter(Mandatory = $true)][string]$OutputPath,
  [string]$OfficeCliPath = (Join-Path $env:LOCALAPPDATA 'OfficeCLI\officecli.exe'),
  [ValidateRange(1, 120)][int]$TimeoutSeconds = 45
)

$ErrorActionPreference = 'Stop'

function Resolve-ExistingFile([string]$Path, [string]$Label) {
  if (-not (Test-Path -LiteralPath $Path -PathType Leaf)) { throw "$Label does not exist: $Path" }
  return [IO.Path]::GetFullPath((Get-Item -LiteralPath $Path).FullName)
}

function Quote-ProcessArgument([string]$Value) {
  if ($Value -notmatch '[\s"]') { return $Value }
  return '"' + ($Value -replace '(\\*)"', '$1$1\\"' -replace '(\\*)$', '$1$1') + '"'
}

$cli = Resolve-ExistingFile $OfficeCliPath 'OfficeCLI executable'
$input = Resolve-ExistingFile $InputPath 'Input file'
$output = [IO.Path]::GetFullPath($OutputPath)
$extension = [IO.Path]::GetExtension($input).ToLowerInvariant()
if ($extension -notin @('.docx', '.xlsx')) { throw "OfficeCLI adapter supports only existing .docx or .xlsx files, got $extension" }
if ([IO.Path]::GetFullPath($input) -eq $output) { throw 'Output path must differ from the input file' }
if (Test-Path -LiteralPath $output) { throw "Refusing to overwrite existing output: $output" }

$outputParent = Split-Path -Parent $output
if (-not (Test-Path -LiteralPath $outputParent -PathType Container)) { throw "Output directory does not exist: $outputParent" }
switch ($Operation) {
  'ViewHtml' { if ([IO.Path]::GetExtension($output).ToLowerInvariant() -ne '.html') { throw 'ViewHtml output must end with .html' }; $arguments = @('view', $input, 'html', '-o', $output) }
  'DumpJson' { if ([IO.Path]::GetExtension($output).ToLowerInvariant() -ne '.json') { throw 'DumpJson output must end with .json' }; $arguments = @('dump', $input, '--json') }
}

$beforeHash = (Get-FileHash -LiteralPath $input -Algorithm SHA256).Hash.ToLowerInvariant()
$priorResidentSetting = $env:OFFICECLI_NO_AUTO_RESIDENT
$env:OFFICECLI_NO_AUTO_RESIDENT = '1'
$psi = [Diagnostics.ProcessStartInfo]::new()
$psi.FileName = $cli
$psi.Arguments = (($arguments | ForEach-Object { Quote-ProcessArgument ([string]$_) }) -join ' ')
$psi.UseShellExecute = $false
$psi.RedirectStandardOutput = $true
$psi.RedirectStandardError = $true
$process = [Diagnostics.Process]::new()
$process.StartInfo = $psi
try {
  if (-not $process.Start()) { throw 'OfficeCLI process did not start' }
  $stdoutTask = $process.StandardOutput.ReadToEndAsync()
  $stderrTask = $process.StandardError.ReadToEndAsync()
  if (-not $process.WaitForExit($TimeoutSeconds * 1000)) {
    & taskkill /PID $process.Id /T /F | Out-Null
    throw "OfficeCLI timed out after $TimeoutSeconds seconds"
  }
  $stdout = $stdoutTask.GetAwaiter().GetResult()
  $stderr = $stderrTask.GetAwaiter().GetResult()
  if ($process.ExitCode -ne 0) { throw "OfficeCLI failed exit=$($process.ExitCode): $stderr$stdout" }
  if ($Operation -eq 'DumpJson') { [IO.File]::WriteAllText($output, $stdout, [Text.UTF8Encoding]::new($false)) }
  if (-not (Test-Path -LiteralPath $output -PathType Leaf) -or (Get-Item -LiteralPath $output).Length -lt 1) { throw "OfficeCLI did not create a non-empty output: $output" }
  $afterHash = (Get-FileHash -LiteralPath $input -Algorithm SHA256).Hash.ToLowerInvariant()
  if ($beforeHash -ne $afterHash) { throw 'Read-only OfficeCLI operation changed the input file' }
  [pscustomobject]@{ operation = $Operation; input = $input; output = $output; input_sha256 = $afterHash; output_sha256 = (Get-FileHash -LiteralPath $output -Algorithm SHA256).Hash.ToLowerInvariant(); stdout = $stdout.Trim(); stderr = $stderr.Trim(); auto_resident_disabled = $true } | ConvertTo-Json -Compress
} finally {
  if ($null -eq $priorResidentSetting) { Remove-Item Env:OFFICECLI_NO_AUTO_RESIDENT -ErrorAction SilentlyContinue } else { $env:OFFICECLI_NO_AUTO_RESIDENT = $priorResidentSetting }
  $process.Dispose()
}
