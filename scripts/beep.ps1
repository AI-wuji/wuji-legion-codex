param(
    [ValidateSet("complete", "error", "notify")]
    [string]$Type = "complete",

    [switch]$NoConsoleBeep,

    [int]$DelayMs = 0,

    [switch]$SpawnDelayed
)

$ErrorActionPreference = "SilentlyContinue"

if ($SpawnDelayed) {
    $scriptPath = $PSCommandPath
    if (-not $scriptPath) {
        $scriptPath = $MyInvocation.MyCommand.Path
    }

    $arguments = @(
        "-NoProfile",
        "-ExecutionPolicy", "Bypass",
        "-File", "`"$scriptPath`"",
        $Type,
        "-DelayMs", $DelayMs.ToString()
    )

    if ($NoConsoleBeep) {
        $arguments += "-NoConsoleBeep"
    }

    Start-Process -FilePath "powershell.exe" -ArgumentList $arguments -WindowStyle Hidden | Out-Null
    Write-Output ("beep:scheduled:{0}:{1}" -f $Type, $DelayMs)
    exit 0
}

function Write-Ascii {
    param(
        [System.IO.BinaryWriter]$Writer,
        [string]$Text
    )
    $Writer.Write([System.Text.Encoding]::ASCII.GetBytes($Text))
}

function New-ToneWav {
    param(
        [string]$Path,
        [int[]]$Frequencies,
        [int[]]$Durations
    )

    $sampleRate = 44100
    $amplitude = 18000
    $channels = 1
    $bitsPerSample = 16
    $blockAlign = [int]($channels * $bitsPerSample / 8)
    $byteRate = $sampleRate * $blockAlign

    $dataStream = New-Object System.IO.MemoryStream
    $dataWriter = New-Object System.IO.BinaryWriter($dataStream)

    for ($i = 0; $i -lt $Frequencies.Count; $i++) {
        $hz = $Frequencies[$i]
        $ms = $Durations[$i]
        $sampleCount = [int]($sampleRate * $ms / 1000)

        for ($n = 0; $n -lt $sampleCount; $n++) {
            $fade = [Math]::Min(1.0, [Math]::Min($n / 500.0, ($sampleCount - $n) / 500.0))
            $value = [int16]([Math]::Sin(2 * [Math]::PI * $hz * $n / $sampleRate) * $amplitude * $fade)
            $dataWriter.Write($value)
        }

        $silenceCount = [int]($sampleRate * 60 / 1000)
        for ($s = 0; $s -lt $silenceCount; $s++) {
            $dataWriter.Write([int16]0)
        }
    }

    $data = $dataStream.ToArray()
    $fileStream = [System.IO.File]::Create($Path)
    $writer = New-Object System.IO.BinaryWriter($fileStream)

    Write-Ascii $writer "RIFF"
    $writer.Write([int](36 + $data.Length))
    Write-Ascii $writer "WAVE"
    Write-Ascii $writer "fmt "
    $writer.Write([int]16)
    $writer.Write([int16]1)
    $writer.Write([int16]$channels)
    $writer.Write([int]$sampleRate)
    $writer.Write([int]$byteRate)
    $writer.Write([int16]$blockAlign)
    $writer.Write([int16]$bitsPerSample)
    Write-Ascii $writer "data"
    $writer.Write([int]$data.Length)
    $writer.Write($data)

    $writer.Close()
    $fileStream.Close()
    $dataWriter.Close()
    $dataStream.Close()
}

function Play-Wav {
    param([string]$Path)

    try {
        $player = New-Object System.Media.SoundPlayer($Path)
        $player.Load()
        $player.PlaySync()
        return $true
    } catch {
        return $false
    }
}

function Play-ConsoleFallback {
    param(
        [int[]]$Frequencies,
        [int[]]$Durations
    )

    if ($NoConsoleBeep) {
        return
    }

    for ($i = 0; $i -lt $Frequencies.Count; $i++) {
        try {
            [Console]::Beep($Frequencies[$i], $Durations[$i])
        } catch {
            return
        }
        Start-Sleep -Milliseconds 60
    }
}

switch ($Type) {
    "complete" {
        $frequencies = @(784, 988, 1175)
        $durations = @(130, 130, 260)
    }
    "error" {
        $frequencies = @(220, 180)
        $durations = @(260, 360)
    }
    "notify" {
        $frequencies = @(740, 988)
        $durations = @(140, 180)
    }
}

if ($DelayMs -gt 0) {
    Start-Sleep -Milliseconds $DelayMs
}

$wavPath = Join-Path $env:TEMP ("wuji-beep-{0}-{1}.wav" -f $Type, ([guid]::NewGuid().ToString("N")))
$played = $false

try {
    New-ToneWav -Path $wavPath -Frequencies $frequencies -Durations $durations
    $played = Play-Wav -Path $wavPath
} finally {
    Remove-Item -LiteralPath $wavPath -Force -ErrorAction SilentlyContinue
}

if (-not $played) {
    Play-ConsoleFallback -Frequencies $frequencies -Durations $durations
}

Write-Output "beep:$Type"
