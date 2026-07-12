package core

import (
	"strings"
	"testing"
)

func TestCodeDelegationRequiresVerifiedBoundedContext(t *testing.T) {
	query := "fix the code and verify it"
	items := []Manifest{{
		ID: "code", Triggers: []string{"code", "fix"}, Status: "callable",
		Experts: []Expert{
			{ID: "implementation", Purpose: "implement", Independent: true, ModelClass: "terra"},
			{ID: "verification", Purpose: "verify", Independent: false, ModelClass: "terra"},
		},
	}}

	direct := Route(query, items)
	if len(direct.Workers) != 0 || direct.DelegationDecision.Reason != "verified-context-artifact-required" {
		t.Fatalf("code delegated without verified context: %#v", direct)
	}

	context := delegationContextForTest(query, 1024)
	delegated := RouteWithContext(query, items, context)
	if len(delegated.Workers) != 1 || delegated.Workers[0].ID != "implementation" || delegated.Workers[0].Model != "gpt-5.6-terra" {
		t.Fatalf("expected one independent Terra implementation worker: %#v", delegated.Workers)
	}
	worker := delegated.Workers[0]
	if worker.AllocatedContextBytes != 1024 || worker.AllocatedTaskContractBytes != len([]byte(query)) || len(worker.ContextHandles) != 1 {
		t.Fatalf("worker handoff costs are incomplete: %#v", worker)
	}
	if delegated.DelegationDecision.EstimatedReplayBytes != 1024+len([]byte(query)) || len(worker.ExecutionEvidenceFields) != 8 {
		t.Fatalf("delegation cost/evidence contract is incomplete: %#v", delegated)
	}
}

func TestDelegationCostAndAffinityGatesStayOnAji(t *testing.T) {
	query := "fix the code"
	items := []Manifest{{
		ID: "code", Triggers: []string{"code"}, Status: "callable",
		Experts: []Expert{{ID: "implementation", Independent: true, ModelClass: "terra"}},
	}}
	tests := []struct {
		name    string
		query   string
		context DelegationContext
		reason  string
	}{
		{"parent context", query, DelegationContext{ParentContextRequired: true}, "parent-context-affinity-requires-Aji"},
		{"oversized context", query, delegationContextForTest(query, maxSharedContextBytes+1), "shared-context-exceeds-per-worker-budget"},
		{"query mismatch", query, delegationContextForTest("another code task", 512), "context-query-fingerprint-mismatch"},
		{"oversized contract", "code " + strings.Repeat("x", maxTaskContractBytes), DelegationContext{}, "task-contract-exceeds-budget"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := RouteWithContext(test.query, items, test.context)
			if len(got.Workers) != 0 || got.DelegationDecision.Allowed || got.DelegationDecision.Reason != test.reason {
				t.Fatalf("gate failed open: %#v", got)
			}
		})
	}
}

func TestTotalReplayBudgetIncludesContractAndContextPerWorker(t *testing.T) {
	query := "ppt parallel"
	items := []Manifest{{
		ID: "presentation", Triggers: []string{"ppt"}, Status: "callable",
		Experts: []Expert{
			{ID: "narrative", Independent: true, ModelClass: "terra"},
			{ID: "visual", Independent: true, ModelClass: "terra"},
		},
	}}
	context := delegationContextForTest(query, maxSharedContextBytes)
	got := RouteWithContext(query, items, context)
	if len(got.Workers) != 0 || got.DelegationDecision.Reason != "estimated-context-replay-exceeds-total-budget" {
		t.Fatalf("total replay gate failed open: %#v", got)
	}
	want := (maxSharedContextBytes + len([]byte(query))) * 2
	if got.DelegationDecision.EstimatedReplayBytes != want {
		t.Fatalf("wrong replay estimate: got %d want %d", got.DelegationDecision.EstimatedReplayBytes, want)
	}
}

func TestSearchSerialGatePrecedesLunaFanout(t *testing.T) {
	items := []Manifest{{
		ID: "search", Triggers: []string{"research"}, Status: "callable",
		Engines: []Engine{{ID: "web-research", Default: true}},
	}}
	serial := Route("research the web serial only", items)
	if len(serial.Workers) != 0 || serial.DelegationDecision.Reason != "serial-requested" {
		t.Fatalf("serial research still fanned out: %#v", serial)
	}
	normal := Route("research the web", items)
	if len(normal.Workers) != 3 || normal.DelegationDecision.Reason != "task-contract-only-research" {
		t.Fatalf("bounded Luna research fanout was lost: %#v", normal)
	}
	for _, worker := range normal.Workers {
		if worker.ContextMode != "task-contract-only" || worker.AllocatedContextBytes != 0 || worker.AllocatedTaskContractBytes == 0 {
			t.Fatalf("research worker received an unbounded handoff: %#v", worker)
		}
	}
}

func delegationContextForTest(query string, selectedBytes int) DelegationContext {
	return DelegationContext{
		Handle: "wuji-context://sha256/test", ArtifactPath: "test.json",
		QueryFingerprint: queryFingerprint(queryTerms(query)), SelectedBytes: selectedBytes,
	}
}
