package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestContextSelectHonorsBudgetAndChineseQuery(t *testing.T) {
	root := t.TempDir()
	content := "package demo\n\nfunc ResolveModel() string {\n  // 修复模型选择失败\n  return \"sol\"\n}\n"
	if err := os.WriteFile(filepath.Join(root, "model.go"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "noise.md"), []byte("unrelated notes"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := SelectContext(root, "修复模型选择失败", 512)
	if err != nil {
		t.Fatal(err)
	}
	if got.SelectedBytes > 512 {
		t.Fatalf("budget exceeded: %d", got.SelectedBytes)
	}
	if len(got.Excerpts) == 0 || got.Excerpts[0].Path != "model.go" {
		t.Fatalf("wrong retrieval result: %#v", got.Excerpts)
	}
}

func TestBoundedWorkspaceScanSkipsExternalSymlink(t *testing.T) {
	workspace := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside.go")
	if err := os.WriteFile(outside, []byte("package outside\nfunc EscapedAnchor() {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(workspace, "escaped.go")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("source symlinks are not available: %v", err)
	}
	files, scanned, truncated, err := boundedWorkspaceScan(workspace, []string{"escapedanchor"}, 512)
	if err != nil {
		t.Fatal(err)
	}
	if scanned != 0 || len(files) != 0 || truncated {
		t.Fatalf("external symlink entered fallback scan: files=%#v scanned=%d truncated=%t", files, scanned, truncated)
	}
}

func TestContextSelectRejectsInvalidInputs(t *testing.T) {
	for _, test := range []struct {
		query  string
		budget int
	}{
		{"", 512},
		{"valid query", 0},
		{"valid query", -1},
	} {
		if _, err := SelectContext(t.TempDir(), test.query, test.budget); err == nil {
			t.Fatalf("invalid input was accepted: %#v", test)
		}
	}
}

func TestContextSelectReportsScannerErrors(t *testing.T) {
	root := t.TempDir()
	content := "needle " + strings.Repeat("x", 600*1024)
	if err := os.WriteFile(filepath.Join(root, "large.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := SelectContext(root, "needle", 1024); err == nil || !strings.Contains(err.Error(), "scan large.md") {
		t.Fatalf("scanner failure was swallowed: %v", err)
	}
}

func TestContextSelectSkipsExcerptThatCannotFit(t *testing.T) {
	root := t.TempDir()
	large := "needle " + strings.Repeat("x", 400*1024)
	if err := os.WriteFile(filepath.Join(root, "needle-big.md"), []byte(large), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "small.md"), []byte("needle fits here\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := SelectContext(root, "needle", 512)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Excerpts) != 1 || got.Excerpts[0].Path != "small.md" {
		t.Fatalf("later fitting excerpt was skipped: %#v", got.Excerpts)
	}
}

func TestContextSelectRejectsQueriesWithoutMatches(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "notes.md"), []byte("unrelated notes\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := SelectContext(root, "missing symbol", 512); err == nil || !strings.Contains(err.Error(), "no matching context excerpts") {
		t.Fatalf("empty context result was accepted: %v", err)
	}
}

func TestContextSelectIncludesFrontendSourcesAndPreservesIndent(t *testing.T) {
	root := t.TempDir()
	content := "<template>\n  <section>\n    BoardAtmosphere stage atmosphere\n  </section>\n</template>\n"
	if err := os.WriteFile(filepath.Join(root, "BoardAtmosphere.vue"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := SelectContext(root, "BoardAtmosphere stage atmosphere", 1024)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Excerpts) != 1 || got.Excerpts[0].Path != "BoardAtmosphere.vue" {
		t.Fatalf("frontend source was not selected: %#v", got.Excerpts)
	}
	if !strings.Contains(got.Excerpts[0].Text, "    BoardAtmosphere") {
		t.Fatalf("excerpt indentation was lost: %q", got.Excerpts[0].Text)
	}
}

func TestContextSelectPrioritizesExplicitCodePathOverGeneratedLocks(t *testing.T) {
	root := t.TempDir()
	codeDir := filepath.Join(root, "internal", "core")
	if err := os.MkdirAll(codeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(codeDir, "route.go"), []byte("package core\n\nfunc delegationEconomics() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	lockNoise := strings.Repeat("delegation economics internal/core/route.go ", 200)
	if err := os.WriteFile(filepath.Join(root, "pnpm-lock.yaml"), []byte(lockNoise), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := SelectContext(root, "implement delegation economics in internal/core/route.go", 2048)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Excerpts) == 0 || got.Excerpts[0].Path != "internal/core/route.go" {
		t.Fatalf("explicit code path lost to generated noise: %#v", got.Excerpts)
	}
	for _, excerpt := range got.Excerpts {
		if excerpt.Path == "pnpm-lock.yaml" {
			t.Fatalf("generated lockfile entered context: %#v", got.Excerpts)
		}
	}
}

func TestContextQualityDoesNotTreatPathOnlyAsContentAnchor(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "auth.go"), []byte("package auth\n\nfunc Login() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := SelectContext(root, "fix auth.go", 1024)
	if err != nil {
		t.Fatal(err)
	}
	if got.ContentAnchorCount != 0 {
		t.Fatalf("path-only match was counted as a content anchor: %#v", got)
	}
}
