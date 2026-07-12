package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func writeGraphFixture(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestWorkspaceGraphDetectsSameMetadataContentChange(t *testing.T) {
	workspace := t.TempDir()
	path := filepath.Join(workspace, "feature.go")
	original := "package feature\nfunc Stable() string { return \"alpha\" }\n"
	changed := "package feature\nfunc Stable() string { return \"bravo\" }\n"
	if len(original) != len(changed) {
		t.Fatal("fixture contents must have the same size")
	}
	writeGraphFixture(t, path, original)
	if _, err := SyncWorkspaceGraph(workspace); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(changed), 0o600); err != nil {
		t.Fatal(err)
	}
	originalTime := info.ModTime()
	if err := os.Chtimes(path, originalTime, originalTime); err != nil {
		t.Fatal(err)
	}
	normalizedWorkspace, err := normalizeWorkspacePath(workspace)
	if err != nil {
		t.Fatal(err)
	}
	files, stats, err := queryWorkspaceGraph(normalizedWorkspace, []string{"stable"})
	if !errors.Is(err, errWorkspaceGraphMissing) || len(files) != 0 || stats.FallbackReason != "stale-index" {
		t.Fatalf("same-metadata content change was not rejected: files=%#v stats=%#v err=%v", files, stats, err)
	}
}

func TestWorkspaceGraphRoutesContextThroughBoundedCandidates(t *testing.T) {
	workspace := t.TempDir()
	alpha := filepath.Join(workspace, "internal", "alpha.go")
	writeGraphFixture(t, alpha, "package internal\n\nfunc AlphaFeature() string { return \"alpha\" }\n")
	writeGraphFixture(t, filepath.Join(workspace, "internal", "alpha_test.go"), "package internal\n\nfunc TestAlphaFeature() { AlphaFeature() }\n")
	for index := 0; index < 20; index++ {
		writeGraphFixture(t, filepath.Join(workspace, "other", fmt.Sprintf("file_%02d.go", index)), fmt.Sprintf("package other\nfunc Unrelated%02d() {}\n", index))
	}
	syncResult, err := SyncWorkspaceGraph(workspace)
	if err != nil {
		t.Fatal(err)
	}
	if syncResult.FileCount != 22 || syncResult.MaxTermsPerFile != workspaceGraphMaxTermsPerFile || syncResult.MaxRefsPerTerm != workspaceGraphMaxRefsPerTerm || syncResult.MaxLookups != workspaceGraphMaxLookups {
		t.Fatalf("unexpected graph sync result: %#v", syncResult)
	}
	result, err := SelectContext(workspace, "fix AlphaFeature in internal/alpha.go", 2048)
	if err != nil {
		t.Fatal(err)
	}
	if result.RetrievalMode != "workspace-graph" || result.GraphLookups == 0 || result.IndexedFiles != 22 {
		t.Fatalf("context did not use the workspace graph: %#v", result)
	}
	if result.CandidateFiles >= result.IndexedFiles || result.ScannedFiles != result.CandidateFiles {
		t.Fatalf("context did not narrow before source reads: candidates=%d indexed=%d scanned=%d", result.CandidateFiles, result.IndexedFiles, result.ScannedFiles)
	}
	if len(result.Excerpts) == 0 || result.Excerpts[0].Path != "internal/alpha.go" {
		t.Fatalf("unexpected graph-selected excerpts: %#v", result.Excerpts)
	}

	writeGraphFixture(t, alpha, "package internal\n\nfunc AlphaFeature() string { return \"alpha-updated\" }\n")
	rebuilt, err := SelectContext(workspace, "fix AlphaFeature in internal/alpha.go", 2048)
	if err != nil {
		t.Fatal(err)
	}
	if rebuilt.RetrievalMode != "workspace-graph-rebuilt" || rebuilt.FallbackReason != "stale-index" || !strings.Contains(rebuilt.Excerpts[0].Text, "alpha-updated") {
		t.Fatalf("stale graph was not rebuilt: %#v", rebuilt)
	}
}

func TestWorkspaceGraphCapsIndexLookups(t *testing.T) {
	workspace := t.TempDir()
	writeGraphFixture(t, filepath.Join(workspace, "feature.go"), "package feature\nfunc Anchor() {}\n")
	if _, err := SyncWorkspaceGraph(workspace); err != nil {
		t.Fatal(err)
	}
	normalizedWorkspace, err := normalizeWorkspacePath(workspace)
	if err != nil {
		t.Fatal(err)
	}
	terms := make([]string, workspaceGraphMaxLookups+20)
	for index := range terms {
		terms[index] = fmt.Sprintf("term-%03d", index)
	}
	_, stats, err := queryWorkspaceGraph(normalizedWorkspace, terms)
	if err != nil {
		t.Fatal(err)
	}
	if stats.GraphLookups != workspaceGraphMaxLookups {
		t.Fatalf("workspace graph lookup budget was not enforced: %#v", stats)
	}
}

func TestWorkspaceGraphRelationsAndDeletionAreRebuilt(t *testing.T) {
	workspace := t.TempDir()
	source := filepath.Join(workspace, "feature.go")
	testFile := filepath.Join(workspace, "feature_test.go")
	writeGraphFixture(t, source, "package feature\ntype Feature struct{}\n")
	writeGraphFixture(t, testFile, "package feature\nfunc TestFeature() {}\n")
	if _, err := SyncWorkspaceGraph(workspace); err != nil {
		t.Fatal(err)
	}
	activeDir, err := activeWorkspaceGraphDir(workspace)
	if err != nil {
		t.Fatal(err)
	}
	nodePath := filepath.Join(activeDir, "nodes", graphHash("feature_test.go")+".json")
	data, err := os.ReadFile(nodePath)
	if err != nil {
		t.Fatal(err)
	}
	var node workspaceGraphNode
	if err := json.Unmarshal(data, &node); err != nil {
		t.Fatal(err)
	}
	foundTestRelation := false
	for _, relation := range node.Relations {
		foundTestRelation = foundTestRelation || relation.Predicate == "tests"
	}
	if !foundTestRelation {
		t.Fatalf("test relation was not indexed: %#v", node.Relations)
	}
	if err := os.Remove(testFile); err != nil {
		t.Fatal(err)
	}
	result, err := SyncWorkspaceGraph(workspace)
	if err != nil {
		t.Fatal(err)
	}
	if result.FileCount != 1 {
		t.Fatalf("deleted source remained in rebuilt graph: %#v", result)
	}
	if _, err := os.Stat(nodePath); !os.IsNotExist(err) {
		t.Fatalf("deleted graph node remained after generation swap: %v", err)
	}
}

func TestWorkspaceGraphRejectsTamperedNodePathAndReference(t *testing.T) {
	workspace := t.TempDir()
	writeGraphFixture(t, filepath.Join(workspace, "feature.go"), "package feature\nfunc Anchor() {}\n")
	if _, err := SyncWorkspaceGraph(workspace); err != nil {
		t.Fatal(err)
	}
	activeDir, err := activeWorkspaceGraphDir(workspace)
	if err != nil {
		t.Fatal(err)
	}
	nodePath := filepath.Join(activeDir, "nodes", graphHash("feature.go")+".json")
	data, err := os.ReadFile(nodePath)
	if err != nil {
		t.Fatal(err)
	}
	var node workspaceGraphNode
	if err := json.Unmarshal(data, &node); err != nil {
		t.Fatal(err)
	}
	node.Path = "../../outside.go"
	tampered, _ := json.Marshal(node)
	if err := os.WriteFile(nodePath, tampered, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := queryWorkspaceGraph(workspace, []string{"anchor"}); !errors.Is(err, errWorkspaceGraphMissing) {
		t.Fatalf("path traversal node was accepted: %v", err)
	}

	if _, err := SyncWorkspaceGraph(workspace); err != nil {
		t.Fatal(err)
	}
	activeDir, err = activeWorkspaceGraphDir(workspace)
	if err != nil {
		t.Fatal(err)
	}
	refsPath := filepath.Join(activeDir, "terms", graphHash("anchor"), "refs.json")
	data, err = os.ReadFile(refsPath)
	if err != nil {
		t.Fatal(err)
	}
	var refs []workspaceGraphRef
	if err := json.Unmarshal(data, &refs); err != nil {
		t.Fatal(err)
	}
	refs[0].ID = strings.Repeat("0", 24)
	tampered, _ = json.Marshal(refs)
	if err := os.WriteFile(refsPath, tampered, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := queryWorkspaceGraph(workspace, []string{"anchor"}); !errors.Is(err, errWorkspaceGraphMissing) {
		t.Fatalf("forged graph reference was accepted: %v", err)
	}
}

func TestWorkspaceGraphRejectsSourceSymlinkEscape(t *testing.T) {
	workspace := t.TempDir()
	source := filepath.Join(workspace, "feature.go")
	content := "package feature\nfunc Anchor() {}\n"
	writeGraphFixture(t, source, content)
	if _, err := SyncWorkspaceGraph(workspace); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(t.TempDir(), "outside.go")
	writeGraphFixture(t, outside, content)
	if err := os.Remove(source); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, source); err != nil {
		t.Skipf("source symlinks are not available in this environment: %v", err)
	}
	if _, _, err := queryWorkspaceGraph(workspace, []string{"anchor"}); !errors.Is(err, errWorkspaceGraphMissing) {
		t.Fatalf("external source symlink was accepted: %v", err)
	}
}

func TestWorkspaceGraphReportsCandidateAndSourceBudgets(t *testing.T) {
	workspace := t.TempDir()
	for index := 0; index < workspaceGraphMaxCandidates+12; index++ {
		writeGraphFixture(t, filepath.Join(workspace, fmt.Sprintf("candidate_%03d.go", index)), "package candidate\n// commonanchor\n")
	}
	if _, err := SyncWorkspaceGraph(workspace); err != nil {
		t.Fatal(err)
	}
	_, stats, err := queryWorkspaceGraph(workspace, []string{"commonanchor"})
	if err != nil {
		t.Fatal(err)
	}
	if !stats.Truncated || stats.CandidateFiles != workspaceGraphMaxCandidates {
		t.Fatalf("candidate budget was not reported: %#v", stats)
	}

	largeWorkspace := t.TempDir()
	largeLine := "// sourcebudget " + strings.Repeat("x", 1000) + "\n"
	largeContent := "package large\n" + strings.Repeat(largeLine, 1900)
	for index := 0; index < 10; index++ {
		writeGraphFixture(t, filepath.Join(largeWorkspace, fmt.Sprintf("large_%02d.go", index)), largeContent)
	}
	if _, err := SyncWorkspaceGraph(largeWorkspace); err != nil {
		t.Fatal(err)
	}
	_, stats, err = queryWorkspaceGraph(largeWorkspace, []string{"sourcebudget"})
	if err != nil {
		t.Fatal(err)
	}
	if !stats.Truncated || stats.SourceBytes > workspaceGraphMaxSourceBytes {
		t.Fatalf("source byte budget was not reported: %#v", stats)
	}
}

func TestWorkspaceGraphConcurrentSyncPublishesOneCompleteGeneration(t *testing.T) {
	workspace := t.TempDir()
	writeGraphFixture(t, filepath.Join(workspace, "feature.go"), "package feature\nfunc Anchor() {}\n")
	const workers = 4
	errs := make(chan error, workers)
	var group sync.WaitGroup
	for index := 0; index < workers; index++ {
		group.Add(1)
		go func() {
			defer group.Done()
			_, err := SyncWorkspaceGraph(workspace)
			errs <- err
		}()
	}
	group.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	activeDir, err := activeWorkspaceGraphDir(workspace)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(activeDir, "meta.json")); err != nil {
		t.Fatalf("active generation is incomplete: %v", err)
	}
	entries, err := os.ReadDir(filepath.Join(workspaceGraphDir(workspace), "generations"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || strings.HasSuffix(entries[0].Name(), ".tmp") {
		t.Fatalf("concurrent sync left partial generations: %#v", entries)
	}
}
