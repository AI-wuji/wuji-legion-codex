package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var modelPolicies = map[string]struct {
	model     string
	fallbacks []string
}{
	"luna":  {model: "gpt-5.6-luna", fallbacks: []string{"gpt-5.6-terra", "gpt-5.6-sol"}},
	"terra": {model: "gpt-5.6-terra", fallbacks: []string{"gpt-5.6-sol"}},
	"sol":   {model: "gpt-5.6-sol"},
}

const (
	routeVersion       = "3.0"
	activeSkillID      = "wuji-legion-codex-3-0"
	ajiMainModel       = "gpt-5.6-terra"
	gptHierarchyMode   = "gpt-hierarchy"
	nonGPTProviderMode = "explicit-non-gpt-provider-mode"
)

var workerExecutionEvidenceFields = []string{
	"schema_version", "worker_id", "requested_model", "session_key", "host_dispatch_id", "write_boundary", "attempts", "effective_model", "model_switch_count", "result_handle",
	"stable_prefix_bytes", "stable_prefix_sha256", "source_execution_bytes", "context_handle_ids", "context_bytes_sent", "context_payload_sha256",
	"task_contract_bytes", "task_contract_sha256", "delegation_gate_reason",
	"input_tokens", "cached_input_tokens", "output_tokens", "retry_count",
	"attempt_failure_kinds", "cache_domain", "billing_unit",
	"total_cost_microunits", "execution_baseline_microunits", "savings_microunits",
}

const (
	maxTaskContractBytes      = 2048
	maxSharedContextBytes     = 4096
	maxTotalReplayBytes       = 8192
	minContextCoverageBPS     = 6000
	priorArtMaxSources        = 3
	priorArtTimeBudgetSec     = 90
	fullResearchTimeBudgetSec = 900
	maxAvailabilityFallbacks  = 2
)

func modelSpec(modelClass string) (string, []string) {
	if spec, ok := modelPolicies[strings.ToLower(strings.TrimSpace(modelClass))]; ok {
		return spec.model, append([]string(nil), spec.fallbacks...)
	}
	return "", nil
}

func modelPolicy(userSelectedModel string) ModelPolicy {
	userSelectedModel = strings.TrimSpace(userSelectedModel)
	if userSelectedModel != "" && !isGPTModel(userSelectedModel) {
		return ModelPolicy{
			RoutingMode:       nonGPTProviderMode,
			UserSelectedModel: userSelectedModel,
			MainModel:         userSelectedModel,
			ClassModels:       map[string]string{},
			FallbackModels:    map[string][]string{},
			Delegation:        "the user selected a non-GPT model, so preserve capability/provider mode routing and do not emit GPT hierarchy worker contracts.",
		}
	}
	classes := map[string]string{}
	fallbacks := map[string][]string{}
	for class, spec := range modelPolicies {
		classes[class] = spec.model
		fallbacks[class] = append([]string(nil), spec.fallbacks...)
	}
	mainModel := ajiMainModel
	mainFallbacks := append([]string(nil), modelPolicies["terra"].fallbacks...)
	if userSelectedModel != "" {
		mainModel = userSelectedModel
		switch strings.ToLower(userSelectedModel) {
		case "gpt-5.6-terra":
			mainFallbacks = append([]string(nil), modelPolicies["terra"].fallbacks...)
		case "gpt-5.6-sol":
			mainFallbacks = nil
		}
	}
	return ModelPolicy{
		RoutingMode:        gptHierarchyMode,
		UserSelectedModel:  userSelectedModel,
		MainModel:          mainModel,
		MainFallbackModels: mainFallbacks,
		ClassModels:        classes,
		FallbackModels:     fallbacks,
		Delegation:         "Aji is the sole user-facing judgment and reporting center; the named General Staff is a deterministic task-state mechanism, and only required execution nodes are created. Aji defaults to Terra and falls back to Sol before generation; task workers retain their declared availability chain and an established session remains sticky.",
	}
}

func isGPTModel(model string) bool {
	model = strings.ToLower(strings.TrimSpace(model))
	return strings.HasPrefix(model, "gpt-") || strings.HasPrefix(model, "openai/") || strings.HasPrefix(model, "openai:")
}

func Route(query string, manifests []Manifest) RouteResult {
	return RouteWithContextAndModel(query, manifests, DelegationContext{}, "")
}

func RouteWithContext(query string, manifests []Manifest, context DelegationContext) RouteResult {
	return RouteWithContextAndModel(query, manifests, context, "")
}

func RouteWithModel(query string, manifests []Manifest, userSelectedModel string) RouteResult {
	return RouteWithContextAndModel(query, manifests, DelegationContext{}, userSelectedModel)
}

func buildAjiTaskIntent(query string, capability Manifest, secondary []string, officers []string, search SearchFirstPolicy, delegated bool) AjiTaskIntent {
	complexity := "direct"
	minimum := "answer or perform the smallest correct action through the selected capability"
	if isSimpleQuestion(query) {
		minimum = "answer directly from available context; do not create a worker"
	} else if search.Required {
		complexity = "search-first"
		minimum = "run the bounded source scan, then reroute only if its evidence changes the approach"
	} else if delegated {
		complexity = "bounded-delegation"
		minimum = "create only the execution branches required by the selected capability"
	}
	if len(secondary) > 0 || len(officers) > 0 {
		complexity = "composed"
	}
	accepted := []string{"the requested outcome is present", "required verification evidence is available", "no unrequested side effect is introduced"}
	if capability.ID == "search" || search.Required {
		accepted = append(accepted, "claims are backed by bounded source evidence")
	}
	if len(officers) > 0 {
		accepted = append(accepted, "independent officer evidence is available before high-risk reporting")
	}
	return AjiTaskIntent{
		Objective:               query,
		Constraints:             []string{"Aji remains the only user-facing communicator", "General Staff is deterministic state and scheduling, not a model worker", "do not claim completion from child creation or self-reported receipts"},
		AcceptanceCriteria:      accepted,
		Complexity:              complexity,
		MinimumCorrectPath:      minimum,
		ReuseCandidates:         []string{"existing primary Skill", "installed plugin or MCP", "local template or dependency", "native Codex capability"},
		SelectedCapabilities:    append([]string{capability.ID}, secondary...),
		RejectedComplexity:      []string{"resident staff model", "default multi-agent panel", "parallel branches without independent work", "automatic fallback after generation"},
		Risks:                   []string{"wrong capability selection", "stale or insufficient context", "provider/model unavailable", "unverified completion claim"},
		EvidenceRequirements:    []string{"native host dispatch identity", "result handle", "task-local verification", "independent verification for high-risk work"},
		SideEffects:             []string{"workers may write only their scoped artifacts", "routing and evidence state may be updated", "no unrelated workspace changes"},
		IndependentVerification: len(officers) > 0 || capability.ID == "security" || capability.ID == "code-review",
	}
}

func RouteWithContextAndModel(query string, manifests []Manifest, context DelegationContext, userSelectedModel string) RouteResult {
	return RouteWithContextModelAndResponseState(query, manifests, context, userSelectedModel, false)
}

// RouteWithContextModelAndResponseState lets the host carry an explicitly
// activated response policy across turns without turning it into a competing
// domain capability.
func RouteWithContextModelAndResponseState(query string, manifests []Manifest, context DelegationContext, userSelectedModel string, responsePolicyActive bool) RouteResult {
	q := strings.ToLower(strings.TrimSpace(query))
	policy := modelPolicy(userSelectedModel)
	selected := Manifest{ID: "core", Status: "primary", PrimarySkill: activeSkillID}
	bestScore := 0
	bestPriority := -1
	for _, item := range manifests {
		if item.ID == responsePolicyCapabilityID {
			continue
		}
		score := scoreCapability(q, item)
		priority := domainPriority(item.ID)
		if score > bestScore || (score == bestScore && score > 0 && priority > bestPriority) {
			bestScore, bestPriority, selected = score, priority, item
		}
	}
	if hasExplicitWebResearchIntent(q) && !isOfflineSearchRequest(q) {
		for _, item := range manifests {
			if item.ID == "search" && rank(item.Status) >= rank("callable") {
				selected = item
				break
			}
		}
	}

	selectedEngine := selectEngine(q, selected.Engines)
	engine := selectedEngine.ID
	mounted := mountSources(q, selected, engine)
	sourceExecution, sourceErr := BuildSourceExecutionContracts(selected, mounted)
	assetContracts, assetErr := selectRouteAssets(q, selected, engine)
	sourceActivationError := ""
	if sourceErr != nil {
		// Do not advertise an unreadable source as an active capability. The
		// source remains in the audit inventory and dispatch will not run it.
		mounted = nil
		sourceExecution = nil
		sourceActivationError = sourceErr.Error()
	}
	if assetErr != nil {
		assetContracts = nil
		sourceActivationError = assetErr.Error()
	}
	var responsePolicy *ResponsePolicyContract
	responsePolicyError := ""
	if responsePolicyActive || responsePolicyRequested(q) {
		if interaction, ok := capabilityManifest(manifests, responsePolicyCapabilityID); ok {
			compiled, err := CompileResponsePolicy(interaction, q, responsePolicyActive)
			if err != nil {
				responsePolicyError = err.Error()
			} else {
				responsePolicy = compiled
			}
		} else {
			responsePolicyError = "interaction response policy capability is unavailable"
		}
	}
	for _, contract := range assetContracts {
		sourceExecution = appendSourceExecution(sourceExecution, contract.Invocation)
	}
	secondary := secondaryCapabilities(q, selected.ID, manifests)
	if responsePolicy != nil && !containsString(secondary, responsePolicyCapabilityID) {
		secondary = append(secondary, responsePolicyCapabilityID)
		sort.Strings(secondary)
	}
	officers := explicitOfficers(q)
	officerRecommendations := SelectOfficerRecommendations(q)
	if len(officers) == 0 && hasCompositeOfficerRecommendation(officerRecommendations) {
		officers = []string{"composite-moe"}
	}
	officerWorkers := officerWorkerPlan(q, officers, selected, engine, context)
	searchFirst, preflightWorkers := searchFirstPlan(q, selected, engine, context)
	if searchFirst.Required && !containsString(secondary, "search") && selected.ID != "search" {
		secondary = append(secondary, "search")
		sort.Strings(secondary)
	}
	workers, delegation := workerPlan(q, selected, engine, context)
	if policy.RoutingMode == nonGPTProviderMode {
		preflightWorkers = nil
		workers = nil
		officerWorkers = nil
		delegation = DelegationDecision{Reason: nonGPTProviderMode}
	}
	if sourceActivationError != "" {
		preflightWorkers = nil
		workers = nil
		officerWorkers = nil
		delegation.Allowed = false
		delegation.Reason = "selected-source-entrypoint-unavailable"
	}
	attachSourceExecution(workers, sourceExecution)
	attachAssetContracts(workers, selected, engine, assetContracts)
	// workerPlan may already have declined the fan-out because its original
	// replay estimate exceeded the hard limit. Do not overwrite that useful
	// evidence with a misleading zero after it clears the worker slice.
	if len(workers) > 0 {
		delegation.EstimatedReplayBytes = estimatedReplayBytes(workers)
	}
	if len(workers) > 0 && delegation.EstimatedReplayBytes > maxTotalReplayBytes {
		workers = nil
		delegation.Allowed = false
		delegation.Reason = "estimated-context-replay-exceeds-total-budget"
	}
	parallel := len(workers) > 1
	provider, providerFallback := selectProvider(q, selected.Providers)
	primarySkill := selected.PrimarySkill
	if selectedEngine.PrimarySkill != "" {
		primarySkill = selectedEngine.PrimarySkill
	}
	if rank(selected.Status) < rank("callable") && selected.Fallback != "" {
		primarySkill = selected.Fallback
	}
	executionLane := executionLane(len(preflightWorkers), len(workers))
	if policy.RoutingMode == nonGPTProviderMode {
		executionLane = "provider-mode-passthrough"
	}
	brain := "aji-terra-with-deterministic-general-staff"
	if policy.RoutingMode == nonGPTProviderMode {
		brain = "aji-provider-mode"
	}
	return RouteResult{
		Version:           routeVersion,
		Brain:             brain,
		MainModel:         policy.MainModel,
		GeneralStaffModel: policy.GeneralStaffModel,
		ModelPolicy:       policy,
		TaskIntent:        buildAjiTaskIntent(q, selected, secondary, officers, searchFirst, len(workers) > 0),
		DelegationPolicy: DelegationPolicy{
			CrossModelCacheAssumed:          false,
			CacheScope:                      "model-local stable-prefix only",
			MaxTaskContractBytes:            maxTaskContractBytes,
			MaxSharedContextBytes:           maxSharedContextBytes,
			MaxTotalReplayBytes:             maxTotalReplayBytes,
			MinContextCoverageBasisPoints:   minContextCoverageBPS,
			RequireCodeExcerpt:              true,
			RequireContentAnchor:            true,
			RequireSelfContainedHandoff:     true,
			FallbackOnlyOnAvailabilityError: true,
			OnGateFailure:                   "return to deterministic General Staff reconciliation",
		},
		DelegationDecision: delegation,
		TaskExecutionPolicy: TaskExecutionPolicy{
			TaskShape: "small", ModelSelectionTiming: "once-at-task-start", SessionAffinity: "sticky-per-worker",
			EscalationPolicy: "availability-only-fallback", MaxModelSwitches: 2,
			DowngradeAfterGeneration: false, PreflightBeforeExecution: len(preflightWorkers) > 0,
		},
		SearchFirstPolicy:       searchFirst,
		ChangeCapsule:           changeCapsuleGate(q, selected),
		Reasoning:               "max",
		WriteAuthority:          "assigned-execution-nodes-only; scoped-artifact-write; staff-and-aji-read-only",
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
		SourceExecution:         sourceExecution,
		ResponsePolicy:          responsePolicy,
		ResponsePolicyError:     responsePolicyError,
		AssetContracts:          assetContracts,
		SourceActivationError:   sourceActivationError,
		ExecutionLane:           executionLane,
		GeneralStaffWorker:      nil,
		Parallel:                parallel,
		PreflightWorkers:        preflightWorkers,
		Workers:                 workers,
		Officers:                officers,
		OfficerRecommendations:  officerRecommendations,
		OfficerWorkers:          officerWorkers,
		InternalAdversarialPass: len(officers) == 0 && needsInternalChallenge(q),
		FinishLine: []string{
			"requested active target changed in place",
			"selected capability behavior verified",
			"task-local verification passes",
			"do not claim fused unless capability_status is behavior-verified or primary",
			"do not claim a worker branch completed without its execution evidence receipt",
			"do not claim an officer executed without a validated officer receipt",
		},
	}
}

// selectRouteAssets binds presentation delivery engines to one trustworthy
// adapter asset. Other capabilities remain compatible with manifests that do
// not yet declare a fusion genome.
func selectRouteAssets(query string, capability Manifest, engine string) ([]AssetInvocationContract, error) {
	if capability.ID != "presentation" || capability.Genome == nil {
		return nil, nil
	}
	compatibility := []string{engine, "default"}
	if engine == "web-deck" && containsAny(query, "stage fluid", "stage-fluid", "流体背景", "烟雾背景") {
		compatibility = []string{engine, "stage-fluid"}
	}
	contract, err := SelectFusionAsset([]Manifest{capability}, AssetSelectionRequest{
		Capability:    capability.ID,
		Domain:        engine,
		Compatibility: compatibility,
	})
	if err != nil {
		return nil, err
	}
	return []AssetInvocationContract{contract}, nil
}

func appendSourceExecution(existing []SourceExecutionContract, candidate SourceExecutionContract) []SourceExecutionContract {
	for _, contract := range existing {
		if contract.SourceID == candidate.SourceID && contract.Entrypoint == candidate.Entrypoint {
			return existing
		}
	}
	return append(existing, candidate)
}

func attachAssetContracts(workers []WorkerTask, capability Manifest, engine string, contracts []AssetInvocationContract) {
	if len(contracts) == 0 {
		return
	}
	prefix := stableCapabilityPrefix(capability, engine, contracts...)
	for index := range workers {
		workers[index].AssetContracts = append([]AssetInvocationContract(nil), contracts...)
		workers[index].StableCapabilityPrefix = prefix
		workers[index].StablePrefixSHA256 = sha256Hex([]byte(prefix))
		workers[index].StablePrefixBytes = len([]byte(prefix))
	}
}

func attachSourceExecution(workers []WorkerTask, contracts []SourceExecutionContract) {
	if len(contracts) == 0 {
		return
	}
	for index := range workers {
		workers[index].SourceExecution = append([]SourceExecutionContract(nil), contracts...)
		for _, contract := range contracts {
			workers[index].SourceExecutionBytes += contract.EntrypointBytes
		}
		workers[index].PromptOrder = append([]string{"stable_capability_prefix", "source_execution"}, workers[index].PromptOrder[1:]...)
	}
}

func changeCapsuleGate(query string, capability Manifest) ChangeCapsuleGate {
	if containsAny(query, "external skill", "from github", "from http") && containsAny(query, "install", "fuse", "integrate") {
		return ChangeCapsuleGate{Required: true, Strict: true, Reason: "external-capability-admission"}
	}
	if capability.ID != "code" && capability.ID != "evolution" && capability.ID != "security" && capability.ID != "context" {
		return ChangeCapsuleGate{}
	}
	if containsAny(query, "架构", "architecture", "迁移", "migration", "路由", "routing", "provider", "供应商", "模型策略", "model policy", "依赖升级", "dependency upgrade", "权限", "permission", "安全策略", "security policy") {
		return ChangeCapsuleGate{Required: true, Strict: true, Reason: "high-risk-change-boundary"}
	}
	return ChangeCapsuleGate{}
}

func officerWorkerPlan(query string, officers []string, capability Manifest, engine string, context DelegationContext) []WorkerTask {
	if len(officers) == 0 {
		return nil
	}
	model, fallbacks := modelSpec("sol")
	prefix := stableCapabilityPrefix(capability, engine)
	workers := make([]WorkerTask, 0, len(officers))
	for _, officer := range officers {
		purpose := "independent read-only adversarial review; identify unsupported assumptions, failure modes, and missing verification"
		if officer == "composite-moe" {
			purpose = "one independent composite-MoE quality inspection; verify requirement coverage, execution evidence, failure modes, and governance risk in a single review"
		}
		worker := newWorkerTask(query, "officer-"+officer, purpose, "sol", model, fallbacks,
			[]string{"task contract", "selected evidence handles", "implementation under review"}, contextMode(context), context,
			"explicit independent officer requested", prefix, false)
		worker.Stage = "officer"
		worker.Writes = false
		workers = append(workers, worker)
	}
	return workers
}

func hasCompositeOfficerRecommendation(recommendations []OfficerRecommendation) bool {
	for _, recommendation := range recommendations {
		if recommendation.Role == "composite-moe-officer" && strings.HasPrefix(recommendation.Decision, "independent-composite-quality-inspection") {
			return true
		}
	}
	return false
}

func executionLane(preflightCount, workerCount int) string {
	if preflightCount > 0 {
		return "bounded-search-first"
	}
	if workerCount > 0 {
		return "bounded-delegation"
	}
	return "direct"
}

func searchFirstPlan(query string, capability Manifest, engine string, context DelegationContext) (SearchFirstPolicy, []WorkerTask) {
	policy := SearchFirstPolicy{}
	if !needsPriorArtSearch(query, capability.ID, engine) || context.ParentContextRequired {
		return policy, nil
	}
	policy = SearchFirstPolicy{
		Required: true, Reason: "existing-solution-scan-before-local-reasoning",
		SourceOrder: []string{"official", "github", "community"}, MaxSources: priorArtMaxSources,
		TimeBudgetSeconds:        priorArtTimeBudgetSec,
		StopConditions:           []string{"official solution found", "maintainer-confirmed implementation found", "source or time budget exhausted"},
		CancelStaleExecutionPlan: true,
	}
	model, fallbacks := modelSpec("luna")
	searchCapability := Manifest{ID: "search", PrimarySkill: "wuji-research-suite"}
	prefix := stableCapabilityPrefix(searchCapability, "web-research")
	worker := newWorkerTask(query, "prior-art", "find an existing solution before local reasoning; official sources first, then GitHub, then community", "luna", model, fallbacks, []string{"query", "technology names", "error signature"}, "task-contract-only", context, policy.Reason, prefix, false)
	worker.Stage = "preflight"
	worker.MaxSources = policy.MaxSources
	worker.TimeBudgetSeconds = policy.TimeBudgetSeconds
	worker.StopConditions = append([]string(nil), policy.StopConditions...)
	return policy, []WorkerTask{worker}
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
		if containsAny(query, "审查", "评审", "review") && containsAny(query, "代码", "code", "diff", "patch") {
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
		if containsAny(query, "生成", "绘制", "制作") && containsAny(query, "图", "海报", "封面") {
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
		if containsAny(query, "analyze data", "dataset", "correlation", "csv", "数据", "数据分析", "异常检测", "统计") {
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
		if !containsAny(query, "search code", "token usage", "代码库", "项目检索", "context") && containsAny(query, "全网", "搜索", "检索", "调研", "github", "官方文档", "联网", "research", "search the web", "url to markdown", "youtube") {
			boost += 30
		}
	case "code":
		// Avoid stealing specialized review/security/search phrasing.
		if containsAny(query, "pull request", "code review", "pr review", "security scan", "search code") || (containsAny(query, "审查", "评审", "review") && containsAny(query, "代码", "code", "diff", "patch")) {
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
		if rank(sourceLifecycle(source)) < rank("callable") {
			continue
		}
		reason, ok := sourceActivationReason(query, source, priority, fullMount)
		if !ok {
			continue
		}
		if path, ok := ResolveCompleteSourceAt(selected.Root, source); ok {
			entrypoint := source.Entrypoint
			if entrypoint != "" && !isSourceEntrypoint(path, entrypoint) {
				continue
			}
			mounted = append(mounted, MountedSource{ID: source.ID, Path: path, Priority: priority, Lifecycle: sourceLifecycle(source), Entrypoint: entrypoint, ActivationReason: reason})
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

func sourceActivationReason(query string, source Source, priority string, fullMount bool) (string, bool) {
	if fullMount {
		return "full-capability-request", true
	}
	switch priority {
	case "primary":
		return "primary-source", true
	}
	for _, trigger := range source.Activation {
		if strings.Contains(query, strings.ToLower(trigger)) {
			return "semantic-trigger:" + trigger, true
		}
	}
	if sourceNamedInQuery(query, source) {
		return "explicit-source-request", true
	}
	return "", false
}

func isSourceEntrypoint(root, entrypoint string) bool {
	path := filepath.Join(root, filepath.FromSlash(entrypoint))
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
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

func workerPlan(query string, capability Manifest, engine string, context DelegationContext) ([]WorkerTask, DelegationDecision) {
	taskContractBytes := len([]byte(strings.TrimSpace(query)))
	decision := DelegationDecision{TaskContractBytes: taskContractBytes, SelectedContextBytes: context.SelectedBytes, ContextCoverageBPS: context.CoverageBPS, CodeExcerptCount: context.CodeExcerptCount, ContentAnchorCount: context.ContentAnchorCount, SelfContained: context.SelfContained}
	if containsAny(query, "串行", "sequential", "serial only") {
		decision.Reason = "serial-task-reasoning"
		return taskJudgmentPlan(query, capability, engine, decision)
	}
	if taskContractBytes > maxTaskContractBytes {
		decision.Reason = "task-contract-exceeds-budget"
		return nil, decision
	}
	if context.Handle != "" && !validContextHandoff(context) {
		decision.Reason = "verified-context-artifact-required"
		return taskJudgmentPlan(query, capability, engine, decision)
	}
	if context.Handle != "" && context.QueryFingerprint != queryFingerprint(queryTerms(query)) {
		decision.Reason = "context-query-fingerprint-mismatch"
		return taskJudgmentPlan(query, capability, engine, decision)
	}
	if capability.ID == "search" && engine == "web-research" && isBroadSearch(query) {
		model, fallbacks := modelSpec("luna")
		decision.Reason = "task-contract-only-research"
		prefix := stableCapabilityPrefix(capability, engine)
		workers := []WorkerTask{
			newWorkerTask(query, "official", "official documentation and primary sources", "luna", model, fallbacks, []string{"query", "source boundary"}, "task-contract-only", context, decision.Reason, prefix, false),
			newWorkerTask(query, "github", "repositories, releases, issues, and implementation evidence", "luna", model, fallbacks, []string{"query", "source boundary"}, "task-contract-only", context, decision.Reason, prefix, false),
			newWorkerTask(query, "community", "independent reports and failure evidence", "luna", model, fallbacks, []string{"query", "source boundary"}, "task-contract-only", context, decision.Reason, prefix, false),
		}
		if hasExplicitWebResearchIntent(query) {
			for index := range workers {
				workers[index].TimeBudgetSeconds = fullResearchTimeBudgetSec
				workers[index].StopConditions = []string{
					"the requested scope has evidence coverage across the assigned source class",
					"two successive relevant sources add no material claim or contradiction",
					"the time budget is exhausted",
				}
			}
		}
		return finalizeWorkerPlan(workers, decision)
	}
	if isMechanicalTask(query) {
		model, fallbacks := modelSpec("luna")
		decision.Reason = "bounded-mechanical-task"
		prefix := stableCapabilityPrefix(capability, engine)
		workers := []WorkerTask{
			newWorkerTask(query, "mechanical", "bounded read-only extraction, inventory, counting, or log parsing", "luna", model, fallbacks, []string{"task contract", "workspace tools"}, "task-contract-only", context, decision.Reason, prefix, false),
		}
		return finalizeWorkerPlan(workers, decision)
	}
	if isSimpleQuestion(query) {
		decision.Reason = "simple-question-direct"
		return nil, decision
	}
	forceParallel := containsAny(query, "并行", "parallel")
	if capability.ID == "code-review" {
		decision.Reason = "two-axis-code-review"
		prefix := stableCapabilityPrefix(capability, engine)
		model, fallbacks := modelSpec("sol")
		workers := []WorkerTask{
			newWorkerTask(query, "spec-conformance", "review only whether the change satisfies the stated behavior, acceptance scenarios, and edge cases; return line-anchored findings", "sol", model, fallbacks, []string{"task contract", "selected context payload"}, contextMode(context), context, decision.Reason, prefix, false),
			newWorkerTask(query, "engineering-quality", "review only correctness, maintainability, security, performance, error recovery, and verification gaps; return line-anchored findings", "sol", model, fallbacks, []string{"task contract", "selected context payload"}, contextMode(context), context, decision.Reason, prefix, false),
		}
		return finalizeWorkerPlan(workers, decision)
	}
	if needsSolJudgment(query) {
		model, fallbacks := modelSpec("sol")
		decision.Reason = "explicit-high-reasoning-judgment"
		prefix := stableCapabilityPrefix(capability, engine)
		workers := []WorkerTask{
			newWorkerTask(query, "sol-judgment", "bounded independent high-reasoning judgment; return options, evidence, and risks for General Staff reconciliation", "sol", model, fallbacks, []string{"task contract", "selected context payload"}, contextMode(context), context, decision.Reason, prefix, false),
		}
		return finalizeWorkerPlan(workers, decision)
	}
	// Every non-conversational task receives a bounded reasoning branch. Routing
	// is the default for tasks rather than an opt-in reserved for code changes.
	if !forceParallel && capability.ID != "code" {
		decision.Reason = "default-task-reasoning"
		if isArtifactMutationTask(query) {
			return taskExecutionPlan(query, capability, engine, decision, context)
		}
		return taskJudgmentPlan(query, capability, engine, decision)
	}
	if capability.ID == "code" {
		if !validContextHandoff(context) {
			decision.Reason = "verified-context-artifact-required"
			return taskJudgmentPlan(query, capability, engine, decision)
		}
		if context.SelectedBytes > maxSharedContextBytes {
			decision.Reason = "shared-context-exceeds-per-worker-budget"
			return taskJudgmentPlan(query, capability, engine, decision)
		}
		if context.CoverageBPS < minContextCoverageBPS {
			decision.Reason = "context-coverage-below-delegation-threshold"
			return taskJudgmentPlan(query, capability, engine, decision)
		}
		if context.CodeExcerptCount == 0 {
			decision.Reason = "code-context-excerpt-required"
			return taskJudgmentPlan(query, capability, engine, decision)
		}
		if context.ContentAnchorCount == 0 {
			decision.Reason = "code-content-anchor-required"
			return taskJudgmentPlan(query, capability, engine, decision)
		}
	} else if forceParallel && !context.SelfContained {
		decision.Reason = "self-contained-handoff-required"
		return taskJudgmentPlan(query, capability, engine, decision)
	}
	workers := []WorkerTask{}
	prefix := stableCapabilityPrefix(capability, engine)
	for _, expert := range capability.Experts {
		if expert.Independent {
			model, fallbacks := modelSpec(expert.ModelClass)
			reason := "bounded-context-handoff"
			if contextMode(context) == "task-contract-only" {
				reason = "self-contained-task-contract"
			}
			workers = append(workers, newWorkerTask(query, expert.ID, expert.Purpose, expert.ModelClass, model, fallbacks, []string{"task contract", "selected context payload"}, contextMode(context), context, reason, prefix, isArtifactMutationTask(query)))
		}
	}
	if len(workers) == 0 {
		decision.Reason = "no-independent-workers"
		return nil, decision
	}
	sort.SliceStable(workers, func(i, j int) bool { return workers[i].ID < workers[j].ID })
	return finalizeWorkerPlan(workers, decision)
}

// taskJudgmentPlan is the safe fallback when a task cannot replay enough
// workspace context for implementation fan-out. It still activates routing,
// but gives Sol only the self-contained user contract rather than leaking a
// stale or incomplete context capsule.
func taskJudgmentPlan(query string, capability Manifest, engine string, decision DelegationDecision) ([]WorkerTask, DelegationDecision) {
	decision.FallbackAllowed = true
	model, fallbacks := modelSpec("terra")
	prefix := stableCapabilityPrefix(capability, engine)
	workers := []WorkerTask{
		newWorkerTask(query, "task-judgment", "bounded task analysis; return an actionable approach, evidence needs, risks, and verification criteria for an assigned execution node", "terra", model, fallbacks, []string{"task contract"}, "task-contract-only", DelegationContext{}, decision.Reason, prefix, false),
	}
	return finalizeWorkerPlan(workers, decision)
}

func taskExecutionPlan(query string, capability Manifest, engine string, decision DelegationDecision, context DelegationContext) ([]WorkerTask, DelegationDecision) {
	model, fallbacks := modelSpec("terra")
	prefix := stableCapabilityPrefix(capability, engine)
	mode := contextMode(context)
	if !validContextHandoff(context) {
		mode = "task-contract-only"
		context = DelegationContext{}
	}
	workers := []WorkerTask{
		newWorkerTask(query, "task-execution", "produce or modify the requested task artifact within the assigned scope and return task-local verification evidence", "terra", model, fallbacks, []string{"task contract", "selected context payload when available"}, mode, context, decision.Reason, prefix, true),
	}
	return finalizeWorkerPlan(workers, decision)
}

func finalizeWorkerPlan(workers []WorkerTask, decision DelegationDecision) ([]WorkerTask, DelegationDecision) {
	maxContract, totalContract := 0, 0
	for _, worker := range workers {
		if worker.AllocatedTaskContractBytes > maxContract {
			maxContract = worker.AllocatedTaskContractBytes
		}
		totalContract += worker.AllocatedTaskContractBytes
	}
	decision.TaskContractBytes = maxContract
	decision.TotalContractBytes = totalContract
	decision.ContextHandle = ""
	if len(workers) > 0 && len(workers[0].ContextHandles) > 0 {
		decision.ContextHandle = workers[0].ContextHandles[0]
	}
	decision.EstimatedReplayBytes = estimatedReplayBytes(workers)
	if maxContract > maxTaskContractBytes {
		decision.Reason = "task-contract-exceeds-budget"
		return nil, decision
	}
	if decision.EstimatedReplayBytes > maxTotalReplayBytes {
		decision.Reason = "estimated-context-replay-exceeds-total-budget"
		return nil, decision
	}
	decision.Allowed = !decision.FallbackAllowed
	decision.ImplementationAllowed = decision.Allowed
	return workers, decision
}

func estimatedReplayBytes(workers []WorkerTask) int {
	total := 0
	for _, worker := range workers {
		total += worker.StablePrefixBytes + worker.SourceExecutionBytes + worker.AllocatedTaskContractBytes + worker.AllocatedContextBytes
	}
	return total
}

func validContextHandoff(context DelegationContext) bool {
	if !context.verified || context.Handle == "" || context.ArtifactPath == "" || context.SelectedBytes <= 0 || context.Payload == "" {
		return false
	}
	return context.SelectedBytes == len([]byte(context.Payload)) && context.PayloadSHA256 == sha256Hex([]byte(context.Payload))
}

func contextMode(context DelegationContext) string {
	if context.Handle != "" && context.ArtifactPath != "" {
		return "shared-content-addressed-handle"
	}
	return "task-contract-only"
}

func newWorkerTask(query, id, purpose, modelClass, model string, fallbacks, inputs []string, mode string, context DelegationContext, reason, stablePrefix string, writes bool) WorkerTask {
	handles := []string{}
	artifact := ""
	payload := ""
	payloadHash := ""
	allocated := 0
	if mode == "shared-content-addressed-handle" {
		handles = []string{context.Handle}
		artifact = context.ArtifactPath
		payload = context.Payload
		payloadHash = context.PayloadSHA256
		allocated = context.SelectedBytes
	}
	sessionKey := taskSessionKey(query, id, context.Handle)
	protocol := workerProtocol(query, id, purpose, stablePrefix)
	contract := marshalWorkerContract(query, id, purpose, handles, sessionKey, protocol, writes)
	availabilityFallbackOn := []string{}
	if len(fallbacks) > 0 {
		availabilityFallbackOn = []string{"model-unavailable", "provider-error-before-generation"}
	}
	return WorkerTask{
		ID:                         id,
		Stage:                      "execution",
		Purpose:                    purpose,
		ModelClass:                 modelClass,
		Model:                      model,
		AvailabilityFallbackModels: append([]string(nil), fallbacks...),
		AvailabilityFallbackOn:     availabilityFallbackOn,
		FallbackModels:             nil,
		SessionKey:                 sessionKey,
		SessionAffinity:            "sticky-per-worker",
		EscalationPolicy:           "availability-only-fallback",
		MaxModelSwitches:           0,
		Inputs:                     append([]string(nil), inputs...),
		Protocol:                   append([]string(nil), protocol...),
		TaskContract:               contract,
		TaskContractSHA256:         sha256Hex([]byte(contract)),
		ContextMode:                mode,
		ContextHandles:             handles,
		ContextArtifact:            artifact,
		ContextPayload:             payload,
		ContextPayloadSHA256:       payloadHash,
		StableCapabilityPrefix:     stablePrefix,
		StablePrefixSHA256:         sha256Hex([]byte(stablePrefix)),
		StablePrefixBytes:          len([]byte(stablePrefix)),
		PromptOrder:                promptOrder(mode),
		AllocatedContextBytes:      allocated,
		AllocatedTaskContractBytes: len([]byte(contract)),
		MaxTaskContractBytes:       maxTaskContractBytes,
		DelegationGateReason:       reason,
		MaxAttempts:                1,
		FallbackOn:                 nil,
		Writes:                     writes,
		ExecutionEvidenceRequired:  true,
		ExecutionEvidenceFields:    append([]string(nil), workerExecutionEvidenceFields...),
	}
}

func stableCapabilityPrefix(capability Manifest, engine string, assets ...AssetInvocationContract) string {
	prefix := struct {
		Schema                 string                    `json:"schema"`
		Capability             string                    `json:"capability"`
		PrimarySkill           string                    `json:"primary_skill"`
		Engine                 string                    `json:"engine,omitempty"`
		ImplementationDoctrine string                    `json:"implementation_doctrine,omitempty"`
		AssetContracts         []AssetInvocationContract `json:"asset_contracts,omitempty"`
		WriteOwner             string                    `json:"write_owner"`
	}{
		Schema: "wuji-stable-capability-prefix-v1", Capability: capability.ID,
		PrimarySkill: primarySkillForEngine(capability, engine), Engine: engine, WriteOwner: "assigned-execution-node-scoped",
	}
	prefix.ImplementationDoctrine = "ponytail-v3: universal-minimum-correct-task-judgment"
	if len(assets) > 0 {
		prefix.AssetContracts = append([]AssetInvocationContract(nil), assets...)
	}
	encoded, _ := json.Marshal(prefix)
	return string(encoded)
}

func primarySkillForEngine(capability Manifest, engineID string) string {
	for _, engine := range capability.Engines {
		if engine.ID == engineID && strings.TrimSpace(engine.PrimarySkill) != "" {
			return engine.PrimarySkill
		}
	}
	return capability.PrimarySkill
}

type workerContract struct {
	Schema        string   `json:"schema"`
	Objective     string   `json:"objective"`
	Branch        string   `json:"branch"`
	Purpose       string   `json:"purpose"`
	Boundaries    []string `json:"boundaries"`
	Acceptance    []string `json:"acceptance"`
	Protocol      []string `json:"protocol,omitempty"`
	ContextHandle string   `json:"context_handle,omitempty"`
	SessionKey    string   `json:"session_key"`
	WriteBoundary string   `json:"write_boundary"`
}

func marshalWorkerContract(query, id, purpose string, handles []string, sessionKey string, protocol []string, writes bool) string {
	writeBoundary := "read-only"
	boundaries := []string{"only assigned execution nodes may perform task work", "do not infer missing parent conversation", "return execution and verification evidence, not a completion claim"}
	if writes {
		writeBoundary = "scoped-artifact-write"
		boundaries = append(boundaries, "write only task-scoped artifacts inside the current workspace; do not modify scheduling or requirement state")
	}
	contract := workerContract{
		Schema:        "wuji-worker-contract-v2",
		Objective:     strings.TrimSpace(query),
		Branch:        id,
		Purpose:       purpose,
		Boundaries:    boundaries,
		Acceptance:    []string{"complete only the named branch", "cite the supplied context handle when used", "report model and token telemetry"},
		Protocol:      append([]string(nil), protocol...),
		SessionKey:    sessionKey,
		WriteBoundary: writeBoundary,
	}
	if len(handles) > 0 {
		contract.ContextHandle = handles[0]
	}
	encoded, _ := json.Marshal(contract)
	return string(encoded)
}

func workerProtocol(query, id, purpose, stablePrefix string) []string {
	value := strings.ToLower(query + " " + id + " " + purpose)
	protocol := []string{
		"first decide whether this needs action, a direct answer, or no action",
		"reuse the existing Skill, plugin, MCP, template, dependency, native tool, or local artifact before inventing anything",
		"choose the smallest correct path that satisfies the stated objective and acceptance criteria",
		"keep scope bounded; reject incorrect premises and unrequested complexity",
		"state the concrete completion evidence and side effects required before reporting success",
	}
	if strings.Contains(value, "review") || strings.Contains(value, "审查") || strings.Contains(value, "评审") {
		protocol = append(protocol, "separate specification conformance from engineering quality", "anchor findings to concrete files, symbols, or evidence", "rank findings by user impact and likelihood", "do not report stylistic preference as a defect")
	}
	if hasPonytailCodeDoctrine(stablePrefix) {
		protocol = append(protocol,
			"trace the actual flow and cite affected file or symbol anchors before choosing",
			"choose the first valid rung: skip, reuse local code, standard library, native platform, installed dependency, one line, minimum code",
			"for bugs, inspect every caller and fix the common root cause once, not each symptom",
			"prefer deletion, fewest files, and the smallest correct diff; no unrequested abstraction, scaffolding, or dependency",
			"for nontrivial logic, name one smallest runnable regression check; trivial one-line edits need no new test",
			"do not weaken validation, error handling, data safety, security, accessibility, or explicit requirements",
		)
	}
	return protocol
}

func hasPonytailCodeDoctrine(stablePrefix string) bool {
	var prefix struct {
		Capability             string `json:"capability"`
		ImplementationDoctrine string `json:"implementation_doctrine"`
	}
	if json.Unmarshal([]byte(stablePrefix), &prefix) != nil {
		return false
	}
	return strings.HasPrefix(prefix.ImplementationDoctrine, "ponytail-v3:")
}

func taskSessionKey(query, workerID, contextHandle string) string {
	payload := strings.Join([]string{"wuji-task-session-v1", strings.TrimSpace(query), workerID, contextHandle}, "\n")
	return "wuji-session://sha256/" + sha256Hex([]byte(payload))
}

func promptOrder(mode string) []string {
	if mode == "shared-content-addressed-handle" {
		return []string{"stable_capability_prefix", "context_payload", "task_contract"}
	}
	return []string{"stable_capability_prefix", "task_contract"}
}

func isBroadSearch(query string) bool {
	return containsAny(query, "全网", "搜索", "检索", "调研", "research", "search the web", "github上看看")
}

func hasExplicitWebResearchIntent(query string) bool {
	return containsAny(query,
		"全网", "联网搜索", "联网调研", "搜遍全网", "全面调研",
		"search the web", "research the web", "web research", "browse the web", "look up online",
	)
}

func isOfflineSearchRequest(query string) bool {
	return containsAny(query, "不要联网", "不要搜索", "do not search", "offline only")
}

func needsPriorArtSearch(query, capabilityID, engine string) bool {
	if capabilityID == "search" || engine == "web-research" || hasExplicitWebResearchIntent(query) || isOfflineSearchRequest(query) {
		return false
	}
	if isLocalExactSkillLookup(query) {
		return false
	}
	if isDeterministicEdit(query) {
		return false
	}
	return containsAny(query,
		"现成方案", "现有方案", "有没有项目", "官方资料", "社区教程", "最佳实践",
		"skill", "mcp", "自动执行", "能力融合",
		"bug", "报错", "错误", "异常", "崩溃", "失败", "根因", "修复", "调试",
		"api", "sdk", "依赖", "插件", "框架", "模型路由", "缓存", "上下文共享",
		"架构", "重构", "迁移", "升级", "性能", "安全", "集成", "兼容",
		"existing solution", "prior art", "official docs", "best practice", "error", "exception", "crash", "debug",
		"integration", "dependency", "plugin", "framework", "routing", "cache", "migration", "upgrade", "performance", "security",
	)
}

func isDeterministicEdit(query string) bool {
	return containsAny(query,
		"错别字", "拼写", "改文案", "修改文案", "改文字", "修改文字", "重命名", "格式化", "改颜色", "修改颜色", "删除注释",
		"typo", "spelling", "copy change", "rename", "format only", "change the text", "update the text", "delete comment",
	)
}

func isArtifactMutationTask(query string) bool {
	return containsAny(query,
		"做", "创建", "制作", "生成", "写入", "编写", "修改", "编辑", "修复", "实现", "重构", "更新", "删除", "替换", "迁移", "安装", "优化",
		"make", "produce", "create", "build", "generate", "write", "edit", "modify", "change", "fix", "implement", "refactor", "update", "delete", "replace", "migrate", "install", "optimize",
	)
}

func isMechanicalTask(query string) bool {
	if containsAny(query, "修改", "修复", "实现", "重构", "写入", "删除", "change", "fix", "implement", "refactor", "write", "delete") {
		return false
	}
	value := strings.ToLower(query)
	if isLocalExactSkillLookup(value) {
		return true
	}
	if strings.Contains(value, "list") && containsAny(value, "file", "path", "directory") {
		return true
	}
	return containsAny(query,
		"列出文件", "文件清单", "统计数量", "统计行数", "提取日志", "解析日志", "查找所有", "扫描所有", "汇总日志",
		"list files", "inventory files", "count lines", "count occurrences", "extract logs", "parse logs", "find all", "scan all", "summarize logs",
	)
}

func isLocalExactSkillLookup(query string) bool {
	value := strings.ToLower(strings.TrimSpace(query))
	return containsAny(value, "find the skill named", "locate the skill named", "find local skill named", "locate local skill named") &&
		!containsAny(value, "install", "add", "fuse", "integrate", "from github", "from http", "external")
}

// isSimpleQuestion deliberately keeps only lightweight conversational Q&A on
// Aji. Requests to inspect, compare, diagnose, plan, search, create, or alter
// anything are tasks and therefore enter the worker routing path.
func isSimpleQuestion(query string) bool {
	value := strings.TrimSpace(strings.ToLower(query))
	if len([]rune(value)) == 0 || len([]rune(value)) > 240 {
		return false
	}
	if value == "hi" || value == "hello" || value == "你好" || value == "谢谢" || value == "thank you" || containsAny(value,
		"你是谁", "who are you",
		"是什么", "什么意思", "what is", "what does", "how are you",
	) && !containsAny(value,
		"检查", "分析", "比较", "诊断", "调试", "搜索", "调研", "设计", "计划", "实现", "修改", "修复", "创建", "安装", "审查", "验证",
		"inspect", "analy", "compare", "diagnos", "debug", "search", "research", "design", "plan", "implement", "change", "fix", "create", "install", "review", "verify",
	) {
		return true
	}
	return false
}

func needsSolJudgment(query string) bool {
	return containsAny(query,
		"explicit sol", "use sol", "high-reasoning", "high reasoning", "architecture decision", "root-cause adjudication", "threat model",
		"使用sol", "调用sol", "高推理", "架构取舍", "根因裁决", "威胁建模",
	)
}

func requiresParentContext(query string) bool {
	return containsAny(query,
		"preceding", "previous conversation", "above context", "earlier context", "parent transcript", "chat history",
		"前文", "上文", "之前的对话", "前面的记录", "聊天记录", "会议原文", "刚才的内容",
	)
}

func explicitOfficers(query string) []string {
	mapTerms := []struct{ term, id string }{
		{"白帽", "white-hat"}, {"white-hat", "white-hat"}, {"根因官", "root-cause-officer"},
		{"根因雷达官", "root-cause-officer"}, {"审计官", "audit"}, {"质检官", "quality-inspection"},
		{"所有独立官", "composite-moe"}, {"全体独立官", "composite-moe"},
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

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
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
