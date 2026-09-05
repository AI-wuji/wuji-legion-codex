package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type GraphRetentionPolicy struct {
	Graph            string `json:"graph"`
	TTLDays          int    `json:"ttl_days"`
	ArchiveAfterDays int    `json:"archive_after_days"`
	GCMode           string `json:"gc_mode"`
	Locking          string `json:"locking"`
	Repair           string `json:"repair"`
	MaxRecords       int    `json:"max_records"`
}

type GraphMaintenanceResult struct {
	Policy          GraphRetentionPolicy `json:"policy"`
	Validated       bool                 `json:"validated"`
	ArchivedRecords int                  `json:"archived_records"`
	GCRecords       int                  `json:"gc_records"`
}

func GraphRetentionPolicies() []GraphRetentionPolicy {
	return []GraphRetentionPolicy{
		{Graph: "acceptance", TTLDays: 365, ArchiveAfterDays: 365, GCMode: "archive-only", Locking: "exclusive-file-lock", Repair: "schema-and-hash-validation", MaxRecords: 512},
		{Graph: "conversation-evidence", TTLDays: 30, ArchiveAfterDays: 30, GCMode: "archive-then-prune", Locking: "exclusive-file-lock", Repair: "schema-and-hash-validation", MaxRecords: 512},
		{Graph: "provenance", TTLDays: 365, ArchiveAfterDays: 365, GCMode: "archive-only", Locking: "exclusive-file-lock", Repair: "schema-and-hash-validation", MaxRecords: 2048},
		{Graph: "source-assessments", TTLDays: 365, ArchiveAfterDays: 365, GCMode: "archive-only", Locking: "exclusive-file-lock", Repair: "schema-and-hash-validation", MaxRecords: 1024},
	}
}

func MaintainGraph(store, graph string, now time.Time) (GraphMaintenanceResult, error) {
	policy, ok := graphRetentionPolicy(graph)
	if !ok {
		return GraphMaintenanceResult{}, fmt.Errorf("graph %q has no governed retention policy", graph)
	}
	result := GraphMaintenanceResult{Policy: policy}
	err := withKnowledgeStoreLock(store, func() error {
		switch policy.Graph {
		case "acceptance":
			if _, err := loadAcceptanceLedger(store); err != nil {
				return err
			}
		case "conversation-evidence":
			index, err := loadConversationEvidenceIndex(store)
			if err != nil {
				return err
			}
			cutoff := now.UTC().AddDate(0, 0, -policy.TTLDays)
			active := make([]ConversationEvidenceRecord, 0, len(index.Records))
			for _, record := range index.Records {
				recordedAt, _ := time.Parse(time.RFC3339, record.RecordedAt)
				if recordedAt.After(cutoff) {
					active = append(active, record)
					continue
				}
				if err := archiveConversationEvidence(store, record); err != nil {
					return err
				}
				result.ArchivedRecords++
				result.GCRecords++
			}
			if result.GCRecords > 0 {
				index.Records = active
				if err := writeConversationEvidenceIndex(store, index); err != nil {
					return err
				}
			}
		case "provenance":
			if _, err := loadProvenanceIndex(store); err != nil {
				return err
			}
		case "source-assessments":
			if _, err := loadSourceAssessmentStore(store); err != nil {
				return err
			}
		}
		result.Validated = true
		return AuditEventRecord(auditStoreFor(store), AuditEvent{EventType: "graph-maintained", Actor: "aji", Authority: "deterministic-governance", Target: policy.Graph, ResultHandle: "wuji-graph-governance://" + policy.Graph})
	})
	if err != nil {
		return GraphMaintenanceResult{}, err
	}
	return result, nil
}

func graphRetentionPolicy(graph string) (GraphRetentionPolicy, bool) {
	graph = strings.TrimSpace(graph)
	for _, policy := range GraphRetentionPolicies() {
		if policy.Graph == graph {
			return policy, true
		}
	}
	return GraphRetentionPolicy{}, false
}

func archiveConversationEvidence(store string, record ConversationEvidenceRecord) error {
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	path := filepath.Join(filepath.Clean(store), "v1", "archive", record.RecordSHA256+".json")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	if existing, err := os.ReadFile(path); err == nil {
		if string(existing) == string(data) {
			return nil
		}
		return fmt.Errorf("conversation evidence archive collision")
	} else if !os.IsNotExist(err) {
		return err
	}
	return atomicWriteFile(path, data, 0o600)
}

func SortGraphRetentionPolicies(policies []GraphRetentionPolicy) {
	sort.Slice(policies, func(left, right int) bool { return policies[left].Graph < policies[right].Graph })
}
