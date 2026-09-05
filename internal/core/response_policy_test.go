package core

import (
	"path/filepath"
	"testing"
)

func loadInteractionFixture(t *testing.T) ([]Manifest, Manifest) {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	items, err := LoadManifests(root)
	if err != nil {
		t.Fatal(err)
	}
	interaction, ok := capabilityManifest(items, responsePolicyCapabilityID)
	if !ok {
		t.Fatal("interaction capability is missing")
	}
	return items, interaction
}

func TestResponsePolicyIsCrossCapabilityOverlay(t *testing.T) {
	items, _ := loadInteractionFixture(t)
	route := RouteWithContextModelAndResponseState("开启专注执行模式，修复这个 bug", items, DelegationContext{}, "", false)
	if route.Capability != "code" {
		t.Fatalf("response policy displaced domain capability: %#v", route)
	}
	if route.ResponsePolicy == nil || !route.ResponsePolicy.Active || route.ResponsePolicy.ActivationReason != "explicit-activation" {
		t.Fatalf("response policy was not activated: %#v", route.ResponsePolicy)
	}
	if !containsString(route.SecondaryCapabilities, "interaction") || route.ResponsePolicy.RulesSHA256 == "" || route.ResponsePolicy.SourceCommit != "01ce5c3747f05d27e4565580254f8efebac7e60d" {
		t.Fatalf("response policy provenance/overlay is incomplete: %#v", route)
	}
}

func TestResponsePolicyCarriesAndExitsSessionState(t *testing.T) {
	items, _ := loadInteractionFixture(t)
	carried := RouteWithContextModelAndResponseState("继续修复 bug", items, DelegationContext{}, "", true)
	if carried.ResponsePolicy == nil || !carried.ResponsePolicy.Active || carried.ResponsePolicy.ActivationReason != "session-active" {
		t.Fatalf("active session was not carried: %#v", carried.ResponsePolicy)
	}
	exit := RouteWithContextModelAndResponseState("恢复正常模式", items, DelegationContext{}, "", true)
	if exit.ResponsePolicy == nil || exit.ResponsePolicy.Active || exit.ResponsePolicy.ActivationReason != "explicit-exit" || len(exit.ResponsePolicy.Directives) != 0 {
		t.Fatalf("explicit exit did not disable policy: %#v", exit.ResponsePolicy)
	}
}

func TestResponsePolicyRequiresAffirmativeExplicitIntent(t *testing.T) {
	items, interaction := loadInteractionFixture(t)
	for _, query := range []string{"do not enable action focus", "explain what action focus means", "我有 ADHD，请解释这种模式"} {
		contract, err := CompileResponsePolicy(interaction, query, false)
		if err != nil {
			t.Fatal(err)
		}
		if contract != nil {
			t.Fatalf("non-affirmative mention activated policy for %q: %#v", query, contract)
		}
	}
	carried := RouteWithContextModelAndResponseState("do not stop focus mode", items, DelegationContext{}, "", true)
	if carried.ResponsePolicy == nil || !carried.ResponsePolicy.Active || carried.ResponsePolicy.ActivationReason != "session-active" {
		t.Fatalf("negated exit disabled policy: %#v", carried.ResponsePolicy)
	}
}

func TestResponsePolicyCompilesConflictOverrides(t *testing.T) {
	_, interaction := loadInteractionFixture(t)
	explanation, err := CompileResponsePolicy(interaction, "开启专注模式，请解释为什么会失败", false)
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"first-action", "bounded-steps", "single-next-action"} {
		if hasResponseDirective(explanation, id) || !containsString(explanation.Suppressed, id) {
			t.Fatalf("explanation did not suppress %q: %#v", id, explanation)
		}
	}
	destructive, err := CompileResponsePolicy(interaction, "开启专注模式，删除这些不明确的文件", false)
	if err != nil {
		t.Fatal(err)
	}
	if hasResponseDirective(destructive, "first-action") || !hasResponseDirective(destructive, "destructive-clarity") {
		t.Fatalf("destructive ambiguity did not override action-first: %#v", destructive)
	}
}

func TestResponsePolicyFiltersConditionalAtoms(t *testing.T) {
	_, interaction := loadInteractionFixture(t)
	contract, err := CompileResponsePolicy(interaction, "enable action focus: what is the current Go version", false)
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"first-action", "bounded-steps", "progress-state", "cause-fix", "specific-estimates", "single-next-action"} {
		if hasResponseDirective(contract, id) {
			t.Fatalf("inapplicable conditional directive %q was compiled: %#v", id, contract)
		}
	}
	for _, id := range []string{"host-safety", "explicit-user-instruction", "topic-filter"} {
		if !hasResponseDirective(contract, id) {
			t.Fatalf("always-on directive %q was omitted: %#v", id, contract)
		}
	}
}

func TestResponsePolicyValidatesObservableBehavior(t *testing.T) {
	_, interaction := loadInteractionFixture(t)
	contract, err := CompileResponsePolicy(interaction, "开启专注执行模式，继续修复这个失败的多步骤任务", false)
	if err != nil {
		t.Fatal(err)
	}
	bad := ValidateResponseDraft(contract, ResponseDraft{
		StepCount: 7, ClosingActionCount: 2, ClosingActionMinutes: 10,
		Continuation: true, Error: true, ContainsOffTopic: true,
	})
	for _, want := range []string{"bounded-steps", "cause-fix", "first-action", "progress-state", "single-next-action", "topic-filter"} {
		if !containsString(bad, want) {
			t.Fatalf("observable violation %q was missed: %#v", want, bad)
		}
	}
	good := ValidateResponseDraft(contract, ResponseDraft{
		FirstLineAction: true, StepCount: 3, ClosingActionCount: 1, ClosingActionMinutes: 2,
		Continuation: true, RestatesCurrentState: true, Error: true, ErrorCausePresent: true, ErrorFixPresent: true,
	})
	if len(good) != 0 {
		t.Fatalf("compliant response was rejected: %#v", good)
	}
}

func TestResponsePolicySourceAuditReportsTrustedEntrypoint(t *testing.T) {
	_, interaction := loadInteractionFixture(t)
	entries := AuditSources([]Manifest{interaction})
	if len(entries) != 1 || entries[0].State != "auto-selectable" || !entries[0].EntrypointReachable || len(entries[0].EntrypointSHA256) != 64 {
		t.Fatalf("trusted response-policy entrypoint was not reported accurately: %#v", entries)
	}
}

func hasResponseDirective(contract *ResponsePolicyContract, id string) bool {
	if contract == nil {
		return false
	}
	for _, directive := range contract.Directives {
		if directive.ID == id {
			return true
		}
	}
	return false
}
