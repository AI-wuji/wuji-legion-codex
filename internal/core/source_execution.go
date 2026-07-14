package core

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// This names a preparation action only. The host still has to execute the
// selected capability and return an independently checkable result.
const sourceEntrypointInvocationKind = "skill-entrypoint-prepared"

// The entrypoint is a bounded instruction surface, not the full retained
// upstream package. The total replay gate in route.go remains stricter and
// decides whether it is economical to delegate a particular task.
const maxInjectedSourceBytes = 32 * 1024

// BuildSourceExecutionContracts reads only the exact entrypoint selected by
// the router. Cold source trees, templates and references are not loaded.
func BuildSourceExecutionContracts(manifest Manifest, sources []MountedSource) ([]SourceExecutionContract, error) {
	contracts := make([]SourceExecutionContract, 0, len(sources))
	for _, source := range sources {
		if strings.TrimSpace(source.Entrypoint) == "" {
			return nil, fmt.Errorf("mounted source %q has no executable entrypoint", source.ID)
		}
		entrypoint, content, err := resolveTrustedSourceEntrypoint([]Manifest{manifest}, manifest.ID, source.ID, source.Entrypoint)
		if err != nil {
			return nil, err
		}
		if len(content) > maxInjectedSourceBytes {
			return nil, fmt.Errorf("source %q entrypoint exceeds %d-byte injection budget", source.ID, maxInjectedSourceBytes)
		}
		digest := sha256.Sum256(content)
		contracts = append(contracts, SourceExecutionContract{
			SourceID: source.ID, Capability: manifest.ID, InvocationKind: sourceEntrypointInvocationKind,
			Entrypoint: source.Entrypoint, EntrypointSHA256: hex.EncodeToString(digest[:]),
			EntrypointBytes: len(content), ActivationReason: source.ActivationReason,
			ResolvedEntrypointPath: entrypoint,
			EntrypointContent:      string(content),
		})
	}
	return contracts, nil
}

// VerifySourceExecutionContracts rejects a stale route before a worker can
// run with a changed or missing Skill body and refreshes prompt-only content.
func VerifySourceExecutionContracts(manifests []Manifest, contracts []SourceExecutionContract) ([]SourceExecutionContract, []SourceEntrypointVerification, error) {
	verified := make([]SourceExecutionContract, 0, len(contracts))
	verification := make([]SourceEntrypointVerification, 0, len(contracts))
	for _, contract := range contracts {
		if contract.InvocationKind != sourceEntrypointInvocationKind {
			return nil, nil, fmt.Errorf("source %q has unsupported invocation kind %q", contract.SourceID, contract.InvocationKind)
		}
		if strings.TrimSpace(contract.SourceID) == "" || strings.TrimSpace(contract.Capability) == "" || !safeRelativePath(contract.Entrypoint) || len(contract.EntrypointSHA256) != 64 {
			return nil, nil, fmt.Errorf("source execution contract is incomplete")
		}
		entrypoint, content, err := resolveTrustedSourceEntrypoint(manifests, contract.Capability, contract.SourceID, contract.Entrypoint)
		if err != nil {
			return nil, nil, err
		}
		if len(content) > maxInjectedSourceBytes {
			return nil, nil, fmt.Errorf("source %q entrypoint exceeds %d-byte injection budget", contract.SourceID, maxInjectedSourceBytes)
		}
		digest := sha256.Sum256(content)
		actual := hex.EncodeToString(digest[:])
		if actual != contract.EntrypointSHA256 || len(content) != contract.EntrypointBytes {
			return nil, nil, fmt.Errorf("source %q entrypoint changed after routing", contract.SourceID)
		}
		contract.ResolvedEntrypointPath, contract.EntrypointContent = entrypoint, string(content)
		verified = append(verified, contract)
		verification = append(verification, SourceEntrypointVerification{
			SourceID: contract.SourceID, Capability: contract.Capability, InvocationKind: contract.InvocationKind, Entrypoint: contract.Entrypoint,
			EntrypointSHA256: actual, EntrypointBytes: len(content),
		})
	}
	return verified, verification, nil
}

// resolveTrustedSourceEntrypoint deliberately ignores every path from a route
// document. It derives the source root and relative entrypoint anew from the
// current trusted registry and rejects symlink/path escapes.
func resolveTrustedSourceEntrypoint(manifests []Manifest, capability, sourceID, entrypoint string) (string, []byte, error) {
	for _, manifest := range manifests {
		if manifest.ID != capability {
			continue
		}
		for _, source := range manifest.Sources {
			if source.ID != sourceID {
				continue
			}
			if source.Entrypoint != entrypoint || rank(sourceLifecycle(source)) < rank("callable") {
				return "", nil, fmt.Errorf("source %q is not a trusted executable entrypoint", sourceID)
			}
			root, ok := ResolveCompleteSourceAt(manifest.Root, source)
			if !ok {
				return "", nil, fmt.Errorf("trusted source %q is unavailable", sourceID)
			}
			resolvedRoot, err := filepath.EvalSymlinks(root)
			if err != nil {
				return "", nil, fmt.Errorf("resolve source %q root: %w", sourceID, err)
			}
			candidate := filepath.Join(resolvedRoot, filepath.FromSlash(entrypoint))
			resolvedCandidate, err := filepath.EvalSymlinks(candidate)
			if err != nil {
				return "", nil, fmt.Errorf("resolve source %q entrypoint: %w", sourceID, err)
			}
			rel, err := filepath.Rel(resolvedRoot, resolvedCandidate)
			if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
				return "", nil, fmt.Errorf("source %q entrypoint escapes its root", sourceID)
			}
			info, err := os.Stat(resolvedCandidate)
			if err != nil || !info.Mode().IsRegular() {
				return "", nil, fmt.Errorf("source %q entrypoint is not a regular file", sourceID)
			}
			content, err := os.ReadFile(resolvedCandidate)
			if err != nil || len(content) == 0 {
				return "", nil, fmt.Errorf("read source %q entrypoint: %w", sourceID, err)
			}
			return resolvedCandidate, content, nil
		}
	}
	return "", nil, fmt.Errorf("trusted source %q for capability %q is not registered", sourceID, capability)
}
