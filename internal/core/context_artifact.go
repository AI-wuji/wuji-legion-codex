package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const contextArtifactSchemaVersion = 3

type contextDigestExcerpt struct {
	Path          string   `json:"path"`
	LineRanges    []string `json:"line_ranges"`
	Text          string   `json:"text"`
	SourceSHA256  string   `json:"source_sha256"`
	ContentSHA256 string   `json:"content_sha256"`
}

type contextDigest struct {
	QueryFingerprint string                 `json:"query_fingerprint"`
	RetrievalTerms   []string               `json:"retrieval_terms"`
	Excerpts         []contextDigestExcerpt `json:"excerpts"`
}

func queryFingerprint(terms []string) string {
	values := append([]string(nil), terms...)
	sort.Strings(values)
	return sha256Hex([]byte(strings.Join(values, "\n")))
}

func contextContentHash(fingerprint string, retrievalTerms []string, excerpts []ContextExcerpt) (string, error) {
	canonicalTerms := append([]string(nil), retrievalTerms...)
	sort.Strings(canonicalTerms)
	payload := make([]contextDigestExcerpt, 0, len(excerpts))
	for _, excerpt := range excerpts {
		payload = append(payload, contextDigestExcerpt{
			Path: excerpt.Path, LineRanges: excerpt.LineRanges, Text: excerpt.Text,
			SourceSHA256: excerpt.SourceSHA256, ContentSHA256: excerpt.ContentSHA256,
		})
	}
	encoded, err := json.Marshal(contextDigest{QueryFingerprint: fingerprint, RetrievalTerms: canonicalTerms, Excerpts: payload})
	if err != nil {
		return "", fmt.Errorf("encode context digest: %w", err)
	}
	return sha256Hex(encoded), nil
}

func selectedContextBytes(excerpts []ContextExcerpt) int {
	return len([]byte(renderContextPayload(excerpts)))
}

func contextHandle(contentSHA256 string) string {
	return "wuji-context://sha256/" + contentSHA256
}

func sha256Hex(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}

func WriteContextArtifact(result ContextResult, artifactDir string) (string, error) {
	if result.ContentSHA256 == "" || result.ContextHandle != contextHandle(result.ContentSHA256) {
		return "", fmt.Errorf("context result has an invalid content handle")
	}
	if strings.TrimSpace(artifactDir) == "" {
		return "", fmt.Errorf("artifact directory is required")
	}
	artifactDir, err := filepath.Abs(artifactDir)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(artifactDir, 0o755); err != nil {
		return "", fmt.Errorf("create context artifact directory: %w", err)
	}
	path := filepath.Join(artifactDir, result.ContentSHA256+".json")
	if _, err := os.Stat(path); err == nil {
		if _, loadErr := LoadContextArtifact(path); loadErr != nil {
			return "", loadErr
		}
		return path, nil
	} else if !os.IsNotExist(err) {
		return "", err
	}
	artifact := ContextArtifact{
		SchemaVersion:      contextArtifactSchemaVersion,
		Workspace:          result.Workspace,
		QueryFingerprint:   result.QueryFingerprint,
		Handle:             result.ContextHandle,
		ContentSHA256:      result.ContentSHA256,
		SelectedBytes:      result.SelectedBytes,
		RetrievalTerms:     append([]string(nil), result.RetrievalTerms...),
		MatchedTerms:       append([]string(nil), result.MatchedTerms...),
		CoverageBPS:        result.CoverageBPS,
		CodeExcerptCount:   result.CodeExcerptCount,
		ContentAnchorCount: result.ContentAnchorCount,
		PayloadSHA256:      result.PayloadSHA256,
		PayloadBytes:       result.PayloadBytes,
		Excerpts:           result.Excerpts,
	}
	temporary, err := os.CreateTemp(artifactDir, result.ContentSHA256+"-*.tmp")
	if err != nil {
		return "", fmt.Errorf("create context artifact: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	encoder := json.NewEncoder(temporary)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(artifact); err != nil {
		temporary.Close()
		return "", fmt.Errorf("write context artifact: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return "", err
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return "", fmt.Errorf("publish context artifact: %w", err)
	}
	return path, nil
}

func LoadContextArtifact(path string) (DelegationContext, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return DelegationContext{}, err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return DelegationContext{}, fmt.Errorf("read context artifact: %w", err)
	}
	var artifact ContextArtifact
	if err := json.Unmarshal(content, &artifact); err != nil {
		return DelegationContext{}, fmt.Errorf("decode context artifact: %w", err)
	}
	if artifact.SchemaVersion != contextArtifactSchemaVersion {
		return DelegationContext{}, fmt.Errorf("unsupported context artifact schema: %d", artifact.SchemaVersion)
	}
	workspace, err := filepath.Abs(artifact.Workspace)
	if err != nil {
		return DelegationContext{}, fmt.Errorf("resolve context workspace: %w", err)
	}
	if resolved, resolveErr := filepath.EvalSymlinks(workspace); resolveErr == nil {
		workspace = resolved
	}
	workspace = filepath.Clean(workspace)
	if strings.TrimSpace(artifact.QueryFingerprint) == "" {
		return DelegationContext{}, fmt.Errorf("context artifact query fingerprint is missing")
	}
	if len(artifact.RetrievalTerms) == 0 {
		return DelegationContext{}, fmt.Errorf("context artifact retrieval terms are missing")
	}
	for _, excerpt := range artifact.Excerpts {
		if excerpt.ContentSHA256 != sha256Hex([]byte(excerpt.Text)) {
			return DelegationContext{}, fmt.Errorf("context artifact excerpt hash mismatch: %s", excerpt.Path)
		}
		sourcePath := filepath.Clean(filepath.Join(workspace, filepath.FromSlash(excerpt.Path)))
		relative, relErr := filepath.Rel(workspace, sourcePath)
		if relErr != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return DelegationContext{}, fmt.Errorf("context artifact path escapes workspace: %s", excerpt.Path)
		}
		source, readErr := os.ReadFile(sourcePath)
		if readErr != nil {
			return DelegationContext{}, fmt.Errorf("read context source %s: %w", excerpt.Path, readErr)
		}
		if excerpt.SourceSHA256 != sha256Hex(source) {
			return DelegationContext{}, fmt.Errorf("context source changed: %s", excerpt.Path)
		}
	}
	actualBytes := selectedContextBytes(artifact.Excerpts)
	payload := renderContextPayload(artifact.Excerpts)
	if artifact.SelectedBytes != actualBytes || artifact.PayloadBytes != actualBytes {
		return DelegationContext{}, fmt.Errorf("context artifact selected byte count mismatch")
	}
	if artifact.PayloadSHA256 != sha256Hex([]byte(payload)) {
		return DelegationContext{}, fmt.Errorf("context artifact payload hash mismatch")
	}
	matchedTerms, coverage, codeFiles, contentAnchors := assessContextQuality(artifact.RetrievalTerms, artifact.Excerpts)
	if !sameStringSlice(artifact.MatchedTerms, matchedTerms) || artifact.CoverageBPS != coverage || artifact.CodeExcerptCount != codeFiles || artifact.ContentAnchorCount != contentAnchors {
		return DelegationContext{}, fmt.Errorf("context artifact quality metadata mismatch")
	}
	digest, err := contextContentHash(artifact.QueryFingerprint, artifact.RetrievalTerms, artifact.Excerpts)
	if err != nil {
		return DelegationContext{}, err
	}
	if artifact.ContentSHA256 != digest || artifact.Handle != contextHandle(digest) {
		return DelegationContext{}, fmt.Errorf("context artifact content hash mismatch")
	}
	if artifact.SelectedBytes <= 0 || len(artifact.Excerpts) == 0 {
		return DelegationContext{}, fmt.Errorf("context artifact is empty")
	}
	return DelegationContext{
		Handle: artifact.Handle, ArtifactPath: path, QueryFingerprint: artifact.QueryFingerprint, SelectedBytes: artifact.SelectedBytes,
		RetrievalTerms: append([]string(nil), artifact.RetrievalTerms...), MatchedTerms: append([]string(nil), artifact.MatchedTerms...),
		CoverageBPS: artifact.CoverageBPS, CodeExcerptCount: artifact.CodeExcerptCount, ContentAnchorCount: artifact.ContentAnchorCount,
		Payload: payload, PayloadSHA256: artifact.PayloadSHA256,
		verified: true,
	}, nil
}

func sameStringSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
