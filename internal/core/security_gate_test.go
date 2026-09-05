package core

import "testing"

func TestSecurityGateRequiresExplicitIntentAndWorkspaceContainment(t *testing.T) {
	workspace := t.TempDir()
	blocked := EvaluateSecurityGate(SecurityAction{Kind: "file-write", Target: "result.json", Workspace: workspace})
	if blocked.Allowed || blocked.Decision != "user-confirmation-required" {
		t.Fatalf("unconfirmed write was allowed: %#v", blocked)
	}
	allowed := EvaluateSecurityGate(SecurityAction{Kind: "file-write", Target: "result.json", Workspace: workspace, ExplicitUserIntent: true})
	if !allowed.Allowed || allowed.Decision != "allow-minimum-scope" {
		t.Fatalf("confirmed workspace write was blocked: %#v", allowed)
	}
	escape := EvaluateSecurityGate(SecurityAction{Kind: "delete", Target: "../outside.txt", Workspace: workspace, ExplicitUserIntent: true})
	if escape.Allowed || escape.Reason != "file target escapes declared workspace" {
		t.Fatalf("workspace escape was allowed: %#v", escape)
	}
}

func TestSecurityGateKeepsSecretsAndPrivilegeEscalationBlocked(t *testing.T) {
	for _, kind := range []string{"secret-handling", "privilege-escalation"} {
		result := EvaluateSecurityGate(SecurityAction{Kind: kind, Target: "provider", ExplicitUserIntent: true})
		if result.Allowed {
			t.Fatalf("%s was allowed: %#v", kind, result)
		}
	}
	local := EvaluateSecurityGate(SecurityAction{Kind: "local-analysis"})
	if !local.Allowed || local.Decision != "not-applicable" {
		t.Fatalf("local analysis should not require a gate: %#v", local)
	}
}
