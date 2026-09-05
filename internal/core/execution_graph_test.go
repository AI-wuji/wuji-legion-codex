package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExecutionGraphBindsExactRequirementsAndInvalidatesDownstream(t *testing.T) {
	root := t.TempDir()
	requirementStore := filepath.Join(root, "requirements")
	executionStore := filepath.Join(root, "execution")
	req, err := UpsertRequirement(requirementStore, RequirementInput{ID: "goal", Summary: "Keep one route.", Goal: "single router", Sources: []string{"message:sha256:" + strings.Repeat("a", 64)}})
	if err != nil {
		t.Fatal(err)
	}
	base, err := RecordExecutionNode(executionStore, requirementStore, ExecutionNodeInput{ID: "base", Authority: "staff", Goal: "inspect route", RequirementRevisions: []string{req.VersionID}, Model: "gpt-5.6-terra", ModelReason: "staff schedules a bounded execution node", NetworkBoundary: "none", WriteBoundary: "scoped-artifact-write", BranchBoundary: "current"})
	if err != nil {
		t.Fatal(err)
	}
	child, err := RecordExecutionNode(executionStore, requirementStore, ExecutionNodeInput{ID: "child", Authority: "deterministic-executor", Goal: "verify route", RequirementRevisions: []string{req.VersionID}, DependsOn: []string{base.VersionID}, Model: "gpt-5.6-terra", ModelReason: "mechanical verification", NetworkBoundary: "none", WriteBoundary: "none", BranchBoundary: "current"})
	if err != nil {
		t.Fatal(err)
	}
	projection, err := ProjectExecutionGraph(executionStore, requirementStore, child.VersionID, 4096)
	if err != nil || len(projection.Nodes) != 2 {
		t.Fatalf("bounded execution projection failed: projection=%#v err=%v", projection, err)
	}
	if _, err := WriteExecutionGraphProjection(projection, filepath.Join(root, "projections")); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadExecutionGraphProjection(filepath.Join(root, "projections", projection.PayloadSHA256+".json")); err != nil {
		t.Fatal(err)
	}

	updated, err := UpsertRequirement(requirementStore, RequirementInput{ID: "goal", Summary: "Keep one route and verify evidence.", Goal: "single router", Sources: []string{"message:sha256:" + strings.Repeat("a", 64)}})
	if err != nil || updated.VersionID == req.VersionID {
		t.Fatalf("requirement revision did not advance: updated=%#v err=%v", updated, err)
	}
	projection, err = ProjectExecutionGraph(executionStore, requirementStore, child.ID, 4096)
	if err != nil || projection.Nodes[0].Status != "invalidated" || projection.Nodes[1].Status != "invalidated" {
		t.Fatalf("requirement change did not invalidate execution chain: projection=%#v err=%v", projection, err)
	}
	if _, err := RecordExecutionNode(executionStore, requirementStore, ExecutionNodeInput{ID: "stale", Authority: "staff", Goal: "stale", RequirementRevisions: []string{req.VersionID}, Model: "gpt-5.6-terra", ModelReason: "test", NetworkBoundary: "none", WriteBoundary: "none", BranchBoundary: "current"}); err == nil {
		t.Fatal("stale requirement revision was accepted")
	}
}

func TestExecutionGraphRejectsAjiAndTamperedProjection(t *testing.T) {
	if _, err := RecordExecutionNode(t.TempDir(), t.TempDir(), ExecutionNodeInput{ID: "node", Authority: "aji", Goal: "blocked", RequirementRevisions: []string{"req@1"}, Model: "gpt-5.6-terra", ModelReason: "test"}); err == nil {
		t.Fatal("Aji was allowed to write execution graph")
	}
	root := t.TempDir()
	reqStore := filepath.Join(root, "req")
	req, err := UpsertRequirement(reqStore, RequirementInput{ID: "r", Summary: "Requirement.", Sources: []string{"message:sha256:" + strings.Repeat("b", 64)}})
	if err != nil {
		t.Fatal(err)
	}
	projection, err := func() (ExecutionGraphProjection, error) {
		store := filepath.Join(root, "exec")
		if _, err := RecordExecutionNode(store, reqStore, ExecutionNodeInput{ID: "n", Authority: "staff", Goal: "run", RequirementRevisions: []string{req.VersionID}, Model: "gpt-5.6-terra", ModelReason: "test"}); err != nil {
			return ExecutionGraphProjection{}, err
		}
		return ProjectExecutionGraph(store, reqStore, "n", 4096)
	}()
	if err != nil {
		t.Fatal(err)
	}
	projection.Payload = projection.Payload + "x"
	if err := validateExecutionProjection(projection); err == nil {
		t.Fatal("tampered execution projection was accepted")
	}
}

func TestVersionedExecutionResultsRejectLateReceiptsAndKeepUnrelatedNodes(t *testing.T) {
	root := t.TempDir()
	requirementStore := filepath.Join(root, "requirements")
	executionStore := filepath.Join(root, "execution")
	req, err := UpsertRequirement(requirementStore, RequirementInput{ID: "goal", Summary: "Keep runtime current.", Sources: []string{"message:sha256:" + strings.Repeat("c", 64)}})
	if err != nil {
		t.Fatal(err)
	}
	baseInput := ExecutionNodeInput{ID: "base", Authority: "staff", Goal: "run", RequirementRevisions: []string{req.VersionID}, Model: "gpt-5.6-terra", ModelReason: "test", NetworkBoundary: "none", WriteBoundary: "none", BranchBoundary: "current"}
	base, err := RecordVersionedExecutionNode(executionStore, requirementStore, VersionedExecutionNodeInput{ExecutionNodeInput: baseInput, TaskInstanceID: "task", GraphVersion: "graph-1", AttemptID: "attempt-1"})
	if err != nil {
		t.Fatal(err)
	}
	child, err := RecordVersionedExecutionNode(executionStore, requirementStore, VersionedExecutionNodeInput{ExecutionNodeInput: ExecutionNodeInput{ID: "child", Authority: "staff", Goal: "follow", RequirementRevisions: []string{req.VersionID}, DependsOn: []string{base.VersionID}, Model: "gpt-5.6-terra", ModelReason: "test", NetworkBoundary: "none", WriteBoundary: "none", BranchBoundary: "current"}, TaskInstanceID: "task", GraphVersion: "graph-1", AttemptID: "attempt-1"})
	if err != nil {
		t.Fatal(err)
	}
	independent, err := RecordExecutionNode(executionStore, requirementStore, ExecutionNodeInput{ID: "independent", Authority: "staff", Goal: "retain", RequirementRevisions: []string{req.VersionID}, Model: "gpt-5.6-terra", ModelReason: "test", NetworkBoundary: "none", WriteBoundary: "none", BranchBoundary: "current"})
	if err != nil {
		t.Fatal(err)
	}

	// A new attempt makes the first receipt stale without revising the node contract.
	if _, err := RecordVersionedExecutionNode(executionStore, requirementStore, VersionedExecutionNodeInput{ExecutionNodeInput: baseInput, TaskInstanceID: "task", GraphVersion: "graph-2", AttemptID: "attempt-2"}); err != nil {
		t.Fatal(err)
	}
	if _, err := RecordVersionedExecutionResult(executionStore, requirementStore, VersionedExecutionResultInput{ExecutionResultInput: ExecutionResultInput{ID: base.ID, Status: "succeeded"}, TaskInstanceID: "task", GraphVersion: "graph-1", AttemptID: "attempt-2"}); err == nil {
		t.Fatal("late graph version receipt was accepted")
	}
	if _, err := RecordVersionedExecutionResult(executionStore, requirementStore, VersionedExecutionResultInput{ExecutionResultInput: ExecutionResultInput{ID: base.ID, Status: "succeeded"}, TaskInstanceID: "task", GraphVersion: "graph-2", AttemptID: "attempt-1"}); err == nil {
		t.Fatal("late attempt receipt was accepted")
	}
	if _, err := RecordVersionedExecutionResult(executionStore, requirementStore, VersionedExecutionResultInput{ExecutionResultInput: ExecutionResultInput{ID: base.ID, Status: "invalidated"}, TaskInstanceID: "task", GraphVersion: "graph-2", AttemptID: "attempt-2"}); err != nil {
		t.Fatal(err)
	}
	projection, err := ProjectExecutionGraph(executionStore, requirementStore, child.ID, 4096)
	if err != nil || projection.Nodes[0].Status != "invalidated" || projection.Nodes[1].Status != "invalidated" {
		t.Fatalf("dependency invalidation did not reach only the downstream node: projection=%#v err=%v", projection, err)
	}
	projection, err = ProjectExecutionGraph(executionStore, requirementStore, independent.ID, 4096)
	if err != nil || projection.Nodes[0].Status != "planned" {
		t.Fatalf("unrelated node was changed: projection=%#v err=%v", projection, err)
	}
}

func TestVersionedExecutionResultsRejectCancelledNode(t *testing.T) {
	root := t.TempDir()
	req, err := UpsertRequirement(filepath.Join(root, "requirements"), RequirementInput{ID: "goal", Summary: "Cancel runtime.", Sources: []string{"message:sha256:" + strings.Repeat("d", 64)}})
	if err != nil {
		t.Fatal(err)
	}
	store := filepath.Join(root, "execution")
	node, err := RecordVersionedExecutionNode(store, filepath.Join(root, "requirements"), VersionedExecutionNodeInput{ExecutionNodeInput: ExecutionNodeInput{ID: "node", Authority: "staff", Goal: "cancel", RequirementRevisions: []string{req.VersionID}, Model: "gpt-5.6-terra", ModelReason: "test", NetworkBoundary: "none", WriteBoundary: "none", BranchBoundary: "current"}, TaskInstanceID: "task", GraphVersion: "graph-1", AttemptID: "attempt-1"})
	if err != nil {
		t.Fatal(err)
	}
	binding := VersionedExecutionResultInput{ExecutionResultInput: ExecutionResultInput{ID: node.ID, Status: "cancelled"}, TaskInstanceID: "task", GraphVersion: "graph-1", AttemptID: "attempt-1"}
	if _, err := RecordVersionedExecutionResult(store, filepath.Join(root, "requirements"), binding); err != nil {
		t.Fatal(err)
	}
	binding.Status = "succeeded"
	if _, err := RecordVersionedExecutionResult(store, filepath.Join(root, "requirements"), binding); err == nil {
		t.Fatal("late receipt changed a cancelled node")
	}
}

func TestVersionedExecutionTerminalResultsAreImmutableAndReplayable(t *testing.T) {
	for _, initial := range []string{"succeeded", "failed"} {
		t.Run(initial, func(t *testing.T) {
			root := t.TempDir()
			requirementStore := filepath.Join(root, "requirements")
			executionStore := filepath.Join(root, "execution")
			req, err := UpsertRequirement(requirementStore, RequirementInput{ID: "goal", Summary: "Keep terminal results immutable.", Sources: []string{"message:sha256:" + strings.Repeat("e", 64)}})
			if err != nil {
				t.Fatal(err)
			}
			node, err := RecordVersionedExecutionNode(executionStore, requirementStore, VersionedExecutionNodeInput{ExecutionNodeInput: ExecutionNodeInput{ID: "node", Authority: "staff", Goal: "run", RequirementRevisions: []string{req.VersionID}, Model: "gpt-5.6-terra", ModelReason: "test", NetworkBoundary: "none", WriteBoundary: "none", BranchBoundary: "current"}, TaskInstanceID: "task", GraphVersion: "graph-1", AttemptID: "attempt-1"})
			if err != nil {
				t.Fatal(err)
			}
			result := VersionedExecutionResultInput{ExecutionResultInput: ExecutionResultInput{ID: node.ID, Status: initial, Failure: "reported failure"}, TaskInstanceID: "task", GraphVersion: "graph-1", AttemptID: "attempt-1"}
			if initial == "succeeded" {
				result.Failure = ""
			}
			if _, err := RecordVersionedExecutionResult(executionStore, requirementStore, result); err != nil {
				t.Fatal(err)
			}
			if _, err := RecordVersionedExecutionResult(executionStore, requirementStore, result); err != nil {
				t.Fatalf("identical terminal replay was rejected: %v", err)
			}
			if initial == "succeeded" {
				result.Status, result.Failure = "failed", "late failure"
			} else {
				result.Status, result.Failure = "succeeded", ""
			}
			if _, err := RecordVersionedExecutionResult(executionStore, requirementStore, result); err == nil {
				t.Fatal("terminal result changed for the same runtime binding")
			}
		})
	}
}

func TestAuditEventIsAppendOnlyAndIdempotent(t *testing.T) {
	store := t.TempDir()
	event := AuditEvent{EventID: "event-1", EventType: "test", Actor: "deterministic-evidence-verifier", Authority: "evidence-reconciliation", Target: "node@1"}
	if err := AuditEventRecord(store, event); err != nil {
		t.Fatal(err)
	}
	if err := AuditEventRecord(store, event); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(store, "v1", "events.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if lines := strings.Count(strings.TrimSpace(string(data)), "\n") + 1; lines != 1 {
		t.Fatalf("duplicate audit event was appended: %d lines", lines)
	}
	var decoded AuditEvent
	if err := json.Unmarshal(bytesBeforeNewline(data), &decoded); err != nil || decoded.RecordSHA256 == "" {
		t.Fatalf("audit record is invalid: %#v err=%v", decoded, err)
	}
}

func bytesBeforeNewline(data []byte) []byte {
	if index := strings.IndexByte(string(data), '\n'); index >= 0 {
		return data[:index]
	}
	return data
}
