package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestTopLevelHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"--help"}, &stdout, &stderr)
	if code != 0 || !strings.Contains(stdout.String(), "context-select") || stderr.Len() != 0 {
		t.Fatalf("unexpected help result: code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
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
