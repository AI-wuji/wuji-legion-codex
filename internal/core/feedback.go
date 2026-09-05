package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	executionFeedbackSchemaVersion = 1
	executionFeedbackMaxRecords    = 1024
	executionFeedbackMaxBytes      = 2 * 1024 * 1024
)

// ExecutionFeedbackRecord is a bounded, unverified routing hint. It is not a
// KnowledgeRecord and it can never change the active capability set.
type ExecutionFeedbackRecord struct {
	SchemaVersion    int    `json:"schema_version"`
	ID               string `json:"id"`
	ExecutionVersion string `json:"execution_version"`
	TaskInstanceID   string `json:"task_instance_id"`
	GraphVersion     string `json:"graph_version"`
	AttemptID        string `json:"attempt_id"`
	Outcome          string `json:"outcome"`
	Status           string `json:"status"`
	RecordedAt       string `json:"recorded_at"`
	RecordSHA256     string `json:"record_sha256"`
}

type ExecutionFeedbackLedger struct {
	SchemaVersion int                       `json:"schema_version"`
	Records       []ExecutionFeedbackRecord `json:"records"`
}

func DefaultExecutionFeedbackStore() string {
	if value := strings.TrimSpace(os.Getenv("WUJI_EXECUTION_FEEDBACK_DIR")); value != "" {
		return filepath.Clean(value)
	}
	return filepath.Join(".wuji", "execution-feedback")
}

func executionFeedbackStoreFor(executionStore string) string {
	if strings.TrimSpace(os.Getenv("WUJI_EXECUTION_FEEDBACK_DIR")) != "" {
		return DefaultExecutionFeedbackStore()
	}
	return filepath.Join(filepath.Dir(filepath.Clean(executionStore)), "execution-feedback")
}

// recordExecutionFeedback stores one candidate event for a terminal,
// versioned execution. The graph and runtime binding are re-read so stale,
// cancelled, invalidated, or rebound attempts cannot be recorded.
func recordExecutionFeedback(store, executionStore, requirementStore, executionVersion string, binding executionRuntimeBinding) (ExecutionFeedbackRecord, error) {
	if err := validateExecutionRuntimeBinding(binding); err != nil {
		return ExecutionFeedbackRecord{}, err
	}
	if strings.TrimSpace(executionVersion) == "" {
		return ExecutionFeedbackRecord{}, fmt.Errorf("execution feedback version is required")
	}
	if len([]byte(strings.TrimSpace(executionVersion))) > executionGraphMaxFieldBytes || strings.ContainsAny(executionVersion, "\r\n\t") {
		return ExecutionFeedbackRecord{}, fmt.Errorf("execution feedback version is invalid")
	}
	if err := rejectKnowledgeSecrets(binding.TaskInstanceID, binding.GraphVersion, binding.AttemptID); err != nil {
		return ExecutionFeedbackRecord{}, fmt.Errorf("execution feedback binding: %w", err)
	}
	if strings.TrimSpace(requirementStore) == "" {
		requirementStore = DefaultRequirementGraphStore()
	}
	requirements, err := loadRequirementGraph(requirementStore)
	if err != nil {
		return ExecutionFeedbackRecord{}, err
	}
	var result ExecutionFeedbackRecord
	err = withKnowledgeStoreLock(executionStore, func() error {
		graph, err := loadExecutionGraph(executionStore)
		if err != nil {
			return err
		}
		refreshExecutionInvalidations(&graph, requirements)
		node := findExecutionVersion(graph, strings.TrimSpace(executionVersion))
		if node == nil {
			return fmt.Errorf("execution feedback version %q is not found", executionVersion)
		}
		if node.Status != "succeeded" && node.Status != "failed" {
			return fmt.Errorf("execution feedback requires a terminal succeeded or failed result")
		}
		runtime, err := loadExecutionRuntimeGraph(executionStore)
		if err != nil {
			return err
		}
		if actual, ok := runtime.Nodes[node.VersionID]; !ok || actual != binding {
			return fmt.Errorf("execution feedback runtime binding is stale")
		}
		result, err = recordExecutionFeedbackCandidate(store, *node, binding)
		return err
	})
	return result, err
}

// recordExecutionFeedbackCandidate receives an execution snapshot while the
// execution-store lock is held. It only writes a cold candidate event.
func recordExecutionFeedbackCandidate(store string, node ExecutionGraphNode, binding executionRuntimeBinding) (ExecutionFeedbackRecord, error) {
	if node.Status != "succeeded" && node.Status != "failed" {
		return ExecutionFeedbackRecord{}, fmt.Errorf("execution feedback requires a terminal succeeded or failed result")
	}
	if err := rejectKnowledgeSecrets(binding.TaskInstanceID, binding.GraphVersion, binding.AttemptID); err != nil {
		return ExecutionFeedbackRecord{}, fmt.Errorf("execution feedback binding: %w", err)
	}
	result := ExecutionFeedbackRecord{}
	err := withKnowledgeStoreLock(store, func() error {
		ledger, err := loadExecutionFeedbackLedger(store)
		if err != nil {
			return err
		}
		id := executionFeedbackID(node.VersionID, binding)
		for _, existing := range ledger.Records {
			if existing.ID == id {
				result = existing
				return nil
			}
		}
		if len(ledger.Records) >= executionFeedbackMaxRecords {
			return fmt.Errorf("execution feedback ledger exceeds %d records", executionFeedbackMaxRecords)
		}
		outcome := "unverified-success"
		if node.Status == "failed" {
			outcome = "unverified-failure"
		}
		result = ExecutionFeedbackRecord{
			SchemaVersion: executionFeedbackSchemaVersion, ID: id, ExecutionVersion: node.VersionID,
			TaskInstanceID: binding.TaskInstanceID, GraphVersion: binding.GraphVersion, AttemptID: binding.AttemptID,
			Outcome: outcome, Status: "candidate", RecordedAt: time.Now().UTC().Format(time.RFC3339Nano),
		}
		result.RecordSHA256 = executionFeedbackDigest(result)
		ledger.Records = append(ledger.Records, result)
		return writeExecutionFeedbackLedger(store, ledger)
	})
	return result, err
}

// RecordVerifiedFailureFeedbackKnowledge is an explicit bridge from an
// unverified failure candidate to the bounded knowledge store. It does not
// trust execution handles: RecordKnowledge independently parses and hashes
// the caller-provided local verification receipt before admitting anything.
// The bridge neither queries knowledge during normal startup nor promotes a
// capability, source, or candidate.
func RecordVerifiedFailureFeedbackKnowledge(feedbackStore, knowledgeStore, feedbackID string, input KnowledgeRecordInput) (KnowledgeRecord, error) {
	feedbackID = strings.TrimSpace(feedbackID)
	if len(feedbackID) == 0 || len([]byte(feedbackID)) > 128 || strings.ContainsAny(feedbackID, "\r\n\t") {
		return KnowledgeRecord{}, fmt.Errorf("execution feedback id is invalid")
	}
	sameStore, err := sameExecutionFeedbackAndKnowledgeStore(feedbackStore, knowledgeStore)
	if err != nil {
		return KnowledgeRecord{}, err
	}
	if sameStore {
		return KnowledgeRecord{}, fmt.Errorf("feedback and knowledge stores must be separate")
	}
	var candidate ExecutionFeedbackRecord
	err = withKnowledgeStoreLock(feedbackStore, func() error {
		ledger, err := loadExecutionFeedbackLedger(feedbackStore)
		if err != nil {
			return err
		}
		for _, record := range ledger.Records {
			if record.ID == feedbackID {
				candidate = record
				return nil
			}
		}
		return fmt.Errorf("execution feedback candidate %q is not found", feedbackID)
	})
	if err != nil {
		return KnowledgeRecord{}, err
	}
	if candidate.Outcome != "unverified-failure" || candidate.Status != "candidate" {
		return KnowledgeRecord{}, fmt.Errorf("execution feedback candidate is not an eligible failure event")
	}
	if strings.ToLower(strings.TrimSpace(input.Kind)) != "failure" || strings.TrimSpace(input.RootCause) == "" {
		return KnowledgeRecord{}, fmt.Errorf("verified feedback knowledge requires failure kind and explicit root cause")
	}
	input.Relations = append(append([]KnowledgeRelation{}, input.Relations...), KnowledgeRelation{Predicate: "derived-from", Target: "wuji-feedback://" + candidate.ID})
	return RecordKnowledge(knowledgeStore, input)
}

func sameExecutionFeedbackAndKnowledgeStore(feedbackStore, knowledgeStore string) (bool, error) {
	feedbackPath, err := filepath.Abs(feedbackStore)
	if err != nil {
		return false, fmt.Errorf("resolve feedback store: %w", err)
	}
	knowledgePath, err := filepath.Abs(knowledgeStore)
	if err != nil {
		return false, fmt.Errorf("resolve knowledge store: %w", err)
	}
	if filepath.Clean(feedbackPath) == filepath.Clean(knowledgePath) {
		return true, nil
	}
	feedbackInfo, feedbackErr := os.Stat(feedbackPath)
	knowledgeInfo, knowledgeErr := os.Stat(knowledgePath)
	if feedbackErr == nil && knowledgeErr == nil {
		return os.SameFile(feedbackInfo, knowledgeInfo), nil
	}
	if feedbackErr != nil && !os.IsNotExist(feedbackErr) {
		return false, fmt.Errorf("inspect feedback store: %w", feedbackErr)
	}
	if knowledgeErr != nil && !os.IsNotExist(knowledgeErr) {
		return false, fmt.Errorf("inspect knowledge store: %w", knowledgeErr)
	}
	return false, nil
}

func executionFeedbackID(executionVersion string, binding executionRuntimeBinding) string {
	value := executionVersion + "\n" + binding.TaskInstanceID + "\n" + binding.GraphVersion + "\n" + binding.AttemptID
	sum := sha256.Sum256([]byte(value))
	return "feedback-" + hex.EncodeToString(sum[:16])
}

func executionFeedbackLedgerPath(store string) string {
	return filepath.Join(filepath.Clean(store), "v1", "records.json")
}

func loadExecutionFeedbackLedger(store string) (ExecutionFeedbackLedger, error) {
	file, err := os.Open(executionFeedbackLedgerPath(store))
	if os.IsNotExist(err) {
		return ExecutionFeedbackLedger{SchemaVersion: executionFeedbackSchemaVersion, Records: []ExecutionFeedbackRecord{}}, nil
	}
	if err != nil {
		return ExecutionFeedbackLedger{}, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, executionFeedbackMaxBytes+1))
	if err != nil {
		return ExecutionFeedbackLedger{}, err
	}
	if len(data) > executionFeedbackMaxBytes {
		return ExecutionFeedbackLedger{}, fmt.Errorf("execution feedback ledger exceeds %d bytes", executionFeedbackMaxBytes)
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	var ledger ExecutionFeedbackLedger
	if err := decoder.Decode(&ledger); err != nil {
		return ExecutionFeedbackLedger{}, fmt.Errorf("decode execution feedback ledger: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return ExecutionFeedbackLedger{}, fmt.Errorf("decode execution feedback ledger: multiple JSON values are not allowed")
		}
		return ExecutionFeedbackLedger{}, err
	}
	if ledger.SchemaVersion != executionFeedbackSchemaVersion || len(ledger.Records) > executionFeedbackMaxRecords {
		return ExecutionFeedbackLedger{}, fmt.Errorf("execution feedback ledger schema or capacity is invalid")
	}
	seen := map[string]bool{}
	for _, record := range ledger.Records {
		if err := validateExecutionFeedbackRecord(record); err != nil {
			return ExecutionFeedbackLedger{}, err
		}
		if seen[record.ID] {
			return ExecutionFeedbackLedger{}, fmt.Errorf("execution feedback ledger has duplicate record %q", record.ID)
		}
		seen[record.ID] = true
	}
	return ledger, nil
}

func writeExecutionFeedbackLedger(store string, ledger ExecutionFeedbackLedger) error {
	data, err := json.MarshalIndent(ledger, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	path := executionFeedbackLedgerPath(store)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return atomicWriteFile(path, data, 0o600)
}

func validateExecutionFeedbackRecord(record ExecutionFeedbackRecord) error {
	binding := executionRuntimeBinding{TaskInstanceID: record.TaskInstanceID, GraphVersion: record.GraphVersion, AttemptID: record.AttemptID}
	if record.SchemaVersion != executionFeedbackSchemaVersion || !strings.HasPrefix(record.ID, "feedback-") || strings.TrimSpace(record.ExecutionVersion) == "" || validateExecutionRuntimeBinding(binding) != nil {
		return fmt.Errorf("execution feedback record identity is invalid")
	}
	if record.ID != executionFeedbackID(record.ExecutionVersion, binding) || (record.Outcome != "unverified-success" && record.Outcome != "unverified-failure") || record.Status != "candidate" {
		return fmt.Errorf("execution feedback record state is invalid")
	}
	if _, err := time.Parse(time.RFC3339Nano, record.RecordedAt); err != nil || record.RecordSHA256 != executionFeedbackDigest(record) {
		return fmt.Errorf("execution feedback record integrity is invalid")
	}
	return nil
}

func executionFeedbackDigest(record ExecutionFeedbackRecord) string {
	record.RecordSHA256 = ""
	data, _ := json.Marshal(record)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
