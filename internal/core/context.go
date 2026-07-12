package core

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
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
	workspace, err := filepath.Abs(workspace)
	if err != nil {
		return ContextResult{}, err
	}
	if resolved, resolveErr := filepath.EvalSymlinks(workspace); resolveErr == nil {
		workspace = resolved
	}
	workspace = filepath.Clean(workspace)
	terms := queryTerms(query)
	if len(terms) == 0 {
		return ContextResult{}, fmt.Errorf("query contains no searchable terms")
	}
	files := []scoredFile{}
	scanned := 0
	err = filepath.WalkDir(workspace, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if path != workspace && excludedDir(entry.Name()) {
				return filepath.SkipDir
			}
			return nil
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
		Workspace: workspace, Query: query, QueryFingerprint: queryFingerprint(terms), BudgetBytes: budget, ScannedFiles: scanned,
		Policy: []string{"rank before read", "emit matched ranges only", "hard byte budget", "content-addressed handoff", "keep raw logs out of context"},
	}
	for _, file := range files {
		remaining := budget - result.SelectedBytes
		overhead := len([]byte(file.path)) + 128
		excerpt := makeExcerpt(file, remaining-overhead)
		if excerpt.Text == "" {
			continue
		}
		used := len([]byte(excerpt.Text)) + overhead
		if result.SelectedBytes+used > budget {
			continue
		}
		result.Excerpts = append(result.Excerpts, excerpt)
		result.SelectedBytes += used
		if len(result.Excerpts) >= 12 {
			break
		}
	}
	if len(result.Excerpts) == 0 {
		return ContextResult{}, fmt.Errorf("no matching context excerpts")
	}
	result.ContentSHA256, err = contextContentHash(result.QueryFingerprint, result.Excerpts)
	if err != nil {
		return ContextResult{}, err
	}
	result.ContextHandle = contextHandle(result.ContentSHA256)
	return result, nil
}

func scoreFile(root, path string, terms []string) (scoredFile, error) {
	rel, _ := filepath.Rel(root, path)
	item := scoredFile{path: filepath.ToSlash(rel)}
	pathLower := strings.ToLower(item.path)
	for _, term := range terms {
		if strings.Contains(pathLower, term) {
			item.score += 20
			item.pathHit = true
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
				item.score += 3
				hit = true
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
	return item, nil
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
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".go", ".ts", ".tsx", ".js", ".jsx", ".mjs", ".cjs", ".vue", ".svelte", ".astro", ".py", ".rs", ".java", ".kt", ".cs", ".cpp", ".c", ".h", ".md", ".json", ".yaml", ".yml", ".toml", ".ps1", ".sh", ".html", ".css", ".scss", ".less", ".sql":
		return true
	}
	return false
}
