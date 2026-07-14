package core

import "testing"

func TestChangeCapsuleKeepsOnlyExplicitBoundaries(t *testing.T) {
	got := NewChangeCapsule("replace routing adapter", "no provider migration", "route emits workers; receipt persists", "go test ./...; smoke route", "restore prior adapter")
	if got.Schema != "wuji-change-capsule-v1" || len(got.AcceptanceScenarios) != 2 || len(got.Verification) != 2 || got.ScopeOut != "no provider migration" {
		t.Fatalf("change capsule did not preserve its bounded contract: %#v", got)
	}
}

func TestChangeCapsuleStrictValidationRequiresEveryBoundary(t *testing.T) {
	invalid := ValidateChangeCapsule(NewChangeCapsule("replace adapter", "", "", "", ""), true)
	if invalid.Valid || len(invalid.Issues) != 4 || invalid.Schema != "wuji-change-capsule-validation-v1" {
		t.Fatalf("strict validation accepted an incomplete capsule: %#v", invalid)
	}
	valid := ValidateChangeCapsule(NewChangeCapsule("replace adapter", "no provider migration", "route passes", "go test ./...", "restore prior adapter"), true)
	if !valid.Valid || len(valid.Issues) != 0 || !valid.Strict {
		t.Fatalf("strict validation rejected a complete capsule: %#v", valid)
	}
}
