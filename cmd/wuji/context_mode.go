package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/AI-wuji/wuji-legion-codex-2.0/internal/core"
)

// runContextModePrepare is kept separate so main.go only needs to dispatch this helper.
func runContextModePrepare(args []string, stdout, stderr io.Writer) int {
	fs := newFlagSet("context-mode-prepare", stderr)
	workspace := fs.String("workspace", ".", "canonical workspace boundary")
	input := fs.String("input", "", "explicit local JSONL file")
	operation := fs.String("operation", "", "jsonl-count, jsonl-filter-eq, or jsonl-extract")
	field := fs.String("field", "", "JSON field for filter/extract")
	value := fs.String("value", "", "exact string value for filter")
	maxMatches := fs.Int("max-matches", 20, "maximum returned line numbers or values (1-100)")
	if code := parseFlags(fs, args, stderr); code >= 0 {
		return code
	}
	if *input == "" {
		return reportError(stderr, 2, errors.New("--input is required"))
	}
	call, err := core.PrepareContextModeExecute(*workspace, *input, core.ContextModeRequest{Operation: *operation, Field: *field, Value: *value, MaxMatches: *maxMatches})
	if err != nil {
		return reportError(stderr, 2, err)
	}
	if err := writeJSON(stdout, call); err != nil {
		return reportError(stderr, 1, err)
	}
	return 0
}

func runContextModeValidate(args []string, stdout, stderr io.Writer) int {
	fs := newFlagSet("context-mode-validate", stderr)
	contractPath := fs.String("contract", "", "prepared call JSON")
	resultPath := fs.String("result", "", "raw ctx_execute result JSON")
	if code := parseFlags(fs, args, stderr); code >= 0 {
		return code
	}
	if *contractPath == "" || *resultPath == "" {
		return reportError(stderr, 2, errors.New("--contract and --result are required"))
	}
	contractBytes, err := readBoundedContextModeFile(*contractPath, 64*1024)
	if err != nil {
		return reportError(stderr, 2, err)
	}
	var call core.ContextModeToolCall
	if err := jsonUnmarshalStrict(contractBytes, &call); err != nil {
		return reportError(stderr, 2, err)
	}
	raw, err := readBoundedContextModeFile(*resultPath, core.ContextModeMaxResultBytes)
	if err != nil {
		return reportError(stderr, 2, err)
	}
	result, err := core.ValidateContextModeResult(call, raw)
	if err != nil {
		return reportError(stderr, 2, err)
	}
	if err := writeJSON(stdout, result); err != nil {
		return reportError(stderr, 1, err)
	}
	return 0
}

func readBoundedContextModeFile(path string, maxBytes int) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	content, err := io.ReadAll(io.LimitReader(file, int64(maxBytes)+1))
	if err != nil {
		return nil, err
	}
	if len(content) > maxBytes {
		return nil, fmt.Errorf("%s exceeds %d bytes", path, maxBytes)
	}
	return content, nil
}
