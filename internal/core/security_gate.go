package core

import (
	"path/filepath"
	"strings"
)

var controlledSecurityActions = map[string]bool{
	"network":              true,
	"external-download":    true,
	"skill-import":         true,
	"mcp-import":           true,
	"code-import":          true,
	"command-execute":      true,
	"file-write":           true,
	"account-access":       true,
	"secret-handling":      true,
	"privilege-escalation": true,
	"publish":              true,
	"payment":              true,
	"delete":               true,
}

// EvaluateSecurityGate is a local deterministic check that runs before a
// controlled side effect. It never invokes a model and never persists action
// details or secrets.
func EvaluateSecurityGate(action SecurityAction) SecurityGateResult {
	action.Kind = strings.ToLower(strings.TrimSpace(action.Kind))
	action.Target = strings.TrimSpace(action.Target)
	action.Workspace = strings.TrimSpace(action.Workspace)
	result := SecurityGateResult{Action: action, Decision: "blocked", Allowed: false, Checks: []string{"deterministic local gate; no model invoked"}}
	if action.Kind == "local-analysis" || action.Kind == "read" {
		result.Decision = "not-applicable"
		result.Allowed = true
		result.Checks = append(result.Checks, "no controlled side effect")
		return result
	}
	if !controlledSecurityActions[action.Kind] {
		result.Reason = "unsupported controlled action"
		return result
	}
	if action.Target == "" || len(action.Target) > 4096 || strings.ContainsAny(action.Target, "\r\n\x00") {
		result.Reason = "action target is required and must be a single bounded value"
		return result
	}
	if (action.Kind == "file-write" || action.Kind == "delete") && action.Workspace != "" && !pathWithinWorkspace(action.Target, action.Workspace) {
		result.Reason = "file target escapes declared workspace"
		return result
	}
	if action.Kind == "secret-handling" {
		result.Decision = "isolated-secret-handler-required"
		result.Reason = "secrets must not enter repository state or task context"
		return result
	}
	if action.Kind == "privilege-escalation" {
		result.Reason = "privilege escalation is blocked by deterministic policy"
		return result
	}
	if !action.ExplicitUserIntent {
		result.Decision = "user-confirmation-required"
		result.Reason = "controlled side effect lacks explicit user intent"
		return result
	}
	result.Decision = "allow-minimum-scope"
	result.Allowed = true
	result.Checks = append(result.Checks, "explicit user intent supplied", "no independent security review required by deterministic policy")
	return result
}

func pathWithinWorkspace(target, workspace string) bool {
	workspacePath, err := filepath.Abs(workspace)
	if err != nil {
		return false
	}
	targetPath := target
	if !filepath.IsAbs(targetPath) {
		targetPath = filepath.Join(workspacePath, targetPath)
	}
	targetPath, err = filepath.Abs(targetPath)
	if err != nil {
		return false
	}
	relative, err := filepath.Rel(workspacePath, targetPath)
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}
