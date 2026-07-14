package core

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRouteCreatesAndDispatchPreparesSourceExecutionContract(t *testing.T) {
	root := t.TempDir()
	sourceRoot := filepath.Join(root, "scenario")
	if err := os.MkdirAll(sourceRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	entrypoint := filepath.Join(sourceRoot, "SKILL.md")
	const body = "# Focused scenario\nUse the dedicated validation path."
	if err := os.WriteFile(entrypoint, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	manifest := Manifest{
		ID: "focused", Root: root, Triggers: []string{"focused scenario"}, Status: "callable", PrimarySkill: "focused-skill",
		Sources: []Source{{ID: "focused-scenario", Priority: "primary", Lifecycle: "callable", Entrypoint: "SKILL.md", Globs: []string{"${ROOT}/scenario"}, Required: []string{"SKILL.md"}}},
	}
	route := Route("perform the focused scenario", []Manifest{manifest})
	if len(route.SourceExecution) != 1 || route.SourceExecution[0].Entrypoint != "SKILL.md" || route.SourceExecution[0].EntrypointContent != body {
		t.Fatalf("route did not create a loaded source contract: %#v", route.SourceExecution)
	}
	if len(route.Workers) != 1 || len(route.Workers[0].SourceExecution) != 1 {
		t.Fatalf("worker did not receive selected source contract: %#v", route.Workers)
	}
	result, err := DispatchWorker(route.Workers[0], DispatchOptions{
		Workspace: root, OutputDir: filepath.Join(root, "output"), TrustedManifests: []Manifest{manifest},
		Runner: func(_ context.Context, _ string, arguments []string) (CodexCommandResult, error) {
			t.Fatalf("native-host preparation must not invoke external CLI: %#v", arguments)
			return CodexCommandResult{}, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.PreparedPromptSHA256 == "" || len(result.SourceContracts) != 1 || result.SourceContracts[0].EntrypointSHA256 != route.SourceExecution[0].EntrypointSHA256 {
		t.Fatalf("source contract was not kept distinct from execution evidence: %#v", result)
	}
}

func TestDispatchRejectsChangedSourceEntrypoint(t *testing.T) {
	root := t.TempDir()
	entrypoint := filepath.Join(root, "SKILL.md")
	if err := os.WriteFile(entrypoint, []byte("original"), 0o600); err != nil {
		t.Fatal(err)
	}
	manifest := Manifest{ID: "focused", Root: root, Sources: []Source{{ID: "focused", Priority: "primary", Lifecycle: "callable", Entrypoint: "SKILL.md", Globs: []string{"${ROOT}"}, Required: []string{"SKILL.md"}}}}
	contracts, err := BuildSourceExecutionContracts(manifest, []MountedSource{{ID: "focused", Entrypoint: "SKILL.md", ActivationReason: "primary-source"}})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(entrypoint, []byte("changed"), 0o600); err != nil {
		t.Fatal(err)
	}
	worker := testReceiptWorker()
	worker.SourceExecution = contracts
	if _, err := DispatchWorker(worker, DispatchOptions{Workspace: root, OutputDir: filepath.Join(root, "output"), DryRun: true, TrustedManifests: []Manifest{manifest}}); err == nil || !strings.Contains(err.Error(), "changed after routing") {
		t.Fatalf("dispatch accepted a changed selected entrypoint: %v", err)
	}
}

func TestEntrypointVerificationDoesNotBecomeReceiptExecutionEvidence(t *testing.T) {
	root := t.TempDir()
	entrypoint := filepath.Join(root, "SKILL.md")
	if err := os.WriteFile(entrypoint, []byte("content"), 0o600); err != nil {
		t.Fatal(err)
	}
	manifest := Manifest{ID: "focused", Root: root, Sources: []Source{{ID: "focused", Priority: "primary", Lifecycle: "callable", Entrypoint: "SKILL.md", Globs: []string{"${ROOT}"}, Required: []string{"SKILL.md"}}}}
	contracts, err := BuildSourceExecutionContracts(manifest, []MountedSource{{ID: "focused", Entrypoint: "SKILL.md", ActivationReason: "primary-source"}})
	if err != nil {
		t.Fatal(err)
	}
	worker := testReceiptWorker()
	worker.SourceExecution = contracts
	receipt := validReceipt(worker)
	_, verification, err := VerifySourceExecutionContracts([]Manifest{manifest}, contracts)
	if err != nil {
		t.Fatal(err)
	}
	if len(verification) != 1 || verification[0].SourceID != "focused" {
		t.Fatalf("entrypoint preparation verification is incomplete: %#v", verification)
	}
	if err := ValidateWorkerReceiptConsistency(worker, receipt); err != nil {
		t.Fatalf("route-consistent receipt was rejected: %v", err)
	}
	// This is still self-reported consistency, never native execution proof.
}

func TestDispatchRejectsTamperedSourceContractWithoutTrustedPathAuthority(t *testing.T) {
	root := t.TempDir()
	entrypoint := filepath.Join(root, "SKILL.md")
	if err := os.WriteFile(entrypoint, []byte("trusted content"), 0o600); err != nil {
		t.Fatal(err)
	}
	manifest := Manifest{ID: "focused", Root: root, Sources: []Source{{ID: "focused", Priority: "primary", Lifecycle: "callable", Entrypoint: "SKILL.md", Globs: []string{"${ROOT}"}, Required: []string{"SKILL.md"}}}}
	contracts, err := BuildSourceExecutionContracts(manifest, []MountedSource{{ID: "focused", Entrypoint: "SKILL.md", ActivationReason: "primary-source"}})
	if err != nil {
		t.Fatal(err)
	}
	tampered := append([]SourceExecutionContract(nil), contracts...)
	tampered[0].Entrypoint = "../outside.txt"
	worker := testReceiptWorker()
	worker.SourceExecution = tampered
	if _, err := DispatchWorker(worker, DispatchOptions{Workspace: root, OutputDir: filepath.Join(root, "output"), DryRun: true, TrustedManifests: []Manifest{manifest}}); err == nil {
		t.Fatal("dispatch accepted a route-controlled source path")
	}
	worker.SourceExecution = contracts
	if _, err := DispatchWorker(worker, DispatchOptions{Workspace: root, OutputDir: filepath.Join(root, "output"), DryRun: true}); err == nil || !strings.Contains(err.Error(), "trusted manifests") {
		t.Fatalf("dispatch did not require a trusted registry: %v", err)
	}
}

func TestDispatchRejectsEntrypointSymlinkEscape(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside-skill.md")
	if err := os.WriteFile(outside, []byte("outside"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "SKILL.md")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlink creation is unavailable on this Windows host: %v", err)
	}
	manifest := Manifest{ID: "focused", Root: root, Sources: []Source{{ID: "focused", Priority: "primary", Lifecycle: "callable", Entrypoint: "SKILL.md", Globs: []string{"${ROOT}"}, Required: []string{"SKILL.md"}}}}
	if _, err := BuildSourceExecutionContracts(manifest, []MountedSource{{ID: "focused", Entrypoint: "SKILL.md", ActivationReason: "primary-source"}}); err == nil || !strings.Contains(err.Error(), "escapes its root") {
		t.Fatalf("source entrypoint symlink escape was accepted: %v", err)
	}
}

func TestOnlyExecutionWorkersReceiveScenarioEntrypoints(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "SKILL.md"), []byte("focused instructions"), 0o600); err != nil {
		t.Fatal(err)
	}
	manifest := Manifest{
		ID: "focused", Root: root, Triggers: []string{"focused"}, Status: "callable",
		Sources: []Source{{ID: "focused", Priority: "primary", Lifecycle: "callable", Entrypoint: "SKILL.md", Globs: []string{"${ROOT}"}, Required: []string{"SKILL.md"}}},
	}
	route := Route("white-hat fix focused SDK bug", []Manifest{manifest})
	if len(route.SourceExecution) != 1 || len(route.Workers) == 0 || len(route.PreflightWorkers) == 0 || len(route.OfficerWorkers) == 0 {
		t.Fatalf("test route did not create every relevant stage: %#v", route)
	}
	for _, worker := range route.Workers {
		if len(worker.SourceExecution) != 1 {
			t.Fatalf("execution worker did not receive its selected source: %#v", worker)
		}
	}
	for _, worker := range append(route.PreflightWorkers, route.OfficerWorkers...) {
		if len(worker.SourceExecution) != 0 || worker.SourceExecutionBytes != 0 {
			t.Fatalf("non-execution worker received a scenario Skill: %#v", worker)
		}
	}
}

func TestRouteDoesNotInjectEmptySourceExecutionSlot(t *testing.T) {
	manifest := Manifest{ID: "code", Triggers: []string{"fix"}, Status: "callable", PrimarySkill: "native"}
	route := Route("fix a small code bug", []Manifest{manifest})
	if len(route.SourceExecution) != 0 || len(route.Workers) == 0 {
		t.Fatalf("test route unexpectedly selected source execution: %#v", route)
	}
	for _, worker := range route.Workers {
		if len(worker.SourceExecution) != 0 || worker.SourceExecutionBytes != 0 {
			t.Fatalf("worker received an empty source execution contract: %#v", worker.SourceExecution)
		}
		if strings.Join(worker.PromptOrder, ",") != "stable_capability_prefix,task_contract" {
			t.Fatalf("worker prompt retained an empty source slot: %#v", worker.PromptOrder)
		}
	}
}
