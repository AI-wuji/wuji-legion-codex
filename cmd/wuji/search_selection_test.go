package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestSearchSelectCommandReadsOneBoundedJSONValue(t *testing.T) {
	input := `{"query":"evidence","incomplete":true,"source_errors":[{"source":"official","error":"unavailable"}],"candidates":[{"url":"https://example.com/a?utm_source=x","title":"first"},{"url":"https://example.com/a","title":"duplicate"}]}`
	var stdout, stderr bytes.Buffer
	code := runSearchSelectionCommand(nil, strings.NewReader(input), &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 || !strings.Contains(stdout.String(), `"duplicate_candidates": 1`) || !strings.Contains(stdout.String(), `"incomplete": true`) {
		t.Fatalf("command output was not bounded selection JSON: code=%d out=%q err=%q", code, stdout.String(), stderr.String())
	}
}

func TestSearchSelectCommandRejectsUnknownAndMultipleJSONValues(t *testing.T) {
	for _, input := range []string{`{"query":"evidence","candidates":[],"unexpected":true}`, `{"query":"evidence","candidates":[]} {}`} {
		var stdout, stderr bytes.Buffer
		code := runSearchSelectionCommand(nil, strings.NewReader(input), &stdout, &stderr)
		if code != 2 || stdout.Len() != 0 || !strings.Contains(stderr.String(), "error:") {
			t.Fatalf("invalid input was accepted: code=%d out=%q err=%q", code, stdout.String(), stderr.String())
		}
	}
}

func TestSearchSelectCommandRejectsOversizeTrailingInput(t *testing.T) {
	input := `{"query":"evidence","candidates":[]}` + strings.Repeat(" ", searchSelectionMaxInputBytes)
	var stdout, stderr bytes.Buffer
	code := runSearchSelectionCommand(nil, strings.NewReader(input), &stdout, &stderr)
	if code != 2 || stdout.Len() != 0 || !strings.Contains(stderr.String(), "exceeds") {
		t.Fatalf("oversize trailing input was accepted: code=%d out=%q err=%q", code, stdout.String(), stderr.String())
	}
}
