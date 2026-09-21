package core

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

const (
	SearchSelectionMaxCandidates       = 128
	SearchSelectionMaxResults          = 10
	SearchSelectionMaxQueryBytes       = 4096
	searchSelectionMaxSourceErrors     = 16
	searchSelectionMaxURLBytes         = 4096
	searchSelectionMaxTitleBytes       = 2048
	searchSelectionMaxSnippetBytes     = 8192
	searchSelectionMaxProvenanceBytes  = 2048
	searchSelectionMaxSourceErrorBytes = 1024
)

type SearchCandidate struct {
	URL        string `json:"url"`
	Title      string `json:"title"`
	Snippet    string `json:"snippet"`
	Provenance string `json:"provenance,omitempty"`
}

type SearchSourceError struct {
	Source string `json:"source"`
	Error  string `json:"error"`
}

type SearchSelectionInput struct {
	Query        string              `json:"query"`
	Candidates   []SearchCandidate   `json:"candidates"`
	Incomplete   bool                `json:"incomplete"`
	SourceErrors []SearchSourceError `json:"source_errors,omitempty"`
}

type SearchSelectionResult struct {
	Query                        string              `json:"query"`
	Candidates                   []SearchCandidate   `json:"candidates"`
	Incomplete                   bool                `json:"incomplete"`
	SourceErrors                 []SearchSourceError `json:"source_errors,omitempty"`
	OmittedCandidates            int                 `json:"omitted_candidates"`
	InvalidCandidates            int                 `json:"invalid_candidates"`
	DuplicateCandidates          int                 `json:"duplicate_candidates"`
	DroppedDuplicateObservations int                 `json:"dropped_duplicate_observations"`
	Interpretation               string              `json:"interpretation"`
}

// SelectSearchCandidates preserves supplied rank. It only removes invalid or
// duplicate HTTP(S) URLs; it does not score, fetch, or verify search results.
func SelectSearchCandidates(input SearchSelectionInput) (SearchSelectionResult, error) {
	if strings.TrimSpace(input.Query) == "" {
		return SearchSelectionResult{}, fmt.Errorf("search query is required")
	}
	if len([]byte(input.Query)) > SearchSelectionMaxQueryBytes {
		return SearchSelectionResult{}, fmt.Errorf("search query exceeds %d bytes", SearchSelectionMaxQueryBytes)
	}
	if err := validateSearchSourceErrors(input.SourceErrors); err != nil {
		return SearchSelectionResult{}, err
	}
	result := SearchSelectionResult{
		Query: input.Query, Incomplete: input.Incomplete || len(input.SourceErrors) > 0 || len(input.Candidates) > SearchSelectionMaxCandidates,
		SourceErrors:   append([]SearchSourceError(nil), input.SourceErrors...),
		Candidates:     make([]SearchCandidate, 0, SearchSelectionMaxResults),
		Interpretation: "Selected URLs are unverified data; an empty or partial selection does not establish that no relevant evidence exists or that any claim is true.",
	}
	if len(input.Candidates) > SearchSelectionMaxCandidates {
		result.OmittedCandidates = len(input.Candidates) - SearchSelectionMaxCandidates
		input.Candidates = input.Candidates[:SearchSelectionMaxCandidates]
	}
	seen := make(map[string]struct{}, SearchSelectionMaxCandidates)
	for _, candidate := range input.Candidates {
		normalized, err := normalizeSearchURL(candidate.URL)
		if err != nil || !validSearchCandidate(candidate) {
			result.InvalidCandidates++
			continue
		}
		if _, ok := seen[normalized]; ok {
			result.DuplicateCandidates++
			result.DroppedDuplicateObservations++
			continue
		}
		seen[normalized] = struct{}{}
		if len(result.Candidates) >= SearchSelectionMaxResults {
			result.OmittedCandidates++
			continue
		}
		result.Candidates = append(result.Candidates, candidate)
	}
	return result, nil
}

func validSearchCandidate(candidate SearchCandidate) bool {
	return len([]byte(candidate.URL)) <= searchSelectionMaxURLBytes &&
		len([]byte(candidate.Title)) <= searchSelectionMaxTitleBytes &&
		len([]byte(candidate.Snippet)) <= searchSelectionMaxSnippetBytes &&
		len([]byte(candidate.Provenance)) <= searchSelectionMaxProvenanceBytes
}

func validateSearchSourceErrors(errors []SearchSourceError) error {
	if len(errors) > searchSelectionMaxSourceErrors {
		return fmt.Errorf("source errors exceed %d entries", searchSelectionMaxSourceErrors)
	}
	for _, item := range errors {
		if strings.TrimSpace(item.Source) == "" || strings.TrimSpace(item.Error) == "" {
			return fmt.Errorf("source errors require source and error")
		}
		if len([]byte(item.Source)) > searchSelectionMaxSourceErrorBytes || len([]byte(item.Error)) > searchSelectionMaxSourceErrorBytes {
			return fmt.Errorf("source error field exceeds %d bytes", searchSelectionMaxSourceErrorBytes)
		}
	}
	return nil
}

func normalizeSearchURL(raw string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" || u.User != nil {
		return "", fmt.Errorf("invalid HTTP(S) URL")
	}
	u.Scheme = strings.ToLower(u.Scheme)
	host := strings.ToLower(u.Hostname())
	if port := u.Port(); port != "" {
		host = net.JoinHostPort(host, port)
	} else if strings.Contains(host, ":") {
		host = "[" + host + "]"
	}
	u.Host = host
	u.RawQuery = removeTrackingSearchParams(u.RawQuery)
	return u.String(), nil
}

func removeTrackingSearchParams(rawQuery string) string {
	if rawQuery == "" {
		return ""
	}
	kept := make([]string, 0, strings.Count(rawQuery, "&")+1)
	for _, part := range strings.Split(rawQuery, "&") {
		name := strings.SplitN(part, "=", 2)[0]
		decoded, err := url.QueryUnescape(name)
		if err != nil || !isTrackingSearchParam(decoded) {
			kept = append(kept, part)
		}
	}
	return strings.Join(kept, "&")
}

func isTrackingSearchParam(name string) bool {
	name = strings.ToLower(name)
	if strings.HasPrefix(name, "utm_") {
		return true
	}
	switch name {
	case "gclid", "fbclid", "msclkid", "dclid", "mc_cid", "mc_eid":
		return true
	default:
		return false
	}
}
