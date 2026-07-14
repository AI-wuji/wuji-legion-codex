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
	if got.ExecutionLane != "bounded-search-first" || len(got.PreflightWorkers) != 1 || len(got.Workers) != 2 {
		t.Fatalf("expected serial preflight followed by root-cause and evidence workers: %#v", got)
	}
	preflight := got.PreflightWorkers[0]
	if preflight.Stage != "preflight" || preflight.Model != "gpt-5.6-luna" || preflight.MaxSources != 3 || preflight.TimeBudgetSeconds != 90 {
		t.Fatalf("search preflight is not bounded or executable: %#v", preflight)
	}
	if got.Workers[0].ID != "root-cause" || got.Workers[0].Stage != "execution" || got.Workers[0].Model != "gpt-5.6-sol" || got.Workers[1].ID != "failure-evidence" || got.Workers[1].Model != "gpt-5.6-luna" {
		t.Fatalf("systematic failure branches were not routed to Sol and Luna: %#v", got.Workers)
	}
	if !got.SearchFirstPolicy.Required || !got.SearchFirstPolicy.CancelStaleExecutionPlan || !containsString(got.SecondaryCapabilities, "search") {
		t.Fatalf("search-first policy is incomplete: %#v", got.SearchFirstPolicy)
	}
	policy := got.TaskExecutionPolicy
	if policy.TaskShape != "small" || policy.ModelSelectionTiming != "once-at-task-start" || policy.SessionAffinity != "sticky-per-worker" || policy.MaxModelSwitches != 0 || policy.DowngradeAfterGeneration || !policy.PreflightBeforeExecution {
		t.Fatalf("task execution policy is incomplete: %#v", policy)
	}
	if preflight.SessionKey == "" || got.Workers[0].SessionKey == "" || preflight.SessionKey == got.Workers[0].SessionKey || got.Workers[0].SessionKey == got.Workers[1].SessionKey {
		t.Fatalf("workers did not receive stable, isolated session keys: preflight=%q workers=%#v", preflight.SessionKey, got.Workers)
	}
}

func TestDeterministicEditSkipsPriorArtSearch(t *testing.T) {
	got := Route("rename button label", nil)
	if got.SearchFirstPolicy.Required || len(got.PreflightWorkers) != 0 || got.TaskExecutionPolicy.PreflightBeforeExecution || got.ExecutionLane != "bounded-delegation" || len(got.Workers) != 1 || got.Workers[0].Model != "gpt-5.6-sol" {
		t.Fatalf("deterministic edit should use bounded task routing without web preflight: %#v", got)
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

func TestFlexibleFileListingStillUsesMechanicalLuna(t *testing.T) {
	got := Route("list the Go source files in the repository root", nil)
	if len(got.Workers) != 1 || got.Workers[0].ID != "mechanical" || got.Workers[0].Model != "gpt-5.6-luna" {
		t.Fatalf("file listing with natural word order did not use Luna: %#v", got)
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

func TestCommunityResearchDoesNotAccidentallyReceiveFailureProtocol(t *testing.T) {
	protocol := workerProtocol("research the web", "community", "independent reports and failure evidence", "")
	if len(protocol) != 0 {
		t.Fatalf("ordinary community research inherited a debugging protocol: %#v", protocol)
	}
}

func TestFailureProtocolRequiresEvidenceBeforeFixAndBoundsRepeatedGuesses(t *testing.T) {
	protocol := workerProtocol("debug timeout error", "root-cause", "", "")
	for _, required := range []string{
		"do not propose a fix before the observation and hypothesis are separately evidenced",
		"after two disproven hypotheses, stop iterative patching and return an architecture-reassessment decision",
		"name a fresh regression verification",
	} {
		if !containsString(protocol, required) {
			t.Fatalf("failure protocol omitted %q: %#v", required, protocol)
		}
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
		"trace the actual flow and cite affected file or symbol anchors before choosing",
		"choose the first valid rung: skip, reuse local code, standard library, native platform, installed dependency, one line, minimum code",
		"for bugs, inspect every caller and fix the common root cause once, not each symptom",
		"prefer deletion, fewest files, and the smallest correct diff; no unrequested abstraction, scaffolding, or dependency",
		"for nontrivial logic, name one smallest runnable regression check; trivial one-line edits need no new test",
		"do not weaken validation, error handling, data safety, security, accessibility, or explicit requirements",
	} {
		if !containsString(worker.Protocol, required) || !strings.Contains(worker.TaskContract, required) {
			t.Fatalf("Ponytail requirement %q was not made executable: %#v", required, worker)
		}
	}
	if worker.StablePrefixBytes > 256 {
		t.Fatalf("Ponytail stable prefix must remain compact: %d bytes", worker.StablePrefixBytes)
	}
}
