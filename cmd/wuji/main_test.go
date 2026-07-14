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
		ID: "mechanical", Model: "gpt-5.6-luna", SessionKey: "session-1", Writes: false,
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
