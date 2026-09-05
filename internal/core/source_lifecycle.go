package core

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const sourceAssessmentSchemaVersion = 1

var sourceVersionPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._+-]{0,127}$`)

type SourceAssessment struct {
	SourceID        string   `json:"source_id"`
	Version         string   `json:"version"`
	Decision        string   `json:"decision"`
	Reason          string   `json:"reason"`
	EvidenceHandles []string `json:"evidence_handles,omitempty"`
	ReanalyzeWhen   []string `json:"reanalyze_when"`
	RecordedAt      string   `json:"recorded_at"`
	RecordSHA256    string   `json:"record_sha256"`
}

type SourceAssessmentInput struct {
	SourceID        string
	Version         string
	Decision        string
	Reason          string
	EvidenceHandles []string
	ReanalyzeWhen   []string
}

type SourceAssessmentStore struct {
	SchemaVersion int                `json:"schema_version"`
	Assessments   []SourceAssessment `json:"assessments"`
}

type SourceImpactResult struct {
	SourceID         string        `json:"source_id"`
	CandidateVersion string        `json:"candidate_version"`
	KnownVersions    []string      `json:"known_versions"`
	ImpactedNodes    []LineageNode `json:"impacted_nodes"`
}

func DefaultSourceAssessmentStore() string {
	if value := strings.TrimSpace(os.Getenv("WUJI_SOURCE_ASSESSMENT_DIR")); value != "" {
		return filepath.Clean(value)
	}
	return filepath.Join(".wuji", "source-assessments")
}

func AssessSource(store string, input SourceAssessmentInput) (SourceAssessment, error) {
	if err := validateSourceAssessmentInput(input); err != nil {
		return SourceAssessment{}, err
	}
	var result SourceAssessment
	err := withKnowledgeStoreLock(store, func() error {
		ledger, err := loadSourceAssessmentStore(store)
		if err != nil {
			return err
		}
		for _, assessment := range ledger.Assessments {
			if assessment.SourceID == input.SourceID && assessment.Version == input.Version {
				result = assessment
				return nil
			}
		}
		result = SourceAssessment{
			SourceID: input.SourceID, Version: input.Version, Decision: strings.ToLower(strings.TrimSpace(input.Decision)),
			Reason: strings.TrimSpace(input.Reason), EvidenceHandles: normalizedGraphList(input.EvidenceHandles),
			ReanalyzeWhen: normalizedSourceReanalysis(input.ReanalyzeWhen), RecordedAt: time.Now().UTC().Format(time.RFC3339),
		}
		result.RecordSHA256 = sourceAssessmentDigest(result)
		ledger.Assessments = append(ledger.Assessments, result)
		if len(ledger.Assessments) > 1024 {
			return fmt.Errorf("source assessment store exceeds 1024 records")
		}
		if err := writeSourceAssessmentStore(store, ledger); err != nil {
			return err
		}
		return AuditEventRecord(auditStoreFor(store), AuditEvent{EventType: "source-assessed", Actor: "deterministic-source-validator", Authority: "governed-source-assessment", Target: input.SourceID + "@" + input.Version, ResultHandle: "wuji-source-assessment://" + sourceAssessmentDigest(result), EvidenceHandles: result.EvidenceHandles})
	})
	if err != nil {
		return SourceAssessment{}, err
	}
	return result, nil
}

func SourceImpact(catalog LineageCatalog, sourceID, candidateVersion string) (SourceImpactResult, error) {
	sourceID, candidateVersion = strings.TrimSpace(sourceID), strings.TrimSpace(candidateVersion)
	if !componentIDPattern.MatchString(sourceID) || !sourceVersionPattern.MatchString(candidateVersion) {
		return SourceImpactResult{}, fmt.Errorf("source id or candidate version is invalid")
	}
	byID := make(map[string]LineageNode, len(catalog.Nodes))
	children := make(map[string][]string, len(catalog.Nodes))
	queue := []string{}
	knownVersions := map[string]bool{}
	for _, node := range catalog.Nodes {
		byID[node.ID] = node
		for _, parent := range node.Parents {
			children[parent] = append(children[parent], node.ID)
		}
		if node.Kind == "source" && node.SourceID == sourceID {
			queue = append(queue, node.ID)
			if node.SourceVersion != "" {
				knownVersions[node.SourceVersion] = true
			}
		}
	}
	if len(queue) == 0 {
		return SourceImpactResult{SourceID: sourceID, CandidateVersion: candidateVersion, KnownVersions: []string{}, ImpactedNodes: []LineageNode{}}, nil
	}
	seen := map[string]bool{}
	impacted := []LineageNode{}
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		if seen[id] {
			continue
		}
		seen[id] = true
		if node, found := byID[id]; found {
			impacted = append(impacted, node)
		}
		queue = append(queue, children[id]...)
	}
	versions := make([]string, 0, len(knownVersions))
	for version := range knownVersions {
		versions = append(versions, version)
	}
	sort.Strings(versions)
	sort.Slice(impacted, func(left, right int) bool { return impacted[left].ID < impacted[right].ID })
	return SourceImpactResult{SourceID: sourceID, CandidateVersion: candidateVersion, KnownVersions: versions, ImpactedNodes: impacted}, nil
}

func LoadLineageCatalog(path string) (LineageCatalog, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return LineageCatalog{}, err
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	var catalog LineageCatalog
	if err := decoder.Decode(&catalog); err != nil {
		return LineageCatalog{}, fmt.Errorf("decode lineage catalog: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return LineageCatalog{}, fmt.Errorf("decode lineage catalog: multiple JSON values are not allowed")
		}
		return LineageCatalog{}, err
	}
	if catalog.SchemaVersion != lineageCatalogSchemaVersion {
		return LineageCatalog{}, fmt.Errorf("unsupported lineage catalog schema")
	}
	return catalog, nil
}

func validateSourceAssessmentInput(input SourceAssessmentInput) error {
	if !componentIDPattern.MatchString(strings.TrimSpace(input.SourceID)) || !sourceVersionPattern.MatchString(strings.TrimSpace(input.Version)) {
		return fmt.Errorf("source id or version is invalid")
	}
	switch strings.ToLower(strings.TrimSpace(input.Decision)) {
	case "adopted", "rejected", "deferred":
	default:
		return fmt.Errorf("source decision must be adopted, rejected, or deferred")
	}
	if value := strings.TrimSpace(input.Reason); value == "" || len(value) > 512 || strings.ContainsAny(value, "\r\n\t") {
		return fmt.Errorf("source assessment reason is invalid")
	}
	for _, pattern := range knowledgeSecretPatterns {
		if pattern.MatchString(input.Reason) {
			return fmt.Errorf("source assessment reason must not contain a secret")
		}
	}
	if len(input.EvidenceHandles) > 16 || len(input.ReanalyzeWhen) > 16 {
		return fmt.Errorf("source assessment has too many evidence or reanalysis entries")
	}
	for _, handle := range input.EvidenceHandles {
		if err := validateBoundedHandle(handle); err != nil {
			return err
		}
	}
	for _, condition := range input.ReanalyzeWhen {
		if err := validateSourceReanalysisCondition(condition); err != nil {
			return err
		}
	}
	return nil
}

func normalizedSourceReanalysis(values []string) []string {
	if len(values) == 0 {
		return []string{"explicit-user-request", "new-version"}
	}
	return normalizedGraphList(values)
}

func validateSourceReanalysisCondition(value string) error {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > 128 || strings.ContainsAny(value, "\r\n\t") {
		return fmt.Errorf("source reanalysis condition is invalid")
	}
	return nil
}

func loadSourceAssessmentStore(store string) (SourceAssessmentStore, error) {
	path := sourceAssessmentPath(store)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return SourceAssessmentStore{SchemaVersion: sourceAssessmentSchemaVersion, Assessments: []SourceAssessment{}}, nil
	}
	if err != nil {
		return SourceAssessmentStore{}, err
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	var ledger SourceAssessmentStore
	if err := decoder.Decode(&ledger); err != nil {
		return SourceAssessmentStore{}, fmt.Errorf("decode source assessment store: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return SourceAssessmentStore{}, fmt.Errorf("decode source assessment store: multiple JSON values are not allowed")
		}
		return SourceAssessmentStore{}, err
	}
	if ledger.SchemaVersion != sourceAssessmentSchemaVersion || len(ledger.Assessments) > 1024 {
		return SourceAssessmentStore{}, fmt.Errorf("source assessment store schema or capacity is invalid")
	}
	seen := map[string]bool{}
	for _, assessment := range ledger.Assessments {
		if err := validateStoredSourceAssessment(assessment); err != nil {
			return SourceAssessmentStore{}, err
		}
		key := assessment.SourceID + "\x00" + assessment.Version
		if seen[key] {
			return SourceAssessmentStore{}, fmt.Errorf("duplicate source assessment")
		}
		seen[key] = true
	}
	return ledger, nil
}

func writeSourceAssessmentStore(store string, ledger SourceAssessmentStore) error {
	data, err := json.MarshalIndent(ledger, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	path := sourceAssessmentPath(store)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return atomicWriteFile(path, data, 0o600)
}

func validateStoredSourceAssessment(assessment SourceAssessment) error {
	if err := validateSourceAssessmentInput(SourceAssessmentInput{SourceID: assessment.SourceID, Version: assessment.Version, Decision: assessment.Decision, Reason: assessment.Reason, EvidenceHandles: assessment.EvidenceHandles, ReanalyzeWhen: assessment.ReanalyzeWhen}); err != nil {
		return err
	}
	if _, err := time.Parse(time.RFC3339, assessment.RecordedAt); err != nil || assessment.RecordSHA256 != sourceAssessmentDigest(assessment) {
		return fmt.Errorf("source assessment integrity is invalid")
	}
	return nil
}

func sourceAssessmentDigest(assessment SourceAssessment) string {
	assessment.RecordSHA256 = ""
	return lineageDigest(assessment)
}

func sourceAssessmentPath(store string) string {
	return filepath.Join(filepath.Clean(store), "v1", "assessments.json")
}
