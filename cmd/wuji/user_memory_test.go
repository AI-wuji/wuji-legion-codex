package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestUserMemoryCommandSmoke(t *testing.T) {
	store, workspace := filepath.Join(t.TempDir(), "store"), t.TempDir()
	var stdout, stderr bytes.Buffer
	code := runUserMemoryCommand([]string{"remember", "--store", store, "--workspace", workspace, "--key", "output format", "--value", "json", "--provenance", "user-confirmed:smoke"}, &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 || !strings.Contains(stdout.String(), `"version": 1`) {
		t.Fatalf("remember smoke failed: code=%d out=%q err=%q", code, stdout.String(), stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	code = runUserMemoryCommand([]string{"recall", "--store", store, "--workspace", workspace, "--query", "format"}, &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 || !strings.Contains(stdout.String(), `"value": "json"`) {
		t.Fatalf("recall smoke failed: code=%d out=%q err=%q", code, stdout.String(), stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	code = runUserMemoryCommand([]string{"remember", "--store", store, "--workspace", workspace, "--key", "output format", "--value", "yaml", "--provenance", "user-confirmed:smoke"}, &stdout, &stderr)
	if code != 2 || !strings.Contains(stderr.String(), "requires expected-version 1") {
		t.Fatalf("silent CLI replacement was accepted: code=%d out=%q err=%q", code, stdout.String(), stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	code = runUserMemoryCommand([]string{"remember", "--store", store, "--workspace", workspace, "--key", "output format", "--value", "yaml", "--provenance", "user-confirmed:smoke", "--expected-version", "1"}, &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 || !strings.Contains(stdout.String(), `"version": 2`) || !strings.Contains(stdout.String(), `"value": "yaml"`) {
		t.Fatalf("versioned CLI replacement failed: code=%d out=%q err=%q", code, stdout.String(), stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	code = runUserMemoryCommand([]string{"revoke", "--store", store, "--workspace", workspace, "--key", "output format", "--expected-version", "2"}, &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 || !strings.Contains(stdout.String(), `"revoked": true`) {
		t.Fatalf("revoke smoke failed: code=%d out=%q err=%q", code, stdout.String(), stderr.String())
	}
}
