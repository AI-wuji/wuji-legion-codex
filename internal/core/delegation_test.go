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
	if worker.AllocatedContextBytes != 1024 || worker.AllocatedTaskContractBytes != len([]byte(worker.TaskContract)) || len(worker.ContextHandles) != 1 {
		t.Fatalf("worker handoff costs are incomplete: %#v", worker)
	}
	if worker.TaskContractSHA256 != sha256Hex([]byte(worker.TaskContract)) || worker.ContextPayload != context.Payload || worker.ContextPayloadSHA256 != context.PayloadSHA256 {
		t.Fatalf("worker prompt payload is not content-addressed: %#v", worker)
	}
	if delegated.DelegationDecision.EstimatedReplayBytes != worker.StablePrefixBytes+1024+len([]byte(worker.TaskContract)) || len(worker.ExecutionEvidenceFields) != len(workerExecutionEvidenceFields) {
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
		{"unverified context struct", query, func() DelegationContext {
			value := delegationContextForTest(query, 512)
			value.verified = false
			return value
		}(), "verified-context-artifact-required"},
		{"oversized context", query, delegationContextForTest(query, maxSharedContextBytes+1), "shared-context-exceeds-per-worker-budget"},
		{"query mismatch", query, delegationContextForTest("another code task", 512), "context-query-fingerprint-mismatch"},
		{"invalid payload", query, func() DelegationContext {
			value := delegationContextForTest(query, 512)
			value.PayloadSHA256 = "bad"
			return value
		}(), "verified-context-artifact-required"},
		{"low coverage", query, func() DelegationContext {
			value := delegationContextForTest(query, 512)
			value.CoverageBPS = minContextCoverageBPS - 1
			return value
		}(), "context-coverage-below-delegation-threshold"},
		{"no code excerpt", query, func() DelegationContext {
			value := delegationContextForTest(query, 512)
			value.CodeExcerptCount = 0
			return value
		}(), "code-context-excerpt-required"},
		{"no content anchor", query, func() DelegationContext {
			value := delegationContextForTest(query, 512)
			value.ContentAnchorCount = 0
			return value
		}(), "code-content-anchor-required"},
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
	prefixBytes := len([]byte(stableCapabilityPrefix(items[0], "")))
	want := prefixBytes*2 + maxSharedContextBytes*2 + len([]byte(marshalWorkerContract(query, "narrative", "", []string{context.Handle}, taskSessionKey(query, "narrative", context.Handle)))) + len([]byte(marshalWorkerContract(query, "visual", "", []string{context.Handle}, taskSessionKey(query, "visual", context.Handle))))
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
	prefix := "WUJI_CONTEXT_CAPSULE_V1\n"
	if selectedBytes < len([]byte(prefix)) {
		selectedBytes = len([]byte(prefix))
	}
	payload := prefix + strings.Repeat("x", selectedBytes-len([]byte(prefix)))
	return DelegationContext{
		Handle: "wuji-context://sha256/test", ArtifactPath: "test.json",
		QueryFingerprint: queryFingerprint(queryTerms(query)), SelectedBytes: selectedBytes,
		CoverageBPS: 10000, CodeExcerptCount: 1, ContentAnchorCount: 1, Payload: payload, PayloadSHA256: sha256Hex([]byte(payload)), SelfContained: true, verified: true,
	}
}
