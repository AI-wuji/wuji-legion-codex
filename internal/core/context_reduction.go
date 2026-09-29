package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

const (
	contextReductionSchemaVersion = 2
	contextReductionMaxRecords    = 128
	contextReductionMaxTextBytes  = 64 * 1024
)

// ContextReductionRecord is a caller-supplied tool observation. Footprint must
// identify the exact viewed resource, including a line range when applicable.
type ContextReductionRecord struct {
	ID        string `json:"id"`
	Role      string `json:"role"`
	Kind      string `json:"kind"`
	Tool      string `json:"tool,omitempty"`
	Operation string `json:"operation,omitempty"`
	Footprint string `json:"footprint,omitempty"`
	PairID    string `json:"pair_id,omitempty"`
	Text      string `json:"text,omitempty"`
	Succeeded bool   `json:"succeeded,omitempty"`
	IsError   bool   `json:"is_error,omitempty"`
	Protected bool   `json:"protected,omitempty"`
	Critical  bool   `json:"critical,omitempty"`
}

type ContextReductionInput struct {
	Records           []ContextReductionRecord `json:"records"`
	PreserveRecent    int                      `json:"preserve_recent,omitempty"`
	ArchiveDuplicates bool                     `json:"archive_duplicates,omitempty"`
	ArchiveStale      bool                     `json:"archive_stale,omitempty"`
}

type ContextReductionDecision struct {
	RecordID string `json:"record_id"`
	Action   string `json:"action"`
	Reason   string `json:"reason"`
	Witness  string `json:"witness,omitempty"`
}

type ContextReductionResult struct {
	SchemaVersion       int                        `json:"schema_version"`
	OriginalSHA256      string                     `json:"original_sha256"`
	PreserveRecent      int                        `json:"preserve_recent"`
	ArchiveDuplicates   bool                       `json:"archive_duplicates"`
	ArchiveStale        bool                       `json:"archive_stale"`
	Visible             []ContextReductionRecord   `json:"visible"`
	Archived            []ContextReductionRecord   `json:"archived"`
	Decisions           []ContextReductionDecision `json:"decisions"`
	ArchivedCount       int                        `json:"archived_count"`
	ProtectedCount      int                        `json:"protected_count"`
	OriginalRecordBytes int                        `json:"original_record_bytes"`
	VisibleRecordBytes  int                        `json:"visible_record_bytes"`
	ArchivedRecordBytes int                        `json:"archived_record_bytes"`
	SavedRecordBytes    int                        `json:"saved_record_bytes"`
	Interpretation      string                     `json:"interpretation"`
}

// ReduceContext prepares a reversible, deterministic projection for a host to
// use. The full archive stays in the result; this command does not edit a host
// transcript or claim that a Laya/Jev score is safe deletion authority.
func ReduceContext(input ContextReductionInput) (ContextReductionResult, error) {
	if err := validateContextReductionInput(input); err != nil {
		return ContextReductionResult{}, err
	}
	records := input.Records
	digest := contextReductionDigest(records)
	decisions := make([]ContextReductionDecision, len(records))
	protected := make([]bool, len(records))
	archived := make([]bool, len(records))
	for index, record := range records {
		protected[index] = record.Protected || record.Critical || record.IsError ||
			record.Kind == "message" || (input.PreserveRecent > 0 && index >= len(records)-input.PreserveRecent)
		reason := "default"
		if protected[index] {
			reason = "protected"
		}
		decisions[index] = ContextReductionDecision{RecordID: record.ID, Action: "keep", Reason: reason}
	}

	// Protection is inherited by both halves of a tool call/result pair.
	for index, record := range records {
		if !protected[index] || record.PairID == "" {
			continue
		}
		for other := range records {
			if records[other].PairID == record.PairID {
				protected[other] = true
				decisions[other].Reason = "protected-pair"
			}
		}
	}

	for index, record := range records {
		if protected[index] || record.Kind != "tool_result" || record.Operation != "view" ||
			!record.Succeeded || record.Footprint == "" || record.Tool == "" {
			continue
		}
		for later := index + 1; later < len(records); later++ {
			next := records[later]
			if next.Kind != "tool_result" || !next.Succeeded || next.IsError || next.Footprint != record.Footprint {
				continue
			}
			reason := ""
			if input.ArchiveDuplicates && next.Operation == "view" &&
				strings.EqualFold(next.Tool, record.Tool) && next.Text == record.Text {
				reason = "identical-observation"
			} else if input.ArchiveStale && next.Operation == "write" {
				reason = "stale-after-write"
			}
			if reason == "" {
				continue
			}
			for other := range records {
				if (other == index || (record.PairID != "" && records[other].PairID == record.PairID)) && protected[other] {
					reason = ""
					break
				}
			}
			if reason == "" {
				break
			}
			for other := range records {
				if other == index || (record.PairID != "" && records[other].PairID == record.PairID) {
					archived[other] = true
					decisions[other] = ContextReductionDecision{
						RecordID: records[other].ID, Action: "archive", Reason: reason, Witness: next.ID,
					}
				}
			}
			break
		}
	}

	result := ContextReductionResult{
		SchemaVersion:     contextReductionSchemaVersion,
		OriginalSHA256:    digest,
		PreserveRecent:    input.PreserveRecent,
		ArchiveDuplicates: input.ArchiveDuplicates,
		ArchiveStale:      input.ArchiveStale,
		Visible:           make([]ContextReductionRecord, 0, len(records)),
		Archived:          make([]ContextReductionRecord, 0),
		Decisions:         decisions,
		Interpretation:    "Projection only: use visible for a bounded handoff, retain archived separately, and restore by hash when needed. No host transcript is edited or completion proven.",
	}
	for index, record := range records {
		if protected[index] {
			result.ProtectedCount++
		}
		if archived[index] {
			result.Archived = append(result.Archived, record)
		} else {
			result.Visible = append(result.Visible, record)
		}
	}
	result.ArchivedCount = len(result.Archived)
	result.OriginalRecordBytes = contextReductionRecordsBytes(records)
	result.VisibleRecordBytes = contextReductionRecordsBytes(result.Visible)
	result.ArchivedRecordBytes = contextReductionRecordsBytes(result.Archived)
	result.SavedRecordBytes = result.OriginalRecordBytes - result.VisibleRecordBytes
	return result, nil
}

// RestoreContext independently checks that a projection reconstructs the
// original ordered records, including archived tool results.
func RestoreContext(result ContextReductionResult) ([]ContextReductionRecord, error) {
	if result.SchemaVersion != contextReductionSchemaVersion ||
		result.ArchivedCount != len(result.Archived) ||
		len(result.Decisions) != len(result.Visible)+len(result.Archived) {
		return nil, fmt.Errorf("context reduction projection is inconsistent")
	}
	visible := make(map[string]ContextReductionRecord, len(result.Visible))
	archived := make(map[string]ContextReductionRecord, len(result.Archived))
	for _, record := range result.Visible {
		if _, exists := visible[record.ID]; exists {
			return nil, fmt.Errorf("duplicate visible record")
		}
		visible[record.ID] = record
	}
	for _, record := range result.Archived {
		if _, exists := archived[record.ID]; exists {
			return nil, fmt.Errorf("duplicate archived record")
		}
		archived[record.ID] = record
	}
	records := make([]ContextReductionRecord, 0, len(result.Decisions))
	for _, decision := range result.Decisions {
		switch decision.Action {
		case "keep":
			record, exists := visible[decision.RecordID]
			if !exists {
				return nil, fmt.Errorf("missing visible record %q", decision.RecordID)
			}
			records = append(records, record)
			delete(visible, decision.RecordID)
		case "archive":
			record, exists := archived[decision.RecordID]
			if !exists {
				return nil, fmt.Errorf("missing archived record %q", decision.RecordID)
			}
			records = append(records, record)
			delete(archived, decision.RecordID)
		default:
			return nil, fmt.Errorf("invalid context decision action")
		}
	}
	if len(visible) != 0 || len(archived) != 0 ||
		validateContextReductionInput(ContextReductionInput{Records: records}) != nil ||
		contextReductionDigest(records) != result.OriginalSHA256 {
		return nil, fmt.Errorf("context reduction integrity mismatch")
	}
	expected, err := ReduceContext(ContextReductionInput{
		Records: records, PreserveRecent: result.PreserveRecent,
		ArchiveDuplicates: result.ArchiveDuplicates, ArchiveStale: result.ArchiveStale,
	})
	if err != nil || !reflect.DeepEqual(expected.Decisions, result.Decisions) ||
		!reflect.DeepEqual(expected.Visible, result.Visible) ||
		!reflect.DeepEqual(expected.Archived, result.Archived) ||
		expected.ProtectedCount != result.ProtectedCount ||
		expected.OriginalRecordBytes != result.OriginalRecordBytes ||
		expected.VisibleRecordBytes != result.VisibleRecordBytes ||
		expected.ArchivedRecordBytes != result.ArchivedRecordBytes ||
		expected.SavedRecordBytes != result.SavedRecordBytes ||
		expected.Interpretation != result.Interpretation {
		return nil, fmt.Errorf("context reduction decisions do not match deterministic policy")
	}
	return records, nil
}

func contextReductionRecordsBytes(records []ContextReductionRecord) int {
	data, _ := json.Marshal(records)
	return len(data)
}

func contextReductionDigest(records []ContextReductionRecord) string {
	data, _ := json.Marshal(records)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func validateContextReductionInput(input ContextReductionInput) error {
	if len(input.Records) == 0 || len(input.Records) > contextReductionMaxRecords {
		return fmt.Errorf("context reduction requires 1-%d records", contextReductionMaxRecords)
	}
	if input.PreserveRecent < 0 || input.PreserveRecent > len(input.Records) {
		return fmt.Errorf("preserve_recent is outside the record range")
	}
	seen := make(map[string]bool, len(input.Records))
	pairs := make(map[string]map[string]bool)
	for _, record := range input.Records {
		if strings.TrimSpace(record.ID) == "" || len(record.ID) > 256 || seen[record.ID] {
			return fmt.Errorf("context record id is missing, duplicate, or oversized")
		}
		seen[record.ID] = true
		if len(record.Text) > contextReductionMaxTextBytes || len(record.Footprint) > 4096 ||
			len(record.PairID) > 256 || len(record.Tool) > 256 {
			return fmt.Errorf("context record %q exceeds a field limit", record.ID)
		}
		switch record.Kind {
		case "message":
			if record.Role != "user" && record.Role != "assistant" && record.Role != "system" {
				return fmt.Errorf("context message %q has invalid role", record.ID)
			}
			if record.PairID != "" {
				return fmt.Errorf("context message %q cannot be paired", record.ID)
			}
		case "tool_call", "tool_result":
			if record.Role != "assistant" && record.Role != "tool" {
				return fmt.Errorf("context tool record %q has invalid role", record.ID)
			}
			if record.Tool == "" || record.Operation == "" {
				return fmt.Errorf("context tool record %q lacks tool or operation", record.ID)
			}
			if record.Operation != "view" && record.Operation != "write" && record.Operation != "other" {
				return fmt.Errorf("context record %q has unsupported operation", record.ID)
			}
			if record.PairID != "" {
				if pairs[record.PairID] == nil {
					pairs[record.PairID] = make(map[string]bool)
				}
				if pairs[record.PairID][record.Kind] {
					return fmt.Errorf("context pair %q has duplicate %s", record.PairID, record.Kind)
				}
				pairs[record.PairID][record.Kind] = true
			}
		default:
			return fmt.Errorf("context record %q has invalid kind", record.ID)
		}
	}
	return nil
}
