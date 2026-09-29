package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var modelPolicies = map[string]struct {
	model              string
	fallbacks          []string
	reasoning          string
	reasoningFallbacks []string
}{
	// Luna/Terra/Sol are Legion reasoning roles, not model IDs. The native
	// host owns the concrete model selected for the current ChatGPT account.
	"luna":  {model: hostSelectedModel, reasoning: "low"},
	"terra": {model: hostSelectedModel, reasoning: "medium"},
	"sol":   {model: hostSelectedModel, reasoning: "max", reasoningFallbacks: []string{"xhigh"}},
}

const (
	routeVersion       = "3.0"
	activeSkillID      = "wuji-legion-codex-3-0"
	hostSelectedModel  = "host-selected"
	ajiMainModel       = hostSelectedModel
	gptHierarchyMode   = "gpt-hierarchy"
	nonGPTProviderMode = "explicit-non-gpt-provider-mode"
	ponyTailDoctrine   = "ponytail-v3: minimum correct; least reasoning, tools and code"
)

var workerExecutionEvidenceFields = []string{
	"schema_version", "worker_id", "requested_model", "requested_reasoning_effort", "session_key", "host_dispatch_id", "write_boundary", "attempts", "effective_model", "effective_reasoning_effort", "model_switch_count", "reasoning_switch_count", "result_handle",
	"stable_prefix_bytes", "stable_prefix_sha256", "source_execution_bytes", "context_handle_ids", "context_bytes_sent", "context_payload_sha256",
	"task_contract_bytes", "task_contract_sha256", "delegation_gate_reason",
	"input_tokens", "cached_input_tokens", "output_tokens", "retry_count",
	"attempt_failure_kinds", "cache_domain", "billing_unit",
	"total_cost_microunits", "execution_baseline_microunits", "savings_microunits",
}

const (
	// Task contracts are compact control-plane records, not prompt transcripts.
	// 4096 bytes leaves room for a selected expert's workflow and acceptance
	// checks while remaining below the per-worker replay budget.
	maxTaskContractBytes      = 4096
	maxSharedContextBytes     = 4096
	maxTotalReplayBytes       = 9216
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

func reasoningSpec(modelClass string) (string, []string) {
	if spec, ok := modelPolicies[strings.ToLower(strings.TrimSpace(modelClass))]; ok {
		return spec.reasoning, append([]string(nil), spec.reasoningFallbacks...)
	}
	return "", nil
}

func modelPolicy(userSelectedModel string) ModelPolicy {
	userSelectedModel = strings.TrimSpace(userSelectedModel)
	if userSelectedModel != "" && !isGPTModel(userSelectedModel) {
		return ModelPolicy{
			RoutingMode:              nonGPTProviderMode,
			UserSelectedModel:        userSelectedModel,
			MainModel:                userSelectedModel,
			MainReasoningEffort:      "provider-defined",
			ClassModels:              map[string]string{},
			FallbackModels:           map[string][]string{},
			ClassReasoningEfforts:    map[string]string{},
			FallbackReasoningEfforts: map[string][]string{},
			Delegation:               "the user selected a non-GPT model, so preserve capability/provider mode routing and do not emit GPT hierarchy worker contracts.",
		}
	}
	classes := map[string]string{}
	fallbacks := map[string][]string{}
	classReasoning := map[string]string{}
	fallbackReasoning := map[string][]string{}
	for class, spec := range modelPolicies {
		classes[class] = spec.model
		fallbacks[class] = append([]string(nil), spec.fallbacks...)
		classReasoning[class] = spec.reasoning
		fallbackReasoning[class] = append([]string(nil), spec.reasoningFallbacks...)
	}
	mainModel := ajiMainModel
	mainReasoning := "medium"
	mainFallbacks := []string(nil)
	mainReasoningFallbacks := []string(nil)
	if userSelectedModel != "" {
		mainModel = userSelectedModel
	}
	return ModelPolicy{
		RoutingMode:              gptHierarchyMode,
		UserSelectedModel:        userSelectedModel,
		MainModel:                mainModel,
		MainReasoningEffort:      mainReasoning,
		MainFallbackModels:       mainFallbacks,
		MainFallbackReasoning:    mainReasoningFallbacks,
		ClassModels:              classes,
		FallbackModels:           fallbacks,
		ClassReasoningEfforts:    classReasoning,
		FallbackReasoningEfforts: fallbackReasoning,
		Delegation:               "Aji is the sole user-facing center and uses the host-selected ChatGPT model with medium reasoning by default. Small tasks and conversation stay on Aji; complex work enters the deterministic General Staff state path. High-reasoning work uses max and falls back to xhigh only before generation; no model switch, A/B test, or post-generation downgrade is allowed.",
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

func shouldStayOnAji(query string, capability Manifest, secondary []string, officers []string, search SearchFirstPolicy, context DelegationContext, policy ModelPolicy) bool {
	if policy.RoutingMode == nonGPTProviderMode {
		return false
	}
	if context.Handle != "" {
		return false
	}
	if context.ParentContextRequired {
		return false
	}
	if search.Required || len(officers) > 0 || len(secondary) > 0 {
		return false
	}
	if containsAny(query, "并行", "parallel", "串行", "sequential", "serial only") ||
		needsSolJudgment(query) || needsInternalChallenge(query) ||
		hasExplicitWebResearchIntent(query) || isBroadSearch(query) ||
		needsPriorArtSearch(query, capability.ID, "") {
		return false
	}
	if isComplexTaskSignal(query) && !isMechanicalTask(query) {
		return false
	}
	if isSimpleQuestion(query) || isMechanicalTask(query) || isDeterministicEdit(query) {
		return true
	}
	// Aji handles only a bounded conversational, mechanical, or deterministic
	// action directly. An unfamiliar request still needs the staff state path
	// so its scope and execution evidence are explicit.
	return false
}

func isComplexTaskSignal(query string) bool {
	return containsAny(query,
		"然后", "接着", "并且", "同时", "之后", "最后", "先", "再",
		"多步", "多阶段", "跨领域", "多个文件", "批量", "完整", "系统性",
		"规划", "拆解", "调度", "工作流", "依赖", "集成", "迁移", "重构",
		"验证并", "测试并", "发布", "部署", "上线", "全套", "端到端",
		"then", "next", "and then", "after that", "multi-step", "multi-stage",
		" and ", " & ", "并", "以及",
		"cross-domain", "multiple files", "batch", "complete", "system-wide",
		"plan", "decompose", "orchestrate", "workflow", "dependency", "integration",
		"migration", "refactor", "verify and", "test and", "release", "deploy",
		"end-to-end",
	) || len([]rune(query)) > 180
}

func routeReasoning(query string, staffRequired bool) (string, []string) {
	if staffRequired && (needsSolJudgment(query) || needsInternalChallenge(query) ||
		containsAny(query, "高难度", "复杂推理", "困难", "complex", "high difficulty", "hard reasoning")) {
		return "max", []string{"xhigh"}
	}
	if isSimpleQuestion(query) {
		return "low", nil
	}
	if isMechanicalTask(query) || isDeterministicEdit(query) {
		return "low", nil
	}
	return "medium", nil
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
	} else {
		complexity = "small-direct"
		minimum = "Aji performs the single bounded action directly; do not create General Staff state or a worker"
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
		Constraints:             []string{"Aji remains the only user-facing communicator", "General Staff is deterministic state and scheduling, not a model worker", "use the minimum correct path: direct answer before plan, one line before many, simple before complex", "do not claim completion from child creation or self-reported receipts"},
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
	directAji := shouldStayOnAji(q, selected, secondaryWithoutResponsePolicy(secondary), officers, searchFirst, context, policy)
	workers := []WorkerTask(nil)
	delegation := DelegationDecision{
		TaskContractBytes:    len([]byte(strings.TrimSpace(q))),
		SelectedContextBytes: context.SelectedBytes,
		ContextCoverageBPS:   context.CoverageBPS,
		CodeExcerptCount:     context.CodeExcerptCount,
		ContentAnchorCount:   context.ContentAnchorCount,
		SelfContained:        context.SelfContained,
	}
	if directAji {
		delegation.Reason = directAjiReason(q)
		preflightWorkers = nil
		searchFirst = SearchFirstPolicy{}
		officerWorkers = nil
	} else {
		workers, delegation = workerPlan(q, selected, engine, context)
	}
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
	var expertRoute *ExpertRouteDecision
	secondaryExpertRoutes := map[string]ExpertRouteDecision{}
	if !directAji && policy.RoutingMode != nonGPTProviderMode && sourceActivationError == "" &&
		selected.Root != "" && rank(selected.Status) >= rank("callable") {
		_, catalogErr := os.Stat(filepath.Join(selected.Root, "capabilities", "experts", "manifest.json"))
		// Standalone capability fixtures and external manifests need not own
		// a Legion roster. An existing but invalid roster must fail closed.
		if catalogErr != nil && !os.IsNotExist(catalogErr) {
			sourceActivationError = "expert catalog: " + catalogErr.Error()
		}
		if catalogErr == nil {
			commander, hasCommander, err := expertCommanderForCapability(selected.Root, selected.ID)
			if err != nil {
				sourceActivationError = "expert catalog: " + err.Error()
			} else if hasCommander {
				selection, err := SelectExpertForCapability(selected.Root, q, selected.ID)
				if err != nil {
					sourceActivationError = "expert selection: " + err.Error()
				} else {
					expertRoute = &ExpertRouteDecision{
						Commander: commander.ID, CommanderName: commander.Name, Capability: selected.ID,
						MoE: "sparse-role-moe", Selection: selection, ExecutionStatus: "contract-only; native-host-receipt-required",
					}
					if selection.State == "selected" {
						catalog, _, err := loadExpertCatalog(selected.Root)
						if err != nil {
							sourceActivationError = "expert catalog: " + err.Error()
						} else {
							for _, expert := range catalog.Experts {
								if expert.ID != selection.ExpertID {
									continue
								}
								for _, sourceID := range expert.SourceIDs {
									alreadyMounted := false
									for _, mountedSource := range mounted {
										if mountedSource.ID == sourceID {
											alreadyMounted = true
											break
										}
									}
									if alreadyMounted {
										continue
									}
									var source *Source
									for index := range selected.Sources {
										if selected.Sources[index].ID == sourceID {
											source = &selected.Sources[index]
											break
										}
									}
									if source == nil || rank(sourceLifecycle(*source)) < rank("callable") {
										sourceActivationError = "expert source is not callable: " + sourceID
										break
									}
									path, ok := ResolveCompleteSourceAt(selected.Root, *source)
									if !ok {
										sourceActivationError = "expert source is unavailable: " + sourceID
										break
									}
									extra := MountedSource{ID: sourceID, Path: path, Priority: sourcePriority(*source), Lifecycle: sourceLifecycle(*source), Entrypoint: source.Entrypoint, ActivationReason: "selected-expert:" + expert.ID}
									contracts, err := BuildSourceExecutionContracts(selected, []MountedSource{extra})
									if err != nil {
										sourceActivationError = err.Error()
										break
									}
									mounted = append(mounted, extra)
									sourceExecution = append(sourceExecution, contracts...)
									attachSourceExecution(workers, contracts)
								}
								if sourceActivationError != "" {
									break
								}
								for index := range workers {
									if workers[index].ID == "task-judgment" && selected.ID == "code" {
										continue // no implementation promise without verified code context
									}
									if err := bindExpertWorker(&workers[index], expert, commander, selected.ID, selection.CatalogSHA256); err != nil {
										sourceActivationError = err.Error()
										break
									}
									expertRoute.BoundWorkers = append(expertRoute.BoundWorkers, workers[index].ID)
								}
								break
							}
						}
					}
				}
			}
		}
	}
	if sourceActivationError != "" {
		workers = nil
		officerWorkers = nil
		delegation.Allowed = false
		delegation.ImplementationAllowed = false
		delegation.Reason = "selected-expert-contract-unavailable"
		if expertRoute != nil {
			expertRoute.BoundWorkers = nil
		}
	}
	if sourceActivationError == "" && !directAji && policy.RoutingMode != nonGPTProviderMode {
		for _, secondaryID := range secondaryWithoutResponsePolicy(secondary) {
			if secondaryID == selected.ID || secondaryID == "search" {
				continue
			}
			secondaryManifest, ok := capabilityManifest(manifests, secondaryID)
			if !ok || rank(secondaryManifest.Status) < rank("callable") {
				continue
			}
			// A secondary capability is one bounded branch of the already
			// composed route. Do not recursively compose it again, otherwise a
			// request such as image+video can grow a second-order route.
			secondaryRoute := routeSingleCapability(q, secondaryManifest, context, userSelectedModel)
			if secondaryRoute.SourceActivationError != "" {
				sourceActivationError = "secondary " + secondaryID + ": " + secondaryRoute.SourceActivationError
				break
			}
			for index := range secondaryRoute.Workers {
				workerID := "secondary-" + secondaryID + "-" + secondaryRoute.Workers[index].ID
				if err := rebaseWorkerTaskID(&secondaryRoute.Workers[index], workerID, q, context.Handle); err != nil {
					sourceActivationError = "secondary " + secondaryID + ": " + err.Error()
					break
				}
				workers = append(workers, secondaryRoute.Workers[index])
			}
			if sourceActivationError != "" {
				break
			}
			if secondaryRoute.ExpertRoute != nil {
				routeCopy := *secondaryRoute.ExpertRoute
				for index, workerID := range routeCopy.BoundWorkers {
					routeCopy.BoundWorkers[index] = "secondary-" + secondaryID + "-" + workerID
				}
				secondaryExpertRoutes[secondaryID] = routeCopy
			}
		}
	}
	if sourceActivationError != "" {
		workers = nil
		officerWorkers = nil
		delegation.Allowed = false
		delegation.ImplementationAllowed = false
		delegation.Reason = "secondary-route-unavailable"
	}
	if len(workers) == 0 {
		if expertRoute != nil {
			expertRoute.BoundWorkers = nil
			expertRoute.Experts = nil
		}
		for id, route := range secondaryExpertRoutes {
			route.BoundWorkers = nil
			route.Experts = nil
			secondaryExpertRoutes[id] = route
		}
	}
	// workerPlan may already have declined the fan-out because its original
	// replay estimate exceeded the hard limit. Do not overwrite that useful
	// evidence with a misleading zero after it clears the worker slice.
	if len(workers) > 0 {
		delegation.EstimatedReplayBytes = estimatedReplayBytes(workers)
		delegation.TaskContractBytes = workers[0].AllocatedTaskContractBytes
		delegation.TotalContractBytes = 0
		for _, worker := range workers {
			delegation.TotalContractBytes += worker.AllocatedTaskContractBytes
			if worker.AllocatedTaskContractBytes > delegation.TaskContractBytes {
				delegation.TaskContractBytes = worker.AllocatedTaskContractBytes
			}
		}
	}
	if len(workers) > 0 && delegation.EstimatedReplayBytes > maxTotalReplayBytes {
		workers = nil
		officerWorkers = nil
		delegation.Allowed = false
		delegation.ImplementationAllowed = false
		delegation.Reason = "estimated-context-replay-exceeds-total-budget"
		if expertRoute != nil {
			expertRoute.BoundWorkers = nil
		}
	}
	// The catalog is cold. A diagnostic selection is not a team activation,
	// and even a unique match must not expose the whole roster or staged SOP.
	// Only a successfully bound worker receives the selected expert's rules.
	if expertRoute != nil && len(expertRoute.BoundWorkers) > 0 && len(workers) > 0 {
		expertRoute.Experts = []string{expertRoute.Selection.ExpertID}
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
	staffRequired := !directAji && policy.RoutingMode != nonGPTProviderMode
	executionLane := executionLane(len(preflightWorkers), len(workers), staffRequired)
	if policy.RoutingMode == nonGPTProviderMode {
		executionLane = "provider-mode-passthrough"
	}
	reasoning, reasoningFallbacks := routeReasoning(q, staffRequired)
	whiteHat := whiteHatDecision(q, staffRequired, selected.ID, secondary, officers, sourceActivationError)
	policy.MainReasoningEffort = reasoning
	policy.MainFallbackReasoning = append([]string(nil), reasoningFallbacks...)
	brain := "aji-direct"
	if staffRequired {
		brain = "aji-with-deterministic-general-staff"
	}
	if policy.RoutingMode == nonGPTProviderMode {
		brain = "aji-provider-mode"
	}
	staffReason := directAjiReason(q)
	writeAuthority := "aji-scoped-artifact-write; no-unrequested-side-effects"
	if staffRequired {
		staffReason = "multi-step-cross-domain-high-risk-or-specialist-work"
		writeAuthority = "assigned-execution-nodes-only; scoped-artifact-write; staff-and-aji-read-only"
	}
	roleGraph, roleGraphs, graphPreflight, graphWorkers, graphOfficerWorkers, graphErr := makeRoleGraphSet(q, selected.ID, secondaryWithoutResponsePolicy(secondary), directAji, context, expertRoute, secondaryExpertRoutes, preflightWorkers, workers, officerWorkers)
	if graphErr != nil {
		sourceActivationError = "role graph: " + graphErr.Error()
		graphPreflight, graphWorkers, graphOfficerWorkers = nil, nil, nil
		delegation.Allowed = false
		delegation.ImplementationAllowed = false
		delegation.Reason = "role-graph-invalid"
		if expertRoute != nil {
			expertRoute.BoundWorkers = nil
			expertRoute.Experts = nil
		}
		for id, route := range secondaryExpertRoutes {
			route.BoundWorkers = nil
			route.Experts = nil
			secondaryExpertRoutes[id] = route
		}
	}
	preflightWorkers, workers, officerWorkers = graphPreflight, graphWorkers, graphOfficerWorkers
	if len(workers) > 0 {
		delegation.EstimatedReplayBytes = estimatedReplayBytes(workers)
		delegation.TaskContractBytes = workers[0].AllocatedTaskContractBytes
		delegation.TotalContractBytes = 0
		for _, worker := range workers {
			delegation.TotalContractBytes += worker.AllocatedTaskContractBytes
			if worker.AllocatedTaskContractBytes > delegation.TaskContractBytes {
				delegation.TaskContractBytes = worker.AllocatedTaskContractBytes
			}
		}
	}
	parallel = len(roleGraph.ParallelGroups) > 0
	return RouteResult{
		Version:               routeVersion,
		Brain:                 brain,
		MainModel:             policy.MainModel,
		MainReasoningEffort:   reasoning,
		MainFallbackReasoning: reasoningFallbacks,
		GeneralStaffModel:     policy.GeneralStaffModel,
		GeneralStaffRequired:  staffRequired,
		GeneralStaffReason:    staffReason,
		ModelPolicy:           policy,
		TaskIntent:            buildAjiTaskIntent(q, selected, secondary, officers, searchFirst, staffRequired),
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
			TaskShape: taskShape(staffRequired), ModelSelectionTiming: "once-at-task-start", SessionAffinity: "sticky-per-worker",
			EscalationPolicy: escalationPolicy(staffRequired), MaxModelSwitches: 0,
			DowngradeAfterGeneration: false, PreflightBeforeExecution: len(preflightWorkers) > 0,
		},
		SearchFirstPolicy:       searchFirst,
		ChangeCapsule:           changeCapsuleGate(q, selected),
		Reasoning:               reasoning,
		WriteAuthority:          writeAuthority,
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
		MoE:                     "sparse-role-moe",
		RoleGraph:               roleGraph,
		RoleGraphs:              roleGraphs,
		GeneralStaffWorker:      nil,
		Parallel:                parallel,
		PreflightWorkers:        preflightWorkers,
		Workers:                 workers,
		ExpertRoute:             expertRoute,
		SecondaryExpertRoutes:   secondaryExpertRoutes,
		Officers:                officers,
		OfficerRecommendations:  officerRecommendations,
		OfficerWorkers:          officerWorkers,
		InternalAdversarialPass: len(officers) == 0 && needsInternalChallenge(q),
		WhiteHat:                whiteHat,
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

func whiteHatDecision(query string, staffRequired bool, capability string, secondary, officers []string, routeError string) WhiteHatDecision {
	checks := []string{"premise-and-scope", "user-constraints", "side-effects", "evidence-boundary"}
	concerns := []string{}
	corrections := []string{}
	if staffRequired {
		checks = append(checks, "delegation-necessity", "completion-evidence")
	}
	if len(secondary) > 0 {
		concerns = append(concerns, "composed-task-dependency")
	}
	if len(officers) > 0 || needsInternalChallenge(query) {
		concerns = append(concerns, "independent-review-needed")
	}
	if routeError != "" {
		concerns = append(concerns, "route-entrypoint-unavailable")
		corrections = append(corrections, "do-not-claim-completion")
	}
	if capability == "security" || capability == "code-review" {
		concerns = append(concerns, "elevated-risk-domain")
	}
	status := "checked"
	if len(concerns) > 0 {
		status = "checked-with-concerns"
	}
	return WhiteHatDecision{
		Required:    true,
		Status:      status,
		Checks:      uniqueStrings(checks),
		Concerns:    uniqueStrings(concerns),
		Corrections: uniqueStrings(corrections),
		Escalate:    len(concerns) > 0,
	}
}

func routeSingleCapability(query string, capability Manifest, context DelegationContext, userSelectedModel string) RouteResult {
	clone := capability
	clone.Triggers = append([]string(nil), capability.Triggers...)
	// A one-item manifest cannot discover another capability, so this call
	// remains a single bounded branch rather than a recursive composition.
	return RouteWithContextModelAndResponseState(query, []Manifest{clone}, context, userSelectedModel, false)
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
		workers[index].SourceExecution = append(workers[index].SourceExecution, contracts...)
		for _, contract := range contracts {
			workers[index].SourceExecutionBytes += contract.EntrypointBytes
		}
		if !containsString(workers[index].PromptOrder, "source_execution") {
			workers[index].PromptOrder = append([]string{"stable_capability_prefix", "source_execution"}, workers[index].PromptOrder[1:]...)
		}
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

func secondaryWithoutResponsePolicy(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value != responsePolicyCapabilityID {
			result = append(result, value)
		}
	}
	return result
}

func directAjiReason(query string) string {
	if isSimpleQuestion(query) {
		return "simple-question-direct"
	}
	return "small-task-direct"
}

func taskShape(staffRequired bool) string {
	if staffRequired {
		return "staff-routed"
	}
	return "aji-direct"
}

func escalationPolicy(staffRequired bool) string {
	if staffRequired {
		return "availability-only-fallback"
	}
	return "none"
}

func executionLane(preflightCount, workerCount int, staffRequired bool) string {
	if preflightCount > 0 {
		return "bounded-search-first"
	}
	if workerCount > 0 {
		return "bounded-delegation"
	}
	if staffRequired {
		return "general-staff"
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
		if containsAny(query, "做课", "课程设计", "课程大纲", "教学大纲", "培训教材", "学习资料", "自测题", "学习辅导", "知识库问答", "飞书知识库", "金山文档知识库") ||
			(containsAny(query, "课件", "courseware") && !containsAny(query, "美化", "修改现有", "排版已有")) {
			boost += 65
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
	reasoning, reasoningFallbacks := reasoningSpec(modelClass)
	if len(fallbacks) > 0 || len(reasoningFallbacks) > 0 {
		availabilityFallbackOn = []string{"model-unavailable", "provider-error-before-generation"}
	}
	return WorkerTask{
		ID:                          id,
		Stage:                       "execution",
		Purpose:                     purpose,
		ModelClass:                  modelClass,
		Model:                       model,
		ReasoningEffort:             reasoning,
		AvailabilityFallbackEfforts: append([]string(nil), reasoningFallbacks...),
		AvailabilityFallbackModels:  append([]string(nil), fallbacks...),
		AvailabilityFallbackOn:      availabilityFallbackOn,
		FallbackModels:              nil,
		SessionKey:                  sessionKey,
		SessionAffinity:             "sticky-per-worker",
		EscalationPolicy:            "availability-only-fallback",
		MaxModelSwitches:            0,
		Inputs:                      append([]string(nil), inputs...),
		Protocol:                    append([]string(nil), protocol...),
		TaskContract:                contract,
		TaskContractSHA256:          sha256Hex([]byte(contract)),
		ContextMode:                 mode,
		ContextHandles:              handles,
		ContextArtifact:             artifact,
		ContextPayload:              payload,
		ContextPayloadSHA256:        payloadHash,
		StableCapabilityPrefix:      stablePrefix,
		StablePrefixSHA256:          sha256Hex([]byte(stablePrefix)),
		StablePrefixBytes:           len([]byte(stablePrefix)),
		PromptOrder:                 promptOrder(mode),
		AllocatedContextBytes:       allocated,
		AllocatedTaskContractBytes:  len([]byte(contract)),
		MaxTaskContractBytes:        maxTaskContractBytes,
		DelegationGateReason:        reason,
		MaxAttempts:                 1,
		FallbackOn:                  nil,
		Writes:                      writes,
		ExecutionEvidenceRequired:   true,
		ExecutionEvidenceFields:     append([]string(nil), workerExecutionEvidenceFields...),
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
	prefix.ImplementationDoctrine = ponyTailDoctrine
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
	Schema          string               `json:"schema"`
	Objective       string               `json:"objective"`
	Branch          string               `json:"branch"`
	Purpose         string               `json:"purpose"`
	RoleID          string               `json:"role_id,omitempty"`
	TaskGraphSHA256 string               `json:"graph_sha256,omitempty"`
	ParentGraphID   string               `json:"parent_graph_id,omitempty"`
	TaskGraphID     string               `json:"task_graph_id,omitempty"`
	Boundaries      []string             `json:"boundaries"`
	Acceptance      []string             `json:"acceptance"`
	Protocol        []string             `json:"protocol,omitempty"`
	ContextHandle   string               `json:"context_handle,omitempty"`
	SessionKey      string               `json:"session_key"`
	WriteBoundary   string               `json:"write_boundary"`
	Expert          *expertTaskDirective `json:"expert,omitempty"`
}

type expertTaskDirective struct {
	Commander      string                `json:"commander"`
	Capability     string                `json:"capability"`
	ID             string                `json:"id"`
	CatalogSHA256  string                `json:"catalog_sha256"`
	PromptCompiler string                `json:"prompt_compiler"`
	Inputs         []string              `json:"inputs,omitempty"`
	Outputs        []string              `json:"outputs,omitempty"`
	Workflow       []string              `json:"workflow"`
	Verification   []string              `json:"verification"`
	Acceptance     []string              `json:"acceptance"`
	Constraints    []string              `json:"constraints"`
	ToolPolicy     []string              `json:"tool_policy"`
	PonyTailRules  []string              `json:"ponytail_rules"`
	WhiteHatChecks []string              `json:"white_hat_checks"`
	TeamMission    string                `json:"team_mission,omitempty"`
	TeamAcceptance []string              `json:"team_acceptance,omitempty"`
	Methods        []expertAppliedMethod `json:"methods,omitempty"`
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
		"answer directly or take no action when sufficient; otherwise use the smallest correct action",
		"reuse existing Skill, plugin, MCP, template, dependency, native tool, or artifact first",
		"one line before many; simple before complex; reason only as much as risk requires",
		"reject wrong premises and unrequested scope; never skip required safety, facts, or verification",
		"report concrete result evidence and side effects, not an unsupported completion claim",
	}
	if strings.Contains(value, "review") || strings.Contains(value, "审查") || strings.Contains(value, "评审") {
		protocol = append(protocol, "separate specification conformance from engineering quality", "anchor findings to concrete files, symbols, or evidence", "rank findings by user impact and likelihood", "do not report stylistic preference as a defect")
	}
	if hasPonytailCodeDoctrine(stablePrefix) {
		protocol = append(protocol,
			"trace affected flow and callers; fix the shared root cause once",
			"prefer skip, local reuse, standard library, platform, dependency, then minimum code",
			"prefer deletion, fewest files and smallest diff; no unrequested abstraction",
			"check nontrivial logic with the smallest runnable regression; preserve safety and requirements",
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
	return strings.HasPrefix(prefix.ImplementationDoctrine, "ponytail-v3:") &&
		(prefix.Capability == "code" || prefix.Capability == "code-review")
}

func taskSessionKey(query, workerID, contextHandle string) string {
	payload := strings.Join([]string{"wuji-task-session-v1", strings.TrimSpace(query), workerID, contextHandle}, "\n")
	return "wuji-session://sha256/" + sha256Hex([]byte(payload))
}

func rebaseWorkerTaskID(worker *WorkerTask, id, query, contextHandle string) error {
	if worker == nil || strings.TrimSpace(id) == "" {
		return fmt.Errorf("secondary worker identity is required")
	}
	var contract workerContract
	if err := json.Unmarshal([]byte(worker.TaskContract), &contract); err != nil {
		return err
	}
	worker.ID = id
	worker.SessionKey = taskSessionKey(query, id, contextHandle)
	contract.Branch = id
	contract.SessionKey = worker.SessionKey
	data, err := json.Marshal(contract)
	if err != nil {
		return err
	}
	if len(data) > maxTaskContractBytes {
		return fmt.Errorf("worker %s task contract exceeds %d bytes", id, maxTaskContractBytes)
	}
	worker.TaskContract = string(data)
	worker.TaskContractSHA256 = sha256Hex(data)
	worker.AllocatedTaskContractBytes = len(data)
	return nil
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
	if strings.Contains(query, "文案") && containsAny(query, "改", "替换", "调整") {
		return true
	}
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
	if (strings.Contains(value, "提取") && strings.Contains(value, "字幕")) ||
		(strings.Contains(value, "extract") && strings.Contains(value, "subtitles")) {
		return true
	}
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
		"你觉得", "你认为", "怎么看", "如何看", "有什么看法", "what do you think", "how do you see",
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
