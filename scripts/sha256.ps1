# Windows PowerShell installations can omit Get-FileHash. Keep probe evidence
# verification available without changing the built-in command when it exists.
if (-not (Get-Command Get-FileHash -ErrorAction SilentlyContinue)) {
  function Get-FileHash {
    [CmdletBinding()]
    param(
      [Parameter(Mandatory = $true)][string]$LiteralPath,
      [ValidateSet('SHA256')][string]$Algorithm = 'SHA256'
    )

    $stream = [IO.File]::OpenRead($LiteralPath)
    $hasher = [Security.Cryptography.SHA256]::Create()
    try {
      $bytes = $hasher.ComputeHash($stream)
    } finally {
      $stream.Dispose()
      $hasher.Dispose()
    }
    [pscustomobject]@{
      Algorithm = $Algorithm
      Hash = ([BitConverter]::ToString($bytes)).Replace('-', '')
      Path = $LiteralPath
    }
  }
}
