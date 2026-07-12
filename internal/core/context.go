package core

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

type scoredFile struct {
	path         string
	score        int
	lines        []string
	hits         []int
	sourceSHA256 string
	pathHit      bool
}

func SelectContext(workspace, query string, budget int) (ContextResult, error) {
	if strings.TrimSpace(query) == "" {
		return ContextResult{}, fmt.Errorf("query is required")
	}
	if budget <= 0 {
		return ContextResult{}, fmt.Errorf("max-bytes must be greater than zero")
	}
	workspace, err := normalizeWorkspacePath(workspace)
	if err != nil {
		return ContextResult{}, err
	}
	fingerprintTerms := queryTerms(query)
	terms := retrievalTerms(query)
	if len(terms) == 0 {
		return ContextResult{}, fmt.Errorf("query contains no retrieval anchors")
	}
	files, scanned, graphStats, err := retrieveWorkspaceFiles(workspace, terms)
	if err != nil {
		return ContextResult{}, err
	}
	sort.Slice(files, func(i, j int) bool {
		if files[i].score == files[j].score {
			return files[i].path < files[j].path
		}
		return files[i].score > files[j].score
	})
	result := ContextResult{
		Workspace: workspace, Query: query, QueryFingerprint: queryFingerprint(fingerprintTerms), BudgetBytes: budget, ScannedFiles: scanned,
		RetrievalTerms: terms,
		IndexedFiles:   graphStats.IndexedFiles, CandidateFiles: graphStats.CandidateFiles, GraphLookups: graphStats.GraphLookups,
		RetrievalTruncated: graphStats.Truncated, GraphSourceBytes: graphStats.SourceBytes,
		RetrievalMode: graphStats.Mode, FallbackReason: graphStats.FallbackReason,
		Policy: []string{"route through workspace relation graph before reading files", "keep graph summaries and indexes out of model context", "rebuild stale derived indexes", "rank unique anchors before read", "exclude generated dependency locks", "emit matched ranges only", "hard byte budget", "deterministic prompt payload", "content-addressed handoff", "keep raw logs out of context"},
	}
	for _, file := range files {
		remaining := budget - len([]byte(renderContextPayload(result.Excerpts)))
		excerpt := makeExcerpt(file, remaining-256)
		if excerpt.Text == "" {
			continue
		}
		candidate := append(append([]ContextExcerpt(nil), result.Excerpts...), excerpt)
		payload := renderContextPayload(candidate)
		if len([]byte(payload)) > budget {
			continue
		}
		result.Excerpts = candidate
		result.SelectedBytes = len([]byte(payload))
		if len(result.Excerpts) >= 12 {
			break
		}
	}
	if len(result.Excerpts) == 0 {
		return ContextResult{}, fmt.Errorf("no matching context excerpts")
	}
	result.MatchedTerms, result.CoverageBPS, result.CodeExcerptCount, result.ContentAnchorCount = assessContextQuality(result.RetrievalTerms, result.Excerpts)
	payload := renderContextPayload(result.Excerpts)
	result.PayloadBytes = len([]byte(payload))
	result.PayloadSHA256 = sha256Hex([]byte(payload))
	result.SelectedBytes = result.PayloadBytes
	result.ContentSHA256, err = contextContentHash(result.QueryFingerprint, result.RetrievalTerms, result.Excerpts)
	if err != nil {
		return ContextResult{}, err
	}
	result.ContextHandle = contextHandle(result.ContentSHA256)
	return result, nil
}

func retrieveWorkspaceFiles(workspace string, terms []string) ([]scoredFile, int, workspaceGraphStats, error) {
	files, stats, err := queryWorkspaceGraph(workspace, terms)
	if err == nil && len(files) > 0 {
		return files, stats.CandidateFiles, stats, nil
	}
	if errors.Is(err, errWorkspaceGraphMissing) {
		reason := stats.FallbackReason
		if _, syncErr := SyncWorkspaceGraph(workspace); syncErr == nil {
			files, rebuiltStats, queryErr := queryWorkspaceGraph(workspace, terms)
			if queryErr == nil && len(files) > 0 {
				rebuiltStats.Mode = "workspace-graph-rebuilt"
				if reason != "" {
					rebuiltStats.FallbackReason = reason
				}
				return files, rebuiltStats.CandidateFiles, rebuiltStats, nil
			}
			stats = rebuiltStats
		}
		if reason == "" {
			reason = "graph-missing"
		}
		stats.FallbackReason = reason
	} else if err != nil {
		stats.FallbackReason = "graph-error"
	}
	files, scanned, scanErr := boundedWorkspaceScan(workspace, terms, 512)
	stats.Mode = "bounded-fallback"
	if stats.FallbackReason == "" {
		stats.FallbackReason = "graph-no-match"
	}
	return files, scanned, stats, scanErr
}

func boundedWorkspaceScan(workspace string, terms []string, limit int) ([]scoredFile, int, error) {
	files := []scoredFile{}
	scanned := 0
	err := filepath.WalkDir(workspace, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if path != workspace && excludedDir(entry.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if scanned >= limit {
			return filepath.SkipAll
		}
		info, statErr := entry.Info()
		if statErr != nil {
			return statErr
		}
		if info.Size() > 2*1024*1024 || !sourceLike(path) {
			return nil
		}
		scanned++
		item, scoreErr := scoreFile(workspace, path, terms)
		if scoreErr != nil {
			return scoreErr
		}
		if item.score > 0 {
			files = append(files, item)
		}
		return nil
	})
	return files, scanned, err
}

func scoreFile(root, path string, terms []string) (scoredFile, error) {
	rel, _ := filepath.Rel(root, path)
	item := scoredFile{path: filepath.ToSlash(rel)}
	pathLower := strings.ToLower(item.path)
	matched := map[string]bool{}
	for _, term := range terms {
		if strings.Contains(pathLower, term) {
			item.score += 100
			item.pathHit = true
			matched[term] = true
		}
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return item, err
	}
	sourceHash := sha256.Sum256(content)
	item.sourceSHA256 = hex.EncodeToString(sourceHash[:])
	scanner := bufio.NewScanner(bytes.NewReader(content))
	buffer := make([]byte, 64*1024)
	scanner.Buffer(buffer, 512*1024)
	lineNo := 0
	for scanner.Scan() && lineNo < 4000 {
		lineNo++
		line := scanner.Text()
		item.lines = append(item.lines, line)
		lower := strings.ToLower(line)
		hit := false
		for _, term := range terms {
			if strings.Contains(lower, term) {
				hit = true
				if !matched[term] {
					item.score += 15
					matched[term] = true
				}
			}
		}
		if hit && len(item.hits) < 12 {
			item.hits = append(item.hits, lineNo)
		}
	}
	if err := scanner.Err(); err != nil {
		return item, fmt.Errorf("scan %s: %w", item.path, err)
	}
	if len(item.hits) == 0 && item.pathHit && len(item.lines) > 0 {
		item.hits = []int{1}
	}
	item.score += minInt(len(item.hits), 6)
	return item, nil
}

func renderContextPayload(excerpts []ContextExcerpt) string {
	var b strings.Builder
	b.WriteString("WUJI_CONTEXT_CAPSULE_V1\n")
	for _, excerpt := range excerpts {
		b.WriteString("\nFILE ")
		b.WriteString(excerpt.Path)
		b.WriteString(" LINES ")
		b.WriteString(strings.Join(excerpt.LineRanges, ","))
		b.WriteByte('\n')
		b.WriteString(excerpt.Text)
	}
	return b.String()
}

func assessContextQuality(terms []string, excerpts []ContextExcerpt) ([]string, int, int, int) {
	matched := []string{}
	contentAnchors := map[string]bool{}
	for _, term := range terms {
		for _, excerpt := range excerpts {
			text := strings.ToLower(excerpt.Text)
			path := strings.ToLower(excerpt.Path)
			contentHit := strings.Contains(text, term)
			exactPathHit := isPathAnchor(term) && (path == term || strings.HasSuffix(path, "/"+term))
			if contentHit || exactPathHit {
				matched = append(matched, term)
				if contentHit && !isPathAnchor(term) {
					contentAnchors[term] = true
				}
				break
			}
		}
	}
	coverage := 0
	if len(terms) > 0 {
		coverage = len(matched) * 10000 / len(terms)
	}
	codeFiles := 0
	for _, excerpt := range excerpts {
		if isCodeSourcePath(excerpt.Path) {
			codeFiles++
		}
	}
	return matched, coverage, codeFiles, len(contentAnchors)
}

func isPathAnchor(term string) bool {
	return strings.ContainsAny(term, "/\\") || strings.Contains(filepath.Base(term), ".")
}

func makeExcerpt(file scoredFile, maxTextBytes int) ContextExcerpt {
	if maxTextBytes <= 0 || len(file.hits) == 0 {
		return ContextExcerpt{}
	}
	selected := map[int]bool{}
	for _, hit := range file.hits {
		for n := hit - 2; n <= hit+2; n++ {
			if n > 0 && n <= len(file.lines) {
				selected[n] = true
			}
		}
	}
	numbers := make([]int, 0, len(selected))
	for n := range selected {
		numbers = append(numbers, n)
	}
	sort.Ints(numbers)
	var b strings.Builder
	ranges := []string{}
	start, last := 0, 0
	for _, n := range numbers {
		line := file.lines[n-1]
		piece := strings.TrimRight(line, " \t")
		if b.Len()+len([]byte(piece))+1 > maxTextBytes {
			break
		}
		b.WriteString(piece)
		b.WriteByte('\n')
		if start == 0 {
			start, last = n, n
		} else if n == last+1 {
			last = n
		} else {
			ranges = append(ranges, lineRange(start, last))
			start, last = n, n
		}
	}
	if start > 0 {
		ranges = append(ranges, lineRange(start, last))
	}
	text := b.String()
	contentHash := sha256.Sum256([]byte(text))
	return ContextExcerpt{
		Path: file.path, Score: file.score, LineRanges: ranges, Text: text,
		SourceSHA256: file.sourceSHA256, ContentSHA256: hex.EncodeToString(contentHash[:]),
	}
}

func lineRange(start, end int) string {
	if start == end {
		return itoa(start)
	}
	return itoa(start) + "-" + itoa(end)
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	digits := [20]byte{}
	i := len(digits)
	for value > 0 {
		i--
		digits[i] = byte('0' + value%10)
		value /= 10
	}
	return string(digits[i:])
}

var tokenPattern = regexp.MustCompile(`[A-Za-z0-9_./-]+|[\p{Han}]+`)

var englishRetrievalStopWords = map[string]bool{
	"a": true, "an": true, "and": true, "are": true, "as": true, "at": true, "be": true, "by": true,
	"for": true, "from": true, "in": true, "into": true, "is": true, "it": true, "of": true, "on": true,
	"or": true, "that": true, "the": true, "this": true, "to": true, "with": true,
	"bug": true, "change": true, "code": true, "coding": true, "continue": true, "fix": true, "implement": true,
	"implementation": true, "improve": true, "issue": true, "optimize": true, "project": true, "task": true,
	"update": true, "verification": true, "verify": true,
}

var chineseRetrievalStopWords = []string{
	"全面", "继续", "项目", "任务", "问题", "代码", "编程", "修复", "实现", "功能", "更新", "验证", "检查", "优化", "完成",
}

func retrievalTerms(query string) []string {
	seen := map[string]bool{}
	result := []string{}
	add := func(value string) {
		value = strings.ToLower(strings.TrimSpace(value))
		if value == "" || seen[value] {
			return
		}
		seen[value] = true
		result = append(result, value)
	}
	for _, raw := range tokenPattern.FindAllString(strings.ToLower(query), -1) {
		runes := []rune(raw)
		if len(runes) > 0 && unicode.Is(unicode.Han, runes[0]) {
			cleaned := raw
			for _, stop := range chineseRetrievalStopWords {
				cleaned = strings.ReplaceAll(cleaned, stop, " ")
			}
			for _, segment := range strings.Fields(cleaned) {
				segmentRunes := []rune(segment)
				if len(segmentRunes) == 2 {
					add(segment)
					continue
				}
				for i := 0; i+1 < len(segmentRunes); i++ {
					add(string(segmentRunes[i : i+2]))
				}
			}
			continue
		}
		pathLike := strings.ContainsAny(raw, "./")
		if !pathLike && (len(runes) < 3 || englishRetrievalStopWords[raw]) {
			continue
		}
		add(raw)
	}
	sort.Strings(result)
	return result
}

func queryTerms(query string) []string {
	seen := map[string]bool{}
	result := []string{}
	for _, raw := range tokenPattern.FindAllString(strings.ToLower(query), -1) {
		runes := []rune(raw)
		candidates := []string{raw}
		if len(runes) > 2 && unicode.Is(unicode.Han, runes[0]) {
			for i := 0; i+1 < len(runes); i++ {
				candidates = append(candidates, string(runes[i:i+2]))
			}
		}
		for _, value := range candidates {
			if len([]rune(value)) >= 2 && !seen[value] {
				seen[value] = true
				result = append(result, value)
			}
		}
	}
	return result
}

func excludedDir(name string) bool {
	switch strings.ToLower(name) {
	case ".git", ".wuji", "node_modules", "vendor", "dist", "build", "bin", "coverage", ".next", ".cache":
		return true
	}
	return false
}

func sourceLike(path string) bool {
	lowerPath := strings.ToLower(filepath.ToSlash(path))
	base := strings.ToLower(filepath.Base(path))
	if base == "package-lock.json" || base == "pnpm-lock.yaml" || base == "yarn.lock" || base == "go.sum" || base == "cargo.lock" ||
		strings.HasSuffix(base, ".min.js") || strings.HasSuffix(base, ".min.css") || strings.HasSuffix(base, ".map") ||
		strings.Contains(lowerPath, "/generated/") {
		return false
	}
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".go", ".ts", ".tsx", ".js", ".jsx", ".mjs", ".cjs", ".vue", ".svelte", ".astro", ".py", ".rs", ".java", ".kt", ".cs", ".cpp", ".c", ".h", ".md", ".json", ".yaml", ".yml", ".toml", ".ps1", ".sh", ".html", ".css", ".scss", ".less", ".sql":
		return true
	}
	return false
}

func isCodeSourcePath(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".go", ".ts", ".tsx", ".js", ".jsx", ".mjs", ".cjs", ".vue", ".svelte", ".astro", ".py", ".rs", ".java", ".kt", ".cs", ".cpp", ".c", ".h", ".ps1", ".sh", ".sql":
		return true
	}
	return false
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
