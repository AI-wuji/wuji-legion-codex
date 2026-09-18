package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRunExpertBridgeSelect(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := runExpertBridge([]string{"select", "--root", root, "--query", "数据分析异常统计"}, &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 || !bytes.Contains(stdout.Bytes(), []byte(`"expert_id": "data-analysis"`)) {
		t.Fatalf("unexpected result code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
}

func TestReadExpertJSONRejectsOversizedInput(t *testing.T) {
	path := filepath.Join(t.TempDir(), "large.json")
	if err := os.WriteFile(path, bytes.Repeat([]byte(" "), int(expertBridgeMaxJSONBytes)+1), 0o600); err != nil {
		t.Fatal(err)
	}
	var target map[string]any
	if err := readExpertJSON(path, &target); err == nil {
		t.Fatal("oversized expert JSON was accepted")
	}
}
