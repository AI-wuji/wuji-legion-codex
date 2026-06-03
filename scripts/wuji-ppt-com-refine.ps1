param(
    [Parameter(Mandatory = $true)]
    [string]$Pptx,

    [Parameter(Mandatory = $true)]
    [string]$Out,

    [string]$Instructions = "",
    [string]$Report = ""
)

$ErrorActionPreference = "Stop"

function Convert-HexToRgbLong {
    param([string]$Hex)
    $clean = $Hex.Trim().TrimStart('#')
    if ($clean.Length -ne 6) {
        throw "Expected 6-digit hex color, got: $Hex"
    }
    $r = [Convert]::ToInt32($clean.Substring(0, 2), 16)
    $g = [Convert]::ToInt32($clean.Substring(2, 2), 16)
    $b = [Convert]::ToInt32($clean.Substring(4, 2), 16)
    return ($b * 65536) + ($g * 256) + $r
}

function Read-Instructions {
    param([string]$Path)
    if ([string]::IsNullOrWhiteSpace($Path)) {
        return @{
            operations = @(@{ type = "remove-empty-placeholders" })
        }
    }
    return [System.IO.File]::ReadAllText((Resolve-Path -LiteralPath $Path).Path, [System.Text.UTF8Encoding]::new($false)) | ConvertFrom-Json
}

function Find-NotesPlaceholderShape {
    param($Slide)

    if (-not $Slide) {
        return $null
    }
    $notesPage = $Slide.NotesPage
    if (-not $notesPage) {
        return $null
    }
    for ($shapeIndex = 1; $shapeIndex -le $notesPage.Shapes.Count; $shapeIndex++) {
        $shape = $notesPage.Shapes.Item($shapeIndex)
        try {
            if ($shape.PlaceholderFormat.Type -eq 2) {
                return $shape
            }
        }
        catch {
        }
        if ($shape.Name -like 'Notes Placeholder*') {
            return $shape
        }
    }
    return $null
}

$resolvedPptx = (Resolve-Path -LiteralPath $Pptx).Path
$resolvedOut = [System.IO.Path]::GetFullPath($Out)
$reportPath = if ($Report) { [System.IO.Path]::GetFullPath($Report) } else { "$resolvedOut.com-refine.json" }
$instructionsData = Read-Instructions -Path $Instructions

$ppt = $null
$presentation = $null
$removedPlaceholders = 0
$replacedText = 0
$coloredShapes = 0
$filledShapes = 0
$updatedSlideNotes = 0

try {
    $ppt = New-Object -ComObject PowerPoint.Application
    try {
        $ppt.Visible = $false
    }
    catch {
    }
    $presentation = $ppt.Presentations.Open($resolvedPptx, $false, $false, $false)

    foreach ($operation in $instructionsData.operations) {
        switch ($operation.type) {
            "remove-empty-placeholders" {
                for ($slideIndex = $presentation.Slides.Count; $slideIndex -ge 1; $slideIndex--) {
                    $slide = $presentation.Slides.Item($slideIndex)
                    for ($shapeIndex = $slide.Shapes.Count; $shapeIndex -ge 1; $shapeIndex--) {
                        $shape = $slide.Shapes.Item($shapeIndex)
                        if (-not $shape.PlaceholderFormat) { continue }
                        $text = ""
                        if ($shape.HasTextFrame -and $shape.TextFrame.HasText) {
                            $text = $shape.TextFrame.TextRange.Text
                        }
                        if ([string]::IsNullOrWhiteSpace($text)) {
                            $shape.Delete()
                            $removedPlaceholders += 1
                        }
                    }
                }
            }
            "replace-text" {
                $findText = [string]$operation.find
                $replaceText = [string]$operation.replace
                if ([string]::IsNullOrWhiteSpace($findText)) { continue }
                for ($slideIndex = 1; $slideIndex -le $presentation.Slides.Count; $slideIndex++) {
                    $slide = $presentation.Slides.Item($slideIndex)
                    for ($shapeIndex = 1; $shapeIndex -le $slide.Shapes.Count; $shapeIndex++) {
                        $shape = $slide.Shapes.Item($shapeIndex)
                        if ($shape.HasTextFrame -and $shape.TextFrame.HasText) {
                            $current = [string]$shape.TextFrame.TextRange.Text
                            if ($current.Contains($findText)) {
                                $shape.TextFrame.TextRange.Text = $current.Replace($findText, $replaceText)
                                $replacedText += 1
                            }
                        }
                    }
                }
            }
            "set-text-color" {
                $contains = [string]$operation.contains
                $color = Convert-HexToRgbLong ([string]$operation.color)
                if ([string]::IsNullOrWhiteSpace($contains)) { continue }
                for ($slideIndex = 1; $slideIndex -le $presentation.Slides.Count; $slideIndex++) {
                    $slide = $presentation.Slides.Item($slideIndex)
                    for ($shapeIndex = 1; $shapeIndex -le $slide.Shapes.Count; $shapeIndex++) {
                        $shape = $slide.Shapes.Item($shapeIndex)
                        if ($shape.HasTextFrame -and $shape.TextFrame.HasText) {
                            $current = [string]$shape.TextFrame.TextRange.Text
                            if ($current.Contains($contains)) {
                                $shape.TextFrame.TextRange.Font.Color.RGB = $color
                                $coloredShapes += 1
                            }
                        }
                    }
                }
            }
            "set-shape-fill" {
                $name = [string]$operation.name
                $color = Convert-HexToRgbLong ([string]$operation.color)
                if ([string]::IsNullOrWhiteSpace($name)) { continue }
                for ($slideIndex = 1; $slideIndex -le $presentation.Slides.Count; $slideIndex++) {
                    $slide = $presentation.Slides.Item($slideIndex)
                    for ($shapeIndex = 1; $shapeIndex -le $slide.Shapes.Count; $shapeIndex++) {
                        $shape = $slide.Shapes.Item($shapeIndex)
                        if ($shape.Name -eq $name) {
                            $shape.Fill.Visible = $true
                            $shape.Fill.Solid()
                            $shape.Fill.ForeColor.RGB = $color
                            $filledShapes += 1
                        }
                    }
                }
            }
            "set-slide-notes" {
                $slideNumber = [int]$operation.slide
                $notesText = [string]$operation.text
                if ($slideNumber -lt 1 -or $slideNumber -gt $presentation.Slides.Count) {
                    throw "Invalid slide number for set-slide-notes: $slideNumber"
                }
                $slide = $presentation.Slides.Item($slideNumber)
                $notesShape = Find-NotesPlaceholderShape -Slide $slide
                if (-not $notesShape) {
                    throw "Could not find notes placeholder on slide $slideNumber"
                }
                if (-not $notesShape.HasTextFrame) {
                    throw "Notes placeholder on slide $slideNumber has no text frame"
                }
                $normalizedNotes = $notesText -replace "`r?`n", "`r`n"
                $notesShape.TextFrame.TextRange.Text = $normalizedNotes
                $updatedSlideNotes += 1
            }
        }
    }

    $saveType = 24
    $presentation.SaveCopyAs($resolvedOut, $saveType)
    $reportData = @{
        input_pptx = $resolvedPptx
        output_pptx = $resolvedOut
        slide_count = $presentation.Slides.Count
        removed_empty_placeholders = $removedPlaceholders
        replaced_text_shapes = $replacedText
        recolored_text_shapes = $coloredShapes
        filled_shapes = $filledShapes
        updated_slide_notes = $updatedSlideNotes
    }
    [System.IO.File]::WriteAllText($reportPath, ($reportData | ConvertTo-Json -Depth 6), [System.Text.UTF8Encoding]::new($false))
    Write-Output $reportPath
}
finally {
    if ($presentation) {
        $presentation.Close()
    }
    if ($ppt) {
        $ppt.Quit()
    }
}
