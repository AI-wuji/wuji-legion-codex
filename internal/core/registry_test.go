package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateManifestRejectsIncompleteContract(t *testing.T) {
	base := validManifest("complete", "callable")
	tests := []struct {
		name   string
		mutate func(*Manifest)
	}{
		{"description", func(item *Manifest) { item.Description = "" }},
		{"triggers", func(item *Manifest) { item.Triggers = nil }},
		{"duplicate trigger", func(item *Manifest) { item.Triggers = []string{"Test", "test"} }},
		{"primary skill", func(item *Manifest) { item.PrimarySkill = "" }},
		{"fallback", func(item *Manifest) { item.Fallback = "" }},
		{"callable entry", func(item *Manifest) { item.HostCallable = false }},
		{"empty probe command", func(item *Manifest) { item.Probe = &Probe{} }},
		{"invalid probe timeout", func(item *Manifest) { item.Probe = &Probe{Command: "test", TimeoutSeconds: -1} }},
		{"activation without entrypoint", func(item *Manifest) {
			item.Sources = []Source{{
				ID: "atom", Activation: []string{"specific task"}, Globs: []string{t.TempDir()}, Required: []string{"SKILL.md"},
			}}
		}},
		{"automatic source must be callable", func(item *Manifest) {
			item.Sources = []Source{{
				ID: "atom", Priority: "primary", Entrypoint: "SKILL.md", Globs: []string{t.TempDir()}, Required: []string{"SKILL.md"},
			}}
		}},
		{"automatic source must declare entrypoint", func(item *Manifest) {
			item.Sources = []Source{{
				ID: "atom", Priority: "primary", Lifecycle: "callable", Globs: []string{t.TempDir()}, Required: []string{"SKILL.md"},
			}}
		}},
		{"invalid source lifecycle", func(item *Manifest) {
			item.Sources = []Source{{
				ID: "atom", Lifecycle: "fused", Globs: []string{t.TempDir()}, Required: []string{"SKILL.md"},
			}}
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			item := base
			test.mutate(&item)
			if err := ValidateManifest(item); err == nil {
				t.Fatalf("invalid manifest was accepted: %#v", item)
			}
		})
	}
}

func TestAuditSourcesSeparatesAutomaticAndColdMaterial(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"auto/SKILL.md", "cold/SKILL.md"} {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("ok"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	entries := AuditSources([]Manifest{{
		ID: "audit", Root: root,
		Sources: []Source{
			{ID: "auto", Priority: "primary", Lifecycle: "callable", Entrypoint: "SKILL.md", Globs: []string{"${ROOT}/auto"}, Required: []string{"SKILL.md"}},
			{ID: "cold", Lifecycle: "assets-retained", Globs: []string{"${ROOT}/cold"}, Required: []string{"SKILL.md"}},
		},
	}})
	if len(entries) != 2 || entries[0].State != "auto-selectable" || entries[0].ExecutionMode != "native-host-entrypoint-contract" || entries[0].ExecutionEvidence != "host-native-agent-or-tool-result-required" || entries[1].State != "cold-retained" {
		t.Fatalf("source audit did not separate route state: %#v", entries)
	}
}

func TestFeishuLinkAutomaticallySelectsOfficialEntrypoint(t *testing.T) {
	root := t.TempDir()
	sourceRoot := filepath.Join(root, "feishu")
	if err := os.MkdirAll(sourceRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sourceRoot, "SKILL.md"), []byte("read through official lark-cli"), 0o600); err != nil {
		t.Fatal(err)
	}
	manifest := Manifest{
		ID: "feishu", Root: root, Status: "callable", PrimarySkill: "feishu-lark", Fallback: "authenticate",
		Triggers: []string{"my.feishu.cn", "feishu"},
		Sources:  []Source{{ID: "official-lark-cli-skill", Priority: "primary", Lifecycle: "callable", Entrypoint: "SKILL.md", Globs: []string{"${ROOT}/feishu"}, Required: []string{"SKILL.md"}}},
	}
	route := Route("read https://my.feishu.cn/wiki/example", []Manifest{manifest})
	if route.Capability != "feishu" || route.PrimarySkill != "feishu-lark" || len(route.MountedSources) != 1 || route.MountedSources[0].ID != "official-lark-cli-skill" {
		t.Fatalf("Feishu link did not select the official automatic route: %#v", route)
	}
}

func TestAuditSourcesRejectsMissingEntrypointEvenWhenRequiredFilesExist(t *testing.T) {
	root := t.TempDir()
	sourceRoot := filepath.Join(root, "source")
	if err := os.MkdirAll(sourceRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sourceRoot, "present.txt"), []byte("present"), 0o600); err != nil {
		t.Fatal(err)
	}
	entries := AuditSources([]Manifest{{
		ID: "audit", Root: root,
		Sources: []Source{{ID: "bad-entrypoint", Priority: "primary", Lifecycle: "callable", Entrypoint: "SKILL.md", Globs: []string{"${ROOT}/source"}, Required: []string{"present.txt"}}},
	}})
	if len(entries) != 1 || entries[0].State != "unavailable" || !strings.Contains(entries[0].Reason, "entrypoint") {
		t.Fatalf("audit accepted a source without its declared entrypoint: %#v", entries)
	}
}

func TestValidateManifestRejectsUnsafeRequiredPath(t *testing.T) {
	item := validManifest("unsafe-source", "callable")
	item.Sources = []Source{{ID: "source", Globs: []string{t.TempDir()}, Required: []string{"../escape.txt"}}}
	if err := ValidateManifest(item); err == nil {
		t.Fatal("required path traversal was accepted")
	}
}

func TestValidateManifestRejectsProviderAmbiguity(t *testing.T) {
	item := validManifest("providers", "callable")
	item.Providers = []Provider{{ID: "one", Default: true}, {ID: "two", Default: true}}
	if err := ValidateManifest(item); err == nil {
		t.Fatal("multiple default providers were accepted")
	}
}

func TestValidateManifestRejectsUnknownProviderFallback(t *testing.T) {
	item := validManifest("provider-fallback", "callable")
	item.Providers = []Provider{{ID: "one", Default: true, Fallback: "missing"}, {ID: "two", Triggers: []string{"two"}, Fallback: "one"}}
	if err := ValidateManifest(item); err == nil {
		t.Fatal("unknown provider fallback was accepted")
	}
}

func TestValidateManifestRejectsDuplicateSourceIDs(t *testing.T) {
	item := validManifest("sources", "callable")
	item.Sources = []Source{
		{ID: "same", Globs: []string{t.TempDir()}, Required: []string{"SKILL.md"}},
		{ID: "same", Globs: []string{t.TempDir()}, Required: []string{"SKILL.md"}},
	}
	if err := ValidateManifest(item); err == nil {
		t.Fatal("duplicate source ids were accepted")
	}
}

func TestLoadManifestsRejectsUnknownFields(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "capabilities", "unknown", "manifest.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	data := `{"id":"unknown","description":"test","triggers":["test"],"status":"callable","primary_skill":"native","host_callable":true,"fallback":"native","typo":true}`
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadManifests(root); err == nil {
		t.Fatal("unknown manifest field was silently accepted")
	}
}

func TestResolveSourceUsesRootAndNaturalVersionOrder(t *testing.T) {
	root := t.TempDir()
	for _, version := range []string{"26.9", "26.10"} {
		if err := os.MkdirAll(filepath.Join(root, "versions", version), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	path, ok := ResolveSourceAt(root, Source{Globs: []string{"${ROOT}/versions/*"}})
	if !ok || filepath.Base(path) != "26.10" {
		t.Fatalf("wrong version selected: %q", path)
	}
}

func TestExpandPathDoesNotInventMachineSpecificProjectsRoot(t *testing.T) {
	t.Setenv("WUJI_PROJECTS", "")
	t.Setenv("USERPROFILE", t.TempDir())
	got := ExpandPathAt("", "${WUJI_PROJECTS}/cold-source")
	if !strings.Contains(got, "${WUJI_PROJECTS}") {
		t.Fatalf("unconfigured projects root was silently replaced: %q", got)
	}
}

func TestValidateManifestRejectsInvalidSourcePriority(t *testing.T) {
	item := validManifest("priority", "callable")
	item.DirectMount = true
	item.HostCallable = false
	item.Sources = []Source{{ID: "source", Priority: "urgent", Globs: []string{t.TempDir()}, Required: []string{"SKILL.md"}}}
	if err := ValidateManifest(item); err == nil {
		t.Fatal("invalid source priority was accepted")
	}
}

func TestValidateManifestRejectsSmokeBehaviorClaim(t *testing.T) {
	item := validBehaviorManifest("smoke-claim", "pass")
	item.Probe.Kind = "smoke"
	if err := ValidateManifest(item); err == nil {
		t.Fatal("smoke probe was accepted as behavior-verified evidence")
	}
}

func TestValidateManifestRejectsBehaviorClaimWithoutEvidenceContract(t *testing.T) {
	item := validBehaviorManifest("missing-evidence-contract", "pass")
	item.Probe.RequiredEvidence = nil
	item.Probe.ComparisonEvidence = ""
	if err := ValidateManifest(item); err == nil {
		t.Fatal("behavior claim without an evidence contract was accepted")
	}
}

func TestValidateManifestRejectsCallableWithoutProbe(t *testing.T) {
	item := validManifest("unprobed-callable", "callable")
	item.Probe = nil
	if err := ValidateManifest(item); err == nil {
		t.Fatal("callable manifest without an executable probe was accepted into the registry")
	}
}

func TestValidateManifestRejectsUnderclaimedBehaviorProbeWithoutEvidence(t *testing.T) {
	item := validManifest("underclaimed-behavior", "callable")
	item.Probe = &Probe{Command: "test", Kind: "behavior", Fixture: "fixture-v1"}
	if err := ValidateManifest(item); err == nil {
		t.Fatal("behavior probe without an evidence contract was accepted for a callable manifest")
	}
}

func TestValidateManifestRejectsUnsafePromotionReceipt(t *testing.T) {
	item := validBehaviorManifest("unsafe-promotion", "success")
	item.Status = "primary"
	item.PromotionReceipt = "../forged.json"
	if err := ValidateManifest(item); err == nil {
		t.Fatal("unsafe promotion receipt path was accepted")
	}
}
