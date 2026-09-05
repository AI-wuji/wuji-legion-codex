package core

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const conversationEvidenceSchemaVersion = 1

type ConversationEvidenceRecord struct {
	Revision       string   `json:"revision"`
	MessageHandles []string `json:"message_handles"`
	RecordedAt     string   `json:"recorded_at"`
	RecordSHA256   string   `json:"record_sha256"`
}

type ConversationEvidenceIndex struct {
	SchemaVersion int                          `json:"schema_version"`
	Records       []ConversationEvidenceRecord `json:"records"`
}

type ConversationEvidenceQuery struct {
	Revision      string `json:"revision,omitempty"`
	MessageHandle string `json:"message_handle,omitempty"`
}

func DefaultConversationEvidenceStore() string {
	if value := strings.TrimSpace(os.Getenv("WUJI_CONVERSATION_EVIDENCE_DIR")); value != "" {
		return filepath.Clean(value)
	}
	return filepath.Join(".wuji", "conversation-evidence")
}

func LinkConversationEvidence(store, requirementStore, revision string, messageHandles []string) (ConversationEvidenceRecord, error) {
	revision = strings.TrimSpace(revision)
	if revision == "" {
		return ConversationEvidenceRecord{}, fmt.Errorf("revision is required")
	}
	if strings.TrimSpace(requirementStore) == "" {
		requirementStore = DefaultRequirementGraphStore()
	}
	graph, err := loadRequirementGraph(requirementStore)
	if err != nil {
		return ConversationEvidenceRecord{}, err
	}
	if !requirementRevisionExists(graph, revision) {
		return ConversationEvidenceRecord{}, fmt.Errorf("revision %q is not found", revision)
	}
	handles, err := normalizeConversationHandles(messageHandles)
	if err != nil {
		return ConversationEvidenceRecord{}, err
	}
	var result ConversationEvidenceRecord
	err = withKnowledgeStoreLock(store, func() error {
		index, err := loadConversationEvidenceIndex(store)
		if err != nil {
			return err
		}
		for recordIndex := range index.Records {
			record := &index.Records[recordIndex]
			if record.Revision != revision {
				continue
			}
			merged := normalizedGraphList(append(record.MessageHandles, handles...))
			if sameStringSlice(merged, record.MessageHandles) {
				result = *record
				return nil
			}
			record.MessageHandles = merged
			record.RecordedAt = time.Now().UTC().Format(time.RFC3339)
			record.RecordSHA256 = conversationEvidenceDigest(*record)
			result = *record
			if err := writeConversationEvidenceIndex(store, index); err != nil {
				return err
			}
			return AuditEventRecord(auditStoreFor(store), AuditEvent{EventType: "conversation-evidence-linked", Actor: "aji", Authority: "aji-merge", Target: revision, ResultHandle: "wuji-conversation-evidence://" + result.RecordSHA256})
		}
		result = ConversationEvidenceRecord{Revision: revision, MessageHandles: handles, RecordedAt: time.Now().UTC().Format(time.RFC3339)}
		result.RecordSHA256 = conversationEvidenceDigest(result)
		index.Records = append(index.Records, result)
		if len(index.Records) > 512 {
			return fmt.Errorf("conversation evidence index exceeds 512 records")
		}
		if err := writeConversationEvidenceIndex(store, index); err != nil {
			return err
		}
		return AuditEventRecord(auditStoreFor(store), AuditEvent{EventType: "conversation-evidence-linked", Actor: "aji", Authority: "aji-merge", Target: revision, ResultHandle: "wuji-conversation-evidence://" + result.RecordSHA256})
	})
	if err != nil {
		return ConversationEvidenceRecord{}, err
	}
	return result, nil
}

func ResolveConversationEvidence(store string, query ConversationEvidenceQuery) ([]ConversationEvidenceRecord, error) {
	query.Revision, query.MessageHandle = strings.TrimSpace(query.Revision), strings.TrimSpace(query.MessageHandle)
	if (query.Revision == "" && query.MessageHandle == "") || (query.Revision != "" && query.MessageHandle != "") {
		return nil, fmt.Errorf("exactly one revision or message handle is required")
	}
	if query.MessageHandle != "" {
		if _, err := normalizeConversationHandles([]string{query.MessageHandle}); err != nil {
			return nil, err
		}
	}
	index, err := loadConversationEvidenceIndex(store)
	if err != nil {
		return nil, err
	}
	result := []ConversationEvidenceRecord{}
	for _, record := range index.Records {
		if query.Revision != "" && record.Revision == query.Revision {
			result = append(result, record)
		}
		if query.MessageHandle != "" && collectionContains(record.MessageHandles, query.MessageHandle) {
			result = append(result, record)
		}
	}
	return result, nil
}

func normalizeConversationHandles(values []string) ([]string, error) {
	if len(values) == 0 || len(values) > 32 {
		return nil, fmt.Errorf("message handles must contain between 1 and 32 entries")
	}
	result := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if !strings.HasPrefix(value, "message:") && !strings.HasPrefix(value, "host-message:") {
			return nil, fmt.Errorf("message handle must be an opaque host handle")
		}
		if len(value) > 512 || strings.ContainsAny(value, " \r\n\t") {
			return nil, fmt.Errorf("message handle is invalid")
		}
		for _, pattern := range knowledgeSecretPatterns {
			if pattern.MatchString(value) {
				return nil, fmt.Errorf("message handle must not contain a secret")
			}
		}
		if !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	return normalizedGraphList(result), nil
}

func loadConversationEvidenceIndex(store string) (ConversationEvidenceIndex, error) {
	path := conversationEvidencePath(store)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return ConversationEvidenceIndex{SchemaVersion: conversationEvidenceSchemaVersion, Records: []ConversationEvidenceRecord{}}, nil
	}
	if err != nil {
		return ConversationEvidenceIndex{}, err
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	var index ConversationEvidenceIndex
	if err := decoder.Decode(&index); err != nil {
		return ConversationEvidenceIndex{}, fmt.Errorf("decode conversation evidence: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return ConversationEvidenceIndex{}, fmt.Errorf("decode conversation evidence: multiple JSON values are not allowed")
		}
		return ConversationEvidenceIndex{}, err
	}
	if index.SchemaVersion != conversationEvidenceSchemaVersion || len(index.Records) > 512 {
		return ConversationEvidenceIndex{}, fmt.Errorf("conversation evidence schema or capacity is invalid")
	}
	seen := map[string]bool{}
	for _, record := range index.Records {
		if record.Revision == "" || seen[record.Revision] {
			return ConversationEvidenceIndex{}, fmt.Errorf("conversation evidence revision is invalid")
		}
		seen[record.Revision] = true
		if _, err := normalizeConversationHandles(record.MessageHandles); err != nil {
			return ConversationEvidenceIndex{}, err
		}
		if _, err := time.Parse(time.RFC3339, record.RecordedAt); err != nil || record.RecordSHA256 != conversationEvidenceDigest(record) {
			return ConversationEvidenceIndex{}, fmt.Errorf("conversation evidence integrity is invalid")
		}
	}
	return index, nil
}

func writeConversationEvidenceIndex(store string, index ConversationEvidenceIndex) error {
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	path := conversationEvidencePath(store)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return atomicWriteFile(path, data, 0o600)
}

func conversationEvidenceDigest(record ConversationEvidenceRecord) string {
	record.RecordSHA256 = ""
	return lineageDigest(record)
}

func requirementRevisionExists(graph RequirementGraph, revision string) bool {
	for _, node := range graph.Nodes {
		if node.VersionID == revision {
			return true
		}
	}
	return false
}

func conversationEvidencePath(store string) string {
	return filepath.Join(filepath.Clean(store), "v1", "index.json")
}
