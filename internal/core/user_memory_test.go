package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestUserMemoryIsolationDedupConflictPersistenceAndRevoke(t *testing.T) {
	store := t.TempDir()
	w1, w2 := filepath.Join(t.TempDir(), "one"), filepath.Join(t.TempDir(), "two")
	_ = os.MkdirAll(w1, 0o700)
	_ = os.MkdirAll(w2, 0o700)
	now := time.Date(2026, 9, 19, 10, 0, 0, 0, time.UTC)
	in := RememberUserMemoryInput{Store: store, Workspace: w1, Key: " Editor Theme ", Value: "dark", Provenance: "user-confirmed:settings", Now: now}
	first, err := RememberUserMemory(in)
	if err != nil || first.Version != 1 || len(first.ProvenanceSHA256) != 64 {
		t.Fatalf("remember failed: %#v %v", first, err)
	}
	duplicate, err := RememberUserMemory(in)
	if err != nil || duplicate.ID != first.ID || duplicate.Version != 1 {
		t.Fatalf("dedup failed: %#v %v", duplicate, err)
	}
	other, err := RecallUserMemories(RecallUserMemoryInput{Store: store, Workspace: w2, Now: now})
	if err != nil || len(other) != 0 {
		t.Fatalf("workspace isolation failed: %#v %v", other, err)
	}
	in.Value, in.ExpectedVersion = "light", 0
	if _, err := RememberUserMemory(in); err == nil || !strings.Contains(err.Error(), "requires expected-version 1") {
		t.Fatalf("silent overwrite was accepted: %v", err)
	}
	unchanged, err := RecallUserMemories(RecallUserMemoryInput{Store: store, Workspace: w1, Now: now})
	if err != nil || len(unchanged) != 1 || unchanged[0].Value != "dark" || unchanged[0].Version != 1 {
		t.Fatalf("rejected overwrite changed persistence: %#v %v", unchanged, err)
	}
	in.Value, in.ExpectedVersion = "light", 9
	if _, err := RememberUserMemory(in); err == nil || !strings.Contains(err.Error(), "version conflict") {
		t.Fatalf("expected conflict, got %v", err)
	}
	in.ExpectedVersion = 1
	updated, err := RememberUserMemory(in)
	if err != nil || updated.Version != 2 {
		t.Fatalf("update failed: %#v %v", updated, err)
	}
	got, err := RecallUserMemories(RecallUserMemoryInput{Store: store, Workspace: w1, Query: "theme", Now: now})
	if err != nil || len(got) != 1 || got[0].Value != "light" {
		t.Fatalf("persistent recall failed: %#v %v", got, err)
	}
	removed, err := RevokeUserMemory(RevokeUserMemoryInput{Store: store, Workspace: w1, Key: "editor theme", ExpectedVersion: 2})
	if err != nil || !removed {
		t.Fatalf("revoke failed: %v %v", removed, err)
	}
	got, _ = RecallUserMemories(RecallUserMemoryInput{Store: store, Workspace: w1, Now: now})
	if len(got) != 0 {
		t.Fatalf("revoked record remained: %#v", got)
	}
}

func TestUserMemoryRejectsTamperedAndDuplicatePersistedRecords(t *testing.T) {
	store, workspace := t.TempDir(), t.TempDir()
	now := time.Date(2026, 9, 19, 10, 0, 0, 0, time.UTC)
	record, err := RememberUserMemory(RememberUserMemoryInput{Store: store, Workspace: workspace, Key: "format", Value: "json", Provenance: "user-confirmed:test", Now: now})
	if err != nil {
		t.Fatal(err)
	}
	record.ProvenanceSHA256 = strings.Repeat("0", 64)
	data, _ := json.Marshal([]UserMemory{record})
	if err := os.WriteFile(userMemoryPath(store), data, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := RecallUserMemories(RecallUserMemoryInput{Store: store, Workspace: workspace, Now: now}); err == nil || !strings.Contains(err.Error(), "provenance") {
		t.Fatalf("tampered provenance was trusted: %v", err)
	}
	record.ProvenanceSHA256 = memoryHash(record.Provenance)
	data, _ = json.Marshal([]UserMemory{record, record})
	if err := os.WriteFile(userMemoryPath(store), data, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := RecallUserMemories(RecallUserMemoryInput{Store: store, Workspace: workspace, Now: now}); err == nil || !strings.Contains(err.Error(), "duplicate id") {
		t.Fatalf("duplicate ids were trusted: %v", err)
	}
}

func TestUserMemorySharedScopeExpiryAndLimits(t *testing.T) {
	store := t.TempDir()
	now := time.Date(2026, 9, 19, 10, 0, 0, 0, time.UTC)
	record, err := RememberUserMemory(RememberUserMemoryInput{Store: store, Workspace: t.TempDir(), SharedScope: "team-a", Key: "format", Value: "json", Provenance: "user-confirmed:task-7", TTL: time.Hour, Now: now})
	if err != nil || record.ExpiresAt == "" {
		t.Fatalf("shared remember failed: %#v %v", record, err)
	}
	got, err := RecallUserMemories(RecallUserMemoryInput{Store: store, Workspace: t.TempDir(), SharedScope: "team-a", Now: now.Add(30 * time.Minute)})
	if err != nil || len(got) != 1 {
		t.Fatalf("shared recall failed: %#v %v", got, err)
	}
	got, err = RecallUserMemories(RecallUserMemoryInput{Store: store, SharedScope: "team-a", Now: now.Add(2 * time.Hour)})
	if err != nil || len(got) != 0 {
		t.Fatalf("expiry failed: %#v %v", got, err)
	}
	_, err = RememberUserMemory(RememberUserMemoryInput{Store: store, Workspace: ".", Key: "credential", Value: "token=abcdefghijklmnop", Provenance: "user-confirmed", Now: now})
	if err == nil || !strings.Contains(err.Error(), "secret") {
		t.Fatalf("secret was accepted: %v", err)
	}
	_, err = RememberUserMemory(RememberUserMemoryInput{Store: store, Workspace: ".", Key: "large", Value: strings.Repeat("x", UserMemoryMaxValueBytes+1), Provenance: "user-confirmed", Now: now})
	if err == nil || !strings.Contains(err.Error(), "exceed") {
		t.Fatalf("oversized value was accepted: %v", err)
	}
	_, err = RecallUserMemories(RecallUserMemoryInput{Store: store, Workspace: ".", Limit: UserMemoryMaxResults + 1, Now: now})
	if err == nil {
		t.Fatal("oversized recall limit was accepted")
	}
}

func TestUserMemoryRejectsSubSecondTTLWithoutCorruptingStore(t *testing.T) {
	store, workspace := t.TempDir(), t.TempDir()
	now := time.Date(2026, 9, 19, 10, 0, 0, 0, time.UTC)
	_, err := RememberUserMemory(RememberUserMemoryInput{Store: store, Workspace: workspace, Key: "valid", Value: "kept", Provenance: "user-confirmed:test", Now: now})
	if err != nil {
		t.Fatal(err)
	}
	_, err = RememberUserMemory(RememberUserMemoryInput{Store: store, Workspace: workspace, Key: "invalid", Value: "rejected", Provenance: "user-confirmed:test", TTL: 500 * time.Millisecond, Now: now})
	if err == nil || !strings.Contains(err.Error(), "whole-second precision") {
		t.Fatalf("sub-second ttl was not rejected: %v", err)
	}
	got, err := RecallUserMemories(RecallUserMemoryInput{Store: store, Workspace: workspace, Now: now})
	if err != nil || len(got) != 1 || got[0].Key != "valid" || got[0].Value != "kept" {
		t.Fatalf("rejected ttl corrupted the existing store: %#v %v", got, err)
	}
}

func TestUserMemoryScopeUsesOSPathCaseSemantics(t *testing.T) {
	parent := t.TempDir()
	upper := filepath.Join(parent, "CaseSensitiveWorkspace")
	lower := filepath.Join(parent, "casesensitiveworkspace")
	if err := os.MkdirAll(upper, 0o700); err != nil {
		t.Fatal(err)
	}
	upperScope, err := UserMemoryScope(upper, "")
	if err != nil {
		t.Fatal(err)
	}
	lowerScope, err := UserMemoryScope(lower, "")
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS == "windows" && upperScope != lowerScope {
		t.Fatalf("Windows path case produced different scopes: %s %s", upperScope, lowerScope)
	}
	if runtime.GOOS != "windows" && upperScope == lowerScope {
		t.Fatalf("case-sensitive platform collapsed distinct path identities: %s", upperScope)
	}
}
