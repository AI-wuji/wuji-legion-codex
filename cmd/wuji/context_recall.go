package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/AI-wuji/wuji-legion-codex-3.0/internal/core"
)

func runContextRecallCommand(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := newFlagSet("context-recall", stderr)
	query := fs.String("query", "", "lexical query for archived records")
	limit := fs.Int("limit", 10, "maximum matched records or atomic tool pairs (1-20)")
	if code := parseFlags(fs, args, stderr); code >= 0 {
		return code
	}
	payload, err := io.ReadAll(io.LimitReader(stdin, contextReductionMaxInputBytes+1))
	if err != nil {
		return reportError(stderr, 2, fmt.Errorf("read context recall input: %w", err))
	}
	if len(payload) > contextReductionMaxInputBytes {
		return reportError(stderr, 2, fmt.Errorf("context recall input exceeds %d bytes", contextReductionMaxInputBytes))
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	var projection core.ContextReductionResult
	if err := decoder.Decode(&projection); err != nil {
		return reportError(stderr, 2, fmt.Errorf("decode context recall projection: %w", err))
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return reportError(stderr, 2, fmt.Errorf("context recall projection must contain one JSON value"))
		}
		return reportError(stderr, 2, fmt.Errorf("decode context recall projection: %w", err))
	}
	result, err := core.RecallArchivedContext(projection, *query, *limit)
	if err != nil {
		return reportError(stderr, 2, err)
	}
	if err := writeJSON(stdout, result); err != nil {
		return reportError(stderr, 1, err)
	}
	return 0
}
