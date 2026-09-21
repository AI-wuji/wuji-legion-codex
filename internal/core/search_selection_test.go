package core

import (
	"strings"
	"testing"
)

func TestSelectSearchCandidatesDeduplicatesConservativelyInOriginalOrder(t *testing.T) {
	input := SearchSelectionInput{Query: "Go URL", Candidates: []SearchCandidate{
		{URL: "https://EXAMPLE.com/a?x=1&utm_source=news#section", Title: "first"},
		{URL: "https://example.com/a?x=1#section", Title: "duplicate"},
		{URL: "https://example.com/a?x=2", Title: "meaningful query"},
		{URL: "https://example.com/A?x=1", Title: "meaningful path"},
		{URL: "https://example.com/a?x=1#other-section", Title: "meaningful fragment"},
		{URL: "https://user:pass@example.com/a", Title: "credentials rejected"},
		{URL: "ftp://example.com/a", Title: "invalid"},
	}}
	got, err := SelectSearchCandidates(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Candidates) != 4 || got.Candidates[0].Title != "first" || got.Candidates[1].Title != "meaningful query" || got.Candidates[2].Title != "meaningful path" || got.Candidates[3].Title != "meaningful fragment" {
		t.Fatalf("selection lost stable rank or distinctions: %#v", got.Candidates)
	}
	if got.DuplicateCandidates != 1 || got.DroppedDuplicateObservations != 1 || got.InvalidCandidates != 2 || got.Incomplete {
		t.Fatalf("unexpected reporting: %#v", got)
	}
}

func TestSelectSearchCandidatesReportsPartialAndBounds(t *testing.T) {
	candidates := make([]SearchCandidate, SearchSelectionMaxCandidates+2)
	for i := range candidates {
		candidates[i] = SearchCandidate{URL: "https://example.com/" + string(rune('a'+i%26)) + "?n=" + string(rune('0'+i%10)), Title: "candidate"}
	}
	candidates[0] = SearchCandidate{URL: "https://example.com/", Title: strings.Repeat("x", searchSelectionMaxTitleBytes+1)}
	got, err := SelectSearchCandidates(SearchSelectionInput{Query: "evidence", Candidates: candidates, SourceErrors: []SearchSourceError{{Source: "community", Error: "timeout"}}})
	if err != nil {
		t.Fatal(err)
	}
	if !got.Incomplete || got.InvalidCandidates != 1 || got.OmittedCandidates == 0 || len(got.Candidates) > SearchSelectionMaxResults || len(got.SourceErrors) != 1 {
		t.Fatalf("partial or bounds state was not preserved: %#v", got)
	}
	if !strings.Contains(got.Interpretation, "does not establish") {
		t.Fatalf("missing non-claim interpretation: %q", got.Interpretation)
	}
}

func TestSelectSearchCandidatesAcceptsChineseAndRejectsOversizeQuery(t *testing.T) {
	got, err := SelectSearchCandidates(SearchSelectionInput{Query: "中文证据", Candidates: []SearchCandidate{{URL: "https://例子.测试/路径?主题=研究", Title: "中文来源", Snippet: "保留原始结果"}}})
	if err != nil || len(got.Candidates) != 1 || got.Candidates[0].Title != "中文来源" {
		t.Fatalf("Chinese candidate failed: result=%#v err=%v", got, err)
	}
	_, err = SelectSearchCandidates(SearchSelectionInput{Query: strings.Repeat("x", SearchSelectionMaxQueryBytes+1)})
	if err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("oversize query was accepted: %v", err)
	}
}

func TestNormalizeSearchURLPreservesIPv6AndHashRoutes(t *testing.T) {
	first, err := normalizeSearchURL("https://[2001:DB8::1]/#/evidence?utm_source=x")
	if err != nil || first != "https://[2001:db8::1]/#/evidence?utm_source=x" {
		t.Fatalf("IPv6 or fragment normalization changed evidence identity: %q %v", first, err)
	}
	second, err := normalizeSearchURL("https://[2001:db8::1]/#/other")
	if err != nil || first == second {
		t.Fatalf("hash routes were collapsed: first=%q second=%q err=%v", first, second, err)
	}
}

func TestSelectSearchCandidatesBoundsSourceErrorsAndCandidateWork(t *testing.T) {
	errors := make([]SearchSourceError, searchSelectionMaxSourceErrors+1)
	for i := range errors {
		errors[i] = SearchSourceError{Source: "source", Error: "unavailable"}
	}
	if _, err := SelectSearchCandidates(SearchSelectionInput{Query: "evidence", SourceErrors: errors}); err == nil || !strings.Contains(err.Error(), "exceed") {
		t.Fatalf("unbounded source errors were accepted: %v", err)
	}
	candidates := make([]SearchCandidate, SearchSelectionMaxCandidates+1)
	for i := range candidates {
		candidates[i] = SearchCandidate{URL: "https://example.com/" + string(rune('a'+i%26)), Title: "candidate"}
	}
	candidates[SearchSelectionMaxCandidates] = SearchCandidate{URL: "not a URL", Title: "must not be inspected"}
	got, err := SelectSearchCandidates(SearchSelectionInput{Query: "evidence", Candidates: candidates})
	if err != nil || got.InvalidCandidates != 0 || got.OmittedCandidates == 0 {
		t.Fatalf("candidate tail was inspected or not reported as omitted: result=%#v err=%v", got, err)
	}
}
