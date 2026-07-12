package core

import "testing"

func TestNonTrivialCodeTaskRunsBoundedSearchBeforeTerra(t *testing.T) {
	query := "fix OAuth SDK timeout bug"
	items := []Manifest{
		{ID: "code", Triggers: []string{"fix", "bug", "sdk"}, Status: "callable", PrimarySkill: "native", Experts: []Expert{{ID: "implementation", Purpose: "implement", Independent: true, ModelClass: "terra"}}},
		{ID: "search", Triggers: []string{"research"}, Status: "callable", PrimarySkill: "wuji-research-suite", Engines: []Engine{{ID: "web-research", Default: true}}},
	}

	got := RouteWithContext(query, items, delegationContextForTest(query, 512))
	if got.ExecutionLane != "bounded-search-first" || len(got.PreflightWorkers) != 1 || len(got.Workers) != 1 {
		t.Fatalf("expected serial preflight followed by one implementation worker: %#v", got)
	}
	preflight := got.PreflightWorkers[0]
	if preflight.Stage != "preflight" || preflight.Model != "gpt-5.6-luna" || preflight.MaxSources != 3 || preflight.TimeBudgetSeconds != 90 {
		t.Fatalf("search preflight is not bounded or executable: %#v", preflight)
	}
	if got.Workers[0].Stage != "execution" || got.Workers[0].Model != "gpt-5.6-terra" {
		t.Fatalf("implementation was not routed to Terra: %#v", got.Workers)
	}
	if !got.SearchFirstPolicy.Required || !got.SearchFirstPolicy.CancelStaleExecutionPlan || !containsString(got.SecondaryCapabilities, "search") {
		t.Fatalf("search-first policy is incomplete: %#v", got.SearchFirstPolicy)
	}
	policy := got.TaskExecutionPolicy
	if policy.TaskShape != "small" || policy.ModelSelectionTiming != "once-at-task-start" || policy.SessionAffinity != "sticky-per-worker" || policy.MaxModelSwitches != 1 || policy.DowngradeAfterGeneration || !policy.PreflightBeforeExecution {
		t.Fatalf("task execution policy is incomplete: %#v", policy)
	}
	if preflight.SessionKey == "" || got.Workers[0].SessionKey == "" || preflight.SessionKey == got.Workers[0].SessionKey {
		t.Fatalf("workers did not receive stable, isolated session keys: preflight=%q implementation=%q", preflight.SessionKey, got.Workers[0].SessionKey)
	}
}

func TestDeterministicEditSkipsPriorArtSearch(t *testing.T) {
	got := Route("rename button label", nil)
	if got.SearchFirstPolicy.Required || len(got.PreflightWorkers) != 0 || got.TaskExecutionPolicy.PreflightBeforeExecution || got.ExecutionLane != "small-task-direct" {
		t.Fatalf("deterministic edit should stay direct: %#v", got)
	}
}

func TestMechanicalReadOnlyTaskUsesLuna(t *testing.T) {
	got := Route("list files and count lines", nil)
	if len(got.PreflightWorkers) != 0 || len(got.Workers) != 1 || got.ExecutionLane != "bounded-delegation" {
		t.Fatalf("mechanical task did not produce one bounded worker: %#v", got)
	}
	worker := got.Workers[0]
	if worker.ID != "mechanical" || worker.Model != "gpt-5.6-luna" || worker.ContextMode != "task-contract-only" || worker.Writes {
		t.Fatalf("mechanical task did not use a read-only Luna worker: %#v", worker)
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
