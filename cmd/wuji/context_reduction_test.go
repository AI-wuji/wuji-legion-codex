package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestContextReduceCommandReturnsReversibleProjection(t *testing.T) {
	input := `{"records":[{"id":"u1","role":"user","kind":"message","text":"keep"},{"id":"r1","role":"tool","kind":"tool_result","tool":"Read","operation":"view","footprint":"a","pair_id":"p1","text":"same","succeeded":true},{"id":"r2","role":"tool","kind":"tool_result","tool":"Read","operation":"view","footprint":"a","pair_id":"p2","text":"same","succeeded":true}],"archive_duplicates":true}`
	var stdout, stderr bytes.Buffer
	code := runContextReductionCommand(nil, strings.NewReader(input), &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 || !strings.Contains(stdout.String(), `"schema_version": 2`) || !strings.Contains(stdout.String(), `"archived_count": 1`) || !strings.Contains(stdout.String(), `"original_sha256"`) {
		t.Fatalf("unexpected context reduction output: code=%d out=%q err=%q", code, stdout.String(), stderr.String())
	}
	projected := stdout.String()
	stdout.Reset()
	code = runContextReductionCommand([]string{"--restore"}, strings.NewReader(projected), &stdout, &stderr)
	if code != 0 || !strings.Contains(stdout.String(), `"id": "r1"`) || !strings.Contains(stdout.String(), `"id": "r2"`) {
		t.Fatalf("projection did not restore: code=%d out=%q err=%q", code, stdout.String(), stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	code = runContextRecallCommand([]string{"--query", "a"}, strings.NewReader(projected), &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 || !strings.Contains(stdout.String(), `"id": "r1"`) {
		t.Fatalf("projection did not recall archived evidence: code=%d out=%q err=%q", code, stdout.String(), stderr.String())
	}
}

func TestContextReduceCommandRejectsTrailingJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runContextReductionCommand(nil, strings.NewReader(`{"records":[{"id":"u1","role":"user","kind":"message"}]} {}`), &stdout, &stderr)
	if code != 2 || stdout.Len() != 0 || !strings.Contains(stderr.String(), "one JSON value") {
		t.Fatalf("trailing JSON was accepted: code=%d out=%q err=%q", code, stdout.String(), stderr.String())
	}
}

func TestContextRecallCommandRejectsForgedProjection(t *testing.T) {
	projection := `{"schema_version":2,"original_sha256":"` + strings.Repeat("0", 64) + `","preserve_recent":0,"archive_duplicates":false,"archive_stale":false,"visible":[{"id":"new","role":"tool","kind":"tool_result","tool":"Read","operation":"view","footprint":"src/a.go","text":"same","succeeded":true}],"archived":[{"id":"old","role":"tool","kind":"tool_result","tool":"Read","operation":"view","footprint":"src/a.go","text":"same","succeeded":true}],"decisions":[{"record_id":"old","action":"archive","reason":"identical-observation"},{"record_id":"new","action":"keep","reason":"default"}],"archived_count":1,"protected_count":0,"original_record_bytes":0,"visible_record_bytes":0,"archived_record_bytes":0,"saved_record_bytes":0,"interpretation":"Projection only: use visible for a bounded handoff, retain archived separately, and restore by hash when needed."}`
	var stdout, stderr bytes.Buffer
	code := runContextRecallCommand([]string{"--query", "src/a.go"}, strings.NewReader(projection), &stdout, &stderr)
	if code != 2 || stdout.Len() != 0 || !strings.Contains(stderr.String(), "integrity") {
		t.Fatalf("forged projection was accepted: code=%d out=%q err=%q", code, stdout.String(), stderr.String())
	}
}
