package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	lineageCatalogSchemaVersion = 1
	lineageCatalogFilename      = "catalog.json"
)

// SyncLineageCatalog stores a compact content-addressed catalog under the
// target workspace. It records only adapter metadata, trusted entrypoints,
// and retained assets; execution and behavior evidence remain Verify's job.
func SyncLineageCatalog(root string, manifests []Manifest) (LineageSyncResult, error) {
	if strings.TrimSpace(root) == "" {
		return LineageSyncResult{}, fmt.Errorf("lineage root is required")
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return LineageSyncResult{}, err
	}
	store := filepath.Join(absRoot, ".wuji", "lineage")
	result := LineageSyncResult{}
	err = withKnowledgeStoreLock(store, func() error {
		catalog := LineageCatalog{SchemaVersion: lineageCatalogSchemaVersion, Nodes: []LineageNode{}, Rejections: []LineageRejection{}}
		for _, manifest := range manifests {
			if manifest.Genome == nil {
				continue
			}
			if manifest.Root == "" {
				manifest.Root = absRoot
			}
			if err := ValidateManifest(manifest); err != nil {
				return fmt.Errorf("capability %q: %w", manifest.ID, err)
			}
			verification, assets, _ := verifyFusionGenome(manifest)
			appendLineageNodes(&catalog, manifest, verification, assets)
		}
		sortLineageCatalog(&catalog)
		path := filepath.Join(store, "v1", lineageCatalogFilename)
		if err := writeLineageCatalog(path, catalog); err != nil {
			return err
		}
		if err := AuditEventRecord(auditStoreFor(store), AuditEvent{EventType: "lineage-synchronized", Actor: "aji", Authority: "aji-merge", Target: path, ResultHandle: "wuji-lineage://" + manifestDigest(catalog)}); err != nil {
			return err
		}
		digest, err := lineageCatalogDigest(catalog)
		if err != nil {
			return err
		}
		result = LineageSyncResult{
			CatalogPath:    path,
			CatalogSHA256:  digest,
			NodeCount:      len(catalog.Nodes),
			RejectionCount: len(catalog.Rejections),
			Catalog:        catalog,
		}
		return nil
	})
	if err != nil {
		return LineageSyncResult{}, err
	}
	return result, nil
}

func manifestDigest(catalog LineageCatalog) string {
	return lineageDigest(catalog)
}

func appendLineageNodes(catalog *LineageCatalog, manifest Manifest, verification FusionGenomeVerification, assets []AssetReachability) {
	assetByPath := make(map[string]AssetReachability, len(assets))
	for _, asset := range assets {
		assetByPath[asset.Path] = asset
	}
	adapterIDs := make([]string, 0, len(manifest.Genome.Adapters))
	sources := make(map[string]Source, len(manifest.Sources))
	for _, source := range manifest.Sources {
		sources[source.ID] = source
	}
	genomeState := "callable"
	for index, adapter := range manifest.Genome.Adapters {
		assetSpecs, _ := fusionAdapterAssets(manifest.ID, adapter)
		entry := verification.Adapters[index]
		sourceNodeID := lineageSourceNodeID(manifest.ID, adapter.Source)
		sourceState := "unavailable"
		if entry.Reachable {
			sourceState = "callable"
		}
		appendLineageNode(catalog, LineageNode{
			ID: sourceNodeID, Kind: "source", Capability: manifest.ID, SourceID: adapter.Source, Entrypoint: adapter.Entrypoint, Path: adapter.Entrypoint, SHA256: entry.EntrypointSHA256, State: sourceState,
			SourceVersion: firstNonEmpty(adapter.SourceVersion, sources[adapter.Source].Version), AtomRevision: firstNonEmpty(adapter.AtomRevision, sources[adapter.Source].Revision), ReleaseID: firstNonEmpty(adapter.ReleaseID, sources[adapter.Source].ReleaseID), License: firstNonEmpty(adapter.License, sources[adapter.Source].License),
		})

		adapterNodeID := lineageAdapterNodeID(manifest.ID, adapter.ID)
		adapterIDs = append(adapterIDs, adapterNodeID)
		adapterState := sourceState
		if !entry.Reachable {
			appendLineageRejection(catalog, adapterNodeID, manifest.ID+":"+adapter.Source, entry.Reason, adapter, manifest.Genome)
		}
		for _, assetSpec := range assetSpecs {
			asset := assetByPath[fusionAssetLabel(adapter.ID, assetSpec.Path)]
			assetNodeID := lineageAssetNodeID(manifest.ID, adapter.ID, assetSpec.Path)
			assetState := "unavailable"
			if asset.Reachable {
				assetState = "assets-retained"
			} else {
				adapterState = "unavailable"
				appendLineageRejection(catalog, assetNodeID, manifest.ID+":"+adapter.Source, asset.Reason, adapter, manifest.Genome)
			}
			appendLineageNode(catalog, LineageNode{
				ID: assetNodeID, Kind: "asset", Capability: manifest.ID, SourceID: adapter.Source, AssetID: manifest.ID + ":" + assetSpec.ID, Entrypoint: adapter.Entrypoint, Compatibility: assetSpec.Compatibility, Parents: []string{adapterNodeID}, Path: assetSpec.Path, SHA256: asset.SHA256, State: assetState,
				SourceVersion: adapter.SourceVersion, AtomRevision: adapter.AtomRevision, ReleaseID: adapter.ReleaseID, License: adapter.License,
			})
		}
		adapterHash := lineageDigest(adapter)
		appendLineageNode(catalog, LineageNode{
			ID: adapterNodeID, Kind: "fusion-adapter", Capability: manifest.ID, SourceID: adapter.Source, Entrypoint: adapter.Entrypoint, Parents: []string{sourceNodeID}, SHA256: adapterHash, Path: adapter.Entrypoint, State: adapterState,
			SourceVersion: adapter.SourceVersion, AtomRevision: adapter.AtomRevision, Species: manifest.Genome.Species, FusionRevision: manifest.Genome.Revision, ReleaseID: firstNonEmpty(adapter.ReleaseID, manifest.Genome.ReleaseID), Generation: manifest.Genome.Generation, License: adapter.License,
		})
		if adapterState != "callable" {
			genomeState = "unavailable"
		}
	}
	appendLineageNode(catalog, LineageNode{
		ID: lineageGenomeNodeID(manifest.ID, manifest.Genome.Revision), Kind: "fusion-genome", Capability: manifest.ID, Parents: adapterIDs,
		SHA256: lineageDigest(manifest.Genome), State: genomeState, Species: manifest.Genome.Species, FusionRevision: manifest.Genome.Revision, ReleaseID: manifest.Genome.ReleaseID, Generation: manifest.Genome.Generation,
	})
}

func appendLineageNode(catalog *LineageCatalog, node LineageNode) {
	for index, existing := range catalog.Nodes {
		if existing.ID == node.ID {
			catalog.Nodes[index] = node
			return
		}
	}
	catalog.Nodes = append(catalog.Nodes, node)
}

func sortLineageCatalog(catalog *LineageCatalog) {
	sort.Slice(catalog.Nodes, func(i, j int) bool { return catalog.Nodes[i].ID < catalog.Nodes[j].ID })
	sort.Slice(catalog.Rejections, func(i, j int) bool { return catalog.Rejections[i].ID < catalog.Rejections[j].ID })
}

func appendLineageRejection(catalog *LineageCatalog, id, source, reason string, adapter FusionAdapter, genome *FusionGenome) {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		reason = "unreachable"
	}
	catalog.Rejections = append(catalog.Rejections, LineageRejection{
		ID: id, Source: source, Reason: reason, State: "rejected", SourceVersion: adapter.SourceVersion, AtomRevision: adapter.AtomRevision, ReleaseID: firstNonEmpty(adapter.ReleaseID, genome.ReleaseID), SHA256: lineageDigest(struct {
			ID     string `json:"id"`
			Source string `json:"source"`
			Reason string `json:"reason"`
		}{ID: id, Source: source, Reason: reason}),
	})
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func writeLineageCatalog(path string, catalog LineageCatalog) error {
	data, err := json.MarshalIndent(catalog, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".lineage-*.json")
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
	return os.Rename(temporaryPath, path)
}

func lineageCatalogDigest(catalog LineageCatalog) (string, error) {
	data, err := json.Marshal(catalog)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:]), nil
}

func lineageDigest(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}

func lineageSourceNodeID(capability, source string) string {
	return "source:" + capability + ":" + source
}

func lineageAdapterNodeID(capability, adapter string) string {
	return "adapter:" + capability + ":" + adapter
}

func lineageAssetNodeID(capability, adapter, asset string) string {
	digest := sha256.Sum256([]byte(capability + "\x00" + adapter + "\x00" + asset))
	return "asset:" + capability + ":" + adapter + ":" + hex.EncodeToString(digest[:])
}

func lineageGenomeNodeID(capability, revision string) string {
	return "genome:" + capability + ":" + revision
}

func verifyFusionGenome(manifest Manifest) (FusionGenomeVerification, []AssetReachability, []string) {
	verification := FusionGenomeVerification{Species: manifest.Genome.Species, Revision: manifest.Genome.Revision, Adapters: make([]FusionAdapterVerification, 0, len(manifest.Genome.Adapters))}
	sources := make(map[string]Source, len(manifest.Sources))
	for _, source := range manifest.Sources {
		sources[source.ID] = source
	}
	assets := []AssetReachability{}
	issues := []string{}
	for _, adapter := range manifest.Genome.Adapters {
		assetSpecs, err := fusionAdapterAssets(manifest.ID, adapter)
		if err != nil {
			issues = append(issues, fmt.Sprintf("fusion adapter %s assets: %s", adapter.ID, err))
			verification.Adapters = append(verification.Adapters, FusionAdapterVerification{ID: adapter.ID, Source: adapter.Source, Entrypoint: adapter.Entrypoint, Reason: err.Error()})
			continue
		}
		entry := FusionAdapterVerification{ID: adapter.ID, Source: adapter.Source, Entrypoint: adapter.Entrypoint}
		_, content, err := resolveTrustedSourceEntrypoint([]Manifest{manifest}, manifest.ID, adapter.Source, adapter.Entrypoint)
		if err != nil {
			entry.Reason = err.Error()
			issues = append(issues, fmt.Sprintf("fusion adapter %s entrypoint: %s", adapter.ID, entry.Reason))
			for _, asset := range assetSpecs {
				assets = append(assets, AssetReachability{Path: fusionAssetLabel(adapter.ID, asset.Path), Kind: "fusion-asset", Reason: entry.Reason})
			}
			verification.Adapters = append(verification.Adapters, entry)
			continue
		}
		digest := sha256.Sum256(content)
		entry.EntrypointSHA256 = hex.EncodeToString(digest[:])
		entry.Reachable = true
		sourceRoot, ok := ResolveCompleteSourceAt(manifest.Root, sources[adapter.Source])
		if !ok {
			entry.Reachable = false
			entry.Reason = "trusted source became unavailable"
			issues = append(issues, fmt.Sprintf("fusion adapter %s entrypoint: %s", adapter.ID, entry.Reason))
			for _, asset := range assetSpecs {
				assets = append(assets, AssetReachability{Path: fusionAssetLabel(adapter.ID, asset.Path), Kind: "fusion-asset", Reason: entry.Reason})
			}
			verification.Adapters = append(verification.Adapters, entry)
			continue
		}
		for _, asset := range assetSpecs {
			reachability := verifyFusionAsset(sourceRoot, adapter.ID, asset.Path)
			assets = append(assets, reachability)
			if !reachability.Reachable {
				issues = append(issues, fmt.Sprintf("fusion adapter %s asset %s: %s", adapter.ID, asset.Path, reachability.Reason))
			}
		}
		verification.Adapters = append(verification.Adapters, entry)
	}
	return verification, assets, issues
}

func verifyFusionAsset(sourceRoot, adapterID, asset string) AssetReachability {
	result := AssetReachability{Path: fusionAssetLabel(adapterID, asset), Kind: "fusion-asset"}
	resolvedRoot, err := filepath.EvalSymlinks(sourceRoot)
	if err != nil {
		result.Reason = "resolve source root: " + err.Error()
		return result
	}
	resolvedAsset, err := filepath.EvalSymlinks(filepath.Join(resolvedRoot, filepath.FromSlash(asset)))
	if err != nil {
		result.Reason = "resolve asset: " + err.Error()
		return result
	}
	relative, err := filepath.Rel(resolvedRoot, resolvedAsset)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		result.Reason = "asset escapes source root"
		return result
	}
	info, err := os.Stat(resolvedAsset)
	if err != nil || !info.Mode().IsRegular() {
		result.Reason = "asset is not a regular file"
		return result
	}
	file, err := os.Open(resolvedAsset)
	if err != nil {
		result.Reason = "open asset: " + err.Error()
		return result
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		result.Reason = "hash asset: " + err.Error()
		return result
	}
	result.Reachable = true
	result.SHA256 = hex.EncodeToString(hash.Sum(nil))
	result.Bytes = info.Size()
	return result
}

func fusionAssetLabel(adapterID, asset string) string {
	return adapterID + ":" + filepath.ToSlash(asset)
}
