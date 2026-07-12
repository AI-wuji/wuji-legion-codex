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
