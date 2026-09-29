package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/AI-wuji/wuji-legion-codex-3.0/internal/core"
)

const contextReductionMaxInputBytes = 4 * 1024 * 1024

func runContextReductionCommand(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := newFlagSet("context-reduce", stderr)
	restore := fs.Bool("restore", false, "validate and restore an existing projection")
	if code := parseFlags(fs, args, stderr); code >= 0 {
		return code
	}
	payload, err := io.ReadAll(io.LimitReader(stdin, contextReductionMaxInputBytes+1))
	if err != nil {
		return reportError(stderr, 2, fmt.Errorf("read context reduction input: %w", err))
	}
	if len(payload) > contextReductionMaxInputBytes {
		return reportError(stderr, 2, fmt.Errorf("context reduction input exceeds %d bytes", contextReductionMaxInputBytes))
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if *restore {
		var projection core.ContextReductionResult
		if err := decoder.Decode(&projection); err != nil {
			return reportError(stderr, 2, fmt.Errorf("decode context reduction projection: %w", err))
		}
		var extra any
		if err := decoder.Decode(&extra); err != io.EOF {
			return reportError(stderr, 2, fmt.Errorf("context reduction projection must contain one JSON value"))
		}
		records, err := core.RestoreContext(projection)
		if err != nil {
			return reportError(stderr, 2, err)
		}
		if err := writeJSON(stdout, core.ContextReductionInput{Records: records}); err != nil {
			return reportError(stderr, 1, err)
		}
		return 0
	}
	var input core.ContextReductionInput
	if err := decoder.Decode(&input); err != nil {
		return reportError(stderr, 2, fmt.Errorf("decode context reduction input: %w", err))
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return reportError(stderr, 2, fmt.Errorf("context reduction input must contain one JSON value"))
		}
		return reportError(stderr, 2, fmt.Errorf("decode context reduction input: %w", err))
	}
	result, err := core.ReduceContext(input)
	if err != nil {
		return reportError(stderr, 2, err)
	}
	if err := writeJSON(stdout, result); err != nil {
		return reportError(stderr, 1, err)
	}
	return 0
}
