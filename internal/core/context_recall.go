package core

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
)

const (
	ContextRecallMaxMatches    = 20 // maximum matched records or atomic tool-pair groups
	ContextRecallMaxQueryBytes = 4096
)

// ContextRecallMatch is an original archived record selected by a bounded,
// deterministic lexical query. It contains no generated replacement text.
type ContextRecallMatch struct {
	Record        ContextReductionRecord `json:"record"`
	Score         int                    `json:"score"`
	MatchedFields []string               `json:"matched_fields,omitempty"`
}

type ContextRecallResult struct {
	Query          string               `json:"query"`
	Matches        []ContextRecallMatch `json:"matches"`
	Examined       int                  `json:"examined"`
	ArchivedOnly   bool                 `json:"archived_only"`
	Interpretation string               `json:"interpretation"`
}

type contextRecallGroup struct {
	key     string
	score   int
	records []ContextRecallMatch
	order   int
}

// RecallArchivedContext validates a projection, then returns only original
// archived records that match the query. This is a cold-store lookup, not a
// semantic judgment and not a substitute for the visible transcript.
func RecallArchivedContext(projection ContextReductionResult, query string, limit int) (ContextRecallResult, error) {
	if strings.TrimSpace(query) == "" {
		return ContextRecallResult{}, fmt.Errorf("context recall query is required")
	}
	if len([]byte(query)) > ContextRecallMaxQueryBytes {
		return ContextRecallResult{}, fmt.Errorf("context recall query exceeds %d bytes", ContextRecallMaxQueryBytes)
	}
	if limit <= 0 || limit > ContextRecallMaxMatches {
		return ContextRecallResult{}, fmt.Errorf("context recall limit must be between 1 and %d", ContextRecallMaxMatches)
	}
	if _, err := RestoreContext(projection); err != nil {
		return ContextRecallResult{}, err
	}
	terms := contextRecallTerms(query)
	if len(terms) == 0 {
		return ContextRecallResult{}, fmt.Errorf("context recall query contains no searchable terms")
	}
	groupsByKey := make(map[string]*contextRecallGroup, len(projection.Archived))
	for index, record := range projection.Archived {
		score, fields := contextRecallScore(record, terms)
		if score == 0 {
			continue
		}
		key := "record:" + record.ID
		if record.PairID != "" {
			key = "pair:" + record.PairID
		}
		group := groupsByKey[key]
		if group == nil {
			group = &contextRecallGroup{key: key, order: index}
			groupsByKey[key] = group
		}
		if score > group.score {
			group.score = score
		}
		group.records = append(group.records, ContextRecallMatch{
			Record:        record,
			Score:         score,
			MatchedFields: fields,
		})
	}
	// A recalled tool result is only safe to rehydrate with its paired call.
	// Group selection therefore counts pairs, not individual records.
	for _, record := range projection.Archived {
		if record.PairID == "" {
			continue
		}
		key := "pair:" + record.PairID
		group := groupsByKey[key]
		if group == nil {
			continue
		}
		found := false
		for _, match := range group.records {
			if match.Record.ID == record.ID {
				found = true
				break
			}
		}
		if !found {
			group.records = append(group.records, ContextRecallMatch{
				Record:        record,
				Score:         group.score,
				MatchedFields: []string{"paired"},
			})
		}
	}
	groups := make([]*contextRecallGroup, 0, len(groupsByKey))
	for _, group := range groupsByKey {
		groups = append(groups, group)
	}
	sort.SliceStable(groups, func(i, j int) bool {
		if groups[i].score != groups[j].score {
			return groups[i].score > groups[j].score
		}
		return groups[i].order < groups[j].order
	})
	if len(groups) > limit {
		groups = groups[:limit]
	}
	selected := make(map[string]bool, len(groups))
	for _, group := range groups {
		selected[group.key] = true
	}
	matches := make([]ContextRecallMatch, 0, len(groups))
	for _, record := range projection.Archived {
		key := "record:" + record.ID
		if record.PairID != "" {
			key = "pair:" + record.PairID
		}
		if !selected[key] {
			continue
		}
		group := groupsByKey[key]
		for _, match := range group.records {
			if match.Record.ID == record.ID {
				matches = append(matches, match)
				break
			}
		}
	}
	return ContextRecallResult{
		Query:          query,
		Matches:        matches,
		Examined:       len(projection.Archived),
		ArchivedOnly:   true,
		Interpretation: "Deterministic lexical recall of original archived records. Matches are evidence for inspection, not a completion claim or a semantic ranking.",
	}, nil
}

func contextRecallTerms(query string) []string {
	var terms []string
	var current []rune
	flush := func() {
		if len(current) == 0 {
			return
		}
		terms = append(terms, strings.ToLower(string(current)))
		current = current[:0]
	}
	for _, r := range query {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' || r == '.' || r == '/' {
			current = append(current, r)
			continue
		}
		flush()
	}
	flush()
	return terms
}

func contextRecallScore(record ContextReductionRecord, terms []string) (int, []string) {
	fields := []struct {
		name   string
		value  string
		weight int
	}{
		{name: "id", value: record.ID, weight: 5},
		{name: "tool", value: record.Tool, weight: 4},
		{name: "operation", value: record.Operation, weight: 3},
		{name: "footprint", value: record.Footprint, weight: 6},
		{name: "text", value: record.Text, weight: 1},
	}
	score := 0
	matched := make([]string, 0, len(fields))
	for _, field := range fields {
		value := strings.ToLower(field.value)
		fieldScore := 0
		for _, term := range terms {
			if strings.Contains(value, term) {
				fieldScore += field.weight
			}
		}
		if fieldScore > 0 {
			score += fieldScore
			matched = append(matched, field.name)
		}
	}
	return score, matched
}
