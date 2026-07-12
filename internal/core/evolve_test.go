package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestEvolutionRejectsLabelOnlyFusion(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "capabilities", "base"), 0o755); err != nil {
		t.Fatal(err)
	}
	base := validManifest("base", "callable")
	writeManifest(t, filepath.Join(root, "capabilities", "base", "manifest.json"), base)
	candidate := validManifest("new-pack", "behavior-verified")
	path := filepath.Join(root, "candidate.json")
	writeManifest(t, path, candidate)
	got, err := EvaluateCandidate(root, path, false)
	if err != nil {
		t.Fatal(err)
	}
	if got.Decision != "reject" || got.CandidateProof.Passed {
		t.Fatalf("label-only fusion was admitted: %#v", got)
	}
}

func TestEvolutionRejectsPathTraversalID(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "capabilities", "base"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeManifest(t, filepath.Join(root, "capabilities", "base", "manifest.json"), validManifest("base", "callable"))
	candidatePath := filepath.Join(root, "candidate.json")
	unsafe := validManifest("safe", "behavior-verified")
	unsafe.ID = "../escaped"
	writeManifest(t, candidatePath, unsafe)

	got, err := EvaluateCandidate(root, candidatePath, true)
	if err != nil {
		t.Fatal(err)
	}
	if got.Decision != "reject" || got.Applied || got.CandidateProof.Passed {
		t.Fatalf("unsafe candidate was admitted: %#v", got)
	}
	if _, err := os.Stat(filepath.Join(root, "escaped", "manifest.json")); !os.IsNotExist(err) {
		t.Fatalf("unsafe candidate escaped capability root: %v", err)
	}
}

func TestLoadManifestsRejectsDirectoryIDMismatch(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "capabilities", "expected", "manifest.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	writeManifest(t, path, validManifest("different", "known"))
	if _, err := LoadManifests(root); err == nil {
		t.Fatal("directory and manifest id mismatch was accepted")
	}
}

func TestEvolutionReplacesOnlyAfterSameFixtureComparison(t *testing.T) {
	t.Setenv("GO_WANT_WUJI_PROBE_HELPER", "1")
	root := t.TempDir()
	target := filepath.Join(root, "capabilities", "shared", "manifest.json")
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatal(err)
	}
	existing := validBehaviorManifest("shared", "success")
	existing.Description = "existing"
	writeManifest(t, target, existing)
	candidate := validBehaviorManifest("shared", "success")
	candidate.Description = "candidate"
	candidatePath := filepath.Join(root, "candidate.json")
	writeManifest(t, candidatePath, candidate)

	got, err := EvaluateCandidate(root, candidatePath, true)
	if err != nil {
		t.Fatal(err)
	}
	if got.Decision != "replace" || !got.Applied || got.ExistingProof == nil || !got.ExistingProof.Passed {
		t.Fatalf("verified replacement did not complete: %#v", got)
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	var installed Manifest
	if err := json.Unmarshal(data, &installed); err != nil {
		t.Fatal(err)
	}
	if installed.Description != "candidate" {
		t.Fatalf("candidate was not installed: %#v", installed)
	}
	if installed.Status != "primary" || installed.PromotionReceipt == "" {
		t.Fatalf("verified replacement was not promoted with evidence: %#v", installed)
	}
	verified := Verify(root, installed)
	if !verified.Passed || verified.Effective != "primary" {
		t.Fatalf("installed promotion receipt did not verify: %#v", verified)
	}
	receiptPath := filepath.Join(root, "capabilities", "shared", filepath.FromSlash(installed.PromotionReceipt))
	if _, err := os.Stat(receiptPath); err != nil {
		t.Fatalf("promotion receipt was not persisted: %v", err)
	}
	archives, err := filepath.Glob(filepath.Join(root, "retired", "shared", "*", "manifest.json"))
	if err != nil || len(archives) != 1 {
		t.Fatalf("existing manifest was not archived: %v %#v", err, archives)
	}
}

func TestEvolutionAdmissionDoesNotSelfPromote(t *testing.T) {
	t.Setenv("GO_WANT_WUJI_PROBE_HELPER", "1")
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "capabilities", "base"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeManifest(t, filepath.Join(root, "capabilities", "base", "manifest.json"), validManifest("base", "known"))
	candidate := validBehaviorManifest("new-pack", "success")
	candidatePath := filepath.Join(root, "candidate.json")
	writeManifest(t, candidatePath, candidate)

	got, err := EvaluateCandidate(root, candidatePath, true)
	if err != nil {
		t.Fatal(err)
	}
	if got.Decision != "admit" || !got.Applied {
		t.Fatalf("verified candidate was not admitted: %#v", got)
	}
	data, err := os.ReadFile(filepath.Join(root, "capabilities", "new-pack", "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	installed, err := decodeManifest(data)
	if err != nil {
		t.Fatal(err)
	}
	if installed.Status != "behavior-verified" || installed.PromotionReceipt != "" {
		t.Fatalf("new admission self-promoted without a prior route: %#v", installed)
	}
}

func TestVerifyRejectsTamperedPromotionReceipt(t *testing.T) {
	t.Setenv("GO_WANT_WUJI_PROBE_HELPER", "1")
	root := t.TempDir()
	target := filepath.Join(root, "capabilities", "shared", "manifest.json")
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatal(err)
	}
	writeManifest(t, target, validBehaviorManifest("shared", "success"))
	candidatePath := filepath.Join(root, "candidate.json")
	writeManifest(t, candidatePath, validBehaviorManifest("shared", "success"))
	if got, err := EvaluateCandidate(root, candidatePath, true); err != nil || !got.Applied {
		t.Fatalf("replacement setup failed: %#v %v", got, err)
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	installed, err := decodeManifest(data)
	if err != nil {
		t.Fatal(err)
	}
	receiptPath := filepath.Join(root, "capabilities", "shared", filepath.FromSlash(installed.PromotionReceipt))
	receipt, err := os.ReadFile(receiptPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(receiptPath, append(receipt, ' '), 0o644); err != nil {
		t.Fatal(err)
	}
	got := Verify(root, installed)
	if got.Passed || got.Effective == "primary" {
		t.Fatalf("tampered promotion receipt was accepted: %#v", got)
	}
}

func TestEvolutionHoldsMismatchedFixture(t *testing.T) {
	t.Setenv("GO_WANT_WUJI_PROBE_HELPER", "1")
	root := t.TempDir()
	target := filepath.Join(root, "capabilities", "shared", "manifest.json")
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatal(err)
	}
	existing := validBehaviorManifest("shared", "success")
	writeManifest(t, target, existing)
	candidate := validBehaviorManifest("shared", "success")
	candidate.Probe.Fixture = "different-fixture"
	candidate.Probe.Args[len(candidate.Probe.Args)-1] = candidate.Probe.Fixture
	candidatePath := filepath.Join(root, "candidate.json")
	writeManifest(t, candidatePath, candidate)

	got, err := EvaluateCandidate(root, candidatePath, true)
	if err != nil {
		t.Fatal(err)
	}
	if got.Decision != "hold" || got.Applied {
		t.Fatalf("mismatched fixture replaced current capability: %#v", got)
	}
}

func TestEvolutionHoldsDifferentBehaviorSignature(t *testing.T) {
	t.Setenv("GO_WANT_WUJI_PROBE_HELPER", "1")
	root := t.TempDir()
	target := filepath.Join(root, "capabilities", "shared", "manifest.json")
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatal(err)
	}
	existing := validBehaviorManifest("shared", "success")
	writeManifest(t, target, existing)
	candidate := validBehaviorManifest("shared", "different-signature")
	candidatePath := filepath.Join(root, "candidate.json")
	writeManifest(t, candidatePath, candidate)

	got, err := EvaluateCandidate(root, candidatePath, true)
	if err != nil {
		t.Fatal(err)
	}
	if got.Decision != "hold" || got.Applied || got.ExistingProof == nil || got.CandidateProof.Probe == nil {
		t.Fatalf("different behavior signature replaced current capability: %#v", got)
	}
}

func TestEvolutionHoldsDifferentVerifiedEvidence(t *testing.T) {
	t.Setenv("GO_WANT_WUJI_PROBE_HELPER", "1")
	root := t.TempDir()
	target := filepath.Join(root, "capabilities", "shared", "manifest.json")
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatal(err)
	}
	writeManifest(t, target, validBehaviorManifest("shared", "success"))
	candidatePath := filepath.Join(root, "candidate.json")
	writeManifest(t, candidatePath, validBehaviorManifest("shared", "different-evidence"))

	got, err := EvaluateCandidate(root, candidatePath, true)
	if err != nil {
		t.Fatal(err)
	}
	if got.Decision != "hold" || got.Applied {
		t.Fatalf("different verified assertion evidence replaced current capability: %#v", got)
	}
}

func TestAtomicWriteFileReplacesExistingContent(t *testing.T) {
	target := filepath.Join(t.TempDir(), "manifest.json")
	if err := os.WriteFile(target, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := atomicWriteFile(target, []byte("new"), 0o644); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(target)
	if err != nil || string(data) != "new" {
		t.Fatalf("atomic replacement failed: %q %v", data, err)
	}
}

func writeManifest(t *testing.T, path string, manifest Manifest) {
	t.Helper()
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}
