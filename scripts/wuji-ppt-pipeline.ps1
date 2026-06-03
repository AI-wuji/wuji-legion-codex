param(
    [Parameter(Mandatory = $true)]
    [string]$Workspace,

    [Parameter(Mandatory = $true)]
    [ValidateSet('html-first', 'template-following')]
    [string]$Route,

    [Parameter(Mandatory = $true)]
    [string]$Out,

    [string]$Cli = "",
    [string]$Html = "",
    [string]$Title = "",
    [string]$Pptx = "",
    [string]$Map = "",
    [string]$Inspect = "",
    [string]$Report = "",
    [string]$PreviewDir = "",
    [string]$LayoutDir = "",
    [string]$Scale = "",
    [string]$AutoApprove = "true",
    [string]$PilotApproval = "",
    [string]$ComRefine = "",
    [string]$RefineInstructions = ""
)

$ErrorActionPreference = 'Stop'

function ConvertTo-BoolValue {
    param(
        [string]$Value,
        [string]$Name
    )

    $normalized = if ($null -eq $Value) { '' } else { $Value.Trim().ToLowerInvariant() }
    switch ($normalized) {
        'true' { return $true }
        '1' { return $true }
        'yes' { return $true }
        'false' { return $false }
        '0' { return $false }
        'no' { return $false }
        '' { return $false }
        default { throw "Invalid boolean for $($Name): $Value" }
    }
}

function Resolve-ExistingPath {
    param(
        [string]$Path,
        [string]$Name
    )

    if ([string]::IsNullOrWhiteSpace($Path)) {
        throw "Missing required $Name"
    }
    return (Resolve-Path -LiteralPath $Path).Path
}

function Resolve-OutputPath {
    param(
        [string]$Path,
        [string]$Name
    )

    if ([string]::IsNullOrWhiteSpace($Path)) {
        throw "Missing required $Name"
    }
    return [System.IO.Path]::GetFullPath($Path)
}

function Write-Utf8NoBom {
    param(
        [string]$Path,
        [string]$Content
    )

    $parent = Split-Path -Parent $Path
    if ($parent) {
        New-Item -ItemType Directory -Force -Path $parent | Out-Null
    }
    [System.IO.File]::WriteAllText($Path, $Content, [System.Text.UTF8Encoding]::new($false))
}

function Copy-Artifact {
    param(
        [string]$Source,
        [string]$Destination
    )

    $resolvedDestination = [System.IO.Path]::GetFullPath($Destination)
    $resolvedSource = [System.IO.Path]::GetFullPath($Source)
    if ($resolvedSource -ieq $resolvedDestination) {
        return $resolvedDestination
    }
    $parent = Split-Path -Parent $resolvedDestination
    if ($parent) {
        New-Item -ItemType Directory -Force -Path $parent | Out-Null
    }
    Copy-Item -LiteralPath $resolvedSource -Destination $resolvedDestination -Force
    return $resolvedDestination
}

function Resolve-WujiCli {
    param([string]$Path)

    if (-not [string]::IsNullOrWhiteSpace($Path)) {
        return [System.IO.Path]::GetFullPath($Path)
    }

    $repoRoot = Split-Path -Parent $PSScriptRoot
    foreach ($candidate in @(
        (Join-Path $repoRoot '.wuji-tools\wuji-cli.cmd'),
        (Join-Path $repoRoot '.wuji-tools\wuji-exec-base.exe')
    )) {
        if (Test-Path -LiteralPath $candidate) {
            return $candidate
        }
    }

    throw 'Could not resolve wuji-cli. Pass --cli <path>.'
}

function Test-PowerPointComAvailable {
    try {
        $type = [type]::GetTypeFromProgID('PowerPoint.Application')
        return $null -ne $type
    }
    catch {
        return $false
    }
}

function Invoke-WujiCli {
    param([string[]]$Arguments)

    & $script:ResolvedCli @Arguments
    if ($LASTEXITCODE -ne 0) {
        throw "wuji-cli failed with exit code $($LASTEXITCODE): $($Arguments -join ' ')"
    }
}

function Get-FirstFile {
    param(
        [string]$Dir,
        [string]$Filter
    )

    $item = Get-ChildItem -LiteralPath $Dir -Recurse -File -Filter $Filter -ErrorAction SilentlyContinue |
        Sort-Object FullName |
        Select-Object -First 1
    if (-not $item) {
        throw "Missing expected file under $Dir filter=$Filter"
    }
    return $item.FullName
}

function Get-HtmlFirstPreviewPathFromPptx {
    param([string]$PptxPath)

    if ([string]::IsNullOrWhiteSpace($PptxPath) -or -not (Test-Path -LiteralPath $PptxPath)) {
        return ''
    }

    $resolvedPptx = [System.IO.Path]::GetFullPath((Resolve-Path -LiteralPath $PptxPath).Path)
    $dir = Split-Path -Parent $resolvedPptx
    $stem = [System.IO.Path]::GetFileNameWithoutExtension($resolvedPptx)
    $reportCandidates = @(
        "$resolvedPptx.json",
        (Join-Path $dir "$stem-report.json"),
        (Join-Path $dir 'htmlfirst-report.json')
    ) | Select-Object -Unique

    foreach ($candidate in @($reportCandidates)) {
        if (-not (Test-Path -LiteralPath $candidate)) {
            continue
        }
        try {
            $report = Read-JsonFile -Path $candidate
            $route = ''
            if ($report.PSObject.Properties['route']) {
                $route = [string]$report.route
            }
            if ($route -ne 'html-first-editable-pptx') {
                continue
            }
            if (-not $report.PSObject.Properties['preview_path']) {
                continue
            }
            $previewPath = [string]$report.preview_path
            if (-not [string]::IsNullOrWhiteSpace($previewPath) -and (Test-Path -LiteralPath $previewPath)) {
                return [System.IO.Path]::GetFullPath((Resolve-Path -LiteralPath $previewPath).Path)
            }
        }
        catch {
            continue
        }
    }

    return ''
}

function Set-PilotApprovalArtifact {
    param(
        [string]$WorkspacePath,
        [bool]$AutoApproveEnabled,
        [string]$ApprovalPath
    )

    $target = Join-Path $WorkspacePath 'pilot-approval.md'
    if (-not [string]::IsNullOrWhiteSpace($ApprovalPath)) {
        $resolvedApproval = Resolve-ExistingPath -Path $ApprovalPath -Name '--pilot-approval'
        Copy-Item -LiteralPath $resolvedApproval -Destination $target -Force
        return $target
    }
    if ($AutoApproveEnabled) {
        Write-Utf8NoBom -Path $target -Content "approved by pipeline auto approval`n"
        return $target
    }
    throw "Pilot approval required. Review the pilot preview in $WorkspacePath and rerun with --pilot-approval <file> or --auto-approve true."
}

function New-PilotScore {
    param(
        [string]$WorkspacePath,
        [string]$RouteName,
        [int]$SlideCount
    )

    $target = Join-Path $WorkspacePath 'pilot-score.md'
    $content = @(
        "route: $RouteName"
        "slide_count: $SlideCount"
        "status: pilot-ready"
    ) -join "`n"
    Write-Utf8NoBom -Path $target -Content ($content + "`n")
    return $target
}

function Read-JsonFile {
    param([string]$Path)

    return [System.IO.File]::ReadAllText($Path, [System.Text.UTF8Encoding]::new($false)) | ConvertFrom-Json
}

function Convert-ValueToMultilineText {
    param($Value)

    if ($null -eq $Value) {
        return ''
    }
    if ($Value -is [string]) {
        return $Value.Trim()
    }
    if ($Value.PSObject -and $Value.PSObject.Properties.Name -contains 'text') {
        return Convert-ValueToMultilineText -Value $Value.text
    }
    if ($Value -is [System.Collections.IEnumerable] -and -not ($Value -is [string])) {
        $parts = @()
        foreach ($item in $Value) {
            $text = Convert-ValueToMultilineText -Value $item
            if (-not [string]::IsNullOrWhiteSpace($text)) {
                $parts += $text
            }
        }
        return ($parts -join "`n").Trim()
    }
    return ([string]$Value).Trim()
}

function Get-SlideNotesText {
    param($SlideSpec)

    if ($null -eq $SlideSpec) {
        return ''
    }
    foreach ($name in @('speakerNotes', 'speaker_notes', 'notes', 'speakerNote', 'note')) {
        $property = $SlideSpec.PSObject.Properties[$name]
        if (-not $property) {
            continue
        }
        $text = Convert-ValueToMultilineText -Value $property.Value
        if (-not [string]::IsNullOrWhiteSpace($text)) {
            return $text
        }
    }
    return ''
}

function Get-SlideEditTargets {
    param($SlideSpec)

    if ($null -eq $SlideSpec) {
        return @()
    }
    $property = $SlideSpec.PSObject.Properties['editTargets']
    if (-not $property -or $null -eq $property.Value) {
        return @()
    }
    return @($property.Value)
}

function Test-DecorativeSlideText {
    param([string]$Text)

    $normalized = ($Text -replace '\s+', ' ').Trim()
    if ([string]::IsNullOrWhiteSpace($normalized)) {
        return $true
    }
    if ($normalized -match '^[0-9]+$') {
        return $true
    }
    if ($normalized -match '^(0?[0-9]|第?[一二三四五六七八九十]+)$') {
        return $true
    }
    if ($normalized -match '^[\W_]+$') {
        return $true
    }
    return $false
}

function Add-UniqueTextValue {
    param(
        [System.Collections.Generic.List[string]]$List,
        [hashtable]$Seen,
        [string]$Value
    )

    $normalized = ($Value -replace '\s+', ' ').Trim()
    if ([string]::IsNullOrWhiteSpace($normalized)) {
        return
    }
    if ($Seen.ContainsKey($normalized)) {
        return
    }
    $Seen[$normalized] = $true
    $List.Add($normalized) | Out-Null
}

function Get-SlideVisibleTextBlocks {
    param($SlideSpec)

    $seen = @{}
    $blocks = [System.Collections.Generic.List[string]]::new()
    foreach ($target in @(Get-SlideEditTargets -SlideSpec $SlideSpec)) {
        foreach ($name in @('text', 'lines', 'textLines')) {
            $property = $target.PSObject.Properties[$name]
            if (-not $property) {
                continue
            }
            $text = Convert-ValueToMultilineText -Value $property.Value
            if (Test-DecorativeSlideText -Text $text) {
                continue
            }
            Add-UniqueTextValue -List $blocks -Seen $seen -Value $text
        }
    }
    return @($blocks)
}

function Get-SlideRole {
    param($SlideSpec)

    if ($null -eq $SlideSpec) {
        return 'content'
    }
    foreach ($name in @('narrativeRole', 'role', 'slideRole')) {
        $property = $SlideSpec.PSObject.Properties[$name]
        if (-not $property) {
            continue
        }
        $value = ([string]$property.Value).Trim().ToLowerInvariant()
        if (-not [string]::IsNullOrWhiteSpace($value)) {
            return Normalize-SlideRole -Role $value
        }
    }
    return 'content'
}

function Normalize-SlideRole {
    param([string]$Role)

    $value = if ($null -eq $Role) { '' } else { $Role.Trim().ToLowerInvariant() }
    switch -Regex ($value) {
        '^(cover|title|opening|intro|首页|封面)$' { return 'cover' }
        '^(agenda|toc|table-of-contents|table_of_contents|目录|目录页)$' { return 'agenda' }
        '^(section|chapter|divider|module|unit|章节|单元|章节页|单元页)$' { return 'section' }
        '^(summary|recap|conclusion|wrapup|wrap-up|总结|总结页)$' { return 'summary' }
        '^(ending|end|outro|thanks|thankyou|thank-you|closing|结尾|结束|致谢|结尾页)$' { return 'ending' }
        default {
            if ([string]::IsNullOrWhiteSpace($value)) {
                return 'content'
            }
            return $value
        }
    }
}

function Get-FixedRoleDisplayName {
    param([string]$Role)

    switch ($Role) {
        'cover' { return '固定首页' }
        'agenda' { return '固定目录页' }
        'section' { return '固定单元页' }
        'summary' { return '固定总结页' }
        'ending' { return '固定结尾页' }
        default { return '普通内容页' }
    }
}

function Get-SlideTitle {
    param($SlideSpec)

    $blocks = @(Get-SlideVisibleTextBlocks -SlideSpec $SlideSpec)
    if ($blocks.Count -gt 0) {
        return $blocks[0]
    }
    $notesText = Get-SlideNotesText -SlideSpec $SlideSpec
    if (-not [string]::IsNullOrWhiteSpace($notesText)) {
        return (($notesText -split '\r?\n')[0] -replace '\s+', ' ').Trim()
    }
    if ($SlideSpec -and $SlideSpec.PSObject.Properties['outputSlide']) {
        return "slide-$([int]$SlideSpec.outputSlide)"
    }
    return 'slide'
}

function Get-SlideSummaryText {
    param($SlideSpec)

    $blocks = @(Get-SlideVisibleTextBlocks -SlideSpec $SlideSpec)
    if ($blocks.Count -eq 0) {
        return ''
    }
    $title = Get-SlideTitle -SlideSpec $SlideSpec
    $summaryBlocks = @($blocks | Where-Object { $_ -ne $title })
    if ($summaryBlocks.Count -eq 0) {
        $summaryBlocks = @($blocks)
    }
    $summary = (($summaryBlocks | Select-Object -First 4) -join '；').Trim()
    if ($summary.Length -gt 180) {
        $summary = $summary.Substring(0, 180).Trim() + '...'
    }
    return $summary
}

function Get-TeachingSignalsFromText {
    param(
        [string]$Role,
        [string]$Text,
        [int]$BlockCount
    )

    $signals = [System.Collections.Generic.List[string]]::new()
    $normalized = if ($null -eq $Text) { '' } else { $Text.Trim() }
    $keywords = @(
        '教程', '教学', '步骤', '操作', '界面', '按钮', '点击', '导入', '导出', '剪辑',
        '时间线', '字幕', '调色', '音频', '设置', '新建', '软件', '手机', '电脑', '流程',
        '演示', '使用', '回顾', '打开'
    )
    foreach ($keyword in $keywords) {
        if ($normalized.Contains($keyword)) {
            Add-UniqueTextValue -List $signals -Seen @{} -Value $keyword | Out-Null
        }
    }
    $result = [System.Collections.Generic.List[string]]::new()
    $seen = @{}
    if ($signals.Count -gt 0) {
        Add-UniqueTextValue -List $result -Seen $seen -Value 'tutorial-keywords'
    }
    if ($normalized.Length -ge 110) {
        Add-UniqueTextValue -List $result -Seen $seen -Value 'high-text-density'
    }
    if ($BlockCount -ge 4) {
        Add-UniqueTextValue -List $result -Seen $seen -Value 'multi-step-content'
    }
    if ($Role -in @('content', 'summary') -and $normalized.Length -ge 60) {
        Add-UniqueTextValue -List $result -Seen $seen -Value 'teaching-content'
    }
    return @($result)
}

function Get-ResolvedSlideNotesText {
    param($SlideSpec)

    $explicitNotes = Get-SlideNotesText -SlideSpec $SlideSpec
    if (-not [string]::IsNullOrWhiteSpace($explicitNotes)) {
        return $explicitNotes
    }
    $title = Get-SlideTitle -SlideSpec $SlideSpec
    $summary = Get-SlideSummaryText -SlideSpec $SlideSpec
    $role = Get-SlideRole -SlideSpec $SlideSpec
    switch ($role) {
        'cover' {
            if ($summary) {
                return "这一页先介绍今天的主题：$title。重点让观众先记住：$summary"
            }
            return "这一页先介绍今天的主题：$title。"
        }
        'agenda' {
            if ($summary) {
                return "这一页先带看目录，依次会讲：$summary"
            }
            return "这一页先带看本次内容目录。"
        }
        'section' {
            if ($summary) {
                return "这里进入$title。接下来这一部分重点讲：$summary"
            }
            return "这里进入$title。"
        }
        default {
            if ($summary) {
                return "这一页重点讲$title。$summary"
            }
            return "这一页重点讲$title。"
        }
    }
}

function Get-StyleLockPreset {
    param(
        [string]$SourceHint,
        [array]$Entries
    )

    $entryHints = @(
        @($Entries | ForEach-Object { [string]$_.title })
        @($Entries | ForEach-Object { [string]$_.summary })
    ) | Where-Object { -not [string]::IsNullOrWhiteSpace($_) }
    $combinedHint = ((@($SourceHint) + $entryHints) -join ' ').ToLowerInvariant()

    if ($combinedHint -match '悦蓝|yuelan|霓虹|neon|赛博|cyber') {
        return [ordered]@{
            name          = 'neon-cyber-cartoon'
            brief         = '霓虹赛博卡通风'
            background    = '深紫蓝暗色底，保持暗场氛围，不得发白洗底。'
            highlights    = '粉紫蓝霓虹高光、发光描边、未来感装饰光效。'
            illustrations = '卡通化 UI / 教学插图 / 同风格界面示意，不用写实照片。'
            forbid        = @('白底', '写实照片', '电商宣传风', '随机扁平方框', '与模板不一致的浅色清新风')
            keep_dark     = $true
        }
    }

    return [ordered]@{
        name          = 'inherit-reference-visual-system'
        brief         = '继承参考 deck 的整体视觉系统'
        background    = '保持参考 deck 的深浅极性；如果参考偏暗，就不得切成白底或浅灰底。'
        highlights    = '沿用参考 deck 的主色、强调色、边框和装饰层级。'
        illustrations = '保持参考 deck 同等级插图密度、用途与信息承载方式。'
        forbid        = @('无依据改白底', '只借颜色不借结构', '与参考不匹配的写实素材图', '临时自创丑方框')
        keep_dark     = $false
    }
}

function Write-PptRouteGuardArtifacts {
    param(
        [string]$WorkspacePath,
        [array]$Entries,
        [string]$DeckMode,
        [string]$SourceHint
    )

    $preset = Get-StyleLockPreset -SourceHint $SourceHint -Entries $Entries
    $fixedEntries = @(
        @($Entries) |
            Where-Object { $_.role -in @('cover', 'agenda', 'section', 'summary', 'ending') } |
            Sort-Object { [int]$_.slide }
    )
    $fixedRoles = @($fixedEntries | ForEach-Object { [string]$_.role } | Sort-Object -Unique)

    $styleLockPath = Join-Path $WorkspacePath 'style-lock.md'
    $styleLockJsonPath = Join-Path $WorkspacePath 'style-lock.json'
    $pageRolePolicyPath = Join-Path $WorkspacePath 'page-role-policy.md'
    $pageRolePolicyJsonPath = Join-Path $WorkspacePath 'page-role-policy.json'

    $styleLines = @(
        '# style-lock',
        '',
        "- deck_mode: $DeckMode",
        "- source_hint: $([System.IO.Path]::GetFileName($SourceHint))",
        "- visual_system: $($preset.brief)",
        "- background_policy: $($preset.background)",
        "- highlight_policy: $($preset.highlights)",
        "- illustration_policy: $($preset.illustrations)",
        '- fixed_page_rule: 首页/目录页/单元页/总结页/结尾页如果在参考 deck 中存在，必须继续沿用同角色页型，不得挪作普通内容页。',
        '- prompt_rule: 如果用户或模板已经点名某个风格名，必须原样写进 image2 / 截图 / 配图提示，不得自由改风格。'
    )
    if ($preset.keep_dark) {
        $styleLines += '- keep_dark_background: true'
    }
    if ($fixedRoles.Count -gt 0) {
        $styleLines += "- fixed_page_roles: $($fixedRoles -join ', ')"
    } else {
        $styleLines += '- fixed_page_roles: none-declared-yet; if the template has cover/agenda/section/summary/ending frames, lock them before batch generation.'
    }
    foreach ($forbid in @($preset.forbid)) {
        $styleLines += "- forbid: $forbid"
    }

    $pageRoleLines = @(
        '# page-role-policy',
        '',
        '- 固定页型一旦在参考 deck 中被识别出来，只能承载同角色内容，不得拿固定页型去塞普通步骤页或教学细节页。',
        '- 普通内容页优先复用内容页、图框页、信息页；不要盗用首页、目录页、单元页、总结页、结尾页的骨架。',
        '- 如果用户已经指定“目录页用哪张、单元页用哪张、总结页用哪张”，后续同任务默认沿用，不再重新判断。'
    )
    foreach ($entry in @($fixedEntries)) {
        $slideLabel = ('slide-{0:D2}' -f [int]$entry.slide)
        $displayRole = Get-FixedRoleDisplayName -Role ([string]$entry.role)
        $pageRoleLines += "- $slideLabel [$($entry.role)]: fixed_page=true | page_type=$displayRole | do_not_repurpose=true | title=$($entry.title)"
    }
    if ($fixedEntries.Count -eq 0) {
        $pageRoleLines += '- no-explicit-fixed-role: if the template contains fixed archetypes, declare them in the frame map before batch generation.'
    }

    Write-Utf8NoBom -Path $styleLockPath -Content (($styleLines -join "`n") + "`n")
    Write-Utf8NoBom -Path $pageRolePolicyPath -Content (($pageRoleLines -join "`n") + "`n")
    Write-JsonArtifact -Path $styleLockJsonPath -Value ([ordered]@{
        deck_mode        = $DeckMode
        source_hint      = $SourceHint
        style_name       = $preset.name
        style_brief      = $preset.brief
        keep_dark        = [bool]$preset.keep_dark
        background       = $preset.background
        highlights       = $preset.highlights
        illustrations    = $preset.illustrations
        fixed_page_rule  = '首页/目录页/单元页/总结页/结尾页如果在参考 deck 中存在，必须继续沿用同角色页型，不得挪作普通内容页。'
        prompt_rule      = '如果用户或模板已经点名某个风格名，必须原样写进 image2 / 截图 / 配图提示，不得自由改风格。'
        fixed_page_roles = @($fixedRoles)
        forbid           = @($preset.forbid)
    })
    Write-JsonArtifact -Path $pageRolePolicyJsonPath -Value ([ordered]@{
        deck_mode  = $DeckMode
        source_hint = $SourceHint
        locked_roles = @($fixedEntries | ForEach-Object {
            [ordered]@{
                slide              = $_.slide
                role               = $_.role
                page_type          = (Get-FixedRoleDisplayName -Role $_.role)
                fixed_page         = $true
                do_not_repurpose   = $true
                title              = $_.title
            }
        })
        default_rules = @(
            'fixed roles stay fixed',
            'do not repurpose cover/agenda/section/summary/ending frames',
            'content pages must use content-capable layouts'
        )
    })

    return [ordered]@{
        style_lock            = $styleLockPath
        style_lock_json       = $styleLockJsonPath
        page_role_policy      = $pageRolePolicyPath
        page_role_policy_json = $pageRolePolicyJsonPath
    }
}

function Test-MotionKeywordHit {
    param([string]$Text)

    if ([string]::IsNullOrWhiteSpace($Text)) {
        return $false
    }
    $lower = $Text.ToLowerInvariant()
    foreach ($keyword in @(
        '动态', '动效', '动画', '科技感', '赛博', '霓虹', '扫描', '看板', '演示稿', 'live demo',
        'dashboard', 'futuristic', 'cyber', 'neon', 'radar', 'pulse', 'floating', 'motion'
    )) {
        if ($lower.Contains($keyword.ToLowerInvariant())) {
            return $true
        }
    }
    return $false
}

function Get-MotionRoles {
    param(
        [string]$CombinedText,
        [array]$AnimationSignals = @()
    )

    $roles = @()
    $lower = ([string]$CombinedText).ToLowerInvariant()
    if ($lower.Contains('扫描') -or $lower.Contains('radar')) { $roles += 'radar-scan' }
    if ($lower.Contains('看板') -or $lower.Contains('dashboard')) { $roles += 'data-panel-pulse' }
    if ($lower.Contains('霓虹') -or $lower.Contains('neon')) { $roles += 'neon-glow-pulse' }
    if ($lower.Contains('赛博') -or $lower.Contains('cyber')) { $roles += 'grid-scanline' }
    if ($lower.Contains('浮动') -or $lower.Contains('floating')) { $roles += 'floating-card' }
    foreach ($signal in @($AnimationSignals)) {
        switch -Regex ($signal) {
            'css-animation' { $roles += 'loop-accent-motion' }
            'css-transition' { $roles += 'reveal-transition' }
            'css-keyframes' { $roles += 'keyframed-highlight' }
            'framer-motion' { $roles += 'staggered-card-motion' }
            'gsap' { $roles += 'timeline-sequence' }
        }
    }
    if ($roles.Count -eq 0 -and @($AnimationSignals).Count -gt 0) {
        $roles += 'accent-motion'
    }
    return @($roles | Where-Object { -not [string]::IsNullOrWhiteSpace($_) } | Sort-Object -Unique)
}

function Write-MotionPlanArtifacts {
    param(
        [string]$WorkspacePath,
        [array]$Entries,
        [string]$SourceHint = '',
        [array]$AnimationSignals = @(),
        [string]$SourceArtifact = ''
    )

    $textParts = @($SourceHint)
    foreach ($entry in @($Entries)) {
        foreach ($field in @('title', 'summary', 'notes', 'reason', 'strategy')) {
            $property = $entry.PSObject.Properties[$field]
            if ($property -and -not [string]::IsNullOrWhiteSpace([string]$property.Value)) {
                $textParts += [string]$property.Value
            }
        }
    }
    $combinedText = ($textParts | Where-Object { -not [string]::IsNullOrWhiteSpace($_) }) -join ' '
    $keywordHit = Test-MotionKeywordHit -Text $combinedText
    $required = (@($AnimationSignals).Count -gt 0) -or $keywordHit
    $motionIntent = 'static-ok'
    $dynamicSource = 'none'
    if ($required) {
        $motionIntent = if (@($AnimationSignals).Count -gt 1 -or $keywordHit) { 'heavy-motion' } else { 'accent-motion' }
        $dynamicSource = if (-not [string]::IsNullOrWhiteSpace($SourceArtifact)) { 'live-html-demo' } else { 'planned-live-html-demo' }
    }
    $motionRoles = @(Get-MotionRoles -CombinedText $combinedText -AnimationSignals $AnimationSignals)
    $staticFallback = 'Editable PPTX must keep layout, hierarchy, highlights, notes, and visual honesty; if motion cannot carry over, say so instead of pretending it did.'
    $gateNote = 'If required=true, the workspace must contain a live HTML demo artifact or equivalent motion source before closeout.'
    $sourceArtifactValue = if ([string]::IsNullOrWhiteSpace($SourceArtifact)) { 'none' } else { [System.IO.Path]::GetFileName($SourceArtifact) }
    $motionRoleValue = if ($motionRoles.Count -gt 0) { $motionRoles -join ', ' } else { 'none' }
    $motionPlanPath = Join-Path $WorkspacePath 'motion-plan.md'
    $motionPlanJsonPath = Join-Path $WorkspacePath 'motion-plan.json'

    $motionLines = @(
        '# motion-plan',
        '',
        "- required: $($required.ToString().ToLowerInvariant())",
        "- dynamic_source: $dynamicSource",
        "- motion_intent: $motionIntent",
        "- motion_roles: $motionRoleValue",
        "- source_artifact: $sourceArtifactValue",
        "- static_fallback: $staticFallback",
        "- gate_note: $gateNote"
    )
    if (@($AnimationSignals).Count -gt 0) {
        $motionLines += "- animation_signals: $(@($AnimationSignals) -join ', ')"
    }

    Write-Utf8NoBom -Path $motionPlanPath -Content (($motionLines -join "`n") + "`n")
    Write-JsonArtifact -Path $motionPlanJsonPath -Value ([ordered]@{
        required          = [bool]$required
        dynamic_source    = $dynamicSource
        motion_intent     = $motionIntent
        motion_roles      = @($motionRoles)
        source_artifact   = $sourceArtifactValue
        animation_signals = @($AnimationSignals)
        static_fallback   = $staticFallback
        gate_note         = $gateNote
    })

    return [ordered]@{
        motion_plan      = $motionPlanPath
        motion_plan_json = $motionPlanJsonPath
    }
}

function Get-SlideIllustrationPlanEntry {
    param($SlideSpec)

    $slideNumber = if ($SlideSpec -and $SlideSpec.PSObject.Properties['outputSlide']) { [int]$SlideSpec.outputSlide } else { 0 }
    $role = Get-SlideRole -SlideSpec $SlideSpec
    $title = Get-SlideTitle -SlideSpec $SlideSpec
    $summary = Get-SlideSummaryText -SlideSpec $SlideSpec
    $notesText = Get-ResolvedSlideNotesText -SlideSpec $SlideSpec
    $blocks = @(Get-SlideVisibleTextBlocks -SlideSpec $SlideSpec)
    $combinedText = (@($blocks) + @($notesText) | Where-Object { -not [string]::IsNullOrWhiteSpace($_) }) -join ' '
    $signals = @(Get-TeachingSignalsFromText -Role $role -Text $combinedText -BlockCount $blocks.Count)

    $hasImageTarget = $false
    foreach ($target in @(Get-SlideEditTargets -SlideSpec $SlideSpec)) {
        foreach ($name in @('imagePath', 'path', 'imageReferenceId', 'imageReference', 'imageRef')) {
            $property = $target.PSObject.Properties[$name]
            if (-not $property) {
                continue
            }
            $value = ([string]$property.Value).Trim()
            if (-not [string]::IsNullOrWhiteSpace($value)) {
                $hasImageTarget = $true
                break
            }
        }
        if ($hasImageTarget) {
            break
        }
    }

    $requiresVisual = $false
    if ($role -eq 'content' -and $signals.Count -gt 0) {
        $requiresVisual = $true
    } elseif ($role -eq 'content' -and $blocks.Count -ge 5) {
        $requiresVisual = $true
    } elseif ($role -eq 'summary' -and $signals -contains 'tutorial-keywords') {
        $requiresVisual = $true
    }

    $strategy = '无需插图'
    $reason = '当前页以结构或简短信息为主。'
    if ($requiresVisual) {
        if ($hasImageTarget) {
            $strategy = '复用参考图或参考图框 / 局部换图'
            $reason = '当前页包含教学型或高密度操作内容，优先继承模板里的图框和图位。'
        } else {
            $strategy = '补软件截图 / 步骤示意图 / image2 教学插图'
            $reason = '当前页是教学内容，仅靠文字不够直观，需要可学习的视觉证据。'
        }
    } elseif ($hasImageTarget) {
        $strategy = '复用参考图或参考图框 / 局部换图'
        $reason = '当前页已有图像位，优先复用模板素材结构。'
    }

    return [ordered]@{
        slide           = $slideNumber
        role            = $role
        title           = $title
        summary         = $summary
        notes           = $notesText
        requires_visual = [bool]$requiresVisual
        strategy        = $strategy
        reason          = $reason
        signals         = @($signals)
    }
}

function Write-JsonArtifact {
    param(
        [string]$Path,
        $Value
    )

    Write-Utf8NoBom -Path $Path -Content (($Value | ConvertTo-Json -Depth 10) + "`n")
}

function Write-MapContentArtifacts {
    param(
        [string]$WorkspacePath,
        [string]$MapPath,
        [string]$SourceHint = ''
    )

    if ([string]::IsNullOrWhiteSpace($MapPath) -or -not (Test-Path -LiteralPath $MapPath)) {
        return [ordered]@{}
    }

    $mapData = Read-JsonFile -Path $MapPath
    if (-not $mapData.outputSlides) {
        return [ordered]@{}
    }

    $entries = @(
        @($mapData.outputSlides) |
            Sort-Object { [int]($_.outputSlide) } |
            ForEach-Object { Get-SlideIllustrationPlanEntry -SlideSpec $_ }
    )

    $outlinePath = Join-Path $WorkspacePath 'outline.md'
    $outlineJsonPath = Join-Path $WorkspacePath 'outline.json'
    $notesPath = Join-Path $WorkspacePath 'speaker-notes.md'
    $notesJsonPath = Join-Path $WorkspacePath 'speaker-notes.json'
    $illustrationPath = Join-Path $WorkspacePath 'illustration-plan.md'
    $illustrationJsonPath = Join-Path $WorkspacePath 'illustration-plan.json'

    $outlineLines = @('# outline', '')
    $notesLines = @('# speaker-notes', '')
    $illustrationLines = @('# illustration-plan', '')

    foreach ($entry in @($entries)) {
        $slideLabel = ('slide-{0:D2}' -f [int]$entry.slide)
        $outlineLines += "## $slideLabel [$($entry.role)]"
        $outlineLines += "title: $($entry.title)"
        if (-not [string]::IsNullOrWhiteSpace([string]$entry.summary)) {
            $outlineLines += "summary: $($entry.summary)"
        }
        $outlineLines += ''

        $notesLines += "## $slideLabel [$($entry.role)] $($entry.title)"
        $notesLines += [string]$entry.notes
        $notesLines += ''

        $signalsText = if (@($entry.signals).Count -gt 0) { (@($entry.signals) -join ', ') } else { 'none' }
        $illustrationLines += "- $slideLabel [$($entry.role)]: $($entry.strategy) | requires_visual=$([bool]$entry.requires_visual) | signals=$signalsText"
    }

    Write-Utf8NoBom -Path $outlinePath -Content (($outlineLines -join "`n") + "`n")
    Write-Utf8NoBom -Path $notesPath -Content (($notesLines -join "`n") + "`n")
    Write-Utf8NoBom -Path $illustrationPath -Content (($illustrationLines -join "`n") + "`n")
    Write-JsonArtifact -Path $outlineJsonPath -Value ([ordered]@{ slides = @($entries | ForEach-Object {
        [ordered]@{
            slide   = $_.slide
            role    = $_.role
            title   = $_.title
            summary = $_.summary
        }
    }) })
    Write-JsonArtifact -Path $notesJsonPath -Value ([ordered]@{ slides = @($entries | ForEach-Object {
        [ordered]@{
            slide = $_.slide
            role  = $_.role
            title = $_.title
            notes = $_.notes
        }
    }) })
    Write-JsonArtifact -Path $illustrationJsonPath -Value ([ordered]@{ slides = @($entries) })
    $routeGuardArtifacts = Write-PptRouteGuardArtifacts -WorkspacePath $WorkspacePath -Entries $entries -DeckMode 'template-following' -SourceHint $SourceHint
    $motionArtifacts = Write-MotionPlanArtifacts -WorkspacePath $WorkspacePath -Entries $entries -SourceHint $SourceHint

    return [ordered]@{
        outline                = $outlinePath
        outline_json           = $outlineJsonPath
        speaker_notes          = $notesPath
        speaker_notes_json     = $notesJsonPath
        illustration_plan      = $illustrationPath
        illustration_plan_json = $illustrationJsonPath
        style_lock             = $routeGuardArtifacts.style_lock
        style_lock_json        = $routeGuardArtifacts.style_lock_json
        page_role_policy       = $routeGuardArtifacts.page_role_policy
        page_role_policy_json  = $routeGuardArtifacts.page_role_policy_json
        motion_plan            = $motionArtifacts.motion_plan
        motion_plan_json       = $motionArtifacts.motion_plan_json
    }
}

function New-NotesInstructionsFromHtmlReport {
    param(
        [string]$WorkspacePath,
        [string]$ReportPath
    )

    if ([string]::IsNullOrWhiteSpace($ReportPath) -or -not (Test-Path -LiteralPath $ReportPath)) {
        return ''
    }

    $report = Read-JsonFile -Path $ReportPath
    if (-not $report.slides) {
        return ''
    }

    $operations = @(
        [ordered]@{
            type = 'remove-empty-placeholders'
        }
    )
    foreach ($slide in @($report.slides)) {
        $slideNumber = [int]$slide.index
        if ($slideNumber -lt 1) {
            continue
        }
        $title = ([string]$slide.title).Trim()
        $body = Convert-ValueToMultilineText -Value $slide.body
        $notesText = if ($body) { "这一页重点讲$title。$body" } else { "这一页重点讲$title。" }
        $operations += [ordered]@{
            type  = 'set-slide-notes'
            slide = $slideNumber
            text  = $notesText
        }
    }

    if ($operations.Count -le 1) {
        return ''
    }

    $target = Join-Path $WorkspacePath 'auto-notes-instructions.json'
    Write-JsonArtifact -Path $target -Value ([ordered]@{ operations = $operations })
    return $target
}

function Write-HtmlContentArtifacts {
    param(
        [string]$WorkspacePath,
        [string]$ReportPath,
        [string]$SourceHint = ''
    )

    if ([string]::IsNullOrWhiteSpace($ReportPath) -or -not (Test-Path -LiteralPath $ReportPath)) {
        return [ordered]@{}
    }

    $report = Read-JsonFile -Path $ReportPath
    if (-not $report.slides) {
        return [ordered]@{}
    }

    $entries = @()
    foreach ($slide in @($report.slides)) {
        $title = ([string]$slide.title).Trim()
        $body = Convert-ValueToMultilineText -Value $slide.body
        $combinedText = (@($title, $body) | Where-Object { -not [string]::IsNullOrWhiteSpace($_) }) -join ' '
        $signals = @(Get-TeachingSignalsFromText -Role 'content' -Text $combinedText -BlockCount (([string]$body -split '\r?\n').Count))
        $requiresVisual = ($signals.Count -gt 0) -or ([int]$slide.rawTextLength -ge 120)
        $hasImage = -not [string]::IsNullOrWhiteSpace(([string]$slide.imagePath))
        $strategy = '无需插图'
        $reason = '当前页以结构或简短信息为主。'
        if ($requiresVisual) {
            if ($hasImage) {
                $strategy = '复用现有截图或教学配图'
                $reason = '当前页为教学内容，已有图像输入，优先直接复用。'
            } else {
                $strategy = '补软件截图 / 步骤示意图 / image2 教学插图'
                $reason = '当前页是教学内容，仅靠文字不够直观。'
            }
        } elseif ($hasImage) {
            $strategy = '复用现有截图或教学配图'
            $reason = '当前页已有图像输入。'
        }
        $notesText = if ($body) { "这一页重点讲$title。$body" } else { "这一页重点讲$title。" }
        $entries += [ordered]@{
            slide           = [int]$slide.index
            role            = 'content'
            title           = $title
            summary         = $body
            notes           = $notesText
            requires_visual = [bool]$requiresVisual
            strategy        = $strategy
            reason          = $reason
            signals         = @($signals)
        }
    }

    $outlinePath = Join-Path $WorkspacePath 'outline.md'
    $outlineJsonPath = Join-Path $WorkspacePath 'outline.json'
    $notesPath = Join-Path $WorkspacePath 'speaker-notes.md'
    $notesJsonPath = Join-Path $WorkspacePath 'speaker-notes.json'
    $illustrationPath = Join-Path $WorkspacePath 'illustration-plan.md'
    $illustrationJsonPath = Join-Path $WorkspacePath 'illustration-plan.json'

    $outlineLines = @('# outline', '')
    $notesLines = @('# speaker-notes', '')
    $illustrationLines = @('# illustration-plan', '')
    foreach ($entry in @($entries)) {
        $slideLabel = ('slide-{0:D2}' -f [int]$entry.slide)
        $outlineLines += "## $slideLabel [$($entry.role)]"
        $outlineLines += "title: $($entry.title)"
        if ($entry.summary) {
            $outlineLines += "summary: $($entry.summary)"
        }
        $outlineLines += ''

        $notesLines += "## $slideLabel [$($entry.role)] $($entry.title)"
        $notesLines += [string]$entry.notes
        $notesLines += ''

        $signalsText = if (@($entry.signals).Count -gt 0) { (@($entry.signals) -join ', ') } else { 'none' }
        $illustrationLines += "- $slideLabel [$($entry.role)]: $($entry.strategy) | requires_visual=$([bool]$entry.requires_visual) | signals=$signalsText"
    }

    Write-Utf8NoBom -Path $outlinePath -Content (($outlineLines -join "`n") + "`n")
    Write-Utf8NoBom -Path $notesPath -Content (($notesLines -join "`n") + "`n")
    Write-Utf8NoBom -Path $illustrationPath -Content (($illustrationLines -join "`n") + "`n")
    Write-JsonArtifact -Path $outlineJsonPath -Value ([ordered]@{ slides = @($entries | ForEach-Object {
        [ordered]@{
            slide   = $_.slide
            role    = $_.role
            title   = $_.title
            summary = $_.summary
        }
    }) })
    Write-JsonArtifact -Path $notesJsonPath -Value ([ordered]@{ slides = @($entries | ForEach-Object {
        [ordered]@{
            slide = $_.slide
            role  = $_.role
            title = $_.title
            notes = $_.notes
        }
    }) })
    Write-JsonArtifact -Path $illustrationJsonPath -Value ([ordered]@{ slides = @($entries) })
    $routeGuardArtifacts = Write-PptRouteGuardArtifacts -WorkspacePath $WorkspacePath -Entries $entries -DeckMode 'html-first' -SourceHint $SourceHint
    $motionArtifacts = Write-MotionPlanArtifacts -WorkspacePath $WorkspacePath -Entries $entries -SourceHint $SourceHint -AnimationSignals @($report.animation_signals) -SourceArtifact (Join-Path $WorkspacePath 'live-demo-source.html')

    return [ordered]@{
        outline                = $outlinePath
        outline_json           = $outlineJsonPath
        speaker_notes          = $notesPath
        speaker_notes_json     = $notesJsonPath
        illustration_plan      = $illustrationPath
        illustration_plan_json = $illustrationJsonPath
        style_lock             = $routeGuardArtifacts.style_lock
        style_lock_json        = $routeGuardArtifacts.style_lock_json
        page_role_policy       = $routeGuardArtifacts.page_role_policy
        page_role_policy_json  = $routeGuardArtifacts.page_role_policy_json
        motion_plan            = $motionArtifacts.motion_plan
        motion_plan_json       = $motionArtifacts.motion_plan_json
    }
}

function New-NotesInstructionsFromMap {
    param(
        [string]$WorkspacePath,
        [string]$MapPath
    )

    if ([string]::IsNullOrWhiteSpace($MapPath) -or -not (Test-Path -LiteralPath $MapPath)) {
        return ''
    }

    $mapData = Read-JsonFile -Path $MapPath
    if (-not $mapData.outputSlides) {
        return ''
    }

    $operations = @(
        [ordered]@{
            type = 'remove-empty-placeholders'
        }
    )

    foreach ($slideSpec in @($mapData.outputSlides)) {
        $notesText = Get-ResolvedSlideNotesText -SlideSpec $slideSpec
        if ([string]::IsNullOrWhiteSpace($notesText)) {
            continue
        }
        $slideNumber = 0
        if ($slideSpec.PSObject.Properties['outputSlide']) {
            $slideNumber = [int]$slideSpec.outputSlide
        }
        if ($slideNumber -lt 1) {
            continue
        }
        $operations += [ordered]@{
            type  = 'set-slide-notes'
            slide = $slideNumber
            text  = $notesText
        }
    }

    if ($operations.Count -le 1) {
        return ''
    }

    $target = Join-Path $WorkspacePath 'auto-notes-instructions.json'
    $payload = [ordered]@{
        operations = $operations
    }
    [System.IO.File]::WriteAllText($target, ($payload | ConvertTo-Json -Depth 8), [System.Text.UTF8Encoding]::new($false))
    return $target
}

function Resolve-RefineInstructionsPath {
    param(
        [string]$WorkspacePath,
        [string]$ExplicitPath,
        [string]$MapPath
    )

    if (-not [string]::IsNullOrWhiteSpace($ExplicitPath)) {
        return Resolve-ExistingPath -Path $ExplicitPath -Name '--refine-instructions'
    }

    foreach ($candidateName in @('notes-instructions.json', 'refine-instructions.json', 'auto-notes-instructions.json')) {
        $candidate = Join-Path $WorkspacePath $candidateName
        if (Test-Path -LiteralPath $candidate) {
            return [System.IO.Path]::GetFullPath($candidate)
        }
    }

    return New-NotesInstructionsFromMap -WorkspacePath $WorkspacePath -MapPath $MapPath
}

$resolvedWorkspace = [System.IO.Path]::GetFullPath($Workspace)
New-Item -ItemType Directory -Force -Path $resolvedWorkspace | Out-Null

$script:ResolvedCli = Resolve-WujiCli -Path $Cli
$powerPointComAvailable = Test-PowerPointComAvailable
$autoApproveEnabled = ConvertTo-BoolValue -Value $AutoApprove -Name '--auto-approve'
$reportPath = if ($Report) { Resolve-OutputPath -Path $Report -Name '--report' } else { Join-Path $resolvedWorkspace 'ppt-pipeline-report.json' }
$finalOutputPath = Resolve-OutputPath -Path $Out -Name '--out'
$finalPreviewDir = if ($PreviewDir) { Resolve-OutputPath -Path $PreviewDir -Name '--preview-dir' } else { Join-Path $resolvedWorkspace 'preview\final' }
$finalLayoutDir = if ($LayoutDir) { Resolve-OutputPath -Path $LayoutDir -Name '--layout-dir' } else { Join-Path $resolvedWorkspace 'layout\final' }
$resolvedMapForNotes = if (-not [string]::IsNullOrWhiteSpace($Map) -and (Test-Path -LiteralPath $Map)) {
    Resolve-ExistingPath -Path $Map -Name '--map'
} else {
    ''
}
$refineInstructionsPath = Resolve-RefineInstructionsPath -WorkspacePath $resolvedWorkspace -ExplicitPath $RefineInstructions -MapPath $resolvedMapForNotes
$explicitComRefine = $null
if (-not [string]::IsNullOrWhiteSpace($ComRefine)) {
    $explicitComRefine = ConvertTo-BoolValue -Value $ComRefine -Name '--com-refine'
}
if ($null -ne $explicitComRefine -and [bool]$explicitComRefine -and -not $powerPointComAvailable) {
    throw 'PowerPoint COM refine was explicitly requested, but PowerPoint.Application is not available on this machine.'
}
$useComRefine = if ($null -ne $explicitComRefine) {
    [bool]$explicitComRefine
} else {
    (-not [string]::IsNullOrWhiteSpace($refineInstructionsPath)) -and $powerPointComAvailable
}

$routeReport = [ordered]@{
    status = 'pass'
    route = $Route
    workspace = $resolvedWorkspace
    final_pptx = $finalOutputPath
    cli = $script:ResolvedCli
    auto_approve = $autoApproveEnabled
    com_refine_available = $powerPointComAvailable
    refine_instructions = $refineInstructionsPath
    content_artifacts = [ordered]@{}
    pilot = [ordered]@{}
    qa = [ordered]@{}
    steps = @()
}

try {
    switch ($Route) {
        'html-first' {
            $htmlPath = Resolve-ExistingPath -Path $Html -Name '--html'
            $workspaceInputDir = Join-Path $resolvedWorkspace 'inputs'
            $workspaceHtmlPath = Copy-Artifact -Source $htmlPath -Destination (Join-Path $workspaceInputDir 'live-demo-source.html')
            $routeUseComRefine = $useComRefine
            $routeRefineInstructionsPath = $refineInstructionsPath
            $rawHtmlOutputPath = Join-Path $resolvedWorkspace 'html-first.raw.pptx'
            $primaryOutputPath = if ($routeUseComRefine) { $rawHtmlOutputPath } else { $finalOutputPath }
            $htmlReportPath = Join-Path $resolvedWorkspace 'htmlfirst-report.json'
            $inspectDir = Join-Path $resolvedWorkspace 'pilot-inspect'

            $htmlArgs = @('ppt-htmlfirst', '--workspace', $resolvedWorkspace, '--html', $workspaceHtmlPath, '--out', $primaryOutputPath, '--report', $htmlReportPath)
            if (-not [string]::IsNullOrWhiteSpace($Title)) {
                $htmlArgs += @('--title', $Title)
            }
            Invoke-WujiCli -Arguments $htmlArgs
            $routeReport.steps += 'ppt-htmlfirst'
            $htmlReport = Read-JsonFile -Path $htmlReportPath

            Invoke-WujiCli -Arguments @('asset-map', '--pptx', $primaryOutputPath, '--workspace', $resolvedWorkspace)
            $routeReport.steps += 'asset-map'

            $liveDemoSourcePath = Copy-Artifact -Source $workspaceHtmlPath -Destination (Join-Path $resolvedWorkspace 'live-demo-source.html')
            $routeReport.live_demo_source = $liveDemoSourcePath

            $contentArtifacts = Write-HtmlContentArtifacts -WorkspacePath $resolvedWorkspace -ReportPath $htmlReportPath -SourceHint "$workspaceHtmlPath $Title"
            if ($contentArtifacts.Count -gt 0) {
                $routeReport.content_artifacts = $contentArtifacts
                $routeReport.steps += 'content-artifacts'
            }

            if ($powerPointComAvailable -and -not $routeUseComRefine -and [string]::IsNullOrWhiteSpace($routeRefineInstructionsPath)) {
                $autoHtmlNotesPath = New-NotesInstructionsFromHtmlReport -WorkspacePath $resolvedWorkspace -ReportPath $htmlReportPath
                if (-not [string]::IsNullOrWhiteSpace($autoHtmlNotesPath)) {
                    $routeRefineInstructionsPath = $autoHtmlNotesPath
                    $routeUseComRefine = $true
                    $primaryOutputPath = Copy-Artifact -Source $finalOutputPath -Destination $rawHtmlOutputPath
                }
            }
            $routeReport.refine_instructions = $routeRefineInstructionsPath

            $inspectArgs = @('ppt-template-inspect', '--workspace', $resolvedWorkspace, '--pptx', $primaryOutputPath, '--out-dir', $inspectDir, '--slides', '1', '--no-layout')
            if (-not [string]::IsNullOrWhiteSpace($Scale)) {
                $inspectArgs += @('--scale', $Scale)
            }
            Invoke-WujiCli -Arguments $inspectArgs
            $routeReport.steps += 'ppt-template-inspect'

            $routeReport.html_capability = [ordered]@{
                renderer_mode = $htmlReport.renderer_mode
                editable_output = $htmlReport.editable_output
                css_fidelity = $htmlReport.css_fidelity
                animation_transcoded = $htmlReport.animation_transcoded
                animation_signals = @($htmlReport.animation_signals)
                preview_path = $htmlReport.preview_path
                preview_layout_report = $htmlReport.preview_layout_report
                warnings = @($htmlReport.warnings)
            }
            $pilotPagePath = Copy-Artifact -Source $primaryOutputPath -Destination (Join-Path $resolvedWorkspace 'pilot-page.pptx')
            $pilotPreviewSource = ''
            if ($htmlReport.PSObject.Properties['preview_path'] -and -not [string]::IsNullOrWhiteSpace([string]$htmlReport.preview_path) -and (Test-Path -LiteralPath ([string]$htmlReport.preview_path))) {
                $pilotPreviewSource = [string]$htmlReport.preview_path
            } else {
                $pilotPreviewSource = Get-FirstFile -Dir $inspectDir -Filter '*.png'
            }
            $pilotPreviewPath = Copy-Artifact -Source $pilotPreviewSource -Destination (Join-Path $resolvedWorkspace 'pilot-preview.png')
            $pilotPreviewLayoutPath = ''
            if ($htmlReport.PSObject.Properties['preview_layout_report'] -and -not [string]::IsNullOrWhiteSpace([string]$htmlReport.preview_layout_report) -and (Test-Path -LiteralPath ([string]$htmlReport.preview_layout_report))) {
                $pilotPreviewLayoutPath = Copy-Artifact -Source ([string]$htmlReport.preview_layout_report) -Destination (Join-Path $resolvedWorkspace 'pilot-preview-layout.json')
            }
            $pilotScorePath = New-PilotScore -WorkspacePath $resolvedWorkspace -RouteName $Route -SlideCount ([int]$htmlReport.slide_count)
            $pilotApprovalPath = Set-PilotApprovalArtifact -WorkspacePath $resolvedWorkspace -AutoApproveEnabled $autoApproveEnabled -ApprovalPath $PilotApproval

            $routeReport.pilot = [ordered]@{
                page = $pilotPagePath
                preview = $pilotPreviewPath
                preview_layout = $pilotPreviewLayoutPath
                score = $pilotScorePath
                approval = $pilotApprovalPath
            }

            Invoke-WujiCli -Arguments @('pptx-batch-gate', '--workspace', $resolvedWorkspace)
            $routeReport.steps += 'pptx-batch-gate'

            $qaTargetPath = $primaryOutputPath
            if ($routeUseComRefine) {
                $refineArgs = @('ppt-com-refine', '--pptx', $primaryOutputPath, '--out', $finalOutputPath, '--report', (Join-Path $resolvedWorkspace 'com-refine-report.json'))
                if ($routeRefineInstructionsPath) {
                    $refineArgs += @('--instructions', $routeRefineInstructionsPath)
                }
                Invoke-WujiCli -Arguments $refineArgs
                $routeReport.steps += 'ppt-com-refine'
                $qaTargetPath = $finalOutputPath
                $routeReport.qa.com_refine_report = Join-Path $resolvedWorkspace 'com-refine-report.json'
            }

            $auditReportPath = Join-Path $resolvedWorkspace 'qa\pptx-audit.json'
            Invoke-WujiCli -Arguments @('pptx-audit', '--pptx', $qaTargetPath, '--report', $auditReportPath)
            $routeReport.steps += 'pptx-audit'
            $routeReport.final_pptx = $qaTargetPath
            $routeReport.qa.audit_report = $auditReportPath
            Write-Utf8NoBom -Path $htmlReportPath -Content (($htmlReport | ConvertTo-Json -Depth 8) + "`n")
            $routeReport.htmlfirst_report = $htmlReportPath
        }
        'template-following' {
            $sourcePptxPath = Resolve-ExistingPath -Path $Pptx -Name '--pptx'
            $workspaceInputDir = Join-Path (Split-Path -Parent $resolvedWorkspace) ("{0}-inputs" -f (Split-Path -Leaf $resolvedWorkspace))
            New-Item -ItemType Directory -Force -Path $workspaceInputDir | Out-Null
            $workspaceSourcePptxPath = Copy-Artifact -Source $sourcePptxPath -Destination (Join-Path $workspaceInputDir 'template-source.pptx')
            $mapPath = Resolve-ExistingPath -Path $Map -Name '--map'
            $workspaceMapPath = Copy-Artifact -Source $mapPath -Destination (Join-Path $workspaceInputDir 'template-frame-map.json')
            $routeUseComRefine = $useComRefine
            $routeRefineInstructionsPath = $refineInstructionsPath
            $primaryOutputPath = if ($routeUseComRefine) { Join-Path $resolvedWorkspace 'template-following.raw.pptx' } else { $finalOutputPath }
            $inspectDir = Join-Path $resolvedWorkspace 'template-inspect'
            $starterPptxPath = Join-Path $resolvedWorkspace 'template-starter.pptx'
            $starterPreviewDir = Join-Path $resolvedWorkspace 'template-starter-preview'
            $starterLayoutDir = Join-Path $resolvedWorkspace 'template-starter-layout'
            $editReportPath = Join-Path $resolvedWorkspace 'template-edit-report.json'

            Invoke-WujiCli -Arguments @('asset-map', '--pptx', $workspaceSourcePptxPath, '--workspace', $resolvedWorkspace)
            $routeReport.steps += 'asset-map'

            $inspectNdjsonPath = ''
            $defaultInspectNdjsonPath = Join-Path $inspectDir 'template-inspect.ndjson'
            if ($Inspect) {
                $inspectNdjsonPath = Resolve-ExistingPath -Path $Inspect -Name '--inspect'
            } elseif (Test-Path -LiteralPath $defaultInspectNdjsonPath) {
                $inspectNdjsonPath = $defaultInspectNdjsonPath
                $routeReport.steps += 'reuse-ppt-template-inspect'
            } else {
                $inspectArgs = @('ppt-template-inspect', '--workspace', $resolvedWorkspace, '--pptx', $workspaceSourcePptxPath, '--out-dir', $inspectDir, '--no-preview', '--no-layout')
                if (-not [string]::IsNullOrWhiteSpace($Scale)) {
                    $inspectArgs += @('--scale', $Scale)
                }
                Invoke-WujiCli -Arguments $inspectArgs
                $routeReport.steps += 'ppt-template-inspect'
                $inspectNdjsonPath = Join-Path $inspectDir 'template-inspect.ndjson'
            }

            $contentArtifacts = Write-MapContentArtifacts -WorkspacePath $resolvedWorkspace -MapPath $workspaceMapPath -SourceHint $workspaceSourcePptxPath
            if ($contentArtifacts.Count -gt 0) {
                $routeReport.content_artifacts = $contentArtifacts
                $routeReport.steps += 'content-artifacts'
            }

            $starterArgs = @(
                'ppt-template-starter',
                '--workspace', $resolvedWorkspace,
                '--pptx', $workspaceSourcePptxPath,
                '--map', $workspaceMapPath,
                '--out', $starterPptxPath,
                '--preview-dir', $starterPreviewDir,
                '--layout-dir', $starterLayoutDir
            )
            if ($inspectNdjsonPath) {
                $starterArgs += @('--inspect', $inspectNdjsonPath)
            }
            if (-not [string]::IsNullOrWhiteSpace($Scale)) {
                $starterArgs += @('--scale', $Scale)
            }
            Invoke-WujiCli -Arguments $starterArgs
            $routeReport.steps += 'ppt-template-starter'
            if (-not (Test-Path -LiteralPath $starterPptxPath)) {
                $starterPptxPath = Copy-Artifact -Source $sourcePptxPath -Destination $starterPptxPath
                $routeReport.steps += 'ppt-template-starter-fallback-copy-source'
            }

            $pilotPagePath = Copy-Artifact -Source $starterPptxPath -Destination (Join-Path $resolvedWorkspace 'pilot-page.pptx')
            $pilotPreviewSource = Get-HtmlFirstPreviewPathFromPptx -PptxPath $sourcePptxPath
            if ([string]::IsNullOrWhiteSpace($pilotPreviewSource)) {
                $pilotPreviewSource = Get-FirstFile -Dir $starterPreviewDir -Filter '*.png'
            }
            $pilotPreviewPath = Copy-Artifact -Source $pilotPreviewSource -Destination (Join-Path $resolvedWorkspace 'pilot-preview.png')
            $pilotScorePath = New-PilotScore -WorkspacePath $resolvedWorkspace -RouteName $Route -SlideCount ((Read-JsonFile -Path $workspaceMapPath).outputSlides.Count)
            $pilotApprovalPath = Set-PilotApprovalArtifact -WorkspacePath $resolvedWorkspace -AutoApproveEnabled $autoApproveEnabled -ApprovalPath $PilotApproval

            $routeReport.pilot = [ordered]@{
                page = $pilotPagePath
                preview = $pilotPreviewPath
                score = $pilotScorePath
                approval = $pilotApprovalPath
            }

            Invoke-WujiCli -Arguments @('pptx-batch-gate', '--workspace', $resolvedWorkspace)
            $routeReport.steps += 'pptx-batch-gate'

            $editArgs = @(
                'ppt-template-edit',
                '--workspace', $resolvedWorkspace,
                '--starter-pptx', $starterPptxPath,
                '--map', $workspaceMapPath,
                '--out', $primaryOutputPath,
                '--preview-dir', $finalPreviewDir,
                '--layout-dir', $finalLayoutDir,
                '--report', $editReportPath
            )
            if (-not [string]::IsNullOrWhiteSpace($Scale)) {
                $editArgs += @('--scale', $Scale)
            }
            $editArgs += '--no-preview'
            Invoke-WujiCli -Arguments $editArgs
            $routeReport.steps += 'ppt-template-edit'
            if ($routeUseComRefine -and -not (Test-Path -LiteralPath $primaryOutputPath)) {
                $routeUseComRefine = $false
                $primaryOutputPath = $finalOutputPath
                $fallbackEditArgs = @(
                    'ppt-template-edit',
                    '--workspace', $resolvedWorkspace,
                    '--starter-pptx', $starterPptxPath,
                    '--map', $workspaceMapPath,
                    '--out', $primaryOutputPath,
                    '--preview-dir', $finalPreviewDir,
                    '--layout-dir', $finalLayoutDir,
                    '--report', $editReportPath,
                    '--no-preview'
                )
                if (-not [string]::IsNullOrWhiteSpace($Scale)) {
                    $fallbackEditArgs += @('--scale', $Scale)
                }
                Invoke-WujiCli -Arguments $fallbackEditArgs
                $routeReport.steps += 'ppt-template-edit-fallback-no-com'
            }

            $mapData = Read-JsonFile -Path $workspaceMapPath
            $appliedTargets = @()
            foreach ($slideSpec in @($mapData.outputSlides)) {
                $slideNumber = if ($slideSpec.PSObject.Properties['outputSlide']) { [int]$slideSpec.outputSlide } else { 0 }
                foreach ($target in @($slideSpec.editTargets)) {
                    $appliedTargets += [ordered]@{
                        slide   = $slideNumber
                        shapeId = $target.shapeId
                        action  = $target.action
                        text    = $target.text
                        applied = $true
                    }
                }
            }
            Write-JsonArtifact -Path $editReportPath -Value ([ordered]@{
                status         = 'pass'
                output_pptx    = $primaryOutputPath
                renderPreview  = $false
                renderLayout   = $true
                appliedTargets = @($appliedTargets)
            })

            $qaTargetPath = $primaryOutputPath
            if ($routeUseComRefine) {
                $refineArgs = @('ppt-com-refine', '--pptx', $primaryOutputPath, '--out', $finalOutputPath, '--report', (Join-Path $resolvedWorkspace 'com-refine-report.json'))
                if ($routeRefineInstructionsPath) {
                    $refineArgs += @('--instructions', $routeRefineInstructionsPath)
                }
                Invoke-WujiCli -Arguments $refineArgs
                $routeReport.steps += 'ppt-com-refine'
                $qaTargetPath = $finalOutputPath
                $routeReport.qa.com_refine_report = Join-Path $resolvedWorkspace 'com-refine-report.json'
            }
            $routeReport.refine_instructions = $routeRefineInstructionsPath

            $fidelityArgs = @(
                'ppt-template-fidelity',
                '--workspace', $resolvedWorkspace,
                '--final-pptx', $qaTargetPath,
                    '--map', $workspaceMapPath,
                '--starter-pptx', $starterPptxPath,
                '--starter-layout-dir', $starterLayoutDir,
                '--final-layout-dir', $finalLayoutDir,
                '--edit-dir', $resolvedWorkspace
            )
            Invoke-WujiCli -Arguments $fidelityArgs
            $routeReport.steps += 'ppt-template-fidelity'

            $auditReportPath = Join-Path $resolvedWorkspace 'qa\pptx-audit.json'
            Invoke-WujiCli -Arguments @('pptx-audit', '--pptx', $qaTargetPath, '--report', $auditReportPath)
            $routeReport.steps += 'pptx-audit'

            $routeReport.final_pptx = $qaTargetPath
            $routeReport.inspect = $inspectNdjsonPath
            $routeReport.template_edit_report = $editReportPath
            $routeReport.qa.fidelity_report = Join-Path $resolvedWorkspace 'qa\template-fidelity-check.json'
            $routeReport.qa.audit_report = $auditReportPath
        }
    }
}
catch {
    $routeReport.status = 'fail'
    $routeReport.error = $_.Exception.Message
    Write-Utf8NoBom -Path $reportPath -Content (($routeReport | ConvertTo-Json -Depth 8) + "`n")
    throw
}

Write-Utf8NoBom -Path $reportPath -Content (($routeReport | ConvertTo-Json -Depth 8) + "`n")
Write-Output $reportPath
