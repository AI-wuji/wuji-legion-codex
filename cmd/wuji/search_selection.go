package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/AI-wuji/wuji-legion-codex-2.0/internal/core"
)

const searchSelectionMaxInputBytes = 256 * 1024

func runSearchSelectionCommand(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := newFlagSet("search-select", stderr)
	if code := parseFlags(fs, args, stderr); code >= 0 {
		return code
	}
	payload, err := io.ReadAll(io.LimitReader(stdin, searchSelectionMaxInputBytes+1))
	if err != nil {
		return reportError(stderr, 2, fmt.Errorf("read search selection input: %w", err))
	}
	if len(payload) > searchSelectionMaxInputBytes {
		return reportError(stderr, 2, fmt.Errorf("search selection input exceeds %d bytes", searchSelectionMaxInputBytes))
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	var input core.SearchSelectionInput
	if err := decoder.Decode(&input); err != nil {
		return reportError(stderr, 2, fmt.Errorf("decode search selection input: %w", err))
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return reportError(stderr, 2, fmt.Errorf("search selection input must contain one JSON value"))
		}
		return reportError(stderr, 2, fmt.Errorf("decode search selection input: %w", err))
	}
	result, err := core.SelectSearchCandidates(input)
	if err != nil {
		return reportError(stderr, 2, err)
	}
	if err := writeJSON(stdout, result); err != nil {
		return reportError(stderr, 1, err)
	}
	return 0
}
