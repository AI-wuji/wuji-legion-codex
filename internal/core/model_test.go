package core

import (
	"strings"
	"testing"
)

func TestRouteEmitsExecutableModelPolicy(t *testing.T) {
	query := "code task parallel"
	items := []Manifest{{
		ID: "code", Triggers: []string{"code"}, Status: "callable", PrimarySkill: "native",
		Experts: []Expert{
			{ID: "implementation", Purpose: "implement", Independent: true, ModelClass: "terra"},
			{ID: "verification", Purpose: "verify", Independent: false, ModelClass: "luna"},
		},
	}}
	context := delegationContextForTest(query, 512)
	got := RouteWithContext(query, items, context)
	if got.Version != "3.0" || got.MainModel != "gpt-5.6-terra" || !equalStrings(got.ModelPolicy.MainFallbackModels, []string{"gpt-5.6-sol"}) || got.GeneralStaffModel != "" || got.GeneralStaffWorker != nil || got.ModelPolicy.RoutingMode != gptHierarchyMode || got.ModelPolicy.ClassModels["sol"] != "gpt-5.6-sol" || got.ModelPolicy.ClassModels["terra"] != "gpt-5.6-terra" {
		t.Fatalf("main model policy is incomplete: %#v", got.ModelPolicy)
	}
	if len(got.Workers) != 1 || got.Workers[0].Model != "gpt-5.6-terra" || !equalStrings(got.Workers[0].AvailabilityFallbackModels, []string{"gpt-5.6-sol"}) || len(got.Workers[0].FallbackModels) != 0 {
		t.Fatalf("route did not emit an executable Terra policy: %#v", got.Workers)
	}
	if got.GeneralStaffWorker != nil {
		t.Fatalf("route incorrectly exposed deterministic General Staff as a worker: %#v", got)
	}
	if got.Workers[0].ContextMode != "shared-content-addressed-handle" || got.ExecutionLane != "bounded-delegation" {
		t.Fatalf("Terra worker did not receive the bounded handoff: %#v", got)
	}
	worker := got.Workers[0]
	if worker.MaxAttempts != 1 || len(worker.FallbackOn) != 0 || worker.MaxModelSwitches != 0 || !equalStrings(worker.AvailabilityFallbackOn, []string{"model-unavailable", "provider-error-before-generation"}) {
		t.Fatalf("Terra worker did not separate availability selection from execution retries: %#v", worker)
	}
	if got.DelegationPolicy.CrossModelCacheAssumed || got.DelegationPolicy.CacheScope != "model-local stable-prefix only" || !got.DelegationPolicy.FallbackOnlyOnAvailabilityError {
		t.Fatalf("cross-model cache or fallback policy is unsafe: %#v", got.DelegationPolicy)
	}
}

func TestRouteKeepsGeneralStaffDeterministicAndDoesNotCreateAStaffWorker(t *testing.T) {
	got := Route("implement a bounded feature", []Manifest{{
		ID: "code", Triggers: []string{"implement", "feature"}, Status: "callable",
	}})
	if got.GeneralStaffWorker != nil || got.GeneralStaffModel != "" {
		t.Fatalf("route created a model-backed General Staff worker: %#v", got)
	}
	if got.ExecutionLane != "bounded-delegation" || !strings.Contains(got.WriteAuthority, "staff-and-aji-read-only") {
		t.Fatalf("general staff deterministic boundary was not preserved: %#v", got)
	}

	simple := Route("你是谁", nil)
	if simple.GeneralStaffWorker != nil || simple.ExecutionLane != "direct" {
		t.Fatalf("simple conversation should stay directly on Luna Aji: %#v", simple)
	}

	nonGPT := RouteWithModel("implement a bounded feature", nil, "grok-4")
	if nonGPT.GeneralStaffWorker != nil || nonGPT.ExecutionLane != "provider-mode-passthrough" {
		t.Fatalf("explicit non-GPT provider mode must not emit GPT General Staff: %#v", nonGPT)
	}
}

func TestSearchWorkersUseConcreteLunaModel(t *testing.T) {
	items := []Manifest{{
		ID: "search", Triggers: []string{"research"}, Status: "callable",
		Engines: []Engine{{ID: "web-research", Default: true}},
	}}
	got := Route("research the web", items)
	if len(got.Workers) != 3 {
		t.Fatalf("expected three research workers: %#v", got.Workers)
	}
	for _, worker := range got.Workers {
		if worker.Model != "gpt-5.6-luna" || !equalStrings(worker.AvailabilityFallbackModels, []string{"gpt-5.6-terra", "gpt-5.6-sol"}) || len(worker.FallbackModels) != 0 {
			t.Fatalf("research worker did not receive an executable Luna policy: %#v", worker)
		}
		if worker.MaxAttempts != 1 || len(worker.FallbackOn) != 0 || worker.MaxModelSwitches != 0 || !equalStrings(worker.AvailabilityFallbackOn, []string{"model-unavailable", "provider-error-before-generation"}) {
			t.Fatalf("research worker can retry after paid generation: %#v", worker)
		}
	}
}

func TestModelClassesAreExplicitlyValidated(t *testing.T) {
	model, fallbacks := modelSpec("unknown")
	if model != "" || fallbacks != nil {
		t.Fatalf("unknown model class must not silently consume Sol: model=%s fallbacks=%#v", model, fallbacks)
	}
	if model, _ := modelSpec("terra"); model != "gpt-5.6-terra" {
		t.Fatalf("Terra model class is not routable: %q", model)
	}
	if err := validateExperts([]Expert{{ID: "general-staff", Purpose: "normal implementation", ModelClass: "terra"}}); err != nil {
		t.Fatalf("manifest validation rejected Terra: %v", err)
	}
	if err := validateExperts([]Expert{{ID: "bad", Purpose: "invalid model", ModelClass: "unknown"}}); err == nil {
		t.Fatal("manifest validation accepted an unsupported model_class")
	}
}

func TestHighReasoningJudgmentUsesOneSolWorker(t *testing.T) {
	got := Route("architecture decision: use Sol", nil)
	if got.MainModel != "gpt-5.6-terra" || got.GeneralStaffModel != "" || got.GeneralStaffWorker != nil || len(got.Workers) != 1 {
		t.Fatalf("high-reasoning route is incomplete: %#v", got)
	}
	worker := got.Workers[0]
	if worker.ID != "sol-judgment" || worker.Model != "gpt-5.6-sol" || len(worker.FallbackModels) != 0 || len(worker.FallbackOn) != 0 || worker.MaxAttempts != 1 || worker.Writes {
		t.Fatalf("Sol must be a bounded read-only judgment worker: %#v", worker)
	}
}

func TestExplicitNonGPTModelPreservesProviderModeWithoutGPTWorkers(t *testing.T) {
	got := RouteWithModel("research the web", []Manifest{{ID: "search", Triggers: []string{"research"}, Status: "callable", Engines: []Engine{{ID: "web-research", Default: true}}}}, "grok-4")
	if got.Version != "3.0" || got.ModelPolicy.RoutingMode != nonGPTProviderMode || got.MainModel != "grok-4" || got.GeneralStaffModel != "" || got.ExecutionLane != "provider-mode-passthrough" {
		t.Fatalf("explicit non-GPT selection did not preserve provider mode: %#v", got)
	}
	if len(got.PreflightWorkers) != 0 || len(got.Workers) != 0 || len(got.OfficerWorkers) != 0 || got.DelegationDecision.Reason != nonGPTProviderMode {
		t.Fatalf("explicit non-GPT selection emitted GPT worker contracts: %#v", got)
	}
}

func equalStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for index := range want {
		if got[index] != want[index] {
			return false
		}
	}
	return true
}
