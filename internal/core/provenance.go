package core

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const provenanceSchemaVersion = 1

type ProvenanceEntry struct {
	ID           string   `json:"id"`
	Scope        string   `json:"scope"`
	Subject      string   `json:"subject"`
	Predicate    string   `json:"predicate"`
	Target       string   `json:"target"`
	Readers      []string `json:"readers"`
	RecordedAt   string   `json:"recorded_at"`
	RecordSHA256 string   `json:"record_sha256"`
}

type ProvenanceInput struct {
	ID        string
	Scope     string
	Subject   string
	Predicate string
	Target    string
	Readers   []string
}

type ProvenanceQuery struct {
	Scope     string
	Subject   string
	Principal string
}

type ProvenanceIndex struct {
	SchemaVersion int               `json:"schema_version"`
	Entries       []ProvenanceEntry `json:"entries"`
}

type ProvenanceResolveResult struct {
	Scope   string            `json:"scope"`
	Subject string            `json:"subject"`
	Entries []ProvenanceEntry `json:"entries"`
	Denied  int               `json:"denied"`
}

func DefaultProvenanceStore() string {
	if value := strings.TrimSpace(os.Getenv("WUJI_PROVENANCE_DIR")); value != "" {
		return filepath.Clean(value)
	}
	return filepath.Join(".wuji", "provenance")
}

func RecordProvenance(store string, input ProvenanceInput) (ProvenanceEntry, error) {
	if err := validateProvenanceInput(&input); err != nil {
		return ProvenanceEntry{}, err
	}
	var result ProvenanceEntry
	err := withKnowledgeStoreLock(store, func() error {
		index, err := loadProvenanceIndex(store)
		if err != nil {
			return err
		}
		for _, entry := range index.Entries {
			if entry.ID == input.ID {
				if sameProvenance(entry, input) {
					result = entry
					return nil
				}
				return fmt.Errorf("provenance id %q is immutable", input.ID)
			}
		}
		result = ProvenanceEntry{ID: input.ID, Scope: input.Scope, Subject: input.Subject, Predicate: input.Predicate, Target: input.Target, Readers: input.Readers, RecordedAt: time.Now().UTC().Format(time.RFC3339)}
		result.RecordSHA256 = provenanceDigest(result)
		index.Entries = append(index.Entries, result)
		if len(index.Entries) > 2048 {
			return fmt.Errorf("provenance index exceeds 2048 entries")
		}
		if err := writeProvenanceIndex(store, index); err != nil {
			return err
		}
		return AuditEventRecord(auditStoreFor(store), AuditEvent{EventType: "provenance-recorded", Actor: "aji", Authority: "aji-merge", Target: input.Subject, ResultHandle: "wuji-provenance://" + result.RecordSHA256})
	})
	if err != nil {
		return ProvenanceEntry{}, err
	}
	return result, nil
}

func ResolveProvenance(store string, query ProvenanceQuery) (ProvenanceResolveResult, error) {
	query.Scope, query.Subject, query.Principal = strings.TrimSpace(query.Scope), strings.TrimSpace(query.Subject), strings.TrimSpace(query.Principal)
	scope, err := normalizeKnowledgeScope(query.Scope)
	if err != nil {
		return ProvenanceResolveResult{}, err
	}
	if err := validateBoundedHandle(query.Subject); err != nil {
		return ProvenanceResolveResult{}, err
	}
	if err := validateProvenancePrincipal(query.Principal, false); err != nil {
		return ProvenanceResolveResult{}, err
	}
	index, err := loadProvenanceIndex(store)
	if err != nil {
		return ProvenanceResolveResult{}, err
	}
	result := ProvenanceResolveResult{Scope: scope, Subject: query.Subject, Entries: []ProvenanceEntry{}}
	for _, entry := range index.Entries {
		if entry.Scope != scope || entry.Subject != query.Subject {
			continue
		}
		if !provenanceReaderAllowed(entry.Readers, query.Principal) {
			result.Denied++
			continue
		}
		result.Entries = append(result.Entries, entry)
	}
	sort.Slice(result.Entries, func(left, right int) bool { return result.Entries[left].ID < result.Entries[right].ID })
	return result, nil
}

func validateProvenanceInput(input *ProvenanceInput) error {
	if !componentIDPattern.MatchString(strings.TrimSpace(input.ID)) {
		return fmt.Errorf("provenance id is invalid")
	}
	scope, err := normalizeKnowledgeScope(input.Scope)
	if err != nil {
		return err
	}
	input.Scope = scope
	if err := validateBoundedHandle(input.Subject); err != nil {
		return err
	}
	if err := validateBoundedHandle(input.Target); err != nil {
		return err
	}
	if !componentIDPattern.MatchString(strings.TrimSpace(input.Predicate)) {
		return fmt.Errorf("provenance predicate is invalid")
	}
	readers, err := normalizeProvenanceReaders(input.Readers)
	if err != nil {
		return err
	}
	input.Readers = readers
	return nil
}

func normalizeProvenanceReaders(values []string) ([]string, error) {
	if len(values) == 0 || len(values) > 32 {
		return nil, fmt.Errorf("provenance readers must contain between 1 and 32 entries")
	}
	seen := map[string]bool{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if err := validateProvenancePrincipal(value, true); err != nil {
			return nil, err
		}
		if !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	sort.Strings(result)
	return result, nil
}

func validateProvenancePrincipal(value string, allowWildcard bool) error {
	if allowWildcard && value == "*" {
		return nil
	}
	if !componentIDPattern.MatchString(value) {
		return fmt.Errorf("provenance principal is invalid")
	}
	return nil
}

func provenanceReaderAllowed(readers []string, principal string) bool {
	return collectionContains(readers, "*") || collectionContains(readers, principal)
}

func loadProvenanceIndex(store string) (ProvenanceIndex, error) {
	path := provenancePath(store)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return ProvenanceIndex{SchemaVersion: provenanceSchemaVersion, Entries: []ProvenanceEntry{}}, nil
	}
	if err != nil {
		return ProvenanceIndex{}, err
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	var index ProvenanceIndex
	if err := decoder.Decode(&index); err != nil {
		return ProvenanceIndex{}, fmt.Errorf("decode provenance index: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return ProvenanceIndex{}, fmt.Errorf("decode provenance index: multiple JSON values are not allowed")
		}
		return ProvenanceIndex{}, err
	}
	if index.SchemaVersion != provenanceSchemaVersion || len(index.Entries) > 2048 {
		return ProvenanceIndex{}, fmt.Errorf("provenance schema or capacity is invalid")
	}
	seen := map[string]bool{}
	for _, entry := range index.Entries {
		input := ProvenanceInput{ID: entry.ID, Scope: entry.Scope, Subject: entry.Subject, Predicate: entry.Predicate, Target: entry.Target, Readers: entry.Readers}
		if err := validateProvenanceInput(&input); err != nil || entry.RecordSHA256 != provenanceDigest(entry) {
			return ProvenanceIndex{}, fmt.Errorf("provenance entry integrity is invalid")
		}
		if _, err := time.Parse(time.RFC3339, entry.RecordedAt); err != nil || seen[entry.ID] {
			return ProvenanceIndex{}, fmt.Errorf("provenance entry timestamp or identity is invalid")
		}
		seen[entry.ID] = true
	}
	return index, nil
}

func writeProvenanceIndex(store string, index ProvenanceIndex) error {
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	path := provenancePath(store)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return atomicWriteFile(path, data, 0o600)
}

func sameProvenance(entry ProvenanceEntry, input ProvenanceInput) bool {
	return entry.Scope == input.Scope && entry.Subject == input.Subject && entry.Predicate == input.Predicate && entry.Target == input.Target && sameStringSlice(entry.Readers, input.Readers)
}

func provenanceDigest(entry ProvenanceEntry) string {
	entry.RecordSHA256 = ""
	return lineageDigest(entry)
}

func provenancePath(store string) string {
	return filepath.Join(filepath.Clean(store), "v1", "index.json")
}
