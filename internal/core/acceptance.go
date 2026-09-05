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

const acceptanceSchemaVersion = 1

// The field remains named AcceptedBy for schema compatibility. New records are
// produced by deterministic evidence reconciliation, never by Aji or staff.
const acceptanceAuthority = "deterministic-evidence-verifier"

type AcceptanceRecord struct {
	ID                  string   `json:"id"`
	VersionID           string   `json:"version_id"`
	Revision            int      `json:"revision"`
	RequirementRevision string   `json:"requirement_revision"`
	ExecutionVersion    string   `json:"execution_version"`
	ArtifactHandles     []string `json:"artifact_handles"`
	VerificationHandles []string `json:"verification_handles"`
	AcceptedBy          string   `json:"accepted_by"`
	AcceptedAt          string   `json:"accepted_at"`
	Supersedes          string   `json:"supersedes,omitempty"`
	RecordSHA256        string   `json:"record_sha256"`
}

type AcceptanceInput struct {
	ID                  string
	RequirementRevision string
	ExecutionVersion    string
	ArtifactHandles     []string
	VerificationHandles []string
	AcceptedBy          string
}

type AcceptanceLedger struct {
	SchemaVersion int                `json:"schema_version"`
	Records       []AcceptanceRecord `json:"records"`
}

func DefaultAcceptanceStore() string {
	if value := strings.TrimSpace(os.Getenv("WUJI_ACCEPTANCE_DIR")); value != "" {
		return filepath.Clean(value)
	}
	return filepath.Join(".wuji", "acceptance")
}

func ReconcileAcceptance(store, requirementStore, executionStore string, input AcceptanceInput) (AcceptanceRecord, error) {
	if err := validateAcceptanceInput(input); err != nil {
		return AcceptanceRecord{}, err
	}
	if strings.TrimSpace(requirementStore) == "" {
		requirementStore = DefaultRequirementGraphStore()
	}
	if strings.TrimSpace(executionStore) == "" {
		executionStore = DefaultExecutionGraphStore()
	}
	requirements, err := loadRequirementGraph(requirementStore)
	if err != nil {
		return AcceptanceRecord{}, err
	}
	if !isActiveRequirementRevision(requirements, input.RequirementRevision) {
		return AcceptanceRecord{}, fmt.Errorf("requirement revision %q is not active", input.RequirementRevision)
	}
	executions, err := loadExecutionGraph(executionStore)
	if err != nil {
		return AcceptanceRecord{}, err
	}
	refreshExecutionInvalidations(&executions, requirements)
	execution := findExecutionVersion(executions, input.ExecutionVersion)
	if execution == nil {
		return AcceptanceRecord{}, fmt.Errorf("execution version %q is not found", input.ExecutionVersion)
	}
	if execution.Status != "succeeded" {
		return AcceptanceRecord{}, fmt.Errorf("execution version %q has not succeeded", input.ExecutionVersion)
	}
	if !collectionContains(execution.RequirementRevisions, input.RequirementRevision) {
		return AcceptanceRecord{}, fmt.Errorf("execution version %q does not bind requirement revision %q", input.ExecutionVersion, input.RequirementRevision)
	}
	artifacts := normalizedGraphList(input.ArtifactHandles)
	if len(artifacts) == 0 {
		artifacts = normalizedGraphList(execution.ArtifactHandles)
	}
	verification := normalizedGraphList(input.VerificationHandles)
	if len(verification) == 0 {
		verification = normalizedGraphList(execution.VerificationHandles)
	}
	if len(artifacts) == 0 || len(verification) == 0 {
		return AcceptanceRecord{}, fmt.Errorf("acceptance requires execution artifacts and verification evidence")
	}
	if !sameStringSlice(artifacts, normalizedGraphList(execution.ArtifactHandles)) || !sameStringSlice(verification, normalizedGraphList(execution.VerificationHandles)) {
		return AcceptanceRecord{}, fmt.Errorf("acceptance evidence does not match the execution result")
	}
	var result AcceptanceRecord
	err = withKnowledgeStoreLock(store, func() error {
		ledger, err := loadAcceptanceLedger(store)
		if err != nil {
			return err
		}
		current := latestAcceptanceRecord(ledger.Records, input.ID)
		if current != nil && sameAcceptance(*current, input, artifacts, verification) {
			result = *current
			return nil
		}
		revision := 1
		supersedes := ""
		if current != nil {
			revision = current.Revision + 1
			supersedes = current.VersionID
		}
		result = AcceptanceRecord{
			ID: input.ID, VersionID: graphVersionID(input.ID, revision), Revision: revision,
			RequirementRevision: input.RequirementRevision, ExecutionVersion: input.ExecutionVersion,
			ArtifactHandles: artifacts, VerificationHandles: verification, AcceptedBy: acceptanceAuthority,
			AcceptedAt: time.Now().UTC().Format(time.RFC3339), Supersedes: supersedes,
		}
		result.RecordSHA256 = acceptanceRecordDigest(result)
		ledger.Records = append(ledger.Records, result)
		if len(ledger.Records) > 512 {
			return fmt.Errorf("acceptance ledger exceeds 512 records")
		}
		if err := writeAcceptanceLedger(store, ledger); err != nil {
			return err
		}
		return AuditEventRecord(auditStoreFor(store), AuditEvent{EventType: "acceptance-reconciled", Actor: acceptanceAuthority, Authority: "evidence-reconciliation", Target: result.VersionID, InputRevision: result.RequirementRevision, ResultHandle: "wuji-acceptance://" + result.VersionID, EvidenceHandles: result.VerificationHandles})
	})
	if err != nil {
		return AcceptanceRecord{}, err
	}
	return result, nil
}

func validateAcceptanceInput(input AcceptanceInput) error {
	if !componentIDPattern.MatchString(strings.TrimSpace(input.ID)) {
		return fmt.Errorf("acceptance id is invalid")
	}
	if strings.TrimSpace(input.RequirementRevision) == "" || strings.TrimSpace(input.ExecutionVersion) == "" {
		return fmt.Errorf("requirement revision and execution version are required")
	}
	if value := strings.TrimSpace(input.AcceptedBy); value != "" && value != acceptanceAuthority {
		return fmt.Errorf("acceptance is derived only by the deterministic evidence verifier")
	}
	for _, handle := range append(append([]string{}, input.ArtifactHandles...), input.VerificationHandles...) {
		if err := validateBoundedHandle(handle); err != nil {
			return err
		}
	}
	return nil
}

func loadAcceptanceLedger(store string) (AcceptanceLedger, error) {
	path := acceptanceLedgerPath(store)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return AcceptanceLedger{SchemaVersion: acceptanceSchemaVersion, Records: []AcceptanceRecord{}}, nil
	}
	if err != nil {
		return AcceptanceLedger{}, err
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	var ledger AcceptanceLedger
	if err := decoder.Decode(&ledger); err != nil {
		return AcceptanceLedger{}, fmt.Errorf("decode acceptance ledger: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return AcceptanceLedger{}, fmt.Errorf("decode acceptance ledger: multiple JSON values are not allowed")
		}
		return AcceptanceLedger{}, err
	}
	if ledger.SchemaVersion != acceptanceSchemaVersion || len(ledger.Records) > 512 {
		return AcceptanceLedger{}, fmt.Errorf("acceptance ledger schema or capacity is invalid")
	}
	versions := map[string]bool{}
	for _, record := range ledger.Records {
		if err := validateAcceptanceRecord(record); err != nil {
			return AcceptanceLedger{}, err
		}
		if versions[record.VersionID] {
			return AcceptanceLedger{}, fmt.Errorf("acceptance ledger has duplicate version %q", record.VersionID)
		}
		versions[record.VersionID] = true
	}
	return ledger, nil
}

func writeAcceptanceLedger(store string, ledger AcceptanceLedger) error {
	data, err := json.MarshalIndent(ledger, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	path := acceptanceLedgerPath(store)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return atomicWriteFile(path, data, 0o600)
}

func validateAcceptanceRecord(record AcceptanceRecord) error {
	if !componentIDPattern.MatchString(record.ID) || record.Revision < 1 || record.VersionID != graphVersionID(record.ID, record.Revision) || record.AcceptedBy != acceptanceAuthority {
		return fmt.Errorf("acceptance record identity is invalid")
	}
	if strings.TrimSpace(record.RequirementRevision) == "" || strings.TrimSpace(record.ExecutionVersion) == "" || len(record.ArtifactHandles) == 0 || len(record.VerificationHandles) == 0 {
		return fmt.Errorf("acceptance record is incomplete")
	}
	if _, err := time.Parse(time.RFC3339, record.AcceptedAt); err != nil {
		return fmt.Errorf("acceptance record timestamp is invalid")
	}
	if record.RecordSHA256 != acceptanceRecordDigest(record) {
		return fmt.Errorf("acceptance record hash is invalid")
	}
	return nil
}

func latestAcceptanceRecord(records []AcceptanceRecord, id string) *AcceptanceRecord {
	var current *AcceptanceRecord
	for index := range records {
		if records[index].ID == id && (current == nil || records[index].Revision > current.Revision) {
			current = &records[index]
		}
	}
	return current
}

func sameAcceptance(record AcceptanceRecord, input AcceptanceInput, artifacts, verification []string) bool {
	return record.RequirementRevision == input.RequirementRevision && record.ExecutionVersion == input.ExecutionVersion && sameStringSlice(record.ArtifactHandles, artifacts) && sameStringSlice(record.VerificationHandles, verification)
}

func acceptanceRecordDigest(record AcceptanceRecord) string {
	record.RecordSHA256 = ""
	return lineageDigest(record)
}

func acceptanceLedgerPath(store string) string {
	return filepath.Join(filepath.Clean(store), "v1", "records.json")
}

func isActiveRequirementRevision(graph RequirementGraph, revision string) bool {
	for _, node := range graph.Nodes {
		if node.Kind == "requirement" && node.Status == "active" && node.VersionID == revision {
			return true
		}
	}
	return false
}

func findExecutionVersion(graph ExecutionGraph, version string) *ExecutionGraphNode {
	for index := range graph.Nodes {
		if graph.Nodes[index].VersionID == version {
			return &graph.Nodes[index]
		}
	}
	return nil
}

func collectionContains(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

func validateBoundedHandle(value string) error {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > 1024 || strings.ContainsAny(value, "\r\n\t") {
		return fmt.Errorf("evidence handle is invalid")
	}
	for _, pattern := range knowledgeSecretPatterns {
		if pattern.MatchString(value) {
			return fmt.Errorf("evidence handle must not contain a secret")
		}
	}
	return nil
}

func sortedUnique(values []string) []string {
	result := normalizedGraphList(values)
	sort.Strings(result)
	return result
}
