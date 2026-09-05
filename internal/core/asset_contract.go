package core

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

func fusionAdapterAssets(capability string, adapter FusionAdapter) ([]FusionAsset, error) {
	if !capabilityIDPattern.MatchString(strings.TrimSpace(capability)) {
		return nil, fmt.Errorf("capability is invalid")
	}
	assetsByPath := map[string]FusionAsset{}
	ids := map[string]bool{}
	for _, item := range adapter.AssetContracts {
		asset, err := normalizeFusionAsset(capability, adapter.ID, item, true)
		if err != nil {
			return nil, err
		}
		if _, exists := assetsByPath[asset.Path]; exists {
			return nil, fmt.Errorf("has duplicate asset path %q", asset.Path)
		}
		if ids[asset.ID] {
			return nil, fmt.Errorf("has duplicate asset_id %q", asset.ID)
		}
		assetsByPath[asset.Path], ids[asset.ID] = asset, true
	}
	for _, path := range adapter.Assets {
		asset, err := normalizeFusionAsset(capability, adapter.ID, FusionAsset{Path: path}, false)
		if err != nil {
			return nil, err
		}
		if _, exists := assetsByPath[asset.Path]; exists {
			continue
		}
		if ids[asset.ID] {
			return nil, fmt.Errorf("has duplicate asset_id %q", asset.ID)
		}
		assetsByPath[asset.Path], ids[asset.ID] = asset, true
	}
	assets := make([]FusionAsset, 0, len(assetsByPath))
	for _, asset := range assetsByPath {
		assets = append(assets, asset)
	}
	sort.Slice(assets, func(left, right int) bool { return assets[left].ID < assets[right].ID })
	return assets, nil
}

func normalizeFusionAsset(capability, adapterID string, item FusionAsset, explicitID bool) (FusionAsset, error) {
	item.ID = strings.TrimSpace(item.ID)
	item.Path = strings.TrimSpace(item.Path)
	if !safeRelativePath(item.Path) || strings.ContainsAny(item.Path, "*?[") {
		return FusionAsset{}, fmt.Errorf("asset %q must be a safe concrete relative path", item.Path)
	}
	if item.ID == "" && explicitID {
		return FusionAsset{}, fmt.Errorf("asset_id is required")
	}
	if item.ID == "" {
		item.ID = stableFusionAssetID(capability, adapterID, item.Path)
	}
	if !componentIDPattern.MatchString(item.ID) {
		return FusionAsset{}, fmt.Errorf("asset_id %q is invalid", item.ID)
	}
	compatibility, err := normalizeAssetCompatibility(item.Compatibility)
	if err != nil {
		return FusionAsset{}, err
	}
	item.Compatibility = compatibility
	return item, nil
}

func stableFusionAssetID(capability, adapterID, path string) string {
	digest := sha256.Sum256([]byte(capability + "\x00" + adapterID + "\x00" + filepathToSlash(path)))
	return "asset-" + hex.EncodeToString(digest[:16])
}

func normalizeAssetCompatibility(values []string) ([]string, error) {
	if len(values) > 16 {
		return nil, fmt.Errorf("compatibility cannot contain more than 16 entries")
	}
	seen := map[string]bool{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || len(value) > 128 || strings.ContainsAny(value, "\r\n\t") {
			return nil, fmt.Errorf("compatibility entry is invalid")
		}
		for _, pattern := range knowledgeSecretPatterns {
			if pattern.MatchString(value) {
				return nil, fmt.Errorf("compatibility must not contain a secret")
			}
		}
		if !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	sort.Strings(result)
	return result, nil
}

func filepathToSlash(value string) string {
	return strings.ReplaceAll(value, "\\", "/")
}

func SelectFusionAsset(manifests []Manifest, request AssetSelectionRequest) (AssetInvocationContract, error) {
	request.Capability = strings.TrimSpace(request.Capability)
	request.Domain = strings.TrimSpace(request.Domain)
	request.AssetID = strings.TrimSpace(request.AssetID)
	if request.Capability != "" && !capabilityIDPattern.MatchString(request.Capability) {
		return AssetInvocationContract{}, fmt.Errorf("capability is invalid")
	}
	if request.Domain == "" && request.AssetID == "" {
		return AssetInvocationContract{}, fmt.Errorf("asset_id or domain is required")
	}
	compatibility, err := normalizeAssetCompatibility(request.Compatibility)
	if err != nil {
		return AssetInvocationContract{}, err
	}
	type candidate struct {
		manifest Manifest
		adapter  FusionAdapter
		asset    FusionAsset
	}
	candidates := []candidate{}
	for _, manifest := range manifests {
		if request.Capability != "" && manifest.ID != request.Capability {
			continue
		}
		if manifest.Genome == nil {
			continue
		}
		for _, adapter := range manifest.Genome.Adapters {
			if request.Domain != "" && adapter.Domain != request.Domain {
				continue
			}
			assets, err := fusionAdapterAssets(manifest.ID, adapter)
			if err != nil {
				return AssetInvocationContract{}, fmt.Errorf("capability %q adapter %q: %w", manifest.ID, adapter.ID, err)
			}
			for _, asset := range assets {
				qualifiedID := manifest.ID + ":" + asset.ID
				if request.AssetID != "" && request.AssetID != asset.ID && request.AssetID != qualifiedID {
					continue
				}
				if !assetSupports(asset.Compatibility, compatibility) {
					continue
				}
				candidates = append(candidates, candidate{manifest: manifest, adapter: adapter, asset: asset})
			}
		}
	}
	if len(candidates) == 0 {
		return AssetInvocationContract{}, fmt.Errorf("no trusted fusion asset matches the selection")
	}
	sort.Slice(candidates, func(left, right int) bool {
		leftKey := candidates[left].manifest.ID + "\x00" + candidates[left].adapter.ID + "\x00" + candidates[left].asset.ID
		rightKey := candidates[right].manifest.ID + "\x00" + candidates[right].adapter.ID + "\x00" + candidates[right].asset.ID
		return leftKey < rightKey
	})
	selected := candidates[0]
	var source Source
	found := false
	for _, item := range selected.manifest.Sources {
		if item.ID == selected.adapter.Source {
			source, found = item, true
			break
		}
	}
	if !found {
		return AssetInvocationContract{}, fmt.Errorf("selected source %q is not registered", selected.adapter.Source)
	}
	sourceRoot, available := ResolveCompleteSourceAt(selected.manifest.Root, source)
	if !available {
		return AssetInvocationContract{}, fmt.Errorf("selected source %q is unavailable", selected.adapter.Source)
	}
	reachability := verifyFusionAsset(sourceRoot, selected.adapter.ID, selected.asset.Path)
	if !reachability.Reachable {
		return AssetInvocationContract{}, fmt.Errorf("selected asset %q is unavailable: %s", selected.asset.ID, reachability.Reason)
	}
	contracts, err := BuildSourceExecutionContracts(selected.manifest, []MountedSource{{ID: source.ID, Entrypoint: selected.adapter.Entrypoint, ActivationReason: "fusion-asset:" + selected.asset.ID}})
	if err != nil {
		return AssetInvocationContract{}, err
	}
	if len(contracts) != 1 {
		return AssetInvocationContract{}, fmt.Errorf("selected source did not resolve one invocation contract")
	}
	return AssetInvocationContract{
		AssetID: selected.manifest.ID + ":" + selected.asset.ID, Capability: selected.manifest.ID, AdapterID: selected.adapter.ID,
		Domain: selected.adapter.Domain, SourceID: source.ID, Entrypoint: selected.adapter.Entrypoint,
		EntrypointSHA256: contracts[0].EntrypointSHA256, AssetPath: selected.asset.Path, AssetSHA256: reachability.SHA256,
		AssetBytes: reachability.Bytes, Compatibility: selected.asset.Compatibility, Invocation: contracts[0],
	}, nil
}

func assetSupports(assetCompatibility, requested []string) bool {
	available := make(map[string]bool, len(assetCompatibility))
	for _, value := range assetCompatibility {
		available[value] = true
	}
	for _, value := range requested {
		if !available[value] {
			return false
		}
	}
	return true
}
