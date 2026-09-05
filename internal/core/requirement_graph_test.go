package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRequirementGraphPersistsRevisionsAndSparseProjection(t *testing.T) {
	store := t.TempDir()
	route, err := UpsertRequirement(store, RequirementInput{ID: "route-contract", Summary: "Preserve the single Aji routing chain.", Sources: []string{"message:sha256:1111111111111111111111111111111111111111111111111111111111111111"}})
	if err != nil || route.VersionID != "route-contract@1" {
		t.Fatalf("initial requirement was not recorded: node=%#v err=%v", route, err)
	}
	if _, err := UpsertRequirement(store, RequirementInput{ID: "unrelated", Summary: "Do not add unrelated scope.", Sources: []string{"message:sha256:2222222222222222222222222222222222222222222222222222222222222222"}}); err != nil {
		t.Fatal(err)
	}
	decision, err := RecordDecision(store, DecisionInput{ID: "preserve-aji", Summary: "Keep Aji as the only router.", Sources: []string{"message:sha256:3333333333333333333333333333333333333333333333333333333333333333"}, Requirements: []string{"route-contract"}, Status: "accepted"})
	if err != nil || decision.VersionID != "preserve-aji@1" || len(decision.DependsOn) != 1 || decision.DependsOn[0] != route.VersionID {
		t.Fatalf("decision did not bind the active requirement revision: node=%#v err=%v", decision, err)
	}
	projection, err := ProjectRequirementGraph(store, "preserve-aji", 1024)
	if err != nil || projection.PayloadBytes > 1024 || len(projection.Nodes) != 2 || projection.Nodes[0].VersionID != decision.VersionID {
		t.Fatalf("projection did not retain only the target and direct dependency: projection=%#v err=%v", projection, err)
	}
	for _, node := range projection.Nodes {
		if node.ID == "unrelated" {
			t.Fatalf("projection included an unrelated requirement: %#v", projection.Nodes)
		}
	}
	updated, err := UpsertRequirement(store, RequirementInput{ID: "route-contract", Summary: "Preserve Aji as the single routing chain.", Sources: []string{"message:sha256:4444444444444444444444444444444444444444444444444444444444444444"}})
	if err != nil || updated.VersionID != "route-contract@2" || updated.Supersedes != route.VersionID {
		t.Fatalf("requirement revision was not preserved: node=%#v err=%v", updated, err)
	}
}

func TestRequirementGraphProjectionRejectsTampering(t *testing.T) {
	store := t.TempDir()
	if _, err := UpsertRequirement(store, RequirementInput{ID: "bounded-context", Summary: "Project only current requirements and dependencies.", Sources: []string{"message:sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}}); err != nil {
		t.Fatal(err)
	}
	projection, err := ProjectRequirementGraph(store, "bounded-context", 1024)
	if err != nil {
		t.Fatal(err)
	}
	path, err := WriteRequirementGraphProjection(projection, filepath.Join(store, "artifacts"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := LoadRequirementGraphProjection(path); err != nil {
		t.Fatalf("valid projection did not load: %v", err)
	}
	if err := os.WriteFile(path, []byte(`{"schema_version":1,"payload":"tampered"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadRequirementGraphProjection(path); err == nil {
		t.Fatal("tampered requirement graph projection was accepted")
	}
}
