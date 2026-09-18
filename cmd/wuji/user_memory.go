package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/AI-wuji/wuji-legion-codex-2.0/internal/core"
)

// runUserMemoryCommand is intentionally separate from main.go. Register it as
// the "user-memory" switch case when the top-level command surface is approved.
func runUserMemoryCommand(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "user-memory requires remember, recall, or revoke")
		return 2
	}
	fs := flag.NewFlagSet("user-memory "+args[0], flag.ContinueOnError)
	fs.SetOutput(stderr)
	workspace := fs.String("workspace", ".", "workspace identity (default isolation boundary)")
	shared := fs.String("shared-scope", "", "explicit shared scope name")
	store := fs.String("store", filepath.Join(".wuji", "memory"), "memory store directory")
	key := fs.String("key", "", "stable memory key")
	value := fs.String("value", "", "memory value")
	provenance := fs.String("provenance", "", "user-confirmation provenance")
	ttl := fs.Duration("ttl", 0, "optional TTL; zero means no automatic expiry")
	expected := fs.Int("expected-version", 0, "required to replace an existing value; zero permits only create or exact dedup")
	query := fs.String("query", "", "bounded recall query")
	limit := fs.Int("limit", 10, "maximum recall results")
	if err := fs.Parse(args[1:]); err != nil {
		return 2
	}
	var output any
	var err error
	switch args[0] {
	case "remember":
		output, err = core.RememberUserMemory(core.RememberUserMemoryInput{Store: *store, Workspace: *workspace, SharedScope: *shared, Key: *key, Value: *value, Provenance: *provenance, TTL: *ttl, ExpectedVersion: *expected})
	case "recall":
		output, err = core.RecallUserMemories(core.RecallUserMemoryInput{Store: *store, Workspace: *workspace, SharedScope: *shared, Query: *query, Limit: *limit})
	case "revoke":
		if strings.TrimSpace(*key) == "" {
			err = errors.New("--key is required")
			break
		}
		var removed bool
		removed, err = core.RevokeUserMemory(core.RevokeUserMemoryInput{Store: *store, Workspace: *workspace, SharedScope: *shared, Key: *key, ExpectedVersion: *expected, Now: time.Now()})
		output = struct {
			Revoked bool `json:"revoked"`
		}{removed}
	default:
		fmt.Fprintln(stderr, "user-memory requires remember, recall, or revoke")
		return 2
	}
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}
