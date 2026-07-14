package core

import "strings"

// ChangeCapsule is a small, executable change-control artifact. It bounds a
// risky change without introducing a second planning workflow.
type ChangeCapsule struct {
	Schema              string   `json:"schema"`
	Intent              string   `json:"intent"`
	ScopeOut            string   `json:"scope_out"`
	AcceptanceScenarios []string `json:"acceptance_scenarios"`
	Verification        []string `json:"verification"`
	Rollback            string   `json:"rollback"`
}

// ChangeCapsuleIssue pinpoints one missing boundary without requiring a
// workflow engine or a second planning authority.
type ChangeCapsuleIssue struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ChangeCapsuleValidation is machine-readable evidence for the strict gate.
type ChangeCapsuleValidation struct {
	Schema string               `json:"schema"`
	Valid  bool                 `json:"valid"`
	Strict bool                 `json:"strict"`
	Issues []ChangeCapsuleIssue `json:"issues,omitempty"`
}

// ChangeCapsuleValidationResult keeps the original capsule and its gate
// result together when a caller explicitly asks for strict validation.
type ChangeCapsuleValidationResult struct {
	Capsule    ChangeCapsule           `json:"capsule"`
	Validation ChangeCapsuleValidation `json:"validation"`
}

func NewChangeCapsule(intent, scopeOut, acceptance, verification, rollback string) ChangeCapsule {
	return ChangeCapsule{
		Schema:              "wuji-change-capsule-v1",
		Intent:              strings.TrimSpace(intent),
		ScopeOut:            strings.TrimSpace(scopeOut),
		AcceptanceScenarios: splitCapsuleLines(acceptance),
		Verification:        splitCapsuleLines(verification),
		Rollback:            strings.TrimSpace(rollback),
	}
}

// ValidateChangeCapsule applies the bounded, executable subset of a change
// contract. Strict mode requires each decision boundary before a caller can
// use the capsule as a change gate.
func ValidateChangeCapsule(capsule ChangeCapsule, strict bool) ChangeCapsuleValidation {
	issues := []ChangeCapsuleIssue{}
	if strings.TrimSpace(capsule.Intent) == "" {
		issues = append(issues, ChangeCapsuleIssue{Field: "intent", Code: "required", Message: "intent is required"})
	}
	if strict && strings.TrimSpace(capsule.ScopeOut) == "" {
		issues = append(issues, ChangeCapsuleIssue{Field: "scope_out", Code: "required", Message: "explicit scope-out is required in strict mode"})
	}
	if strict && len(capsule.AcceptanceScenarios) == 0 {
		issues = append(issues, ChangeCapsuleIssue{Field: "acceptance_scenarios", Code: "required", Message: "at least one acceptance scenario is required in strict mode"})
	}
	if strict && len(capsule.Verification) == 0 {
		issues = append(issues, ChangeCapsuleIssue{Field: "verification", Code: "required", Message: "at least one verification command or evidence handle is required in strict mode"})
	}
	if strict && strings.TrimSpace(capsule.Rollback) == "" {
		issues = append(issues, ChangeCapsuleIssue{Field: "rollback", Code: "required", Message: "rollback boundary is required in strict mode"})
	}
	return ChangeCapsuleValidation{
		Schema: "wuji-change-capsule-validation-v1",
		Valid:  len(issues) == 0,
		Strict: strict,
		Issues: issues,
	}
}

func splitCapsuleLines(value string) []string {
	var items []string
	for _, item := range strings.FieldsFunc(value, func(r rune) bool { return r == '\n' || r == ';' }) {
		if item = strings.TrimSpace(item); item != "" {
			items = append(items, item)
		}
	}
	return items
}
