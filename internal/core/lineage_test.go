package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateManifestRejectsInvalidFusionGenomeContract(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Manifest)
	}{
		{"unknown source", func(manifest *Manifest) { manifest.Genome.Adapters[0].Source = "missing" }},
		{"different entrypoint", func(manifest *Manifest) { manifest.Genome.Adapters[0].Entrypoint = "other.md" }},
		{"cold source", func(manifest *Manifest) { manifest.Sources[0].Lifecycle = "assets-retained" }},
		{"unsafe asset", func(manifest *Manifest) { manifest.Genome.Adapters[0].Assets = []string{"../escape.md"} }},
		{"wildcard asset", func(manifest *Manifest) { manifest.Genome.Adapters[0].Assets = []string{"assets/*.md"} }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manifest := newLineageManifest(t, t.TempDir())
			test.mutate(&manifest)
			if err := ValidateManifest(manifest); err == nil {
				t.Fatalf("invalid fusion genome was accepted: %#v", manifest.Genome)
			}
		})
	}
}

func TestVerifyChecksFusionGenomeEntrypointAndAssets(t *testing.T) {
	t.Setenv("GO_WANT_WUJI_PROBE_HELPER", "1")
	root := t.TempDir()
	manifest := newLineageManifest(t, root)
	got := Verify(root, manifest)
	if !got.Passed || got.Effective != "behavior-verified" || got.Genome == nil || len(got.Genome.Adapters) != 1 {
		t.Fatalf("valid fusion genome did not verify: %#v", got)
	}
	if len(got.Assets) != 1 || !got.Assets[0].Reachable || len(got.Assets[0].SHA256) != 64 || got.Assets[0].Bytes == 0 {
		t.Fatalf("fusion asset reachability was not recorded: %#v", got.Assets)
	}

	manifest.Genome.Adapters[0].Assets = append(manifest.Genome.Adapters[0].Assets, "assets/missing.md")
	got = Verify(root, manifest)
	if got.Passed || got.Genome == nil || len(got.Assets) != 2 || got.Assets[1].Reachable || !strings.Contains(strings.Join(got.Errors, "\n"), "assets/missing.md") {
		t.Fatalf("missing fusion asset was not rejected: %#v", got)
	}
}

func TestSyncLineageCatalogPersistsCallableMetadataAndMinimalRejections(t *testing.T) {
	root := t.TempDir()
	manifest := newLineageManifest(t, root)
	result, err := SyncLineageCatalog(root, []Manifest{manifest})
	if err != nil {
		t.Fatal(err)
	}
	if result.NodeCount != 4 || result.RejectionCount != 0 || len(result.CatalogSHA256) != 64 {
		t.Fatalf("lineage catalog is incomplete: %#v", result)
	}
	if _, err := os.Stat(result.CatalogPath); err != nil {
		t.Fatalf("lineage catalog was not persisted: %v", err)
	}
	if !lineageCatalogHasNode(result.Catalog, "genome:lineage-fixture:rev-1", "callable") {
		t.Fatalf("lineage catalog overstated or omitted the callable genome: %#v", result.Catalog)
	}
	for _, node := range result.Catalog.Nodes {
		if node.ID == "genome:lineage-fixture:rev-1" && (node.Species != "lineage-fixture" || node.ReleaseID != "release-1" || node.Generation != 2) {
			t.Fatalf("lineage generation metadata missing: %#v", node)
		}
	}
	manifest.Genome.Adapters = append(manifest.Genome.Adapters, FusionAdapter{
		ID: "workflow-copy", Domain: "workflow-copy", Source: "native-source", Entrypoint: "SKILL.md", Assets: []string{"assets/workflow.md"},
	})
	result, err = SyncLineageCatalog(root, []Manifest{manifest})
	if err != nil {
		t.Fatal(err)
	}
	if result.NodeCount != 6 {
		t.Fatalf("shared source should be represented once: %#v", result.Catalog.Nodes)
	}

	manifest.Genome.Adapters[0].Assets = append(manifest.Genome.Adapters[0].Assets, "assets/missing.md")
	result, err = SyncLineageCatalog(root, []Manifest{manifest})
	if err != nil {
		t.Fatal(err)
	}
	if result.RejectionCount != 1 || !lineageCatalogHasNode(result.Catalog, "genome:lineage-fixture:rev-1", "unavailable") {
		t.Fatalf("unreachable asset was not reduced to a minimal lineage rejection: %#v", result)
	}
}

func newLineageManifest(t *testing.T, root string) Manifest {
	t.Helper()
	sourceRoot := filepath.Join(root, "source")
	if err := os.MkdirAll(filepath.Join(sourceRoot, "assets"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sourceRoot, "SKILL.md"), []byte("entrypoint"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sourceRoot, "assets", "workflow.md"), []byte("retained workflow"), 0o600); err != nil {
		t.Fatal(err)
	}
	manifest := validBehaviorManifest("lineage-fixture", "success")
	manifest.Root = root
	manifest.Sources = []Source{{
		ID: "native-source", Lifecycle: "callable", Version: "2.0.1", Revision: "atom-7", ReleaseID: "release-1", License: "MIT", Entrypoint: "SKILL.md",
		Globs: []string{"${ROOT}/source"}, Required: []string{"SKILL.md"},
	}}
	manifest.Genome = &FusionGenome{
		SchemaVersion: 1, Species: "lineage-fixture", Revision: "rev-1", ReleaseID: "release-1", Generation: 2,
		Adapters: []FusionAdapter{{
			ID: "workflow", Domain: "workflow", Source: "native-source", Entrypoint: "SKILL.md", SourceVersion: "2.0.1", AtomRevision: "atom-7", ReleaseID: "release-1", License: "MIT", Assets: []string{"assets/workflow.md"},
		}},
	}
	return manifest
}

func lineageCatalogHasNode(catalog LineageCatalog, id, state string) bool {
	for _, node := range catalog.Nodes {
		if node.ID == id && node.State == state {
			return true
		}
	}
	return false
}
