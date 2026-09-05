package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const workspaceGraphSchemaVersion = 1

const (
	workspaceGraphMaxTermsPerFile   = 512
	workspaceGraphMaxRefsPerTerm    = 256
	workspaceGraphMaxLookups        = 64
	workspaceGraphMaxCandidates     = 128
	workspaceGraphMaxSourceFiles    = 4096
	workspaceGraphMaxScanDuration   = 5 * time.Second
	workspaceGraphMaxSourceBytes    = 16 * 1024 * 1024
	workspaceGraphMaxBuildBytes     = 32 * 1024 * 1024
	workspaceGraphMaxRefBytes       = 512 * 1024
	workspaceGraphMaxTermBytes      = 256
	workspaceGraphMaxCleanupEntries = 512
	workspaceGraphMaxGenerations    = 8
)

var (
	errWorkspaceGraphMissing      = errors.New("workspace relation graph is missing")
	errWorkspaceGraphCleanupLimit = errors.New("workspace graph cleanup entry limit exceeded")
)

var graphSymbolPattern = regexp.MustCompile(`(?m)^\s*(?:func|type|class|def|interface|struct|enum|export\s+function|const|var)\s+([A-Za-z_][A-Za-z0-9_]*)`)

type WorkspaceGraphRelation struct {
	Predicate string `json:"predicate"`
	Target    string `json:"target"`
}

type workspaceGraphNode struct {
	ID              string                   `json:"id"`
	Path            string                   `json:"path"`
	Size            int64                    `json:"size"`
	ModTimeUnixNano int64                    `json:"mod_time_unix_nano"`
	SourceSHA256    string                   `json:"source_sha256"`
	Terms           []string                 `json:"terms"`
	Symbols         []string                 `json:"symbols,omitempty"`
	Relations       []WorkspaceGraphRelation `json:"relations,omitempty"`
}

type workspaceGraphMeta struct {
	SchemaVersion int      `json:"schema_version"`
	Workspace     string   `json:"workspace"`
	FileCount     int      `json:"file_count"`
	TermCount     int      `json:"term_count"`
	OverflowTerms []string `json:"overflow_terms,omitempty"`
	CreatedAt     string   `json:"created_at"`
}

type workspaceGraphActive struct {
	SchemaVersion int    `json:"schema_version"`
	Workspace     string `json:"workspace"`
	Generation    string `json:"generation"`
}

type WorkspaceGraphSyncResult struct {
	SchemaVersion   int    `json:"schema_version"`
	Workspace       string `json:"workspace"`
	GraphPath       string `json:"graph_path"`
	FileCount       int    `json:"file_count"`
	TermCount       int    `json:"term_count"`
	Rebuilt         bool   `json:"rebuilt"`
	MaxTermsPerFile int    `json:"max_terms_per_file"`
	MaxRefsPerTerm  int    `json:"max_refs_per_term"`
	MaxLookups      int    `json:"max_lookups"`
	MaxCandidates   int    `json:"max_candidates"`
	MaxSourceFiles  int    `json:"max_source_files"`
	MaxSourceBytes  int64  `json:"max_source_bytes"`
}

type workspaceGraphRef struct {
	ID   string `json:"id"`
	Path string `json:"path"`
}

type workspaceGraphStats struct {
	Mode           string
	IndexedFiles   int
	CandidateFiles int
	GraphLookups   int
	FallbackReason string
	Truncated      bool
	SourceBytes    int64
}

func workspaceGraphDir(workspace string) string {
	return filepath.Join(workspace, ".wuji", "graph", "v1")
}

func workspaceGraphGenerationDir(workspace, generation string) string {
	return filepath.Join(workspaceGraphDir(workspace), "generations", generation)
}

func SyncWorkspaceGraph(workspace string) (WorkspaceGraphSyncResult, error) {
	workspace, err := normalizeWorkspacePath(workspace)
	if err != nil {
		return WorkspaceGraphSyncResult{}, err
	}
	var result WorkspaceGraphSyncResult
	deadline := time.Now().Add(workspaceGraphMaxScanDuration)
	err = withWorkspaceGraphLock(workspace, func() error {
		if err := checkWorkspaceGraphDeadline(deadline); err != nil {
			return err
		}
		graphDir := workspaceGraphDir(workspace)
		generation := fmt.Sprintf("%d-%d", time.Now().UnixNano(), os.Getpid())
		temporary := workspaceGraphGenerationDir(workspace, generation+".tmp")
		finalGeneration := workspaceGraphGenerationDir(workspace, generation)
		if err := cleanupWorkspaceGraphTemporary(graphDir, deadline); err != nil {
			return err
		}
		if activeDir, activeErr := activeWorkspaceGraphDir(workspace); activeErr == nil {
			if err := cleanupWorkspaceGraphGenerations(graphDir, filepath.Base(activeDir), deadline); err != nil {
				return err
			}
		} else if !errors.Is(activeErr, errWorkspaceGraphMissing) {
			return activeErr
		}
		if err := ensureWorkspaceGraphGenerationCapacity(graphDir); err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Join(temporary, "nodes"), 0o700); err != nil {
			return err
		}
		termRefs := map[string][]workspaceGraphRef{}
		relationRefs := map[string][]workspaceGraphRef{}
		overflow := map[string]bool{}
		fileCount := 0
		var buildBytes int64
		err := filepath.WalkDir(workspace, func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if err := checkWorkspaceGraphDeadline(deadline); err != nil {
				return err
			}
			if entry.IsDir() {
				if path != workspace && excludedDir(entry.Name()) {
					return filepath.SkipDir
				}
				return nil
			}
			if entry.Type()&os.ModeSymlink != 0 || !sourceLike(path) {
				return nil
			}
			info, err := entry.Info()
			if err != nil {
				return err
			}
			if info.Size() > 2*1024*1024 {
				return nil
			}
			if buildBytes+info.Size() > workspaceGraphMaxBuildBytes {
				return fmt.Errorf("workspace graph build source byte limit exceeded: %d", workspaceGraphMaxBuildBytes)
			}
			if fileCount >= workspaceGraphMaxSourceFiles {
				return fmt.Errorf("workspace graph source file limit exceeded: %d", workspaceGraphMaxSourceFiles)
			}
			node, err := buildWorkspaceGraphNode(workspace, path)
			if err != nil {
				return err
			}
			data, err := json.Marshal(node)
			if err != nil {
				return err
			}
			if err := writeWorkspaceGraphFile(filepath.Join(temporary, "nodes", node.ID+".json"), append(data, '\n'), deadline); err != nil {
				return err
			}
			ref := workspaceGraphRef{ID: node.ID, Path: node.Path}
			for _, term := range node.Terms {
				appendWorkspaceGraphRef(termRefs, overflow, "term:"+term, term, ref)
			}
			for _, relation := range node.Relations {
				for _, term := range retrievalTerms(relation.Target) {
					appendWorkspaceGraphRef(relationRefs, overflow, "relation:"+term, term, ref)
				}
			}
			fileCount++
			buildBytes += info.Size()
			return nil
		})
		if err != nil {
			_ = removeWorkspaceGraphTreeBounded(temporary, deadline)
			return err
		}
		if err := writeWorkspaceGraphRefs(temporary, "terms", termRefs, deadline); err != nil {
			_ = removeWorkspaceGraphTreeBounded(temporary, deadline)
			return err
		}
		if err := writeWorkspaceGraphRefs(temporary, "relations", relationRefs, deadline); err != nil {
			_ = removeWorkspaceGraphTreeBounded(temporary, deadline)
			return err
		}
		overflowTerms := make([]string, 0, len(overflow))
		for term := range overflow {
			overflowTerms = append(overflowTerms, term)
		}
		sort.Strings(overflowTerms)
		meta := workspaceGraphMeta{SchemaVersion: workspaceGraphSchemaVersion, Workspace: workspace, FileCount: fileCount, TermCount: len(termRefs) + len(relationRefs), OverflowTerms: overflowTerms, CreatedAt: time.Now().UTC().Format(time.RFC3339Nano)}
		data, err := json.MarshalIndent(meta, "", "  ")
		if err != nil {
			_ = removeWorkspaceGraphTreeBounded(temporary, deadline)
			return err
		}
		if err := writeWorkspaceGraphFile(filepath.Join(temporary, "meta.json"), append(data, '\n'), deadline); err != nil {
			_ = removeWorkspaceGraphTreeBounded(temporary, deadline)
			return err
		}
		if err := checkWorkspaceGraphDeadline(deadline); err != nil {
			return err
		}
		if err := os.Rename(temporary, finalGeneration); err != nil {
			_ = removeWorkspaceGraphTreeBounded(temporary, deadline)
			return err
		}
		active := workspaceGraphActive{SchemaVersion: workspaceGraphSchemaVersion, Workspace: workspace, Generation: generation}
		activeData, err := json.Marshal(active)
		if err != nil {
			return err
		}
		if err := writeWorkspaceGraphFile(filepath.Join(graphDir, "active.json"), append(activeData, '\n'), deadline); err != nil {
			return err
		}
		// Publication is already atomic. Reclaim small prior generations when the
		// remaining budget allows, but never turn a published graph into failure.
		_ = cleanupWorkspaceGraphGenerations(graphDir, generation, deadline)
		result = WorkspaceGraphSyncResult{
			SchemaVersion: workspaceGraphSchemaVersion, Workspace: workspace, GraphPath: graphDir,
			FileCount: fileCount, TermCount: len(termRefs) + len(relationRefs), Rebuilt: true,
			MaxTermsPerFile: workspaceGraphMaxTermsPerFile, MaxRefsPerTerm: workspaceGraphMaxRefsPerTerm, MaxLookups: workspaceGraphMaxLookups,
			MaxCandidates: workspaceGraphMaxCandidates, MaxSourceFiles: workspaceGraphMaxSourceFiles, MaxSourceBytes: workspaceGraphMaxSourceBytes,
		}
		return nil
	})
	return result, err
}

func queryWorkspaceGraph(workspace string, terms []string) ([]scoredFile, workspaceGraphStats, error) {
	var files []scoredFile
	var stats workspaceGraphStats
	normalized, err := normalizeWorkspacePath(workspace)
	if err != nil {
		return files, stats, err
	}
	workspace = normalized
	err = withWorkspaceGraphLock(workspace, func() error {
		graphDir, err := activeWorkspaceGraphDir(workspace)
		if err != nil {
			return err
		}
		metaData, err := readWorkspaceGraphFileBounded(filepath.Join(graphDir, "meta.json"), workspaceGraphMaxRefBytes)
		if err != nil {
			if os.IsNotExist(err) {
				return errWorkspaceGraphMissing
			}
			return err
		}
		var meta workspaceGraphMeta
		if err := json.Unmarshal(metaData, &meta); err != nil {
			return err
		}
		if meta.SchemaVersion != workspaceGraphSchemaVersion || filepath.Clean(meta.Workspace) != filepath.Clean(workspace) {
			return errWorkspaceGraphMissing
		}
		stats = workspaceGraphStats{Mode: "workspace-graph", IndexedFiles: meta.FileCount}
		overflow := map[string]bool{}
		for _, term := range meta.OverflowTerms {
			overflow[term] = true
		}
		refsByID := map[string]workspaceGraphRef{}
		lookups := boundedWorkspaceGraphTerms(terms)
		if len(lookups) > workspaceGraphMaxLookups {
			lookups = lookups[:workspaceGraphMaxLookups]
			stats.Truncated = true
		}
	lookupLoop:
		for _, term := range lookups {
			stats.GraphLookups++
			for _, indexType := range []string{"terms", "relations"} {
				if overflow[strings.TrimSuffix(indexType, "s")+":"+term] {
					stats.Truncated = true
				}
				data, err := readWorkspaceGraphFileBounded(filepath.Join(graphDir, indexType, graphHash(term), "refs.json"), workspaceGraphMaxRefBytes)
				if os.IsNotExist(err) {
					continue
				}
				if err != nil {
					return err
				}
				var refs []workspaceGraphRef
				if err := json.Unmarshal(data, &refs); err != nil {
					return err
				}
				for _, ref := range refs {
					if err := validateWorkspaceGraphRef(ref); err != nil {
						return errWorkspaceGraphMissing
					}
					refsByID[ref.ID] = ref
					if len(refsByID) >= workspaceGraphMaxCandidates {
						stats.Truncated = true
						break lookupLoop
					}
				}
			}
		}
		stats.CandidateFiles = len(refsByID)
		files = make([]scoredFile, 0, len(refsByID))
		for _, ref := range refsByID {
			nodeData, err := readWorkspaceGraphFileBounded(filepath.Join(graphDir, "nodes", ref.ID+".json"), workspaceGraphMaxRefBytes)
			if err != nil {
				return errWorkspaceGraphMissing
			}
			var node workspaceGraphNode
			if err := json.Unmarshal(nodeData, &node); err != nil {
				return errWorkspaceGraphMissing
			}
			if node.ID != ref.ID || node.Path != ref.Path || node.ID != graphHash(node.Path) {
				return errWorkspaceGraphMissing
			}
			path, err := resolveWorkspaceGraphPath(workspace, node.Path)
			if err != nil {
				return errWorkspaceGraphMissing
			}
			info, err := os.Stat(path)
			if os.IsNotExist(err) {
				stats.FallbackReason = "stale-index"
				return errWorkspaceGraphMissing
			}
			if err != nil {
				return err
			}
			if info.Size() != node.Size || info.ModTime().UnixNano() != node.ModTimeUnixNano {
				stats.FallbackReason = "stale-index"
				return errWorkspaceGraphMissing
			}
			if stats.SourceBytes+info.Size() > workspaceGraphMaxSourceBytes {
				stats.Truncated = true
				break
			}
			item, err := scoreFile(workspace, path, terms)
			if err != nil {
				return err
			}
			stats.SourceBytes += info.Size()
			if item.sourceSHA256 != node.SourceSHA256 {
				stats.FallbackReason = "stale-index"
				return errWorkspaceGraphMissing
			}
			if item.score > 0 {
				files = append(files, item)
			}
		}
		return nil
	})
	return files, stats, err
}

func buildWorkspaceGraphNode(workspace, path string) (workspaceGraphNode, error) {
	resolved, err := resolveWorkspaceGraphPath(workspace, filepath.ToSlash(strings.TrimPrefix(strings.TrimPrefix(path, workspace), string(filepath.Separator))))
	if err != nil || filepath.Clean(resolved) != filepath.Clean(path) {
		return workspaceGraphNode{}, fmt.Errorf("source path escapes workspace: %s", path)
	}
	info, err := os.Stat(path)
	if err != nil {
		return workspaceGraphNode{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return workspaceGraphNode{}, err
	}
	relative, err := filepath.Rel(workspace, path)
	if err != nil {
		return workspaceGraphNode{}, err
	}
	relative = filepath.ToSlash(relative)
	text := string(data)
	terms := boundedWorkspaceGraphTerms(append(retrievalTerms(relative), retrievalTerms(text)...))
	if len(terms) > workspaceGraphMaxTermsPerFile {
		terms = terms[:workspaceGraphMaxTermsPerFile]
	}
	symbols := []string{}
	for _, match := range graphSymbolPattern.FindAllStringSubmatch(text, 128) {
		if len(match) > 1 {
			symbols = append(symbols, match[1])
		}
	}
	symbols = uniqueKnowledgeStrings(symbols)
	sourceHash := sha256.Sum256(data)
	return workspaceGraphNode{
		ID: graphHash(relative), Path: relative, Size: info.Size(), ModTimeUnixNano: info.ModTime().UnixNano(),
		SourceSHA256: hex.EncodeToString(sourceHash[:]), Terms: terms, Symbols: symbols,
		Relations: workspaceGraphRelations(relative, symbols),
	}, nil
}

func boundedWorkspaceGraphTerms(terms []string) []string {
	bounded := make([]string, 0, len(terms))
	for _, term := range terms {
		term = strings.TrimSpace(term)
		if term != "" && len([]byte(term)) <= workspaceGraphMaxTermBytes {
			bounded = append(bounded, term)
		}
	}
	return uniqueKnowledgeStrings(bounded)
}

func workspaceGraphRelations(path string, symbols []string) []WorkspaceGraphRelation {
	result := []WorkspaceGraphRelation{}
	if strings.HasSuffix(path, "_test.go") || strings.HasSuffix(path, ".test.ts") || strings.HasSuffix(path, ".spec.ts") || strings.HasSuffix(path, ".test.js") {
		result = append(result, WorkspaceGraphRelation{Predicate: "tests", Target: strings.TrimSuffix(strings.TrimSuffix(path, "_test.go"), ".test.ts")})
	}
	for _, symbol := range symbols {
		result = append(result, WorkspaceGraphRelation{Predicate: "defines", Target: symbol})
	}
	return result
}

func graphHash(value string) string {
	hash := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(value))))
	return hex.EncodeToString(hash[:12])
}

func checkWorkspaceGraphDeadline(deadline time.Time) error {
	if time.Now().After(deadline) {
		return fmt.Errorf("workspace graph rebuild timed out after %s", workspaceGraphMaxScanDuration)
	}
	return nil
}

func writeWorkspaceGraphFile(path string, data []byte, deadline time.Time) error {
	if err := checkWorkspaceGraphDeadline(deadline); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".wuji-graph-*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return err
	}
	if _, err := temporary.Write(data); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := checkWorkspaceGraphDeadline(deadline); err != nil {
		return err
	}
	return os.Rename(temporaryPath, path)
}

func appendWorkspaceGraphRef(index map[string][]workspaceGraphRef, overflow map[string]bool, overflowKey, term string, ref workspaceGraphRef) {
	if len(index[term]) >= workspaceGraphMaxRefsPerTerm {
		overflow[overflowKey] = true
		return
	}
	index[term] = append(index[term], ref)
}

func writeWorkspaceGraphRefs(root, indexType string, index map[string][]workspaceGraphRef, deadline time.Time) error {
	for term, refs := range index {
		if err := checkWorkspaceGraphDeadline(deadline); err != nil {
			return err
		}
		sort.Slice(refs, func(i, j int) bool { return refs[i].Path < refs[j].Path })
		data, err := json.Marshal(refs)
		if err != nil {
			return err
		}
		if len(data) > workspaceGraphMaxRefBytes {
			return fmt.Errorf("workspace graph reference index exceeds %d bytes", workspaceGraphMaxRefBytes)
		}
		path := filepath.Join(root, indexType, graphHash(term), "refs.json")
		if err := writeWorkspaceGraphFile(path, append(data, '\n'), deadline); err != nil {
			return err
		}
	}
	return nil
}

func activeWorkspaceGraphDir(workspace string) (string, error) {
	normalized, err := normalizeWorkspacePath(workspace)
	if err != nil {
		return "", err
	}
	workspace = normalized
	base := workspaceGraphDir(workspace)
	data, err := readWorkspaceGraphFileBounded(filepath.Join(base, "active.json"), 16*1024)
	if os.IsNotExist(err) {
		return "", errWorkspaceGraphMissing
	}
	if err != nil {
		return "", err
	}
	var active workspaceGraphActive
	if err := json.Unmarshal(data, &active); err != nil {
		return "", fmt.Errorf("%w: invalid active pointer", errWorkspaceGraphMissing)
	}
	if active.SchemaVersion != workspaceGraphSchemaVersion || filepath.Clean(active.Workspace) != filepath.Clean(workspace) || strings.Contains(active.Generation, "/") || strings.Contains(active.Generation, "\\") || strings.Contains(active.Generation, "..") {
		return "", fmt.Errorf("%w: active pointer identity mismatch", errWorkspaceGraphMissing)
	}
	path := workspaceGraphGenerationDir(workspace, active.Generation)
	if info, err := os.Stat(path); err != nil || !info.IsDir() {
		return "", fmt.Errorf("%w: active generation unavailable: %s", errWorkspaceGraphMissing, path)
	}
	return path, nil
}

func validateWorkspaceGraphRef(ref workspaceGraphRef) error {
	if len(ref.ID) != 24 || ref.ID != graphHash(ref.Path) || filepath.IsAbs(filepath.FromSlash(ref.Path)) || ref.Path == "" {
		return fmt.Errorf("invalid workspace graph reference")
	}
	if _, err := hex.DecodeString(ref.ID); err != nil {
		return fmt.Errorf("invalid workspace graph reference")
	}
	return nil
}

func resolveWorkspaceGraphPath(workspace, relative string) (string, error) {
	if filepath.IsAbs(filepath.FromSlash(relative)) || strings.TrimSpace(relative) == "" {
		return "", fmt.Errorf("workspace graph path must be relative")
	}
	path := filepath.Clean(filepath.Join(workspace, filepath.FromSlash(relative)))
	rel, err := filepath.Rel(workspace, path)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("workspace graph path escapes workspace")
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", err
	}
	rel, err = filepath.Rel(workspace, resolved)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("workspace graph symlink escapes workspace")
	}
	return filepath.Clean(resolved), nil
}

func readWorkspaceGraphFileBounded(path string, maxBytes int64) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if info.Size() > maxBytes {
		return nil, fmt.Errorf("workspace graph file exceeds %d bytes", maxBytes)
	}
	return io.ReadAll(io.LimitReader(file, maxBytes+1))
}

func withWorkspaceGraphLock(workspace string, fn func() error) error {
	base := workspaceGraphDir(workspace)
	if err := os.MkdirAll(base, 0o700); err != nil {
		return err
	}
	lockPath := filepath.Join(base, ".lock")
	deadline := time.Now().Add(knowledgeLockWait)
	for {
		file, err := os.OpenFile(lockPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if err == nil {
			_ = file.Close()
			defer os.Remove(lockPath)
			return fn()
		}
		if !os.IsExist(err) {
			return err
		}
		if info, statErr := os.Stat(lockPath); statErr == nil && time.Since(info.ModTime()) > knowledgeLockWait*4 {
			_ = os.Remove(lockPath)
			continue
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("workspace graph lock timed out")
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func cleanupWorkspaceGraphTemporary(base string, deadline time.Time) error {
	entries, err := os.ReadDir(filepath.Join(base, "generations"))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err := checkWorkspaceGraphDeadline(deadline); err != nil {
			return err
		}
		if entry.IsDir() && strings.HasSuffix(entry.Name(), ".tmp") {
			if err := removeWorkspaceGraphTreeBounded(filepath.Join(base, "generations", entry.Name()), deadline); err != nil {
				if errors.Is(err, errWorkspaceGraphCleanupLimit) {
					continue
				}
				return err
			}
		}
	}
	return nil
}

func cleanupWorkspaceGraphGenerations(base, active string, deadline time.Time) error {
	entries, err := os.ReadDir(filepath.Join(base, "generations"))
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err := checkWorkspaceGraphDeadline(deadline); err != nil {
			return err
		}
		if entry.IsDir() && entry.Name() != active {
			if err := removeWorkspaceGraphTreeBounded(filepath.Join(base, "generations", entry.Name()), deadline); err != nil {
				if errors.Is(err, errWorkspaceGraphCleanupLimit) {
					continue
				}
				return err
			}
		}
	}
	return nil
}

// removeWorkspaceGraphTreeBounded only touches disposable graph generations.
// It refuses a large recursive sweep so a corrupted cache cannot turn rebuild
// into an unbounded cleanup operation.
func removeWorkspaceGraphTreeBounded(root string, deadline time.Time) error {
	if err := checkWorkspaceGraphDeadline(deadline); err != nil {
		return err
	}
	directories := make([]string, 0, 8)
	entries := 0
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if err := checkWorkspaceGraphDeadline(deadline); err != nil {
			return err
		}
		if entries >= workspaceGraphMaxCleanupEntries {
			return fmt.Errorf("%w: %d", errWorkspaceGraphCleanupLimit, workspaceGraphMaxCleanupEntries)
		}
		entries++
		if entry.IsDir() {
			directories = append(directories, path)
			return nil
		}
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	})
	for index := len(directories) - 1; index >= 0; index-- {
		if err := checkWorkspaceGraphDeadline(deadline); err != nil {
			return err
		}
		if err := os.Remove(directories[index]); err != nil && !os.IsNotExist(err) && !errors.Is(err, os.ErrExist) {
			return err
		}
	}
	return err
}

func ensureWorkspaceGraphGenerationCapacity(base string) error {
	entries, err := os.ReadDir(filepath.Join(base, "generations"))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if len(entries) >= workspaceGraphMaxGenerations {
		return fmt.Errorf("workspace graph generation limit exceeded: %d", workspaceGraphMaxGenerations)
	}
	return nil
}

func normalizeWorkspacePath(workspace string) (string, error) {
	workspace, err := filepath.Abs(workspace)
	if err != nil {
		return "", err
	}
	if resolved, resolveErr := filepath.EvalSymlinks(workspace); resolveErr == nil {
		workspace = resolved
	}
	info, err := os.Stat(workspace)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("workspace is not a directory: %s", workspace)
	}
	return filepath.Clean(workspace), nil
}
