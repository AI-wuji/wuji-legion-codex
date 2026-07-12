package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const knowledgeSchemaVersion = 1

const (
	knowledgeMaxLookups       = 12
	knowledgeMaxCandidates    = 128
	knowledgeMaxResults       = 10
	knowledgeMaxRefsPerIndex  = 256
	knowledgeMaxTags          = 32
	knowledgeMaxRelations     = 32
	knowledgeMaxRecordBytes   = 64 * 1024
	knowledgeMaxReferenceSize = 1024
	knowledgeMaxKeyBytes      = 512
	knowledgeMaxScopeBytes    = 512
	knowledgeMaxSummaryBytes  = 4096
	knowledgeMaxRootCause     = 4096
	knowledgeMaxTagBytes      = 256
	knowledgeMaxTargetBytes   = 2048
	knowledgeMaxNodes         = 4096
	knowledgeMaxStoreBytes    = 256 * 1024 * 1024
	knowledgeMaxObjectBytes   = 2 * 1024 * 1024
	knowledgeRetention        = 365 * 24 * time.Hour
	knowledgeLockWait         = 2 * time.Second
)

var (
	knowledgeKindPattern        = regexp.MustCompile(`^[a-z][a-z0-9-]{0,31}$`)
	knowledgeSpacePattern       = regexp.MustCompile(`\s+`)
	knowledgeUUIDPattern        = regexp.MustCompile(`(?i)\b[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}\b`)
	knowledgeHexPattern         = regexp.MustCompile(`(?i)\b(?:0x)?[0-9a-f]{12,}\b`)
	knowledgeLongNumberPattern  = regexp.MustCompile(`\b\d{5,}\b`)
	knowledgeWindowsPathPattern = regexp.MustCompile(`(?i)\b[a-z]:\\[^\s"']+`)
	knowledgeWorkspaceScope     = regexp.MustCompile(`^workspace:[0-9a-f]{64}$`)
	knowledgeObjectReference    = regexp.MustCompile(`^object:sha256:([0-9a-f]{64})$`)
	knowledgeSecretPatterns     = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(?:api[_-]?key|authorization|bearer|password|secret)\s*[:=]\s*\S+`),
		regexp.MustCompile(`\bsk-[A-Za-z0-9_-]{16,}\b`),
		regexp.MustCompile(`\bgh[pousr]_[A-Za-z0-9]{16,}\b`),
		regexp.MustCompile(`\bxox[baprs]-[A-Za-z0-9-]{12,}\b`),
	}
	knowledgeRelationPredicates = map[string]bool{
		"caused-by": true, "depends-on": true, "derived-from": true, "implements": true,
		"located-at": true, "solved-by": true, "supersedes": true, "tests": true,
		"uses": true, "verified-by": true,
	}
)

type KnowledgeRelation struct {
	Predicate string `json:"predicate"`
	Target    string `json:"target"`
}

type KnowledgeRecordInput struct {
	Kind         string
	Key          string
	Scope        string
	Summary      string
	RootCause    string
	Location     string
	Verification string
	Tags         []string
	Relations    []KnowledgeRelation
}

type KnowledgeRecord struct {
	SchemaVersion      int                 `json:"schema_version"`
	ID                 string              `json:"id"`
	Kind               string              `json:"kind"`
	Key                string              `json:"key"`
	NormalizedKey      string              `json:"normalized_key"`
	Scope              string              `json:"scope,omitempty"`
	NormalizedScope    string              `json:"normalized_scope,omitempty"`
	Summary            string              `json:"summary"`
	RootCause          string              `json:"root_cause,omitempty"`
	Location           string              `json:"location"`
	Verification       string              `json:"verification"`
	VerificationSHA256 string              `json:"verification_sha256"`
	Tags               []string            `json:"tags,omitempty"`
	Relations          []KnowledgeRelation `json:"relations,omitempty"`
	Status             string              `json:"status"`
	Revision           int                 `json:"revision"`
	CreatedAt          string              `json:"created_at"`
	UpdatedAt          string              `json:"updated_at"`
	RecordSHA256       string              `json:"record_sha256"`
	Capacity           *KnowledgeCapacity  `json:"capacity,omitempty"`
}

type KnowledgeQuery struct {
	Trigger   string
	Kind      string
	Key       string
	Scope     string
	Tags      []string
	RelatedTo string
	Relation  string
	Limit     int
}

type KnowledgeMatch struct {
	ID                 string              `json:"id"`
	Kind               string              `json:"kind"`
	Score              int                 `json:"score"`
	Match              string              `json:"match"`
	Summary            string              `json:"summary"`
	RootCause          string              `json:"root_cause,omitempty"`
	Location           string              `json:"location"`
	Verification       string              `json:"verification"`
	VerificationSHA256 string              `json:"verification_sha256"`
	Relations          []KnowledgeRelation `json:"relations,omitempty"`
	Revision           int                 `json:"revision"`
	UpdatedAt          string              `json:"updated_at"`
}

type KnowledgeQueryResult struct {
	SchemaVersion       int               `json:"schema_version"`
	Trigger             string            `json:"trigger"`
	ExactMatch          bool              `json:"exact_match"`
	IndexLookups        int               `json:"index_lookups"`
	CandidateRecords    int               `json:"candidate_records"`
	FullScan            bool              `json:"full_scan"`
	Truncated           bool              `json:"truncated"`
	MaxIndexLookups     int               `json:"max_index_lookups"`
	MaxCandidateRecords int               `json:"max_candidate_records"`
	MaxResults          int               `json:"max_results"`
	MaxRefsPerIndex     int               `json:"max_refs_per_index"`
	Capacity            KnowledgeCapacity `json:"capacity"`
	Matches             []KnowledgeMatch  `json:"matches"`
}

type KnowledgeCapacity struct {
	NodeCount      int   `json:"node_count"`
	StoreBytes     int64 `json:"store_bytes"`
	MaxNodes       int   `json:"max_nodes"`
	MaxStoreBytes  int64 `json:"max_store_bytes"`
	RetentionDays  int   `json:"retention_days"`
	EvictedRecords int   `json:"evicted_records,omitempty"`
}

type knowledgeVerificationReceipt struct {
	SchemaVersion int    `json:"schema_version"`
	Type          string `json:"type"`
	Passed        bool   `json:"passed"`
	Verifier      string `json:"verifier"`
	VerifiedAt    string `json:"verified_at"`
}

func DefaultKnowledgeStore() string {
	if value := strings.TrimSpace(os.Getenv("WUJI_KNOWLEDGE_DIR")); value != "" {
		return filepath.Clean(value)
	}
	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return filepath.Join(".wuji", "knowledge")
	}
	return filepath.Join(home, ".wuji", "knowledge")
}

func RecordKnowledge(store string, input KnowledgeRecordInput) (KnowledgeRecord, error) {
	return recordKnowledgeAt(store, input, time.Now().UTC())
}

func recordKnowledgeAt(store string, input KnowledgeRecordInput, now time.Time) (KnowledgeRecord, error) {
	var result KnowledgeRecord
	err := withKnowledgeStoreLock(store, func() error {
		if err := repairKnowledgeStore(store); err != nil {
			return err
		}
		var err error
		result, err = recordKnowledgeLocked(store, input, now)
		return err
	})
	return result, err
}

func recordKnowledgeLocked(store string, input KnowledgeRecordInput, now time.Time) (KnowledgeRecord, error) {
	if err := validateKnowledgeInputBounds(input); err != nil {
		return KnowledgeRecord{}, err
	}
	kind := strings.ToLower(strings.TrimSpace(input.Kind))
	if !knowledgeKindPattern.MatchString(kind) {
		return KnowledgeRecord{}, fmt.Errorf("invalid knowledge kind %q", input.Kind)
	}
	key := strings.TrimSpace(input.Key)
	scope, err := normalizeKnowledgeScope(input.Scope)
	if err != nil {
		return KnowledgeRecord{}, err
	}
	summary := strings.TrimSpace(input.Summary)
	rootCause := strings.TrimSpace(input.RootCause)
	if key == "" || summary == "" {
		return KnowledgeRecord{}, fmt.Errorf("knowledge key and summary are required")
	}
	if scope == "" {
		return KnowledgeRecord{}, fmt.Errorf("knowledge scope is required; use an explicit repository identity or global")
	}
	if kind == "failure" && rootCause == "" {
		return KnowledgeRecord{}, fmt.Errorf("failure knowledge requires a verified root cause")
	}
	relationText := make([]string, 0, len(input.Relations))
	for _, relation := range input.Relations {
		relationText = append(relationText, relation.Predicate+"="+relation.Target)
	}
	if err := rejectKnowledgeSecrets(key, scope, summary, rootCause, strings.Join(input.Tags, " "), strings.Join(relationText, " ")); err != nil {
		return KnowledgeRecord{}, err
	}
	location, err := storeKnowledgeLocation(store, input.Location)
	if err != nil {
		return KnowledgeRecord{}, fmt.Errorf("solution location: %w", err)
	}
	verification, verificationSHA256, err := storeKnowledgeVerification(store, input.Verification, now)
	if err != nil {
		return KnowledgeRecord{}, fmt.Errorf("verification evidence: %w", err)
	}
	relations, err := normalizeKnowledgeRelations(input.Relations)
	if err != nil {
		return KnowledgeRecord{}, err
	}
	normalizedKey := normalizeKnowledgeKey(key)
	normalizedScope := normalizeKnowledgeText(scope)
	id := knowledgeRecordID(kind, normalizedKey, normalizedScope)
	recordPath := knowledgeRecordPath(store, kind, id)

	createdAt := now.Format(time.RFC3339Nano)
	revision := 1
	var previous *KnowledgeRecord
	if prior, readErr := readKnowledgeRecord(recordPath); readErr == nil {
		previous = &prior
		createdAt = prior.CreatedAt
		revision = prior.Revision + 1
	} else if !os.IsNotExist(readErr) {
		return KnowledgeRecord{}, readErr
	}

	record := KnowledgeRecord{
		SchemaVersion:      knowledgeSchemaVersion,
		ID:                 id,
		Kind:               kind,
		Key:                key,
		NormalizedKey:      normalizedKey,
		Scope:              scope,
		NormalizedScope:    normalizedScope,
		Summary:            summary,
		RootCause:          rootCause,
		Location:           location,
		Verification:       verification,
		VerificationSHA256: verificationSHA256,
		Tags:               normalizeKnowledgeList(input.Tags),
		Relations:          relations,
		Status:             "evidence-bound",
		Revision:           revision,
		CreatedAt:          createdAt,
		UpdatedAt:          now.Format(time.RFC3339Nano),
	}
	record.RecordSHA256, err = hashKnowledgeRecord(record)
	if err != nil {
		return KnowledgeRecord{}, err
	}
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return KnowledgeRecord{}, err
	}
	if len(data) > knowledgeMaxRecordBytes {
		return KnowledgeRecord{}, fmt.Errorf("knowledge record exceeds %d bytes", knowledgeMaxRecordBytes)
	}
	if err := ensureKnowledgeIndexCapacity(store, record); err != nil {
		return KnowledgeRecord{}, err
	}
	transactionPath := knowledgeTransactionPath(store)
	if err := writeKnowledgeFile(transactionPath, []byte("updating\n")); err != nil {
		return KnowledgeRecord{}, err
	}
	if err := writeKnowledgeFile(recordPath, append(data, '\n')); err != nil {
		return KnowledgeRecord{}, err
	}
	if previous != nil {
		if err := removeKnowledgeReferences(store, *previous); err != nil {
			return KnowledgeRecord{}, err
		}
	}
	for _, term := range knowledgeRecordTerms(record) {
		if err := writeKnowledgeReference(store, "terms", knowledgeScopedIndexValue(record.NormalizedScope, term), record); err != nil {
			return KnowledgeRecord{}, err
		}
	}
	for _, relation := range record.Relations {
		if err := writeKnowledgeReference(store, "relations", knowledgeScopedIndexValue(record.NormalizedScope, relation.Target), record); err != nil {
			return KnowledgeRecord{}, err
		}
		if err := writeKnowledgeReference(store, "relations", knowledgeScopedIndexValue(record.NormalizedScope, knowledgeRelationKey(relation.Predicate, relation.Target)), record); err != nil {
			return KnowledgeRecord{}, err
		}
	}
	if err := os.Remove(transactionPath); err != nil && !os.IsNotExist(err) {
		return KnowledgeRecord{}, err
	}
	capacity, err := enforceKnowledgeStoreCapacity(store, now, knowledgeMaxNodes, knowledgeMaxStoreBytes)
	if err != nil {
		return KnowledgeRecord{}, err
	}
	record.Capacity = &capacity
	return record, nil
}

func QueryKnowledge(store string, query KnowledgeQuery) (KnowledgeQueryResult, error) {
	var result KnowledgeQueryResult
	err := withKnowledgeStoreLock(store, func() error {
		if err := repairKnowledgeStore(store); err != nil {
			return err
		}
		var err error
		result, err = queryKnowledgeLocked(store, query)
		return err
	})
	return result, err
}

func queryKnowledgeLocked(store string, query KnowledgeQuery) (KnowledgeQueryResult, error) {
	if err := validateKnowledgeQueryBounds(query); err != nil {
		return KnowledgeQueryResult{}, err
	}
	trigger := strings.ToLower(strings.TrimSpace(query.Trigger))
	allowedTriggers := map[string]bool{
		"failure":            true,
		"reported-failure":   true,
		"explicit-reuse":     true,
		"capability-miss":    true,
		"verification-trace": true,
	}
	if !allowedTriggers[trigger] {
		return KnowledgeQueryResult{}, fmt.Errorf("knowledge query requires an event trigger; normal task startup is not allowed")
	}
	kind := strings.ToLower(strings.TrimSpace(query.Kind))
	if kind != "" && !knowledgeKindPattern.MatchString(kind) {
		return KnowledgeQueryResult{}, fmt.Errorf("invalid knowledge kind %q", query.Kind)
	}
	if strings.TrimSpace(query.Key) == "" && strings.TrimSpace(query.RelatedTo) == "" && len(query.Tags) == 0 {
		return KnowledgeQueryResult{}, fmt.Errorf("knowledge query requires a key, relation target, or tag")
	}
	limit := query.Limit
	if limit <= 0 {
		limit = 3
	}
	if limit > knowledgeMaxResults {
		return KnowledgeQueryResult{}, fmt.Errorf("knowledge query limit cannot exceed %d", knowledgeMaxResults)
	}
	result := KnowledgeQueryResult{
		SchemaVersion: knowledgeSchemaVersion, Trigger: trigger, FullScan: false, Matches: []KnowledgeMatch{},
		MaxIndexLookups: knowledgeMaxLookups, MaxCandidateRecords: knowledgeMaxCandidates,
		MaxResults: knowledgeMaxResults, MaxRefsPerIndex: knowledgeMaxRefsPerIndex,
	}
	normalizedKey := normalizeKnowledgeKey(query.Key)
	normalizedScope, err := normalizeKnowledgeScope(query.Scope)
	if err != nil {
		return KnowledgeQueryResult{}, err
	}
	capacity, err := knowledgeStoreCapacity(store)
	if err != nil {
		return KnowledgeQueryResult{}, err
	}
	result.Capacity = capacity
	candidates := map[string]KnowledgeRecord{}

	if kind != "" && normalizedKey != "" {
		result.IndexLookups++
		id := knowledgeRecordID(kind, normalizedKey, normalizedScope)
		record, err := readKnowledgeRecord(knowledgeRecordPath(store, kind, id))
		if err == nil {
			if err := validateKnowledgeRecord(store, record); err != nil {
				return KnowledgeQueryResult{}, err
			}
			result.ExactMatch = true
			result.CandidateRecords = 1
			result.Matches = []KnowledgeMatch{knowledgeMatch(record, 1000, "exact")}
			return result, nil
		}
		if !os.IsNotExist(err) {
			return KnowledgeQueryResult{}, err
		}
	}

	lookupTerms := knowledgeTerms(normalizedKey + " " + strings.Join(query.Tags, " "))
	if related := strings.TrimSpace(query.RelatedTo); related != "" {
		relation := normalizeKnowledgeText(query.Relation)
		if relation != "" && !knowledgeRelationPredicates[relation] {
			return KnowledgeQueryResult{}, fmt.Errorf("unsupported knowledge relation predicate %q", query.Relation)
		}
		if relation != "" {
			lookupTerms = append(lookupTerms, "relation:"+knowledgeRelationKey(relation, related))
		} else {
			lookupTerms = append(lookupTerms, "relation:"+normalizeKnowledgeText(related))
		}
	}
	lookupTerms = uniqueKnowledgeStrings(lookupTerms)
	if len(lookupTerms) > knowledgeMaxLookups {
		lookupTerms = lookupTerms[:knowledgeMaxLookups]
		result.Truncated = true
	}

lookupLoop:
	for _, term := range lookupTerms {
		indexType := "terms"
		indexTerm := term
		if strings.HasPrefix(term, "relation:") {
			indexType = "relations"
			indexTerm = strings.TrimPrefix(term, "relation:")
		}
		result.IndexLookups++
		refs, err := readKnowledgeReferences(store, indexType, knowledgeScopedIndexValue(normalizedScope, indexTerm), 64)
		if err != nil {
			return KnowledgeQueryResult{}, err
		}
		for _, ref := range refs {
			if _, exists := candidates[ref.ID]; exists {
				continue
			}
			record, err := readKnowledgeRecord(knowledgeRecordPath(store, ref.Kind, ref.ID))
			if err != nil {
				if os.IsNotExist(err) {
					continue
				}
				return KnowledgeQueryResult{}, err
			}
			if err := validateKnowledgeRecord(store, record); err != nil {
				return KnowledgeQueryResult{}, err
			}
			candidates[record.ID] = record
			if len(candidates) >= knowledgeMaxCandidates {
				result.Truncated = true
				break lookupLoop
			}
		}
	}
	result.CandidateRecords = len(candidates)

	type scoredRecord struct {
		record KnowledgeRecord
		score  int
	}
	scored := make([]scoredRecord, 0, len(candidates))
	queryTermSet := knowledgeTermSet(normalizedKey + " " + strings.Join(query.Tags, " "))
	for _, record := range candidates {
		if kind != "" && record.Kind != kind {
			continue
		}
		if record.NormalizedScope != normalizedScope {
			continue
		}
		score := 0
		if record.NormalizedKey == normalizedKey && normalizedKey != "" {
			score += 700
		}
		if normalizedScope != "" && record.NormalizedScope == normalizedScope {
			score += 200
		}
		for term := range knowledgeTermSet(record.NormalizedKey + " " + strings.Join(record.Tags, " ")) {
			if queryTermSet[term] {
				score += 20
			}
		}
		if strings.TrimSpace(query.RelatedTo) != "" && knowledgeHasRelation(record, query.Relation, query.RelatedTo) {
			score += 300
		}
		if score > 0 {
			scored = append(scored, scoredRecord{record: record, score: score})
		}
	}
	sort.Slice(scored, func(i, j int) bool {
		if scored[i].score == scored[j].score {
			return scored[i].record.UpdatedAt > scored[j].record.UpdatedAt
		}
		return scored[i].score > scored[j].score
	})
	for index, item := range scored {
		if index >= limit {
			break
		}
		result.Matches = append(result.Matches, knowledgeMatch(item.record, item.score, "indexed-candidate"))
	}
	return result, nil
}

// KnowledgeRecordPath exposes only the node location for behavior probes and
// repair tooling; callers still need to validate a record before using it.
func KnowledgeRecordPath(store, kind, id string) string {
	return knowledgeRecordPath(store, kind, id)
}

type knowledgeReference struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
}

func knowledgeRecordID(kind, key, scope string) string {
	hash := sha256.Sum256([]byte(kind + "\n" + key + "\n" + scope))
	return kind + "-" + hex.EncodeToString(hash[:12])
}

func knowledgeRecordPath(store, kind, id string) string {
	return filepath.Join(filepath.Clean(store), "v1", "nodes", kind, id+".json")
}

func knowledgeIndexPath(store, indexType, value string) string {
	hash := sha256.Sum256([]byte(normalizeKnowledgeText(value)))
	return filepath.Join(filepath.Clean(store), "v1", "indexes", indexType, hex.EncodeToString(hash[:]), "refs")
}

func knowledgeScopedIndexValue(scope, value string) string {
	return normalizeKnowledgeText(scope) + "\n" + normalizeKnowledgeText(value)
}

func knowledgeTransactionPath(store string) string {
	return filepath.Join(filepath.Clean(store), "v1", ".transaction")
}

func writeKnowledgeReference(store, indexType, value string, record KnowledgeRecord) error {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	data, err := json.Marshal(knowledgeReference{ID: record.ID, Kind: record.Kind})
	if err != nil {
		return err
	}
	return writeKnowledgeFile(filepath.Join(knowledgeIndexPath(store, indexType, value), record.ID+".json"), append(data, '\n'))
}

func ensureKnowledgeIndexCapacity(store string, record KnowledgeRecord) error {
	indexes := map[string]string{}
	for _, term := range knowledgeRecordTerms(record) {
		indexes["terms\n"+knowledgeScopedIndexValue(record.NormalizedScope, term)] = knowledgeScopedIndexValue(record.NormalizedScope, term)
	}
	for _, relation := range record.Relations {
		indexes["relations\n"+knowledgeScopedIndexValue(record.NormalizedScope, relation.Target)] = knowledgeScopedIndexValue(record.NormalizedScope, relation.Target)
		key := knowledgeRelationKey(relation.Predicate, relation.Target)
		indexes["relations\n"+knowledgeScopedIndexValue(record.NormalizedScope, key)] = knowledgeScopedIndexValue(record.NormalizedScope, key)
	}
	for composite, value := range indexes {
		parts := strings.SplitN(composite, "\n", 2)
		dir := knowledgeIndexPath(store, parts[0], value)
		entries, err := os.ReadDir(dir)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return err
		}
		count := 0
		exists := false
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
				continue
			}
			count++
			exists = exists || entry.Name() == record.ID+".json"
		}
		if !exists && count >= knowledgeMaxRefsPerIndex {
			return fmt.Errorf("knowledge index reference limit reached for %q", value)
		}
	}
	return nil
}

func removeKnowledgeReferences(store string, record KnowledgeRecord) error {
	for _, term := range knowledgeRecordTerms(record) {
		if err := removeKnowledgeReference(store, "terms", knowledgeScopedIndexValue(record.NormalizedScope, term), record.ID); err != nil {
			return err
		}
	}
	for _, relation := range record.Relations {
		if err := removeKnowledgeReference(store, "relations", knowledgeScopedIndexValue(record.NormalizedScope, relation.Target), record.ID); err != nil {
			return err
		}
		if err := removeKnowledgeReference(store, "relations", knowledgeScopedIndexValue(record.NormalizedScope, knowledgeRelationKey(relation.Predicate, relation.Target)), record.ID); err != nil {
			return err
		}
	}
	return nil
}

func removeKnowledgeReference(store, indexType, value, id string) error {
	err := os.Remove(filepath.Join(knowledgeIndexPath(store, indexType, value), id+".json"))
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func readKnowledgeReferences(store, indexType, value string, max int) ([]knowledgeReference, error) {
	dir := knowledgeIndexPath(store, indexType, value)
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if max <= 0 || max > knowledgeMaxRefsPerIndex {
		max = knowledgeMaxRefsPerIndex
	}
	refs := make([]knowledgeReference, 0, minInt(len(entries), max))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}
		if info.Size() > knowledgeMaxReferenceSize {
			return nil, fmt.Errorf("knowledge reference exceeds %d bytes: %s", knowledgeMaxReferenceSize, path)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		var ref knowledgeReference
		if err := json.Unmarshal(data, &ref); err != nil {
			return nil, err
		}
		refs = append(refs, ref)
		if len(refs) >= max {
			break
		}
	}
	return refs, nil
}

func readKnowledgeRecord(path string) (KnowledgeRecord, error) {
	info, err := os.Stat(path)
	if err != nil {
		return KnowledgeRecord{}, err
	}
	if info.Size() > knowledgeMaxRecordBytes {
		return KnowledgeRecord{}, fmt.Errorf("knowledge record exceeds %d bytes: %s", knowledgeMaxRecordBytes, path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return KnowledgeRecord{}, err
	}
	var record KnowledgeRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return KnowledgeRecord{}, err
	}
	return record, nil
}

func validateKnowledgeRecord(store string, record KnowledgeRecord) error {
	if record.SchemaVersion != knowledgeSchemaVersion || record.ID != knowledgeRecordID(record.Kind, record.NormalizedKey, record.NormalizedScope) {
		return fmt.Errorf("knowledge record identity is invalid: %s", record.ID)
	}
	if _, err := normalizeKnowledgeScope(record.Scope); err != nil || record.NormalizedScope != normalizeKnowledgeText(record.Scope) {
		return fmt.Errorf("knowledge record scope is invalid: %s", record.ID)
	}
	expected, err := hashKnowledgeRecord(record)
	if err != nil {
		return err
	}
	if expected != record.RecordSHA256 {
		return fmt.Errorf("knowledge record hash mismatch: %s", record.ID)
	}
	verificationSHA256, err := hashKnowledgeObject(store, record.Verification)
	if err != nil {
		return fmt.Errorf("knowledge verification evidence is unavailable for %s: %w", record.ID, err)
	}
	if verificationSHA256 != record.VerificationSHA256 {
		return fmt.Errorf("knowledge verification evidence hash mismatch: %s", record.ID)
	}
	return nil
}

func hashKnowledgeFile(path string) (string, error) {
	data, err := readKnowledgeFileBounded(path, knowledgeMaxObjectBytes)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

func hashKnowledgeRecord(record KnowledgeRecord) (string, error) {
	record.RecordSHA256 = ""
	record.Capacity = nil
	data, err := json.Marshal(record)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

func writeKnowledgeFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".wuji-knowledge-*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o600); err != nil {
		temporary.Close()
		return err
	}
	if _, err := temporary.Write(data); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		if removeErr := os.Remove(path); removeErr != nil && !os.IsNotExist(removeErr) {
			return err
		}
		return os.Rename(temporaryPath, path)
	}
	return nil
}

func normalizeKnowledgeKey(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = knowledgeWindowsPathPattern.ReplaceAllString(value, "<path>")
	value = knowledgeUUIDPattern.ReplaceAllString(value, "<uuid>")
	value = knowledgeHexPattern.ReplaceAllString(value, "<hex>")
	value = knowledgeLongNumberPattern.ReplaceAllString(value, "<number>")
	return knowledgeSpacePattern.ReplaceAllString(value, " ")
}

func normalizeKnowledgeText(value string) string {
	return knowledgeSpacePattern.ReplaceAllString(strings.ToLower(strings.TrimSpace(value)), " ")
}

func normalizeKnowledgeLocation(value string, requireLocal bool) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("path is required")
	}
	if !filepath.IsAbs(value) {
		if parsed, err := url.Parse(value); err == nil && parsed.Scheme != "" {
			if parsed.Scheme != "https" || requireLocal {
				return "", fmt.Errorf("only HTTPS solution locations are allowed; verification must be a local artifact")
			}
			return value, nil
		}
	}
	absolute, err := filepath.Abs(value)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(absolute)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return "", fmt.Errorf("expected a file, got directory %s", absolute)
	}
	return filepath.Clean(absolute), nil
}

func normalizeKnowledgeList(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = normalizeKnowledgeText(value)
		if value != "" {
			result = append(result, value)
		}
	}
	return uniqueKnowledgeStrings(result)
}

func normalizeKnowledgeRelations(relations []KnowledgeRelation) ([]KnowledgeRelation, error) {
	result := make([]KnowledgeRelation, 0, len(relations))
	seen := map[string]bool{}
	for _, relation := range relations {
		predicate := normalizeKnowledgeText(relation.Predicate)
		target := strings.TrimSpace(relation.Target)
		if predicate == "" || target == "" {
			continue
		}
		if !knowledgeRelationPredicates[predicate] {
			return nil, fmt.Errorf("unsupported knowledge relation predicate %q", relation.Predicate)
		}
		if filepath.IsAbs(target) || strings.Contains(target, ":\\") || strings.HasPrefix(strings.ToLower(target), "file:") {
			return nil, fmt.Errorf("knowledge relation targets must not contain local absolute paths")
		}
		key := predicate + "\n" + normalizeKnowledgeText(target)
		if !seen[key] {
			seen[key] = true
			result = append(result, KnowledgeRelation{Predicate: predicate, Target: target})
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Predicate == result[j].Predicate {
			return result[i].Target < result[j].Target
		}
		return result[i].Predicate < result[j].Predicate
	})
	return result, nil
}

func knowledgeRecordTerms(record KnowledgeRecord) []string {
	values := []string{record.Kind}
	values = append(values, knowledgeTerms(record.NormalizedKey)...)
	for _, tag := range record.Tags {
		values = append(values, knowledgeTerms(tag)...)
	}
	return uniqueKnowledgeStrings(values)
}

func knowledgeTerms(value string) []string {
	parts := strings.FieldsFunc(normalizeKnowledgeText(value), func(r rune) bool {
		return !(r >= 'a' && r <= 'z') && !(r >= '0' && r <= '9') && !(r >= '\u4e00' && r <= '\u9fff') && r != '<' && r != '>' && r != '-' && r != '_'
	})
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if len([]rune(part)) >= 2 {
			result = append(result, part)
		}
	}
	return result
}

func knowledgeTermSet(value string) map[string]bool {
	result := map[string]bool{}
	for _, term := range knowledgeTerms(value) {
		result[term] = true
	}
	return result
}

func uniqueKnowledgeStrings(values []string) []string {
	seen := map[string]bool{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value != "" && !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	sort.Strings(result)
	return result
}

func knowledgeHasRelation(record KnowledgeRecord, predicate, target string) bool {
	predicate = normalizeKnowledgeText(predicate)
	target = normalizeKnowledgeText(target)
	for _, relation := range record.Relations {
		if normalizeKnowledgeText(relation.Target) == target && (predicate == "" || relation.Predicate == predicate) {
			return true
		}
	}
	return false
}

func knowledgeRelationKey(predicate, target string) string {
	return normalizeKnowledgeText(predicate) + "\n" + normalizeKnowledgeText(target)
}

func knowledgeMatch(record KnowledgeRecord, score int, match string) KnowledgeMatch {
	return KnowledgeMatch{
		ID:                 record.ID,
		Kind:               record.Kind,
		Score:              score,
		Match:              match,
		Summary:            record.Summary,
		RootCause:          record.RootCause,
		Location:           record.Location,
		Verification:       record.Verification,
		VerificationSHA256: record.VerificationSHA256,
		Relations:          record.Relations,
		Revision:           record.Revision,
		UpdatedAt:          record.UpdatedAt,
	}
}

func rejectKnowledgeSecrets(values ...string) error {
	joined := strings.Join(values, "\n")
	for _, pattern := range knowledgeSecretPatterns {
		if pattern.MatchString(joined) {
			return fmt.Errorf("knowledge records must not contain credentials or secret values")
		}
	}
	return nil
}

func validateKnowledgeInputBounds(input KnowledgeRecordInput) error {
	fields := []struct {
		name  string
		value string
		max   int
	}{
		{"key", input.Key, knowledgeMaxKeyBytes},
		{"scope", input.Scope, knowledgeMaxScopeBytes},
		{"summary", input.Summary, knowledgeMaxSummaryBytes},
		{"root cause", input.RootCause, knowledgeMaxRootCause},
		{"location", input.Location, knowledgeMaxTargetBytes},
		{"verification", input.Verification, knowledgeMaxTargetBytes},
	}
	for _, field := range fields {
		if len([]byte(field.value)) > field.max {
			return fmt.Errorf("knowledge %s exceeds %d bytes", field.name, field.max)
		}
	}
	if len(input.Tags) > knowledgeMaxTags {
		return fmt.Errorf("knowledge tags cannot exceed %d", knowledgeMaxTags)
	}
	for _, tag := range input.Tags {
		if len([]byte(tag)) > knowledgeMaxTagBytes {
			return fmt.Errorf("knowledge tag exceeds %d bytes", knowledgeMaxTagBytes)
		}
	}
	if len(input.Relations) > knowledgeMaxRelations {
		return fmt.Errorf("knowledge relations cannot exceed %d", knowledgeMaxRelations)
	}
	for _, relation := range input.Relations {
		if len([]byte(relation.Target)) > knowledgeMaxTargetBytes {
			return fmt.Errorf("knowledge relation target exceeds %d bytes", knowledgeMaxTargetBytes)
		}
	}
	return nil
}

func validateKnowledgeQueryBounds(query KnowledgeQuery) error {
	if len([]byte(query.Key)) > knowledgeMaxKeyBytes {
		return fmt.Errorf("knowledge query key exceeds %d bytes", knowledgeMaxKeyBytes)
	}
	if len([]byte(query.Scope)) > knowledgeMaxScopeBytes {
		return fmt.Errorf("knowledge query scope exceeds %d bytes", knowledgeMaxScopeBytes)
	}
	if len(query.Tags) > knowledgeMaxTags {
		return fmt.Errorf("knowledge query tags cannot exceed %d", knowledgeMaxTags)
	}
	for _, tag := range query.Tags {
		if len([]byte(tag)) > knowledgeMaxTagBytes {
			return fmt.Errorf("knowledge query tag exceeds %d bytes", knowledgeMaxTagBytes)
		}
	}
	if len([]byte(query.RelatedTo)) > knowledgeMaxTargetBytes {
		return fmt.Errorf("knowledge query relation target exceeds %d bytes", knowledgeMaxTargetBytes)
	}
	return nil
}
