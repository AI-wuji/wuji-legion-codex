package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/AI-wuji/wuji-legion-codex-2.0/internal/core"
)

func TestTopLevelHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"--help"}, &stdout, &stderr)
	if code != 0 || !strings.Contains(stdout.String(), "context-select") || stderr.Len() != 0 {
		t.Fatalf("unexpected help result: code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}

func TestFindRouteWorkerIncludesEveryExecutionStage(t *testing.T) {
	route := core.RouteResult{
		PreflightWorkers: []core.WorkerTask{{ID: "prior-art"}},
		Workers:          []core.WorkerTask{{ID: "implementation"}},
		OfficerWorkers:   []core.WorkerTask{{ID: "officer-white-hat"}},
	}
	if got := findRouteWorker(route, "general-staff"); got != nil {
		t.Fatalf("deterministic General Staff must not be exposed as a worker: %#v", got)
	}
	if got := findRouteWorker(route, "prior-art"); got == nil || got.ID != "prior-art" {
		t.Fatalf("preflight worker was not addressable: %#v", got)
	}
	if got := findRouteWorker(route, "implementation"); got == nil || got.ID != "implementation" {
		t.Fatalf("execution worker was not addressable: %#v", got)
	}
	if got := findRouteWorker(route, "officer-white-hat"); got == nil || got.ID != "officer-white-hat" {
		t.Fatalf("officer worker was not addressable: %#v", got)
	}
}

func TestRouteRequiresQuery(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"route"}, &stdout, &stderr)
	if code != 2 || !strings.Contains(stderr.String(), "--query is required") || stdout.Len() != 0 {
		t.Fatalf("missing query was not diagnosed: code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}

func TestModelSelectionUsesGPTHierarchyOrProviderMode(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"route", "--query", "list files", "--model", "gpt-5.6-sol"}, &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 {
		t.Fatalf("GPT route failed: code=%d stderr=%q", code, stderr.String())
	}
	var gptRoute core.RouteResult
	if err := json.Unmarshal(stdout.Bytes(), &gptRoute); err != nil {
		t.Fatalf("GPT route output is invalid: %v", err)
	}
	if gptRoute.Version != "3.0" || gptRoute.ModelPolicy.RoutingMode != "gpt-hierarchy" || gptRoute.ModelPolicy.UserSelectedModel != "gpt-5.6-sol" || gptRoute.MainModel != "gpt-5.6-sol" || gptRoute.GeneralStaffModel != "" || gptRoute.GeneralStaffWorker != nil || len(gptRoute.Workers) != 1 || gptRoute.Workers[0].Model != "gpt-5.6-luna" {
		t.Fatalf("GPT model selection did not use the explicit Aji model and bounded execution worker: %#v", gptRoute)
	}

	stdout.Reset()
	stderr.Reset()
	code = run([]string{"orchestrate", "--query", "research the web", "--model", "grok-4", "--dry-run"}, &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 {
		t.Fatalf("non-GPT orchestration failed: code=%d stderr=%q", code, stderr.String())
	}
	var orchestration core.OrchestrationResult
	if err := json.Unmarshal(stdout.Bytes(), &orchestration); err != nil {
		t.Fatalf("non-GPT orchestration output is invalid: %v", err)
	}
	route := orchestration.InitialRoute
	if route.ModelPolicy.RoutingMode != "explicit-non-gpt-provider-mode" || route.ModelPolicy.UserSelectedModel != "grok-4" || route.MainModel != "grok-4" || route.GeneralStaffModel != "" || route.ExecutionLane != "provider-mode-passthrough" || len(route.PreflightWorkers) != 0 || len(route.Workers) != 0 || len(route.OfficerWorkers) != 0 {
		t.Fatalf("non-GPT model selection did not preserve provider mode: %#v", route)
	}
}

func TestChangeCapsuleRequiresIntentAndProducesBoundedArtifact(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"change-capsule"}, &stdout, &stderr)
	if code != 2 || !strings.Contains(stderr.String(), "--intent is required") {
		t.Fatalf("missing capsule intent was not diagnosed: code=%d stderr=%q", code, stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	code = run([]string{"change-capsule", "--intent", "replace adapter", "--acceptance", "route passes; receipt persists"}, &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 || !strings.Contains(stdout.String(), "wuji-change-capsule-v1") {
		t.Fatalf("change capsule command did not produce an artifact: code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}

func TestStrictChangeCapsuleReturnsMachineReadableGateEvidence(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"change-capsule", "--intent", "replace adapter", "--strict"}, &stdout, &stderr)
	if code != 1 || stderr.Len() != 0 {
		t.Fatalf("incomplete strict capsule did not fail cleanly: code=%d stderr=%q", code, stderr.String())
	}
	var invalid core.ChangeCapsuleValidationResult
	if err := json.Unmarshal(stdout.Bytes(), &invalid); err != nil || invalid.Validation.Valid || len(invalid.Validation.Issues) != 4 {
		t.Fatalf("strict failure did not return validation evidence: result=%#v err=%v", invalid, err)
	}
	stdout.Reset()
	code = run([]string{"change-capsule", "--intent", "replace adapter", "--scope-out", "no provider migration", "--acceptance", "route passes", "--verification", "go test ./...", "--rollback", "restore prior adapter", "--strict"}, &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 {
		t.Fatalf("complete strict capsule failed: code=%d stderr=%q", code, stderr.String())
	}
	var valid core.ChangeCapsuleValidationResult
	if err := json.Unmarshal(stdout.Bytes(), &valid); err != nil || !valid.Validation.Valid || valid.Capsule.Intent != "replace adapter" {
		t.Fatalf("strict success did not return validation evidence: result=%#v err=%v", valid, err)
	}
}

func TestContextSelectReportsInvalidBudget(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"context-select", "--query", "valid query", "--max-bytes", "-1"}, &stdout, &stderr)
	if code != 2 || !strings.Contains(stderr.String(), "greater than zero") || stdout.Len() != 0 {
		t.Fatalf("invalid budget was not diagnosed: code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}

func TestContextSelectReportsMissingWorkspace(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"context-select", "--query", "valid query", "--workspace", t.TempDir() + "-missing"}, &stdout, &stderr)
	if code != 2 || !strings.Contains(stderr.String(), "error:") || stdout.Len() != 0 {
		t.Fatalf("missing workspace was not diagnosed: code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}

func TestUnexpectedArgumentsAreRejected(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"route", "extra"}, &stdout, &stderr)
	if code != 2 || !strings.Contains(stderr.String(), "unexpected arguments") {
		t.Fatalf("unexpected argument was accepted: code=%d stderr=%q", code, stderr.String())
	}
}

func TestValidateReceiptRequiresEvidenceFiles(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"validate-receipt"}, &stdout, &stderr)
	if code != 2 || !strings.Contains(stderr.String(), "--route, --receipt and --worker are required") || stdout.Len() != 0 {
		t.Fatalf("missing receipt evidence was not diagnosed: code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}

func TestDispatchPreparesExactRouteSelectedNativeHostContract(t *testing.T) {
	routePath := filepath.Join(t.TempDir(), "route.json")
	route := core.RouteResult{Workers: []core.WorkerTask{{
		ID: "mechanical", Model: "gpt-5.6-luna", SessionKey: "session-1", Writes: false, MaxAttempts: 1,
	}}}
	routeData, err := json.Marshal(route)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(routePath, routeData, 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := run([]string{
		"dispatch", "--route", routePath, "--worker", "mechanical", "--workspace", t.TempDir(),
		"--output-dir", filepath.Join(t.TempDir(), "dispatch"), "--dry-run",
	}, &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 {
		t.Fatalf("dispatch dry run failed: code=%d stderr=%q", code, stderr.String())
	}
	var result core.DispatchResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil || result.Status != "native-host-dispatch-required" || !result.NativeHostRequired || result.RequestedModel != "gpt-5.6-luna" || result.PreparedPromptSHA256 == "" {
		t.Fatalf("dispatch did not preserve worker route: result=%#v err=%v", result, err)
	}
}

func TestDecodeRouteAcceptsPowerShellUTF8BOM(t *testing.T) {
	route, err := decodeRoute(append([]byte{0xEF, 0xBB, 0xBF}, []byte(`{"workers":[{"id":"mechanical"}]}`)...))
	if err != nil || len(route.Workers) != 1 || route.Workers[0].ID != "mechanical" {
		t.Fatalf("BOM route was not decoded: route=%#v err=%v", route, err)
	}
}

func TestKnowledgeCommandsRecordAndQueryWithoutScanning(t *testing.T) {
	store := t.TempDir()
	evidence := filepath.Join(t.TempDir(), "evidence.json")
	solution := filepath.Join(t.TempDir(), "solution.md")
	receipt := `{"schema_version":1,"type":"wuji-verification-receipt","passed":true,"verifier":"go-test","verified_at":"` + time.Now().UTC().Format(time.RFC3339) + `"}`
	if err := os.WriteFile(evidence, []byte(receipt), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(solution, []byte("verified solution"), 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := run([]string{
		"knowledge-record", "--store", store, "--kind", "solution", "--key", "browser timeout", "--scope", "global",
		"--summary", "Use the verified bounded wait.", "--location", solution, "--verification", evidence, "--tags", "browser,timeout",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("knowledge-record failed: code=%d stderr=%q", code, stderr.String())
	}
	var record core.KnowledgeRecord
	if err := json.Unmarshal(stdout.Bytes(), &record); err != nil || record.ID == "" {
		t.Fatalf("knowledge record output is invalid: record=%#v err=%v", record, err)
	}
	stdout.Reset()
	stderr.Reset()
	code = run([]string{
		"knowledge-query", "--store", store, "--trigger", "explicit-reuse", "--kind", "solution", "--key", "browser timeout", "--scope", "global",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("knowledge-query failed: code=%d stderr=%q", code, stderr.String())
	}
	var result core.KnowledgeQueryResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil || !result.ExactMatch || result.FullScan || len(result.Matches) != 1 {
		t.Fatalf("knowledge query output is invalid: result=%#v err=%v", result, err)
	}
}

func TestKnowledgeQueryRejectsNormalStartup(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"knowledge-query", "--key", "browser timeout"}, &stdout, &stderr)
	if code != 2 || !strings.Contains(stderr.String(), "event trigger") {
		t.Fatalf("normal startup queried experience graph: code=%d stderr=%q", code, stderr.String())
	}
}

func TestKnowledgeRecordFailureFeedbackBridge(t *testing.T) {
	root := t.TempDir()
	requirements, executions := filepath.Join(root, "requirements"), filepath.Join(root, "execution")
	requirement, err := core.UpsertRequirement(requirements, core.RequirementInput{ID: "goal", Summary: "Persist verified failures.", Sources: []string{"message:sha256:" + strings.Repeat("f", 64)}})
	if err != nil {
		t.Fatal(err)
	}
	feedbackID := cliFeedbackCandidate(t, root, requirements, executions, requirement.VersionID, "failed-node", "failed")
	feedbackStore, knowledgeStore := filepath.Join(root, "execution-feedback"), filepath.Join(root, "knowledge")
	location, evidence := filepath.Join(root, "solution.md"), filepath.Join(root, "verification.json")
	if err := os.WriteFile(location, []byte("bounded recovery"), 0o600); err != nil {
		t.Fatal(err)
	}
	receipt := `{"schema_version":1,"type":"wuji-verification-receipt","passed":true,"verifier":"go-test","verified_at":"` + time.Now().UTC().Format(time.RFC3339) + `"}`
	if err := os.WriteFile(evidence, []byte(receipt), 0o600); err != nil {
		t.Fatal(err)
	}
	base := []string{"knowledge-record", "--store", knowledgeStore, "--feedback-store", feedbackStore, "--feedback-id", feedbackID, "--kind", "failure", "--key", "provider unavailable", "--scope", "global", "--summary", "Use bounded availability fallback.", "--root-cause", "Provider unavailable before generation.", "--location", location, "--verification", evidence}
	var stdout, stderr bytes.Buffer
	if code := run(base, &stdout, &stderr); code != 0 {
		t.Fatalf("verified failed feedback was rejected: code=%d stderr=%q", code, stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	if code := run([]string{"knowledge-query", "--store", knowledgeStore, "--trigger", "failure", "--kind", "failure", "--key", "provider unavailable", "--scope", "global"}, &stdout, &stderr); code != 0 || !strings.Contains(stdout.String(), "bounded availability fallback") {
		t.Fatalf("failure-triggered reuse could not read admitted feedback: code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	successID := cliFeedbackCandidate(t, root, requirements, executions, requirement.VersionID, "success-node", "succeeded")
	successArgs := append([]string{}, base...)
	for index := range successArgs {
		if successArgs[index] == feedbackID {
			successArgs[index] = successID
		}
	}
	stdout.Reset()
	stderr.Reset()
	if code := run(successArgs, &stdout, &stderr); code != 2 {
		t.Fatalf("success feedback was admitted: code=%d stderr=%q", code, stderr.String())
	}
	missingEvidence := append([]string{}, base...)
	for index := range missingEvidence {
		if missingEvidence[index] == evidence {
			missingEvidence[index] = filepath.Join(root, "missing.json")
		}
	}
	stdout.Reset()
	stderr.Reset()
	if code := run(missingEvidence, &stdout, &stderr); code != 2 {
		t.Fatalf("feedback without independent evidence was admitted: code=%d stderr=%q", code, stderr.String())
	}
}

func cliFeedbackCandidate(t *testing.T, root, requirements, executions, requirementVersion, id, status string) string {
	t.Helper()
	node, err := core.RecordVersionedExecutionNode(executions, requirements, core.VersionedExecutionNodeInput{ExecutionNodeInput: core.ExecutionNodeInput{ID: id, Authority: "staff", Goal: "test", RequirementRevisions: []string{requirementVersion}, Model: "gpt-5.6-terra", ModelReason: "test", NetworkBoundary: "none", WriteBoundary: "none", BranchBoundary: "current"}, TaskInstanceID: "task", GraphVersion: "graph-" + id, AttemptID: "attempt-1"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := core.RecordVersionedExecutionResult(executions, requirements, core.VersionedExecutionResultInput{ExecutionResultInput: core.ExecutionResultInput{ID: node.ID, Status: status}, TaskInstanceID: "task", GraphVersion: "graph-" + id, AttemptID: "attempt-1"}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, "execution-feedback", "v1", "records.json"))
	if err != nil {
		t.Fatal(err)
	}
	var ledger struct {
		Records []core.ExecutionFeedbackRecord `json:"records"`
	}
	if err := json.Unmarshal(data, &ledger); err != nil {
		t.Fatal(err)
	}
	for _, record := range ledger.Records {
		if record.ExecutionVersion == node.VersionID {
			return record.ID
		}
	}
	t.Fatal("execution feedback candidate was not written")
	return ""
}

func TestTaskCircuitCommandsPersistAndBlockNoProgress(t *testing.T) {
	store := t.TempDir()
	arguments := []string{"--store", store, "--task", "repair-routing", "--strategy", "direct-fix", "--policy", "bounded-repair-v1", "--max-no-progress", "1", "--attempt", "attempt-a"}
	var stdout, stderr bytes.Buffer
	code := run(append([]string{"task-gate"}, arguments...), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("task-gate failed: code=%d stderr=%q", code, stderr.String())
	}
	var gate core.TaskCircuitResult
	if err := json.Unmarshal(stdout.Bytes(), &gate); err != nil || !gate.Allowed {
		t.Fatalf("task-gate returned invalid result: result=%#v err=%v", gate, err)
	}
	stdout.Reset()
	stderr.Reset()
	code = run(append(append([]string{"task-record"}, arguments...), "--outcome", "no-progress"), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("task-record failed: code=%d stderr=%q", code, stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	code = run(append(append([]string{"task-gate"}, arguments...), "--attempt", "attempt-b"), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("blocked task-gate returned an error: code=%d stderr=%q", code, stderr.String())
	}
	if err := json.Unmarshal(stdout.Bytes(), &gate); err != nil || gate.Allowed || gate.Reason != "no-progress-limit" {
		t.Fatalf("task-gate did not expose the persisted circuit: result=%#v err=%v", gate, err)
	}
}

func TestLineageSyncPersistsCatalogInsideSelectedRoot(t *testing.T) {
	root := t.TempDir()
	sourceRoot := filepath.Join(root, "source")
	if err := os.MkdirAll(filepath.Join(sourceRoot, "assets"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sourceRoot, "SKILL.md"), []byte("entrypoint"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sourceRoot, "assets", "workflow.md"), []byte("asset"), 0o600); err != nil {
		t.Fatal(err)
	}
	manifest := core.Manifest{
		ID: "lineage", Description: "lineage fixture", Triggers: []string{"lineage"}, Status: "callable",
		PrimarySkill: "native", HostCallable: true, Fallback: "native fallback",
		Probe:   &core.Probe{Command: "smoke", Kind: "smoke"},
		Sources: []core.Source{{ID: "atom", Lifecycle: "callable", Entrypoint: "SKILL.md", Globs: []string{"${ROOT}/source"}, Required: []string{"SKILL.md"}}},
		Genome:  &core.FusionGenome{SchemaVersion: 1, Species: "lineage", Revision: "rev-1", Adapters: []core.FusionAdapter{{ID: "workflow", Domain: "workflow", Source: "atom", Entrypoint: "SKILL.md", Assets: []string{"assets/workflow.md"}}}},
	}
	manifestPath := filepath.Join(root, "capabilities", "lineage", "manifest.json")
	if err := os.MkdirAll(filepath.Dir(manifestPath), 0o755); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifestPath, data, 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := run([]string{"lineage-sync", "--root", root}, &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 {
		t.Fatalf("lineage-sync failed: code=%d stderr=%q", code, stderr.String())
	}
	var result core.LineageSyncResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil || result.NodeCount != 4 || result.RejectionCount != 0 || !strings.HasPrefix(result.CatalogPath, root) {
		t.Fatalf("lineage-sync did not return the persisted root-owned catalog: result=%#v err=%v", result, err)
	}
	stdout.Reset()
	stderr.Reset()
	code = run([]string{"asset-select", "--root", root, "--capability", "lineage", "--domain", "workflow"}, &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 {
		t.Fatalf("asset-select failed: code=%d stderr=%q", code, stderr.String())
	}
	var contract core.AssetInvocationContract
	if err := json.Unmarshal(stdout.Bytes(), &contract); err != nil || contract.AssetID == "" || contract.Invocation.SourceID != "atom" || contract.EntrypointSHA256 == "" {
		t.Fatalf("asset-select did not return a trusted invocation contract: contract=%#v err=%v", contract, err)
	}
}

func TestGovernanceCommandsBindReceiptsAndRespectACL(t *testing.T) {
	root := t.TempDir()
	requirementStore := filepath.Join(root, "requirements")
	executionStore := filepath.Join(root, "execution")
	acceptanceStore := filepath.Join(root, "acceptance")
	conversationStore := filepath.Join(root, "conversation")
	provenanceStore := filepath.Join(root, "provenance")
	sourceStore := filepath.Join(root, "sources")
	source := "message:sha256:" + strings.Repeat("a", 64)
	artifact := "artifact:sha256:" + strings.Repeat("b", 64)
	verification := "verification:sha256:" + strings.Repeat("c", 64)
	var stdout, stderr bytes.Buffer

	code := run([]string{"requirement-record", "--store", requirementStore, "--id", "governance", "--summary", "Keep governance evidence bounded.", "--sources", source}, &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 {
		t.Fatalf("requirement-record failed: code=%d stderr=%q", code, stderr.String())
	}
	var requirement core.RequirementGraphNode
	if err := json.Unmarshal(stdout.Bytes(), &requirement); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	stderr.Reset()
	code = run([]string{"execution-record", "--store", executionStore, "--requirements-store", requirementStore, "--id", "verify-governance", "--authority", "deterministic-executor", "--goal", "verify bounded evidence", "--requirements", requirement.VersionID, "--model", "gpt-5.6-terra", "--model-reason", "deterministic verification", "--network-boundary", "none", "--write-boundary", "scoped-artifact-write", "--branch-boundary", "current", "--task-instance", "task-governance", "--graph-version", "graph-1", "--attempt", "attempt-1"}, &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 {
		t.Fatalf("execution-record failed: code=%d stderr=%q", code, stderr.String())
	}
	var execution core.ExecutionGraphNode
	if err := json.Unmarshal(stdout.Bytes(), &execution); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	stderr.Reset()
	code = run([]string{"execution-result", "--store", executionStore, "--requirements-store", requirementStore, "--id", execution.ID, "--status", "succeeded", "--artifacts", artifact, "--verification", verification, "--task-instance", "task-governance", "--graph-version", "graph-1", "--attempt", "attempt-1"}, &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 {
		t.Fatalf("execution-result failed: code=%d stderr=%q", code, stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	code = run([]string{"acceptance-reconcile", "--store", acceptanceStore, "--requirements-store", requirementStore, "--execution-store", executionStore, "--id", "governance-acceptance", "--requirement", requirement.VersionID, "--execution", execution.VersionID}, &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 {
		t.Fatalf("acceptance-reconcile failed: code=%d stderr=%q", code, stderr.String())
	}
	var acceptance core.AcceptanceRecord
	if err := json.Unmarshal(stdout.Bytes(), &acceptance); err != nil || acceptance.AcceptedBy != "deterministic-evidence-verifier" || len(acceptance.VerificationHandles) != 1 {
		t.Fatalf("acceptance output is invalid: acceptance=%#v err=%v", acceptance, err)
	}
	stdout.Reset()
	stderr.Reset()
	code = run([]string{"conversation-link", "--store", conversationStore, "--requirements-store", requirementStore, "--revision", requirement.VersionID, "--messages", "host-message:opaque-001"}, &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 {
		t.Fatalf("conversation-link failed: code=%d stderr=%q", code, stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	code = run([]string{"conversation-resolve", "--store", conversationStore, "--message", "host-message:opaque-001"}, &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 {
		t.Fatalf("conversation-resolve failed: code=%d stderr=%q", code, stderr.String())
	}
	var evidence []core.ConversationEvidenceRecord
	if err := json.Unmarshal(stdout.Bytes(), &evidence); err != nil || len(evidence) != 1 || evidence[0].Revision != requirement.VersionID {
		t.Fatalf("conversation-resolve output is invalid: evidence=%#v err=%v", evidence, err)
	}
	stdout.Reset()
	stderr.Reset()
	code = run([]string{"provenance-record", "--store", provenanceStore, "--id", "governance-origin", "--subject", requirement.VersionID, "--predicate", "derived-from", "--target", "host-message:opaque-001", "--readers", "aji"}, &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 {
		t.Fatalf("provenance-record failed: code=%d stderr=%q", code, stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	code = run([]string{"provenance-resolve", "--store", provenanceStore, "--subject", requirement.VersionID, "--principal", "staff"}, &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 {
		t.Fatalf("provenance-resolve failed: code=%d stderr=%q", code, stderr.String())
	}
	var resolved core.ProvenanceResolveResult
	if err := json.Unmarshal(stdout.Bytes(), &resolved); err != nil || resolved.Denied != 1 || len(resolved.Entries) != 0 {
		t.Fatalf("provenance-resolve bypassed ACL: result=%#v err=%v", resolved, err)
	}
	stdout.Reset()
	stderr.Reset()
	code = run([]string{"source-assess", "--store", sourceStore, "--source", "native-source", "--version", "2.0.1", "--decision", "deferred", "--reason", "wait for an explicit reuse"}, &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 {
		t.Fatalf("source-assess failed: code=%d stderr=%q", code, stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	code = run([]string{"graph-govern", "--graph", "source-assessments", "--store", sourceStore}, &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 {
		t.Fatalf("graph-govern failed: code=%d stderr=%q", code, stderr.String())
	}
	var maintenance core.GraphMaintenanceResult
	if err := json.Unmarshal(stdout.Bytes(), &maintenance); err != nil || !maintenance.Validated || maintenance.Policy.Graph != "source-assessments" {
		t.Fatalf("graph-govern output is invalid: result=%#v err=%v", maintenance, err)
	}
}

func TestSecurityGateAndOfficerSelectCommandsStayDeterministic(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"security-gate", "--kind", "file-write", "--target", "result.json"}, &stdout, &stderr)
	if code != 1 || stderr.Len() != 0 {
		t.Fatalf("unconfirmed security gate did not block cleanly: code=%d stderr=%q", code, stderr.String())
	}
	var gate core.SecurityGateResult
	if err := json.Unmarshal(stdout.Bytes(), &gate); err != nil || gate.Allowed || gate.Decision != "user-confirmation-required" {
		t.Fatalf("security gate result is invalid: result=%#v err=%v", gate, err)
	}
	stdout.Reset()
	stderr.Reset()
	code = run([]string{"officer-select", "--query", "diagnose a timeout failure and provide test evidence"}, &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 {
		t.Fatalf("officer-select failed: code=%d stderr=%q", code, stderr.String())
	}
	var recommendations []core.OfficerRecommendation
	if err := json.Unmarshal(stdout.Bytes(), &recommendations); err != nil || len(recommendations) != 1 || recommendations[0].Role != "internal-quality-check" {
		t.Fatalf("officer-select returned invalid recommendations: %#v err=%v", recommendations, err)
	}
}

func TestGeneralStaffCommandsPersistIncrementalUpdatesAndVetoReplacement(t *testing.T) {
	store := t.TempDir()
	var stdout, stderr bytes.Buffer
	code := run([]string{"staff-create", "--store", store, "--task-instance", "task-1", "--session-key", "session-1", "--requirement-version", "requirements-1", "--graph-version", "graph-1"}, &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 {
		t.Fatalf("staff-create failed: code=%d stderr=%q", code, stderr.String())
	}
	var created core.GeneralStaffState
	if err := json.Unmarshal(stdout.Bytes(), &created); err != nil || created.Current.SessionKey != "session-1" {
		t.Fatalf("staff-create output is invalid: state=%#v err=%v", created, err)
	}

	stdout.Reset()
	stderr.Reset()
	code = run([]string{"staff-update", "--store", store, "--task-instance", "task-1", "--session-key", "session-1", "--requirement-version", "requirements-2", "--graph-version", "graph-2"}, &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 {
		t.Fatalf("staff-update failed: code=%d stderr=%q", code, stderr.String())
	}
	var updated core.GeneralStaffState
	if err := json.Unmarshal(stdout.Bytes(), &updated); err != nil || updated.Current.SessionKey != "session-1" || updated.Current.GraphVersion != "graph-2" || len(updated.Replaced) != 0 {
		t.Fatalf("incremental staff update output is invalid: state=%#v err=%v", updated, err)
	}

	stdout.Reset()
	stderr.Reset()
	code = run([]string{"staff-status", "--store", store}, &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 {
		t.Fatalf("staff-status failed: code=%d stderr=%q", code, stderr.String())
	}
	var status core.GeneralStaffState
	if err := json.Unmarshal(stdout.Bytes(), &status); err != nil || status.Current.RequirementVersion != "requirements-2" {
		t.Fatalf("staff-status output is invalid: state=%#v err=%v", status, err)
	}

	stdout.Reset()
	stderr.Reset()
	code = run([]string{"staff-update", "--store", store, "--task-instance", "task-1", "--session-key", "session-2", "--requirement-version", "requirements-3", "--graph-version", "graph-3", "--veto"}, &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 {
		t.Fatalf("staff veto update failed: code=%d stderr=%q", code, stderr.String())
	}
	if err := json.Unmarshal(stdout.Bytes(), &updated); err != nil || updated.Current.SessionKey != "session-2" || len(updated.Replaced) != 1 || updated.Replaced[0].SessionKey != "session-1" {
		t.Fatalf("staff veto did not replace the instance: state=%#v err=%v", updated, err)
	}
}

func TestRequirementGraphCommandsPersistAndProjectBoundedContext(t *testing.T) {
	store := t.TempDir()
	var stdout, stderr bytes.Buffer
	code := run([]string{"requirement-record", "--store", store, "--id", "single-router", "--summary", "Keep Aji as the only router.", "--sources", "message:sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("requirement-record failed: code=%d stderr=%q", code, stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	code = run([]string{"decision-record", "--store", store, "--id", "accept-single-router", "--summary", "Accept the single-router policy.", "--sources", "message:sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", "--requirements", "single-router", "--status", "accepted"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("decision-record failed: code=%d stderr=%q", code, stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	code = run([]string{"requirement-project", "--store", store, "--id", "accept-single-router", "--max-bytes", "1024"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("requirement-project failed: code=%d stderr=%q", code, stderr.String())
	}
	var projection core.RequirementGraphProjection
	if err := json.Unmarshal(stdout.Bytes(), &projection); err != nil || projection.PayloadBytes > 1024 || len(projection.Nodes) != 2 || projection.ArtifactPath == "" {
		t.Fatalf("requirement-project output is invalid: projection=%#v err=%v", projection, err)
	}
	if _, err := core.LoadRequirementGraphProjection(projection.ArtifactPath); err != nil {
		t.Fatalf("projection artifact was not independently loadable: %v", err)
	}
}

func TestGraphSyncCommandBuildsBoundedWorkspaceIndex(t *testing.T) {
	workspace := t.TempDir()
	if err := os.WriteFile(filepath.Join(workspace, "main.go"), []byte("package main\nfunc MainFeature() {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := run([]string{"graph-sync", "--workspace", workspace}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("graph-sync failed: code=%d stderr=%q", code, stderr.String())
	}
	var result core.WorkspaceGraphSyncResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil || result.FileCount != 1 || result.MaxRefsPerTerm == 0 {
		t.Fatalf("graph sync output is invalid: result=%#v err=%v", result, err)
	}
}

func TestRequirementFlagsAndExecutionCommandsPersistCommonContract(t *testing.T) {
	root := t.TempDir()
	reqStore := filepath.Join(root, "requirements")
	execStore := filepath.Join(root, "execution")
	source := "message:sha256:" + strings.Repeat("c", 64)
	var stdout, stderr bytes.Buffer
	code := run([]string{"requirement-record", "--store", reqStore, "--id", "common-contract", "--summary", "Preserve shared intent.", "--goal", "single durable contract", "--wants", "sparse context,deterministic gates", "--avoids", "parallel router", "--constraints", "only assigned execution nodes write artifacts", "--preferences", "bounded artifacts", "--priority", "high", "--acceptance", "go test passes", "--decisions", "keep wuji CLI", "--open-questions", "which host owns execution", "--conflicts", "Pi-only session features", "--source-messages", "msg-1", "--sources", source}, &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 {
		t.Fatalf("requirement flags failed: code=%d stderr=%q", code, stderr.String())
	}
	var requirement core.RequirementGraphNode
	if err := json.Unmarshal(stdout.Bytes(), &requirement); err != nil || requirement.Goal != "single durable contract" || len(requirement.Wants) != 2 || requirement.Priority != "high" {
		t.Fatalf("requirement flags were not persisted: node=%#v err=%v", requirement, err)
	}
	stdout.Reset()
	code = run([]string{"decision-record", "--store", reqStore, "--id", "common-decision", "--summary", "Accept shared contract.", "--goal", "align Codex and Pi principles", "--acceptance", "same invariants", "--source-messages", "msg-2", "--sources", source, "--requirements", requirement.VersionID, "--status", "accepted"}, &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 {
		t.Fatalf("decision flags failed: code=%d stderr=%q", code, stderr.String())
	}
	var decision core.RequirementGraphNode
	if err := json.Unmarshal(stdout.Bytes(), &decision); err != nil || len(decision.DependsOn) != 1 || decision.DependsOn[0] != requirement.VersionID {
		t.Fatalf("decision did not bind exact requirement: node=%#v err=%v", decision, err)
	}
	stdout.Reset()
	code = run([]string{"execution-record", "--store", execStore, "--requirements-store", reqStore, "--id", "execute-common", "--authority", "deterministic-executor", "--goal", "verify common contract", "--requirements", requirement.VersionID, "--model", "gpt-5.6-terra", "--model-reason", "execution-node boundary", "--network-boundary", "none", "--write-boundary", "scoped-artifact-write", "--branch-boundary", "current", "--task-instance", "task-common", "--graph-version", "graph-1", "--attempt", "attempt-1"}, &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 {
		t.Fatalf("execution-record failed: code=%d stderr=%q", code, stderr.String())
	}
	var execution core.ExecutionGraphNode
	if err := json.Unmarshal(stdout.Bytes(), &execution); err != nil || execution.VersionID == "" || execution.RequirementRevisions[0] != requirement.VersionID {
		t.Fatalf("execution node output is invalid: node=%#v err=%v", execution, err)
	}
	stdout.Reset()
	code = run([]string{"execution-result", "--store", execStore, "--requirements-store", reqStore, "--id", execution.ID, "--status", "succeeded", "--verification", "evidence:sha256:" + strings.Repeat("d", 64), "--task-instance", "task-common", "--graph-version", "graph-1", "--attempt", "attempt-1"}, &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 {
		t.Fatalf("execution-result failed: code=%d stderr=%q", code, stderr.String())
	}
	stdout.Reset()
	code = run([]string{"execution-project", "--store", execStore, "--requirements-store", reqStore, "--id", execution.ID, "--max-bytes", "1024"}, &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 {
		t.Fatalf("execution-project failed: code=%d stderr=%q", code, stderr.String())
	}
	var projection core.ExecutionGraphProjection
	if err := json.Unmarshal(stdout.Bytes(), &projection); err != nil || projection.PayloadBytes > 1024 || projection.ArtifactPath == "" {
		t.Fatalf("execution projection output is invalid: projection=%#v err=%v", projection, err)
	}
}
