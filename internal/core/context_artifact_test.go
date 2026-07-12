package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestContextArtifactIsStableAndVerifiable(t *testing.T) {
	root := t.TempDir()
	writeTestSource(t, root, "task.go", "package task\n\nfunc verifyCode() {}\n")
	first, err := SelectContext(root, "fix verifyCode task.go", 2048)
	if err != nil {
		t.Fatal(err)
	}
	second, err := SelectContext(root, "task.go verifyCode fix", 2048)
	if err != nil {
		t.Fatal(err)
	}
	if first.QueryFingerprint != second.QueryFingerprint || first.ContextHandle != second.ContextHandle {
		t.Fatalf("equivalent queries produced unstable handles: %#v %#v", first, second)
	}
	path, err := WriteContextArtifact(first, filepath.Join(root, ".wuji", "context"))
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadContextArtifact(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Handle != first.ContextHandle || loaded.QueryFingerprint != first.QueryFingerprint || loaded.SelectedBytes != first.SelectedBytes {
		t.Fatalf("loaded context changed: %#v", loaded)
	}
	if loaded.Payload != renderContextPayload(first.Excerpts) || loaded.PayloadSHA256 != first.PayloadSHA256 || loaded.SelectedBytes != len([]byte(loaded.Payload)) {
		t.Fatalf("loaded prompt payload is not deterministic: %#v", loaded)
	}
}

func TestContextArtifactRejectsTamperingAndStaleSources(t *testing.T) {
	t.Run("selected byte count", func(t *testing.T) {
		root, path := writeTestContextArtifact(t)
		_ = root
		artifact := readTestContextArtifact(t, path)
		artifact.SelectedBytes = 1
		writeTestContextArtifactJSON(t, path, artifact)
		if _, err := LoadContextArtifact(path); err == nil || !strings.Contains(err.Error(), "byte count mismatch") {
			t.Fatalf("tampered byte count was accepted: %v", err)
		}
	})

	t.Run("query fingerprint", func(t *testing.T) {
		_, path := writeTestContextArtifact(t)
		artifact := readTestContextArtifact(t, path)
		artifact.QueryFingerprint = strings.Repeat("0", 64)
		writeTestContextArtifactJSON(t, path, artifact)
		if _, err := LoadContextArtifact(path); err == nil || !strings.Contains(err.Error(), "content hash mismatch") {
			t.Fatalf("tampered query fingerprint was accepted: %v", err)
		}
	})

	t.Run("payload hash", func(t *testing.T) {
		_, path := writeTestContextArtifact(t)
		artifact := readTestContextArtifact(t, path)
		artifact.PayloadSHA256 = strings.Repeat("0", 64)
		writeTestContextArtifactJSON(t, path, artifact)
		if _, err := LoadContextArtifact(path); err == nil || !strings.Contains(err.Error(), "payload hash mismatch") {
			t.Fatalf("tampered payload hash was accepted: %v", err)
		}
	})

	t.Run("quality metadata", func(t *testing.T) {
		_, path := writeTestContextArtifact(t)
		artifact := readTestContextArtifact(t, path)
		artifact.CoverageBPS--
		writeTestContextArtifactJSON(t, path, artifact)
		if _, err := LoadContextArtifact(path); err == nil || !strings.Contains(err.Error(), "quality metadata mismatch") {
			t.Fatalf("tampered quality metadata was accepted: %v", err)
		}
	})

	t.Run("source snapshot", func(t *testing.T) {
		root, path := writeTestContextArtifact(t)
		writeTestSource(t, root, "task.go", "package task\n\nfunc changed() {}\n")
		if _, err := LoadContextArtifact(path); err == nil || !strings.Contains(err.Error(), "source changed") {
			t.Fatalf("stale source artifact was accepted: %v", err)
		}
	})

	t.Run("workspace escape", func(t *testing.T) {
		_, path := writeTestContextArtifact(t)
		artifact := readTestContextArtifact(t, path)
		artifact.Excerpts[0].Path = "../outside.go"
		writeTestContextArtifactJSON(t, path, artifact)
		if _, err := LoadContextArtifact(path); err == nil || !strings.Contains(err.Error(), "escapes workspace") {
			t.Fatalf("escaping source path was accepted: %v", err)
		}
	})
}

func TestContextPathQueryReturnsContent(t *testing.T) {
	root := t.TempDir()
	writeTestSource(t, root, "router-special.go", "package router\n\nconst mode = \"bounded\"\n")
	got, err := SelectContext(root, "router-special.go", 1024)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Excerpts) != 1 || got.Excerpts[0].Text == "" || got.Excerpts[0].LineRanges[0] != "1-3" {
		t.Fatalf("path-only query returned no useful excerpt: %#v", got.Excerpts)
	}
}

func writeTestContextArtifact(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	writeTestSource(t, root, "task.go", "package task\n\nfunc verifyCode() {}\n")
	result, err := SelectContext(root, "task.go verifyCode", 2048)
	if err != nil {
		t.Fatal(err)
	}
	path, err := WriteContextArtifact(result, filepath.Join(root, ".wuji", "context"))
	if err != nil {
		t.Fatal(err)
	}
	return root, path
}

func writeTestSource(t *testing.T, root, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readTestContextArtifact(t *testing.T, path string) ContextArtifact {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var artifact ContextArtifact
	if err := json.Unmarshal(content, &artifact); err != nil {
		t.Fatal(err)
	}
	return artifact
}

func writeTestContextArtifactJSON(t *testing.T, path string, artifact ContextArtifact) {
	t.Helper()
	content, err := json.Marshal(artifact)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}
}
