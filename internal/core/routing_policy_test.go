package core

import (
	"strings"
	"testing"
)

func TestNonTrivialCodeTaskRunsBoundedSearchBeforeSol(t *testing.T) {
	query := "fix OAuth SDK timeout bug"
	items := []Manifest{
		{ID: "code", Triggers: []string{"fix", "bug", "sdk"}, Status: "callable", PrimarySkill: "native", Experts: []Expert{{ID: "implementation", Purpose: "implement", Independent: true, ModelClass: "sol"}}},
		{ID: "search", Triggers: []string{"research"}, Status: "callable", PrimarySkill: "wuji-research-suite", Engines: []Engine{{ID: "web-research", Default: true}}},
	}

	got := RouteWithContext(query, items, delegationContextForTest(query, 512))
	if got.GeneralStaffWorker != nil || got.ExecutionLane != "bounded-search-first" || len(got.PreflightWorkers) != 1 || len(got.Workers) != 1 {
		t.Fatalf("expected serial preflight followed by one implementation worker: %#v", got)
	}
	preflight := got.PreflightWorkers[0]
	if preflight.Stage != "preflight" || preflight.Model != hostSelectedModel || preflight.ReasoningEffort != "low" || preflight.MaxSources != 3 || preflight.TimeBudgetSeconds != 90 {
		t.Fatalf("search preflight is not bounded or executable: %#v", preflight)
	}
	if got.Workers[0].Stage != "execution" || got.Workers[0].Model != hostSelectedModel || got.Workers[0].ReasoningEffort != "max" {
		t.Fatalf("implementation was not routed to Sol: %#v", got.Workers)
	}
	if !got.SearchFirstPolicy.Required || !got.SearchFirstPolicy.CancelStaleExecutionPlan || !containsString(got.SecondaryCapabilities, "search") {
		t.Fatalf("search-first policy is incomplete: %#v", got.SearchFirstPolicy)
	}
	policy := got.TaskExecutionPolicy
	if policy.TaskShape != "staff-routed" || policy.ModelSelectionTiming != "once-at-task-start" || policy.SessionAffinity != "sticky-per-worker" || policy.MaxModelSwitches != 0 || policy.DowngradeAfterGeneration || !policy.PreflightBeforeExecution {
		t.Fatalf("task execution policy is incomplete: %#v", policy)
	}
	if preflight.SessionKey == "" || got.Workers[0].SessionKey == "" || preflight.SessionKey == got.Workers[0].SessionKey {
		t.Fatalf("workers did not receive stable, isolated session keys: preflight=%q workers=%#v", preflight.SessionKey, got.Workers)
	}
}

func TestDeterministicEditSkipsPriorArtSearch(t *testing.T) {
	got := Route("rename button label", nil)
	if got.GeneralStaffWorker != nil || got.SearchFirstPolicy.Required || len(got.PreflightWorkers) != 0 || got.TaskExecutionPolicy.PreflightBeforeExecution || got.ExecutionLane != "direct" || len(got.Workers) != 0 || got.DelegationDecision.Reason != "small-task-direct" {
		t.Fatalf("deterministic edit should use bounded task routing without web preflight: %#v", got)
	}
}

func TestMechanicalReadOnlyTaskUsesLuna(t *testing.T) {
	got := Route("list files and count lines", nil)
	if got.GeneralStaffWorker != nil || len(got.PreflightWorkers) != 0 || len(got.Workers) != 0 || got.ExecutionLane != "direct" || got.MainReasoningEffort != "low" {
		t.Fatalf("mechanical task did not stay on Aji: %#v", got)
	}
}

func TestExactLocalSkillLookupSkipsExternalAdmissionRouting(t *testing.T) {
	got := Route("find the skill named superpowers", nil)
	if got.SearchFirstPolicy.Required || len(got.PreflightWorkers) != 0 || len(got.Workers) != 0 || got.ExecutionLane != "direct" {
		t.Fatalf("exact local Skill lookup should stay on Aji: %#v", got)
	}
}

func TestExternalSkillInstallKeepsAdmissionGates(t *testing.T) {
	got := Route("install the external skill named superpowers from github", nil)
	if !got.SearchFirstPolicy.Required || len(got.PreflightWorkers) != 1 || !got.ChangeCapsule.Required || len(got.Workers) != 1 || got.Workers[0].ID != "task-execution" {
		t.Fatalf("external Skill installation lost preflight or mutation gates: %#v", got)
	}
}

func TestFlexibleFileListingStillUsesMechanicalLuna(t *testing.T) {
	got := Route("list the Go source files in the repository root", nil)
	if len(got.Workers) != 0 || got.ExecutionLane != "direct" {
		t.Fatalf("file listing with natural word order did not stay on Aji: %#v", got)
	}
}

func TestExplicitWebResearchDoesNotCreateNestedPreflight(t *testing.T) {
	items := []Manifest{{
		ID: "search", Triggers: []string{"research"}, Status: "callable", PrimarySkill: "wuji-research-suite",
		Engines: []Engine{{ID: "web-research", Default: true}},
	}}
	got := Route("research the web", items)
	if got.SearchFirstPolicy.Required || len(got.PreflightWorkers) != 0 || len(got.Workers) != 3 {
		t.Fatalf("explicit research should use its own bounded source workers: %#v", got)
	}
}

func TestOfflineRequestSuppressesPriorArtSearch(t *testing.T) {
	got := Route("debug this SDK error, offline only", nil)
	if got.SearchFirstPolicy.Required || len(got.PreflightWorkers) != 0 {
		t.Fatalf("offline request unexpectedly created a web preflight: %#v", got)
	}
}

func TestRetiredSuperpowersProtocolIsNotInjected(t *testing.T) {
	protocol := workerProtocol("debug timeout error", "task-judgment", "analyze the task", "")
	if len(protocol) == 0 {
		t.Fatalf("universal PonyTail protocol was not injected: %#v", protocol)
	}
}

func TestPonytailProtocolIsIncludedInCodeWorkerAndContract(t *testing.T) {
	query := "fix the shared parser root cause with the smallest correct change"
	items := []Manifest{{
		ID: "code", Triggers: []string{"fix"}, Status: "behavior-verified", PrimarySkill: "native Codex coding route",
		Experts: []Expert{{ID: "implementation", Purpose: "smallest complete implementation", Independent: true, ModelClass: "sol"}},
	}}
	worker := RouteWithContext(query, items, delegationContextForTest(query, 512)).Workers[0]
	for _, required := range []string{
		"trace affected flow and callers; fix the shared root cause once",
		"prefer skip, local reuse, standard library, platform, dependency, then minimum code",
		"prefer deletion, fewest files and smallest diff; no unrequested abstraction",
		"check nontrivial logic with the smallest runnable regression; preserve safety and requirements",
	} {
		if !containsString(worker.Protocol, required) || !strings.Contains(worker.TaskContract, required) {
			t.Fatalf("Ponytail requirement %q was not made executable: %#v", required, worker)
		}
	}
	if worker.StablePrefixBytes > 256 {
		t.Fatalf("Ponytail stable prefix must remain compact: %d bytes", worker.StablePrefixBytes)
	}
}

func TestPonytailAppliesAcrossConversationAndNonCodeExperts(t *testing.T) {
	for _, query := range []string{"你好", "你觉得这个方案怎么样？", "rename button label", "改一个按钮文案", "list files"} {
		route := Route(query, nil)
		if route.GeneralStaffRequired || len(route.Workers) != 0 || route.MainReasoningEffort != "low" || route.ModelPolicy.MainReasoningEffort != route.MainReasoningEffort {
			t.Fatalf("small task unnecessarily entered staff or used high reasoning: %q %#v", query, route)
		}
	}
	route := Route("制作一个产品演示视频", []Manifest{{ID: "video", Triggers: []string{"视频"}, Status: "callable", PrimarySkill: "video-skill"}})
	if !route.GeneralStaffRequired || len(route.Workers) != 1 || route.Workers[0].Model != hostSelectedModel {
		t.Fatalf("non-code expert was not routed: %#v", route)
	}
	worker := route.Workers[0]
	for _, rule := range []string{
		"answer directly or take no action when sufficient; otherwise use the smallest correct action",
		"one line before many; simple before complex; reason only as much as risk requires",
		"reject wrong premises and unrequested scope; never skip required safety, facts, or verification",
	} {
		if !containsString(worker.Protocol, rule) || !strings.Contains(worker.TaskContract, rule) {
			t.Fatalf("cross-domain PonyTail rule %q missing from worker contract", rule)
		}
	}
	if containsString(worker.Protocol, "trace affected flow and callers; fix the shared root cause once") {
		t.Fatal("code-only protocol was mounted on the video expert")
	}
	if worker.AllocatedTaskContractBytes > maxTaskContractBytes {
		t.Fatalf("PonyTail protocol exceeded bounded contract: %d", worker.AllocatedTaskContractBytes)
	}
}

func TestPonytailNeverBypassesRiskOrMultiStepRouting(t *testing.T) {
	for _, query := range []string{
		"rename button label and deploy to production",
		"改一个颜色并部署到生产",
		"list files and audit security",
		"what is the threat model for this production migration",
	} {
		route := Route(query, nil)
		if !route.GeneralStaffRequired || route.ExecutionLane == "direct" || route.MainReasoningEffort == "low" {
			t.Fatalf("risk or multi-step request bypassed staff: %q %#v", query, route)
		}
	}
}
