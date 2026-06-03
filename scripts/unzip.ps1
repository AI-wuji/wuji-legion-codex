param(
    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]]$ArgsList
)

$ErrorActionPreference = "Stop"

if ($ArgsList.Count -lt 2) {
    throw "Usage: unzip -Z1 <zip> | unzip -p <zip> <entry>"
}

Add-Type -AssemblyName System.IO.Compression
Add-Type -AssemblyName System.IO.Compression.FileSystem

$mode = $ArgsList[0]
$zipPath = [System.IO.Path]::GetFullPath($ArgsList[1])

if (-not (Test-Path -LiteralPath $zipPath)) {
    throw "Zip file not found: $zipPath"
}

$archive = [System.IO.Compression.ZipFile]::OpenRead($zipPath)
try {
    switch ($mode) {
        "-Z1" {
            foreach ($entry in $archive.Entries) {
                [Console]::Out.WriteLine($entry.FullName)
            }
        }
        "-p" {
            if ($ArgsList.Count -lt 3) {
                throw "Usage: unzip -p <zip> <entry>"
            }
            $entryName = ($ArgsList[2] -replace '\\', '/')
            $entry = $archive.Entries | Where-Object { $_.FullName -eq $entryName } | Select-Object -First 1
            if (-not $entry) {
                throw "Zip entry not found: $entryName"
            }
            $stream = $entry.Open()
            try {
                $stdout = [Console]::OpenStandardOutput()
                $stream.CopyTo($stdout)
                $stdout.Flush()
            }
            finally {
                $stream.Dispose()
            }
        }
        default {
            throw "Unsupported unzip mode: $mode"
        }
    }
}
finally {
    $archive.Dispose()
}
