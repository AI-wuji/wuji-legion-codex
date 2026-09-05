package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAcceptanceReconciliationRequiresActiveVerifiedExecution(t *testing.T) {
	root := t.TempDir()
	requirementStore := filepath.Join(root, "requirements")
	executionStore := filepath.Join(root, "execution")
	acceptanceStore := filepath.Join(root, "acceptance")
	requirement, err := UpsertRequirement(requirementStore, RequirementInput{
		ID: "single-route", Summary: "Keep the only Aji route.", Sources: []string{"message:sha256:" + strings.Repeat("a", 64)},
	})
	if err != nil {
		t.Fatal(err)
	}
	execution, err := RecordExecutionNode(executionStore, requirementStore, ExecutionNodeInput{
		ID: "verify-route", Authority: "staff", Goal: "verify the only route", RequirementRevisions: []string{requirement.VersionID},
		Model: "gpt-5.6-terra", ModelReason: "bounded implementation", NetworkBoundary: "none", WriteBoundary: "scoped-artifact-write", BranchBoundary: "current",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := RecordExecutionResult(executionStore, requirementStore, ExecutionResultInput{ID: execution.ID, Status: "succeeded"}); err != nil {
		t.Fatal(err)
	}
	if _, err := ReconcileAcceptance(acceptanceStore, requirementStore, executionStore, AcceptanceInput{ID: "route-acceptance", RequirementRevision: requirement.VersionID, ExecutionVersion: execution.VersionID}); err == nil {
		t.Fatal("acceptance accepted a succeeded execution without artifacts and verification")
	}
	artifact := "artifact:sha256:" + strings.Repeat("b", 64)
	verification := "verification:sha256:" + strings.Repeat("c", 64)
	if _, err := RecordExecutionResult(executionStore, requirementStore, ExecutionResultInput{ID: execution.ID, Status: "succeeded", ArtifactHandles: []string{artifact}, VerificationHandles: []string{verification}}); err != nil {
		t.Fatal(err)
	}
	if _, err := ReconcileAcceptance(acceptanceStore, requirementStore, executionStore, AcceptanceInput{ID: "aji-acceptance", RequirementRevision: requirement.VersionID, ExecutionVersion: execution.VersionID, AcceptedBy: "aji"}); err == nil {
		t.Fatal("new acceptance record permitted Aji as the accepter")
	}
	accepted, err := ReconcileAcceptance(acceptanceStore, requirementStore, executionStore, AcceptanceInput{ID: "route-acceptance", RequirementRevision: requirement.VersionID, ExecutionVersion: execution.VersionID})
	if err != nil || accepted.AcceptedBy != "deterministic-evidence-verifier" || accepted.RecordSHA256 == "" || len(accepted.VerificationHandles) != 1 {
		t.Fatalf("acceptance did not bind real execution evidence: record=%#v err=%v", accepted, err)
	}
	legacy := accepted
	legacy.AcceptedBy = "aji"
	legacy.RecordSHA256 = acceptanceRecordDigest(legacy)
	if err := writeAcceptanceLedger(acceptanceStore, AcceptanceLedger{SchemaVersion: acceptanceSchemaVersion, Records: []AcceptanceRecord{legacy}}); err != nil {
		t.Fatal(err)
	}
	if _, err := loadAcceptanceLedger(acceptanceStore); err == nil {
		t.Fatal("legacy Aji acceptance record was still readable")
	}
	if _, err := UpsertRequirement(requirementStore, RequirementInput{ID: "single-route", Summary: "Keep the only Aji route with a verified receipt.", Sources: []string{"message:sha256:" + strings.Repeat("d", 64)}}); err != nil {
		t.Fatal(err)
	}
	if _, err := ReconcileAcceptance(acceptanceStore, requirementStore, executionStore, AcceptanceInput{ID: "stale-acceptance", RequirementRevision: requirement.VersionID, ExecutionVersion: execution.VersionID}); err == nil {
		t.Fatal("acceptance accepted a stale requirement revision")
	}
}

func TestConversationEvidenceRetentionAndProvenanceACL(t *testing.T) {
	root := t.TempDir()
	requirementStore := filepath.Join(root, "requirements")
	conversationStore := filepath.Join(root, "conversation")
	provenanceStore := filepath.Join(root, "provenance")
	requirement, err := UpsertRequirement(requirementStore, RequirementInput{
		ID: "evidence-scope", Summary: "Keep chat evidence cold and opaque.", Sources: []string{"message:sha256:" + strings.Repeat("e", 64)},
	})
	if err != nil {
		t.Fatal(err)
	}
	record, err := LinkConversationEvidence(conversationStore, requirementStore, requirement.VersionID, []string{"host-message:opaque-001"})
	if err != nil || len(record.MessageHandles) != 1 {
		t.Fatalf("conversation evidence was not linked: record=%#v err=%v", record, err)
	}
	byMessage, err := ResolveConversationEvidence(conversationStore, ConversationEvidenceQuery{MessageHandle: "host-message:opaque-001"})
	if err != nil || len(byMessage) != 1 || byMessage[0].Revision != requirement.VersionID {
		t.Fatalf("opaque message reverse lookup failed: records=%#v err=%v", byMessage, err)
	}
	if _, err := RecordProvenance(provenanceStore, ProvenanceInput{
		ID: "requirement-origin", Scope: "global", Subject: requirement.VersionID, Predicate: "derived-from", Target: "host-message:opaque-001", Readers: []string{"aji"},
	}); err != nil {
		t.Fatal(err)
	}
	denied, err := ResolveProvenance(provenanceStore, ProvenanceQuery{Scope: "global", Subject: requirement.VersionID, Principal: "staff"})
	if err != nil || denied.Denied != 1 || len(denied.Entries) != 0 {
		t.Fatalf("provenance ACL exposed a restricted edge: result=%#v err=%v", denied, err)
	}
	allowed, err := ResolveProvenance(provenanceStore, ProvenanceQuery{Scope: "global", Subject: requirement.VersionID, Principal: "aji"})
	if err != nil || len(allowed.Entries) != 1 {
		t.Fatalf("provenance ACL denied the permitted reader: result=%#v err=%v", allowed, err)
	}

	index, err := loadConversationEvidenceIndex(conversationStore)
	if err != nil {
		t.Fatal(err)
	}
	index.Records[0].RecordedAt = time.Now().UTC().AddDate(0, 0, -31).Format(time.RFC3339)
	index.Records[0].RecordSHA256 = conversationEvidenceDigest(index.Records[0])
	archivedDigest := index.Records[0].RecordSHA256
	if err := writeConversationEvidenceIndex(conversationStore, index); err != nil {
		t.Fatal(err)
	}
	maintenance, err := MaintainGraph(conversationStore, "conversation-evidence", time.Now().UTC())
	if err != nil || !maintenance.Validated || maintenance.ArchivedRecords != 1 || maintenance.GCRecords != 1 {
		t.Fatalf("conversation evidence governance did not archive expired data: result=%#v err=%v", maintenance, err)
	}
	if records, err := ResolveConversationEvidence(conversationStore, ConversationEvidenceQuery{Revision: requirement.VersionID}); err != nil || len(records) != 0 {
		t.Fatalf("expired conversation evidence remained hot: records=%#v err=%v", records, err)
	}
	if _, err := os.Stat(filepath.Join(conversationStore, "v1", "archive", archivedDigest+".json")); err != nil {
		t.Fatalf("expired conversation evidence was not preserved in its archive: %v", err)
	}
}

func TestSourceLifecycleImpactAndFusionAssetInvocation(t *testing.T) {
	assessmentStore := t.TempDir()
	input := SourceAssessmentInput{SourceID: "native-source", Version: "2.0.1", Decision: "adopted", Reason: "verified callable source", ReanalyzeWhen: []string{"new-version"}}
	first, err := AssessSource(assessmentStore, input)
	if err != nil {
		t.Fatal(err)
	}
	second, err := AssessSource(assessmentStore, input)
	if err != nil || second.RecordSHA256 != first.RecordSHA256 {
		t.Fatalf("same source version was reassessed instead of reused: first=%#v second=%#v err=%v", first, second, err)
	}
	if _, err := AssessSource(assessmentStore, SourceAssessmentInput{SourceID: "native-source", Version: "2.1.0", Decision: "deferred", Reason: "wait for an explicit reuse or version review"}); err != nil {
		t.Fatal(err)
	}
	store, err := loadSourceAssessmentStore(assessmentStore)
	if err != nil || len(store.Assessments) != 2 {
		t.Fatalf("source lifecycle did not keep minimal version decisions: store=%#v err=%v", store, err)
	}

	root := t.TempDir()
	manifest := newLineageManifest(t, root)
	manifest.Genome.Adapters[0].Assets = nil
	manifest.Genome.Adapters[0].AssetContracts = []FusionAsset{{ID: "workflow-template", Path: "assets/workflow.md", Compatibility: []string{"codex", "go"}}}
	impact, err := SourceImpact(LineageCatalog{SchemaVersion: lineageCatalogSchemaVersion, Nodes: []LineageNode{
		{ID: "source:lineage-fixture:native-source", Kind: "source", SourceID: "native-source", SourceVersion: "2.0.1"},
		{ID: "adapter:lineage-fixture:workflow", Kind: "fusion-adapter", Parents: []string{"source:lineage-fixture:native-source"}},
		{ID: "asset:lineage-fixture:workflow:template", Kind: "asset", Parents: []string{"adapter:lineage-fixture:workflow"}},
		{ID: "genome:lineage-fixture:rev-1", Kind: "fusion-genome", Parents: []string{"adapter:lineage-fixture:workflow"}},
	}}, "native-source", "2.1.0")
	if err != nil || len(impact.KnownVersions) != 1 || impact.KnownVersions[0] != "2.0.1" || len(impact.ImpactedNodes) != 4 {
		t.Fatalf("source impact did not return only lineage-related differences: impact=%#v err=%v", impact, err)
	}
	contract, err := SelectFusionAsset([]Manifest{manifest}, AssetSelectionRequest{Capability: manifest.ID, AssetID: "workflow-template", Compatibility: []string{"codex"}})
	if err != nil || contract.AssetID != manifest.ID+":workflow-template" || contract.EntrypointSHA256 == "" || contract.Invocation.SourceID != "native-source" {
		t.Fatalf("fusion asset did not produce a trusted invocation contract: contract=%#v err=%v", contract, err)
	}
	if _, err := SelectFusionAsset([]Manifest{manifest}, AssetSelectionRequest{Capability: manifest.ID, Domain: "workflow", Compatibility: []string{"pi-only"}}); err == nil {
		t.Fatal("fusion asset selection ignored declared compatibility")
	}
}
