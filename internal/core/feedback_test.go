package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestVersionedExecutionResultRecordsBoundedUnverifiedFeedback(t *testing.T) {
	root := t.TempDir()
	requirementStore := filepath.Join(root, "requirements")
	executionStore := filepath.Join(root, "execution")
	requirement, err := UpsertRequirement(requirementStore, RequirementInput{ID: "goal", Summary: "Record a bounded result.", Sources: []string{"message:sha256:" + strings.Repeat("a", 64)}})
	if err != nil {
		t.Fatal(err)
	}
	node, err := RecordVersionedExecutionNode(executionStore, requirementStore, VersionedExecutionNodeInput{ExecutionNodeInput: feedbackTestNode("node", requirement.VersionID), TaskInstanceID: "task", GraphVersion: "graph-1", AttemptID: "attempt-1"})
	if err != nil {
		t.Fatal(err)
	}
	binding := executionRuntimeBinding{TaskInstanceID: "task", GraphVersion: "graph-1", AttemptID: "attempt-1"}
	if _, err := RecordVersionedExecutionResult(executionStore, requirementStore, VersionedExecutionResultInput{ExecutionResultInput: ExecutionResultInput{ID: node.ID, Status: "succeeded"}, TaskInstanceID: binding.TaskInstanceID, GraphVersion: binding.GraphVersion, AttemptID: binding.AttemptID}); err != nil {
		t.Fatal(err)
	}
	feedbackStore := executionFeedbackStoreFor(executionStore)
	ledger, err := loadExecutionFeedbackLedger(feedbackStore)
	if err != nil || len(ledger.Records) != 1 {
		t.Fatalf("feedback was not recorded: %#v err=%v", ledger, err)
	}
	record := ledger.Records[0]
	if record.Outcome != "unverified-success" || record.Status != "candidate" || record.ExecutionVersion != node.VersionID {
		t.Fatalf("unexpected feedback record: %#v", record)
	}
	if _, err := recordExecutionFeedback(feedbackStore, executionStore, requirementStore, node.VersionID, binding); err != nil {
		t.Fatal(err)
	}
	ledger, err = loadExecutionFeedbackLedger(feedbackStore)
	if err != nil || len(ledger.Records) != 1 {
		t.Fatalf("feedback was not deduplicated: %#v err=%v", ledger, err)
	}
	if _, err := filepath.Glob(filepath.Join(root, "knowledge", "*")); err != nil {
		t.Fatal(err)
	}
	if _, err := RecordKnowledge(filepath.Join(root, "knowledge"), KnowledgeRecordInput{Kind: "solution", Key: "x", Scope: "global", Summary: "x"}); err == nil {
		t.Fatal("candidate feedback was treated as verification evidence")
	}
}

func TestExecutionFeedbackRecordsFailuresAndRejectsStaleVersions(t *testing.T) {
	root := t.TempDir()
	requirementStore := filepath.Join(root, "requirements")
	executionStore := filepath.Join(root, "execution")
	requirement, err := UpsertRequirement(requirementStore, RequirementInput{ID: "goal", Summary: "Reject stale feedback.", Sources: []string{"message:sha256:" + strings.Repeat("b", 64)}})
	if err != nil {
		t.Fatal(err)
	}
	node, err := RecordVersionedExecutionNode(executionStore, requirementStore, VersionedExecutionNodeInput{ExecutionNodeInput: feedbackTestNode("node", requirement.VersionID), TaskInstanceID: "task", GraphVersion: "graph-1", AttemptID: "attempt-1"})
	if err != nil {
		t.Fatal(err)
	}
	binding := executionRuntimeBinding{TaskInstanceID: "task", GraphVersion: "graph-1", AttemptID: "attempt-1"}
	if _, err := RecordVersionedExecutionResult(executionStore, requirementStore, VersionedExecutionResultInput{ExecutionResultInput: ExecutionResultInput{ID: node.ID, Status: "failed", Failure: "provider unavailable"}, TaskInstanceID: binding.TaskInstanceID, GraphVersion: binding.GraphVersion, AttemptID: binding.AttemptID}); err != nil {
		t.Fatal(err)
	}
	feedbackStore := executionFeedbackStoreFor(executionStore)
	ledger, err := loadExecutionFeedbackLedger(feedbackStore)
	if err != nil || len(ledger.Records) != 1 || ledger.Records[0].Outcome != "unverified-failure" {
		t.Fatalf("failure feedback was not a candidate: %#v err=%v", ledger, err)
	}
	if _, err := UpsertRequirement(requirementStore, RequirementInput{ID: "goal", Summary: "Changed requirement.", Sources: []string{"message:sha256:" + strings.Repeat("c", 64)}}); err != nil {
		t.Fatal(err)
	}
	if _, err := recordExecutionFeedback(feedbackStore, executionStore, requirementStore, node.VersionID, binding); err == nil {
		t.Fatal("feedback accepted an invalidated execution version")
	}
	if _, err := recordExecutionFeedback(feedbackStore, executionStore, requirementStore, node.VersionID, executionRuntimeBinding{TaskInstanceID: "task", GraphVersion: "graph-2", AttemptID: "attempt-2"}); err == nil {
		t.Fatal("feedback accepted a stale runtime binding")
	}
}

func TestVerifiedFailureFeedbackKnowledgeRequiresIndependentEvidence(t *testing.T) {
	root := t.TempDir()
	requirementStore, executionStore := filepath.Join(root, "requirements"), filepath.Join(root, "execution")
	requirement, err := UpsertRequirement(requirementStore, RequirementInput{ID: "goal", Summary: "Persist verified failure learning.", Sources: []string{"message:sha256:" + strings.Repeat("d", 64)}})
	if err != nil {
		t.Fatal(err)
	}
	node, err := RecordVersionedExecutionNode(executionStore, requirementStore, VersionedExecutionNodeInput{ExecutionNodeInput: feedbackTestNode("node", requirement.VersionID), TaskInstanceID: "task", GraphVersion: "graph-1", AttemptID: "attempt-1"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := RecordVersionedExecutionResult(executionStore, requirementStore, VersionedExecutionResultInput{ExecutionResultInput: ExecutionResultInput{ID: node.ID, Status: "failed", Failure: "provider unavailable"}, TaskInstanceID: "task", GraphVersion: "graph-1", AttemptID: "attempt-1"}); err != nil {
		t.Fatal(err)
	}
	feedbackStore, knowledgeStore := executionFeedbackStoreFor(executionStore), filepath.Join(root, "knowledge")
	ledger, err := loadExecutionFeedbackLedger(feedbackStore)
	if err != nil || len(ledger.Records) != 1 {
		t.Fatalf("missing feedback candidate: %#v err=%v", ledger, err)
	}
	location, evidence := filepath.Join(root, "solution.md"), filepath.Join(root, "verification.json")
	if err := os.WriteFile(location, []byte("bounded recovery"), 0o600); err != nil {
		t.Fatal(err)
	}
	receipt := `{"schema_version":1,"type":"wuji-verification-receipt","passed":true,"verifier":"go-test","verified_at":"` + time.Now().UTC().Format(time.RFC3339) + `"}`
	if err := os.WriteFile(evidence, []byte(receipt), 0o600); err != nil {
		t.Fatal(err)
	}
	input := KnowledgeRecordInput{Kind: "failure", Key: "provider unavailable", Scope: "global", Summary: "Use the bounded availability fallback.", RootCause: "The provider was unavailable before generation.", Location: location, Verification: evidence}
	record, err := RecordVerifiedFailureFeedbackKnowledge(feedbackStore, knowledgeStore, ledger.Records[0].ID, input)
	if err != nil {
		t.Fatal(err)
	}
	if record.VerificationSHA256 == "" || !knowledgeHasRelation(record, "derived-from", "wuji-feedback://"+ledger.Records[0].ID) {
		t.Fatalf("verified feedback provenance missing: %#v", record)
	}
	result, err := QueryKnowledge(knowledgeStore, KnowledgeQuery{Trigger: "failure", Kind: "failure", Key: "provider unavailable", Scope: "global"})
	if err != nil || len(result.Matches) != 1 {
		t.Fatalf("failure reuse query failed: %#v err=%v", result, err)
	}
	if _, err := RecordVerifiedFailureFeedbackKnowledge(feedbackStore, knowledgeStore, ledger.Records[0].ID, KnowledgeRecordInput{Kind: "failure", Key: "x", Scope: "global", Summary: "x", RootCause: "x", Location: location, Verification: filepath.Join(root, "missing.json")}); err == nil {
		t.Fatal("missing independent verification evidence was accepted")
	}
	ledger.Records[0].Outcome = "unverified-success"
	ledger.Records[0].RecordSHA256 = executionFeedbackDigest(ledger.Records[0])
	if err := writeExecutionFeedbackLedger(feedbackStore, ledger); err != nil {
		t.Fatal(err)
	}
	if _, err := RecordVerifiedFailureFeedbackKnowledge(feedbackStore, knowledgeStore, ledger.Records[0].ID, input); err == nil {
		t.Fatal("success feedback was admitted as failure knowledge")
	}
	relativeRoot, err := os.MkdirTemp(".", ".feedback-alias-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(relativeRoot)
	relativeFeedbackStore := filepath.Join(relativeRoot, "store")
	if err := os.MkdirAll(relativeFeedbackStore, 0o700); err != nil {
		t.Fatal(err)
	}
	absoluteFeedbackStore, err := filepath.Abs(relativeFeedbackStore)
	if err != nil {
		t.Fatal(err)
	}
	sameStore, err := sameExecutionFeedbackAndKnowledgeStore(relativeFeedbackStore, absoluteFeedbackStore)
	if err != nil || !sameStore {
		t.Fatalf("relative and absolute aliases were not rejected: same=%v err=%v", sameStore, err)
	}
}

func feedbackTestNode(id, requirementVersion string) ExecutionNodeInput {
	return ExecutionNodeInput{ID: id, Authority: "staff", Goal: "test feedback", RequirementRevisions: []string{requirementVersion}, Model: "gpt-5.6-terra", ModelReason: "test", NetworkBoundary: "none", WriteBoundary: "none", BranchBoundary: "current"}
}
