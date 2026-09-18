package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AI-wuji/wuji-legion-codex-2.0/internal/core"
)

func TestReadBoundedContextModeFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "result.json")
	if err := os.WriteFile(path, []byte(strings.Repeat("x", core.ContextModeMaxResultBytes+1)), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := readBoundedContextModeFile(path, core.ContextModeMaxResultBytes); err == nil {
		t.Fatal("oversized CLI result file accepted")
	}
}

func TestJSONUnmarshalStrictRejectsTrailingContent(t *testing.T) {
	var value map[string]interface{}
	if err := jsonUnmarshalStrict([]byte(`{} {}`), &value); err == nil {
		t.Fatal("trailing JSON accepted")
	}
}
