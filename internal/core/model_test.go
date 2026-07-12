package core

import "testing"

func TestRouteEmitsExecutableModelPolicy(t *testing.T) {
	query := "code task parallel"
	items := []Manifest{{
		ID: "code", Triggers: []string{"code"}, Status: "callable", PrimarySkill: "native",
		Experts: []Expert{
			{ID: "implementation", Purpose: "implement", Independent: true, ModelClass: "terra"},
			{ID: "verification", Purpose: "verify", Independent: false, ModelClass: "terra"},
		},
	}}
	context := delegationContextForTest(query, 512)
	got := RouteWithContext(query, items, context)
	if got.MainModel != "gpt-5.6-terra" || got.ModelPolicy.ClassModels["terra"] != "gpt-5.6-terra" {
		t.Fatalf("main model policy is incomplete: %#v", got.ModelPolicy)
	}
	if len(got.Workers) != 1 || got.Workers[0].Model != "gpt-5.6-terra" || !equalStrings(got.Workers[0].FallbackModels, []string{"gpt-5.6-sol"}) {
		t.Fatalf("route did not emit an executable Terra policy: %#v", got.Workers)
	}
	if got.Workers[0].ContextMode != "shared-content-addressed-handle" || got.ExecutionLane != "bounded-delegation" {
		t.Fatalf("Terra worker did not receive the bounded handoff: %#v", got)
	}
	worker := got.Workers[0]
	if worker.MaxAttempts != 2 || !equalStrings(worker.FallbackOn, []string{"model-unavailable", "provider-error-before-generation"}) {
		t.Fatalf("Terra retry policy can spend on quality retries: %#v", worker)
	}
	if got.DelegationPolicy.CrossModelCacheAssumed || got.DelegationPolicy.CacheScope != "model-local stable-prefix only" || !got.DelegationPolicy.FallbackOnlyOnAvailabilityError {
		t.Fatalf("cross-model cache or fallback policy is unsafe: %#v", got.DelegationPolicy)
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
		if worker.Model != "gpt-5.6-luna" || !equalStrings(worker.FallbackModels, []string{"gpt-5.6-terra"}) {
			t.Fatalf("research worker did not receive an executable Luna policy: %#v", worker)
		}
		if worker.MaxAttempts != 2 || !equalStrings(worker.FallbackOn, []string{"model-unavailable", "provider-error-before-generation"}) {
			t.Fatalf("research worker can retry after paid generation: %#v", worker)
		}
	}
}

func TestUnknownModelClassIsNotSilentlyRoutedToSol(t *testing.T) {
	model, fallbacks := modelSpec("unknown")
	if model != "" || fallbacks != nil {
		t.Fatalf("unknown model class must not silently consume Sol: model=%s fallbacks=%#v", model, fallbacks)
	}
	if err := validateExperts([]Expert{{ID: "bad", Purpose: "invalid model", ModelClass: "terra-"}}); err == nil {
		t.Fatal("manifest validation accepted an unsupported model_class")
	}
}

func TestHighReasoningJudgmentUsesOneSolWorker(t *testing.T) {
	got := Route("architecture decision: use Sol", nil)
	if got.MainModel != "gpt-5.6-terra" || len(got.Workers) != 1 {
		t.Fatalf("high-reasoning route is incomplete: %#v", got)
	}
	worker := got.Workers[0]
	if worker.ID != "sol-judgment" || worker.Model != "gpt-5.6-sol" || len(worker.FallbackModels) != 0 || len(worker.FallbackOn) != 0 || worker.MaxAttempts != 1 || worker.Writes {
		t.Fatalf("Sol must be a bounded read-only, no-retry judgment worker: %#v", worker)
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
