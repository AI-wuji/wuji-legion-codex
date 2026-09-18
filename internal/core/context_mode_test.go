package core

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestPrepareContextModeExecuteChineseJSONL(t *testing.T) {
	root := t.TempDir()
	content := "{\"状态\":\"失败\",\"message\":\"中文错误\"}\n{\"状态\":\"成功\"}\n"
	path := filepath.Join(root, "日志.jsonl")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	call, err := PrepareContextModeExecute(root, path, ContextModeRequest{Operation: "jsonl-filter-eq", Field: "状态", Value: "失败"})
	if err != nil {
		t.Fatal(err)
	}
	raw := []byte(`{"schema_version":1,"input_bytes":` + itoa(len([]byte(content))) + `,"input_sha256":"` + call.InputSHA256 + `","operation":"jsonl-filter-eq","line_count":2,"invalid_lines":0,"match_count":1,"truncated":false,"matches":[1]}`)
	got, err := ValidateContextModeResult(call, raw)
	if err != nil || got.MatchCount != 1 {
		t.Fatalf("Chinese filter validation failed: %#v %v", got, err)
	}
	call, err = PrepareContextModeExecute(root, path, ContextModeRequest{Operation: "jsonl-extract", Field: "message", MaxMatches: 5})
	if err != nil {
		t.Fatal(err)
	}
	if call.InputBytes != len([]byte(content)) || call.Tool != "ctx_execute" || !call.PreparedOnly || call.Arguments["timeout"] != ContextModeTimeoutMS || call.Arguments["language"] != "javascript" {
		t.Fatalf("bad call: %#v", call)
	}
	raw = []byte(`{"schema_version":1,"input_bytes":` + itoa(len([]byte(content))) + `,"input_sha256":"` + call.InputSHA256 + `","operation":"jsonl-extract","line_count":2,"invalid_lines":0,"match_count":1,"truncated":false,"matches":["中文错误"]}`)
	got, err = ValidateContextModeResult(call, raw)
	if err != nil || got.MatchCount != 1 {
		t.Fatalf("validation failed: %#v %v", got, err)
	}
}

func TestGeneratedContextModeJavaScriptRunsInNode(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("Node is not installed")
	}
	root := t.TempDir()
	content := strings.Join([]string{
		`{"message":"中文错误"}`,
		`null`,
		`[]`,
		`42`,
		`{"message":null}`,
		`{"message":{"nested":true}}`,
		`{"message":"尾随"} garbage`,
	}, "\n") + "\n"
	path := filepath.Join(root, "edge.jsonl")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	call, err := PrepareContextModeExecute(root, path, ContextModeRequest{Operation: "jsonl-extract", Field: "message", MaxMatches: 10})
	if err != nil {
		t.Fatal(err)
	}
	code, ok := call.Arguments["code"].(string)
	if !ok {
		t.Fatal("generated code is missing")
	}
	output, err := exec.Command(node, "-e", code).Output()
	if err != nil {
		t.Fatalf("generated JavaScript failed: %v", err)
	}
	output = bytes.TrimSpace(output)
	result, err := ValidateContextModeResult(call, output)
	if err != nil {
		t.Fatalf("Node result failed independent validation: %s: %v", output, err)
	}
	if result.LineCount != 7 || result.InvalidLines != 4 || result.MatchCount != 3 || !result.Truncated {
		t.Fatalf("unexpected Node result: %#v", result)
	}
	expected := []interface{}{"中文错误", "null"}
	if !sameContextModeMatches(result.Matches, expected) {
		t.Fatalf("unexpected exact matches: %#v", result.Matches)
	}
}

func TestPrepareContextModeExecuteCaps(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "large.jsonl")
	if err := os.WriteFile(path, []byte(strings.Repeat("x", ContextModeMaxInputBytes+1)), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := PrepareContextModeExecute(root, path, ContextModeRequest{Operation: "jsonl-count"}); err == nil {
		t.Fatal("oversized input accepted")
	}
	small := filepath.Join(root, "small.jsonl")
	if err := os.WriteFile(small, []byte("{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	call, err := PrepareContextModeExecute(root, small, ContextModeRequest{Operation: "jsonl-count"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ValidateContextModeResult(call, []byte(strings.Repeat("x", ContextModeMaxResultBytes+1))); err == nil {
		t.Fatal("oversized result accepted")
	}
	valid, _ := json.Marshal(ContextModeResult{SchemaVersion: 1, InputBytes: call.InputBytes, InputSHA256: call.InputSHA256, Operation: call.Operation, LineCount: 1, MatchCount: 1})
	if _, err := ValidateContextModeResult(call, valid); err != nil {
		t.Fatal(err)
	}
}

func TestContextModeRuntimeValidationEdgeCases(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("Node unavailable")
	}
	root := t.TempDir()
	path := filepath.Join(root, "edge.jsonl")
	content := "{\"x\":null}\n{}\n{\"x\":1.0}\n{\"x\":1e3}\n{\"x\":-0}\n"
	for i := 0; i < 40; i++ {
		content += `{"x":"` + strings.Repeat("<&>", 40) + `"}` + "\n"
	}
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	for _, request := range []ContextModeRequest{
		{Operation: "jsonl-filter-eq", Field: "x", Value: "absent", MaxMatches: 100},
		{Operation: "jsonl-filter-eq", Field: "x", Value: "null", MaxMatches: 100},
		{Operation: "jsonl-extract", Field: "missing", MaxMatches: 100},
		{Operation: "jsonl-extract", Field: "x", MaxMatches: 100},
	} {
		call, err := PrepareContextModeExecute(root, path, request)
		if err != nil {
			t.Fatal(err)
		}
		output, err := exec.Command(node, "-e", call.Arguments["code"].(string)).Output()
		if err != nil {
			t.Fatal(err)
		}
		if _, err := ValidateContextModeResult(call, bytes.TrimSpace(output)); err != nil {
			t.Fatalf("%+v: %v", request, err)
		}
		if _, err := ValidateContextModeResult(call, append(bytes.TrimSpace(output), []byte(" {}")...)); err == nil {
			t.Fatal("trailing JSON accepted")
		}
	}
}

func TestPrepareContextModeExecuteRejectsSymlinkEscape(t *testing.T) {
	root, outside := t.TempDir(), t.TempDir()
	target := filepath.Join(outside, "outside.jsonl")
	if err := os.WriteFile(target, []byte("{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "escape.jsonl")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if _, err := PrepareContextModeExecute(root, link, ContextModeRequest{Operation: "jsonl-count"}); err == nil {
		t.Fatal("symlink escape accepted")
	}
}
