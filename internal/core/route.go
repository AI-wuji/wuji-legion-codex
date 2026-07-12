package core

import (
	"sort"
	"strings"
)

func Route(query string, manifests []Manifest) RouteResult {
	q := strings.ToLower(strings.TrimSpace(query))
	selected := Manifest{ID: "core", Status: "primary", PrimarySkill: "wuji-legion-codex-2-0"}
	bestScore := 0
	bestPriority := -1
	for _, item := range manifests {
		score := scoreCapability(q, item)
		priority := domainPriority(item.ID)
		if score > bestScore || (score == bestScore && score > 0 && priority > bestPriority) {
			bestScore, bestPriority, selected = score, priority, item
		}
	}

	selectedEngine := selectEngine(q, selected.Engines)
	engine := selectedEngine.ID
	mounted := mountSources(q, selected, engine)
	secondary := secondaryCapabilities(q, selected.ID, manifests)
	officers := explicitOfficers(q)
	workers := workerPlan(q, selected, engine)
	parallel := len(workers) > 1
	provider, providerFallback := selectProvider(q, selected.Providers)
	primarySkill := selected.PrimarySkill
	if selectedEngine.PrimarySkill != "" {
		primarySkill = selectedEngine.PrimarySkill
	}
	if rank(selected.Status) < rank("callable") && selected.Fallback != "" {
		primarySkill = selected.Fallback
	}
	return RouteResult{
		Version:                 "2.0",
		Brain:                   "aji-general-staff",
		MainModel:               "gpt-5.6-sol",
		Reasoning:               "max",
		WriteAuthority:          "aji-only",
		Nuwa:                    false,
		Capability:              selected.ID,
		CapabilityStatus:        selected.Status,
		PrimarySkill:            primarySkill,
		Fallback:                selected.Fallback,
		Engine:                  engine,
		Provider:                provider,
		ProviderFallback:        providerFallback,
		SecondaryCapabilities:   secondary,
		MountedSources:          mounted,
		ExecutionLane:           "small-task-direct",
		Parallel:                parallel,
		Workers:                 workers,
		Officers:                officers,
		InternalAdversarialPass: len(officers) == 0 && needsInternalChallenge(q),
		FinishLine: []string{
			"requested active target changed in place",
			"selected capability behavior verified",
			"task-local verification passes",
			"do not claim fused unless capability_status is behavior-verified or primary",
		},
	}
}

func scoreCapability(query string, item Manifest) int {
	score := 0
	for _, trigger := range item.Triggers {
		if trigger == "" {
			continue
		}
		lower := strings.ToLower(trigger)
		if !strings.Contains(query, lower) {
			continue
		}
		score += len([]rune(trigger)) * 4
		if strings.Contains(lower, " ") || len([]rune(trigger)) >= 6 {
			score += 8
		}
	}
	score += intentBoosts(query, item.ID)
	return score
}

func intentBoosts(query, capabilityID string) int {
	boost := 0
	switch capabilityID {
	case "code-review":
		if containsAny(query, "pull request", "code review", "pr review", "review this pr", "review this pull", "代码审查", "代码评审", "行级评论") {
			boost += 80
		}
		if containsAny(query, "review") && containsAny(query, "pr", "pull request", "merge request", "diff", "patch") {
			boost += 50
		}
	case "security":
		if containsAny(query, "security scan", "sast", "sca", "secret scan", "漏洞", "安全扫描", "安全审查", "devsecops") {
			boost += 70
		}
	case "presentation":
		if containsAny(query, "ppt", "pptx", "powerpoint", "slide", "幻灯片", "演示文稿", "keynote", "slidev") {
			boost += 40
		}
	case "image":
		if containsAny(query, "generate an image", "create image", "image generation", "生图", "生成图", "生成一张图", "生成一个图", "生成图片", "插图") {
			boost += 45
		}
		// When the true deliverable is video, stay secondary.
		if containsAny(query, "做成视频", "into a video", "make a video", "and video") {
			boost -= 30
		}
	case "video":
		if containsAny(query, "video", "hyperframes", "remotion", "视频", "生视频", "动画视频", "做成视频", "into a video", "make a video") {
			boost += 45
		}
		if containsAny(query, "做成视频", "into a video", "and video", "再做成视频") && containsAny(query, "图", "image", "图片", "插图") {
			boost += 40
		}
	case "documents":
		if containsAny(query, "docx", "word", "pdf", "xlsx", "excel", "电子表格", "报告文件") {
			boost += 40
		}
	case "data":
		if containsAny(query, "analyze data", "dataset", "correlation", "数据分析", "异常检测", "统计") {
			boost += 40
		}
	case "writing":
		if containsAny(query, "写文章", "文案", "润色", "翻译", "去ai味", "humanize", "copywriting", "translate", "article") {
			boost += 30
		}
	case "visual":
		if containsAny(query, "design system", "polish the design", "polish design", "taste", "design critique", "设计系统", "视觉设计", "审美", "信息图", "infographic", "动态看板", "hud") {
			boost += 55
		}
		if containsAny(query, "电子看板") && containsAny(query, "动态", "hud", "看板") {
			boost += 30
		}
		// Prefer frontend for generic page polish unless design-system language is present.
		if containsAny(query, "美化") && containsAny(query, "页面", "网页", "ui") && !containsAny(query, "设计系统", "design system", "视觉", "审美") {
			boost -= 40
		}
	case "frontend":
		if containsAny(query, "页面", "前端", "网页", "frontend", "dashboard", "ui", "ux") {
			boost += 35
		}
		if containsAny(query, "美化这个页面", "美化页面", "改这个页面") {
			boost += 50
		}
		if containsAny(query, "电子看板") && !containsAny(query, "动态", "hud", "视觉", "审美") {
			boost += 20
		}
	case "context":
		if containsAny(query, "上下文", "项目检索", "代码库检索", "repo map", "context-select") {
			boost += 50
		}
		if containsAny(query, "token") && containsAny(query, "search", "code", "usage", "检索", "代码") {
			boost += 40
		}
	case "search":
		// Demote generic "search" when the query is about local code/token context.
		if containsAny(query, "search code", "token usage", "代码库", "repo map", "context") && !containsAny(query, "全网", "web", "research", "github上看看") {
			boost -= 60
		}
		if containsAny(query, "全网", "research", "search the web", "url to markdown", "youtube") {
			boost += 30
		}
	case "code":
		// Avoid stealing specialized review/security/search phrasing.
		if containsAny(query, "pull request", "code review", "pr review", "security scan", "search code") {
			boost -= 40
		}
		if containsAny(query, "review this") && !containsAny(query, "code review", "pull request", "pr ") {
			boost -= 20
		}
	case "evolution":
		if containsAny(query, "evolve", "distill", "蒸馏", "融合", "替换skill", "能力包") {
			boost += 40
		}
	}
	return boost
}

func domainPriority(id string) int {
	order := map[string]int{
		"code-review":  100,
		"security":     95,
		"presentation": 90,
		"image":        85,
		"video":        84,
		"documents":    80,
		"data":         78,
		"visual":       75,
		"frontend":     70,
		"writing":      68,
		"search":       60,
		"context":      55,
		"evolution":    50,
		"code":         40,
	}
	if value, ok := order[id]; ok {
		return value
	}
	return 0
}

func secondaryCapabilities(query, primary string, manifests []Manifest) []string {
	hints := []struct {
		terms []string
		id    string
	}{
		{[]string{"并配图", "配图", "插图", "生成图片", "生成图", "生图", "图片", "and image", "with image", "generate image", "generate an image", "create image"}, "image"},
		{[]string{"并做成视频", "做成视频", "再做成视频", "生成视频", "生视频", "and video", "into a video", "make a video", "generate video"}, "video"},
		{[]string{"再写文档", "并写文档", "and document", "write docs"}, "documents"},
		{[]string{"再做ppt", "并做ppt", "and ppt", "with slides"}, "presentation"},
		{[]string{"并分析数据", "and analyze data"}, "data"},
		{[]string{"并做安全扫描", "and security scan"}, "security"},
	}
	available := map[string]bool{}
	for _, item := range manifests {
		if rank(item.Status) >= rank("callable") {
			available[item.ID] = true
		}
	}
	result := []string{}
	seen := map[string]bool{primary: true}
	for _, hint := range hints {
		if !containsAny(query, hint.terms...) || seen[hint.id] || !available[hint.id] {
			continue
		}
		seen[hint.id] = true
		result = append(result, hint.id)
	}
	return result
}

func mountSources(query string, selected Manifest, engine string) []MountedSource {
	mounted := []MountedSource{}
	if rank(selected.Status) < rank("callable") {
		return mounted
	}
	fullMount := containsAny(query, "完整能力", "full capability", "mount all sources", "全部来源")
	for _, source := range selected.Sources {
		if source.Engine != "" && engine != "" && source.Engine != engine {
			continue
		}
		priority := sourcePriority(source)
		if !shouldMountSource(query, source, priority, fullMount) {
			continue
		}
		if path, ok := ResolveCompleteSourceAt(selected.Root, source); ok {
			mounted = append(mounted, MountedSource{ID: source.ID, Path: path, Priority: priority})
		}
	}
	return mounted
}

func sourcePriority(source Source) string {
	priority := strings.ToLower(strings.TrimSpace(source.Priority))
	switch priority {
	case "primary", "secondary", "optional":
		return priority
	}
	id := strings.ToLower(source.ID)
	if strings.Contains(id, "unified") || strings.Contains(id, "catalog") {
		return "primary"
	}
	return "secondary"
}

func shouldMountSource(query string, source Source, priority string, fullMount bool) bool {
	if fullMount {
		return true
	}
	named := sourceNamedInQuery(query, source)
	switch priority {
	case "primary":
		return true
	case "optional":
		return named
	default: // secondary
		return named
	}
}

func sourceNamedInQuery(query string, source Source) bool {
	id := strings.ToLower(source.ID)
	tokens := []string{id}
	// Prefer multi-part atom names over single generic words like "code"/"design".
	parts := strings.Split(id, "-")
	for i := 0; i+1 < len(parts); i++ {
		pair := parts[i] + "-" + parts[i+1]
		if len([]rune(pair)) >= 7 {
			tokens = append(tokens, pair)
			tokens = append(tokens, strings.ReplaceAll(pair, "-", " "))
		}
	}
	for _, part := range parts {
		// Long distinctive tokens only; avoid matching generic query words.
		if len([]rune(part)) >= 8 {
			tokens = append(tokens, part)
		}
	}
	aliases := map[string][]string{
		"humanize-ppt":     {"humanize-ppt", "humanize ppt"},
		"ppt-master":       {"ppt master", "ppt-master"},
		"slidev":           {"slidev"},
		"baoyu":            {"baoyu", "宝玉"},
		"html-ppt":         {"html-ppt", "html ppt"},
		"huashu":           {"huashu", "华数"},
		"open-code-review": {"open-code-review", "opencodereview"},
		"impeccable":       {"impeccable"},
		"khazix":           {"khazix"},
		"stage-fluid":      {"stage fluid", "stage-fluid", "流体背景", "烟雾背景"},
	}
	for key, values := range aliases {
		if strings.Contains(id, key) {
			tokens = append(tokens, values...)
		}
	}
	return containsAny(query, tokens...)
}

func selectEngine(query string, engines []Engine) Engine {
	selected := Engine{}
	for _, engine := range engines {
		if engine.Default {
			selected = engine
			break
		}
	}
	best := 0
	for _, engine := range engines {
		for _, trigger := range engine.Triggers {
			if strings.Contains(query, strings.ToLower(trigger)) && len([]rune(trigger)) > best {
				selected, best = engine, len([]rune(trigger))
			}
		}
	}
	return selected
}

func selectProvider(query string, providers []Provider) (string, string) {
	selected := Provider{}
	for _, provider := range providers {
		if provider.Default {
			selected = provider
		}
	}
	best := 0
	for _, provider := range providers {
		for _, trigger := range provider.Triggers {
			if strings.Contains(query, strings.ToLower(trigger)) && len([]rune(trigger)) > best {
				selected, best = provider, len([]rune(trigger))
			}
		}
	}
	return selected.ID, selected.Fallback
}

func workerPlan(query string, capability Manifest, engine string) []WorkerTask {
	if capability.ID == "search" && engine == "web-research" && isBroadSearch(query) {
		return []WorkerTask{
			{ID: "official", Purpose: "official documentation and primary sources", ModelClass: "luna", Inputs: []string{"query", "source boundary"}},
			{ID: "github", Purpose: "repositories, releases, issues, and implementation evidence", ModelClass: "luna", Inputs: []string{"query", "source boundary"}},
			{ID: "community", Purpose: "independent reports and failure evidence", ModelClass: "luna", Inputs: []string{"query", "source boundary"}},
		}
	}
	if containsAny(query, "串行", "sequential", "serial only") {
		return nil
	}
	forceParallel := containsAny(query, "并行", "parallel")
	defaultFanout := map[string]bool{
		"presentation": true,
		"writing":      true,
		"code":         true,
		"frontend":     true,
	}
	if !forceParallel && !defaultFanout[capability.ID] {
		return nil
	}
	workers := []WorkerTask{}
	for _, expert := range capability.Experts {
		if expert.Independent {
			workers = append(workers, WorkerTask{
				ID:         expert.ID,
				Purpose:    expert.Purpose,
				ModelClass: expert.ModelClass,
				Inputs:     []string{"task contract", "selected context handles"},
			})
		}
	}
	if len(workers) < 2 {
		return nil
	}
	sort.SliceStable(workers, func(i, j int) bool { return workers[i].ID < workers[j].ID })
	return workers
}

func isBroadSearch(query string) bool {
	return containsAny(query, "全网", "搜索", "检索", "调研", "research", "search the web", "github上看看")
}

func explicitOfficers(query string) []string {
	mapTerms := []struct{ term, id string }{
		{"白帽", "white-hat"}, {"white-hat", "white-hat"}, {"根因官", "root-cause-officer"},
		{"根因雷达官", "root-cause-officer"}, {"审计官", "audit"}, {"质检官", "quality-inspection"},
		{"所有独立官", "all-independent-officers"}, {"全体独立官", "all-independent-officers"},
	}
	seen := map[string]bool{}
	result := []string{}
	for _, item := range mapTerms {
		if strings.Contains(query, item.term) && !seen[item.id] {
			seen[item.id] = true
			result = append(result, item.id)
		}
	}
	return result
}

func needsInternalChallenge(query string) bool {
	return containsAny(query, "架构", "重构", "删除", "安全", "权限", "生产", "发布", "迁移", "architecture", "security", "delete", "deploy")
}

func containsAny(query string, terms ...string) bool {
	for _, term := range terms {
		if term != "" && strings.Contains(query, strings.ToLower(term)) {
			return true
		}
	}
	return false
}

func rank(status string) int {
	for i, value := range []string{"known", "doctrine-only", "assets-retained", "callable", "behavior-verified", "primary"} {
		if status == value {
			return i
		}
	}
	return -1
}
