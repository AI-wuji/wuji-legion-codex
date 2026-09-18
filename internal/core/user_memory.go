package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	UserMemoryMaxValueBytes      = 4096
	UserMemoryMaxKeyBytes        = 256
	UserMemoryMaxProvenanceBytes = 2048
	UserMemoryMaxQueryBytes      = 256
	UserMemoryMaxResults         = 50
	UserMemoryMaxEntries         = 1024
	UserMemoryMaxStoreBytes      = 16 * 1024 * 1024
	UserMemoryMaxTTL             = 365 * 24 * time.Hour
)

var (
	userMemoryLocks       sync.Map
	userMemorySharedScope = regexp.MustCompile(`^shared:[a-z0-9][a-z0-9._-]{0,63}$`)
	userMemorySecrets     = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(?:api[_-]?key|authorization|bearer|password|passwd|secret|token)\s*[:=]\s*\S+`),
		regexp.MustCompile(`\bsk-[A-Za-z0-9_-]{16,}\b`),
		regexp.MustCompile(`\bgh[pousr]_[A-Za-z0-9]{16,}\b`),
		regexp.MustCompile(`\bxox[baprs]-[A-Za-z0-9-]{12,}\b`),
	}
)

type userMemoryProcessLock struct {
	semaphore chan struct{}
}

type UserMemory struct {
	SchemaVersion    int    `json:"schema_version"`
	ID               string `json:"id"`
	Scope            string `json:"scope"`
	Key              string `json:"key"`
	NormalizedKey    string `json:"normalized_key"`
	Value            string `json:"value"`
	Version          int    `json:"version"`
	Provenance       string `json:"provenance"`
	ProvenanceSHA256 string `json:"provenance_sha256"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
	ExpiresAt        string `json:"expires_at,omitempty"`
	TTLSeconds       int64  `json:"ttl_seconds,omitempty"`
}

type RememberUserMemoryInput struct {
	Store, Workspace, SharedScope, Key, Value, Provenance string
	TTL                                                   time.Duration
	ExpectedVersion                                       int
	Now                                                   time.Time
}

type RecallUserMemoryInput struct {
	Store, Workspace, SharedScope, Query string
	Limit                                int
	Now                                  time.Time
}

type RevokeUserMemoryInput struct {
	Store, Workspace, SharedScope, Key string
	ExpectedVersion                    int
	Now                                time.Time
}

func UserMemoryScope(workspace, shared string) (string, error) {
	if strings.TrimSpace(shared) != "" {
		s := strings.ToLower(strings.TrimSpace(shared))
		if !strings.HasPrefix(s, "shared:") {
			s = "shared:" + s
		}
		if !userMemorySharedScope.MatchString(s) {
			return "", errors.New("shared scope must contain only letters, digits, dot, underscore, or hyphen")
		}
		return s, nil
	}
	abs, err := filepath.Abs(strings.TrimSpace(workspace))
	if err != nil {
		return "", err
	}
	if resolved, resolveErr := filepath.EvalSymlinks(abs); resolveErr == nil {
		abs = resolved
	}
	abs = filepath.Clean(abs)
	if runtime.GOOS == "windows" {
		abs = strings.ToLower(abs)
	}
	h := sha256.Sum256([]byte(abs))
	return "workspace:" + hex.EncodeToString(h[:]), nil
}

func RememberUserMemory(in RememberUserMemoryInput) (UserMemory, error) {
	now := memoryNow(in.Now)
	scope, key, err := validateUserMemoryInput(in.Workspace, in.SharedScope, in.Key, in.Value, in.Provenance, in.TTL)
	if err != nil {
		return UserMemory{}, err
	}
	var result UserMemory
	err = withUserMemoryLock(in.Store, func() error {
		items, err := loadUserMemories(in.Store)
		if err != nil {
			return err
		}
		items = discardExpired(items, now)
		id := userMemoryID(scope, key)
		for i := range items {
			if items[i].ID != id {
				continue
			}
			if in.ExpectedVersion > 0 && items[i].Version != in.ExpectedVersion {
				return fmt.Errorf("memory version conflict: expected %d, found %d", in.ExpectedVersion, items[i].Version)
			}
			if items[i].Value == strings.TrimSpace(in.Value) && items[i].Provenance == strings.TrimSpace(in.Provenance) && items[i].TTLSeconds == int64(in.TTL/time.Second) {
				result = items[i]
				return writeUserMemories(in.Store, items)
			}
			if in.ExpectedVersion <= 0 {
				return fmt.Errorf("memory update requires expected-version %d; recall the existing record first", items[i].Version)
			}
			items[i].Value = strings.TrimSpace(in.Value)
			items[i].Provenance = strings.TrimSpace(in.Provenance)
			items[i].ProvenanceSHA256 = memoryHash(items[i].Provenance)
			items[i].Version++
			items[i].UpdatedAt = now.Format(time.RFC3339Nano)
			items[i].ExpiresAt = expiry(in.TTL, now)
			items[i].TTLSeconds = int64(in.TTL / time.Second)
			result = items[i]
			return writeUserMemories(in.Store, items)
		}
		if in.ExpectedVersion > 0 {
			return fmt.Errorf("memory version conflict: expected %d, found no record", in.ExpectedVersion)
		}
		if len(items) >= UserMemoryMaxEntries {
			return fmt.Errorf("memory store limit of %d entries reached", UserMemoryMaxEntries)
		}
		result = UserMemory{SchemaVersion: 1, ID: id, Scope: scope, Key: strings.TrimSpace(in.Key), NormalizedKey: key, Value: strings.TrimSpace(in.Value), Version: 1, Provenance: strings.TrimSpace(in.Provenance), ProvenanceSHA256: memoryHash(strings.TrimSpace(in.Provenance)), CreatedAt: now.Format(time.RFC3339Nano), UpdatedAt: now.Format(time.RFC3339Nano), ExpiresAt: expiry(in.TTL, now), TTLSeconds: int64(in.TTL / time.Second)}
		items = append(items, result)
		return writeUserMemories(in.Store, items)
	})
	return result, err
}

func RecallUserMemories(in RecallUserMemoryInput) ([]UserMemory, error) {
	if len(in.Query) > UserMemoryMaxQueryBytes {
		return nil, fmt.Errorf("query exceeds %d bytes", UserMemoryMaxQueryBytes)
	}
	if in.Limit <= 0 {
		in.Limit = 10
	}
	if in.Limit > UserMemoryMaxResults {
		return nil, fmt.Errorf("limit exceeds %d", UserMemoryMaxResults)
	}
	scope, err := UserMemoryScope(in.Workspace, in.SharedScope)
	if err != nil {
		return nil, err
	}
	var out []UserMemory
	err = withUserMemoryLock(in.Store, func() error {
		items, err := loadUserMemories(in.Store)
		if err != nil {
			return err
		}
		active := discardExpired(items, memoryNow(in.Now))
		if len(active) != len(items) {
			if err := writeUserMemories(in.Store, active); err != nil {
				return err
			}
		}
		q := normalizeMemoryKey(in.Query)
		for _, item := range active {
			if item.Scope == scope && (q == "" || strings.Contains(item.NormalizedKey, q) || strings.Contains(strings.ToLower(item.Value), q)) {
				out = append(out, item)
			}
		}
		sort.Slice(out, func(i, j int) bool { return out[i].NormalizedKey < out[j].NormalizedKey })
		if len(out) > in.Limit {
			out = out[:in.Limit]
		}
		return nil
	})
	return out, err
}

func RevokeUserMemory(in RevokeUserMemoryInput) (bool, error) {
	scope, err := UserMemoryScope(in.Workspace, in.SharedScope)
	if err != nil {
		return false, err
	}
	key := normalizeMemoryKey(in.Key)
	if key == "" {
		return false, errors.New("key is required")
	}
	removed := false
	err = withUserMemoryLock(in.Store, func() error {
		items, err := loadUserMemories(in.Store)
		if err != nil {
			return err
		}
		id := userMemoryID(scope, key)
		kept := items[:0]
		for _, item := range items {
			if item.ID != id {
				kept = append(kept, item)
				continue
			}
			if in.ExpectedVersion > 0 && item.Version != in.ExpectedVersion {
				return fmt.Errorf("memory version conflict: expected %d, found %d", in.ExpectedVersion, item.Version)
			}
			removed = true
		}
		return writeUserMemories(in.Store, kept)
	})
	return removed, err
}

func validateUserMemoryInput(workspace, shared, key, value, provenance string, ttl time.Duration) (string, string, error) {
	scope, err := UserMemoryScope(workspace, shared)
	if err != nil {
		return "", "", err
	}
	key = normalizeMemoryKey(key)
	if key == "" || len(key) > UserMemoryMaxKeyBytes {
		return "", "", fmt.Errorf("key is required and must not exceed %d bytes", UserMemoryMaxKeyBytes)
	}
	value, provenance = strings.TrimSpace(value), strings.TrimSpace(provenance)
	if value == "" || len(value) > UserMemoryMaxValueBytes {
		return "", "", fmt.Errorf("value is required and must not exceed %d bytes", UserMemoryMaxValueBytes)
	}
	if provenance == "" || len(provenance) > UserMemoryMaxProvenanceBytes {
		return "", "", fmt.Errorf("provenance is required and must not exceed %d bytes", UserMemoryMaxProvenanceBytes)
	}
	if ttl < 0 || ttl > UserMemoryMaxTTL {
		return "", "", fmt.Errorf("ttl must be between zero and %s", UserMemoryMaxTTL)
	}
	if ttl%time.Second != 0 {
		return "", "", errors.New("ttl must use whole-second precision")
	}
	for _, p := range userMemorySecrets {
		if p.MatchString(key + " " + value + " " + provenance) {
			return "", "", errors.New("memory rejected because it appears to contain a secret")
		}
	}
	return scope, key, nil
}

func loadUserMemories(store string) ([]UserMemory, error) {
	p := userMemoryPath(store)
	f, err := os.Open(p)
	if os.IsNotExist(err) {
		return []UserMemory{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, UserMemoryMaxStoreBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > UserMemoryMaxStoreBytes {
		return nil, fmt.Errorf("memory store exceeds %d bytes", UserMemoryMaxStoreBytes)
	}
	var items []UserMemory
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, fmt.Errorf("invalid memory store: %w", err)
	}
	if len(items) > UserMemoryMaxEntries {
		return nil, fmt.Errorf("memory store exceeds %d entries", UserMemoryMaxEntries)
	}
	seen := make(map[string]bool, len(items))
	for index, item := range items {
		if err := validateStoredUserMemory(item); err != nil {
			return nil, fmt.Errorf("invalid memory record %d: %w", index, err)
		}
		if seen[item.ID] {
			return nil, fmt.Errorf("invalid memory store: duplicate id %s", item.ID)
		}
		seen[item.ID] = true
	}
	return items, nil
}

func validateStoredUserMemory(item UserMemory) error {
	if item.SchemaVersion != 1 {
		return fmt.Errorf("unsupported schema version %d", item.SchemaVersion)
	}
	if !knowledgeWorkspaceScope.MatchString(item.Scope) && !userMemorySharedScope.MatchString(item.Scope) {
		return errors.New("invalid scope")
	}
	if item.NormalizedKey == "" || item.NormalizedKey != normalizeMemoryKey(item.Key) || len(item.NormalizedKey) > UserMemoryMaxKeyBytes {
		return errors.New("invalid normalized key")
	}
	if item.ID != userMemoryID(item.Scope, item.NormalizedKey) {
		return errors.New("id does not match scope and key")
	}
	if item.Value == "" || len(item.Value) > UserMemoryMaxValueBytes {
		return errors.New("invalid value")
	}
	if item.Provenance == "" || len(item.Provenance) > UserMemoryMaxProvenanceBytes || item.ProvenanceSHA256 != memoryHash(item.Provenance) {
		return errors.New("invalid provenance or provenance hash")
	}
	if item.Version < 1 {
		return errors.New("invalid version")
	}
	created, createdErr := time.Parse(time.RFC3339Nano, item.CreatedAt)
	updated, updatedErr := time.Parse(time.RFC3339Nano, item.UpdatedAt)
	if createdErr != nil || updatedErr != nil || updated.Before(created) {
		return errors.New("invalid timestamps")
	}
	if item.TTLSeconds < 0 || item.TTLSeconds > int64(UserMemoryMaxTTL/time.Second) {
		return errors.New("invalid ttl")
	}
	if item.TTLSeconds == 0 && item.ExpiresAt != "" {
		return errors.New("expiry requires a positive ttl")
	}
	if item.TTLSeconds > 0 {
		expires, err := time.Parse(time.RFC3339Nano, item.ExpiresAt)
		if err != nil || !expires.After(updated) {
			return errors.New("invalid expiry")
		}
	}
	for _, pattern := range userMemorySecrets {
		if pattern.MatchString(item.NormalizedKey + " " + item.Value + " " + item.Provenance) {
			return errors.New("record appears to contain a secret")
		}
	}
	return nil
}

func writeUserMemories(store string, items []UserMemory) error {
	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}
	if len(data) > UserMemoryMaxStoreBytes {
		return fmt.Errorf("memory store would exceed %d bytes", UserMemoryMaxStoreBytes)
	}
	p := userMemoryPath(store)
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return err
	}
	return atomicWriteFile(p, append(data, '\n'), 0o600)
}

func withUserMemoryLock(store string, fn func() error) error {
	k := filepath.Clean(store)
	candidate := &userMemoryProcessLock{semaphore: make(chan struct{}, 1)}
	candidate.semaphore <- struct{}{}
	v, _ := userMemoryLocks.LoadOrStore(k, candidate)
	processLock := v.(*userMemoryProcessLock)
	deadline := time.Now().Add(2 * time.Second)
	remaining := time.Until(deadline)
	select {
	case <-processLock.semaphore:
		defer func() { processLock.semaphore <- struct{}{} }()
	case <-time.After(remaining):
		return errors.New("user memory process lock timed out; retry after the active operation completes")
	}
	root := filepath.Dir(userMemoryPath(store))
	if err := os.MkdirAll(root, 0o700); err != nil {
		return err
	}
	lock := filepath.Join(root, ".lock")
	for {
		file, err := os.OpenFile(lock, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if err == nil {
			_ = file.Close()
			defer os.Remove(lock)
			return fn()
		}
		if !os.IsExist(err) {
			return err
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("user memory store lock timed out at %s; verify no process is using the store, then remove the lock file manually if it is stale", lock)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func userMemoryPath(store string) string {
	return filepath.Join(filepath.Clean(store), "user-memory", "v1", "records.json")
}
func normalizeMemoryKey(v string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(v))), " ")
}
func userMemoryID(scope, key string) string { return memoryHash(scope + "\x00" + key) }
func memoryHash(v string) string            { h := sha256.Sum256([]byte(v)); return hex.EncodeToString(h[:]) }
func memoryNow(v time.Time) time.Time {
	if v.IsZero() {
		return time.Now().UTC()
	}
	return v.UTC()
}
func expiry(ttl time.Duration, now time.Time) string {
	if ttl == 0 {
		return ""
	}
	return now.Add(ttl).Format(time.RFC3339Nano)
}
func discardExpired(items []UserMemory, now time.Time) []UserMemory {
	out := make([]UserMemory, 0, len(items))
	for _, item := range items {
		if item.ExpiresAt != "" {
			if t, err := time.Parse(time.RFC3339Nano, item.ExpiresAt); err != nil || !t.After(now) {
				continue
			}
		}
		out = append(out, item)
	}
	return out
}
