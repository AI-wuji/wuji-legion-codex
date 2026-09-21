package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func expertTestRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func expertEvidence(t *testing.T, dir, name, content string) ExpertEvidenceFile {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	d := sha256.Sum256([]byte(content))
	return ExpertEvidenceFile{Path: path, SHA256: hex.EncodeToString(d[:]), Bytes: int64(len(content))}
}

func expertTestWorker() WorkerTask {
	return WorkerTask{ID: "expert-worker", Model: "gpt-5.6-terra", ModelClass: "terra", SessionKey: "expert-bridge-20260919", MaxAttempts: 1, DelegationGateReason: "expert workflow selected", TaskContract: "repair reproducible bug", TaskContractSHA256: strings.Repeat("a", 64), StablePrefixSHA256: strings.Repeat("b", 64), ContextPayloadSHA256: strings.Repeat("c", 64)}
}

func writeExpertTestCatalog(t *testing.T, root string, experts ...expertDefinition) {
	t.Helper()
	for len(experts) < 6 {
		experts = append(experts, expertDefinition{ID: fmt.Sprintf("filler-%d", len(experts)), Triggers: []string{fmt.Sprintf("filler-%d", len(experts))}})
	}
	data, err := json.Marshal(expertCatalog{Experts: experts})
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(root, "capabilities", "experts")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestSelectExpertUsesTriggersAndAntiTriggers(t *testing.T) {
	root := expertTestRoot(t)
	selected, err := SelectExpert(root, "测试失败，请修复代码 bug")
	if err != nil || selected.State != "selected" || selected.ExpertID != "code-repair" || len(selected.Matched) != 3 || selected.CandidateCount != 1 {
		t.Fatalf("unexpected selection: %#v err=%v", selected, err)
	}
	vetoed, err := SelectExpert(root, "修复代码，但这只是纯解释")
	if err != nil || vetoed.State != "none" || vetoed.Reason != "anti-trigger" || vetoed.FallbackHint != "return-to-aji-original-route" {
		t.Fatalf("anti-trigger did not return original route: %#v err=%v", vetoed, err)
	}
}

func TestSelectExpertReturnsNoMatchToOriginalRoute(t *testing.T) {
	selection, err := SelectExpert(expertTestRoot(t), "请解释这个概念的历史背景")
	if err != nil || selection.State != "none" || selection.Reason != "no-match" || selection.ExpertID != "" || selection.FallbackHint != "return-to-aji-original-route" {
		t.Fatalf("no lexical match must abstain without claiming no skill: %#v err=%v", selection, err)
	}
}

func TestSelectExpertReturnsAmbiguousIndependentOfCatalogOrder(t *testing.T) {
	first := t.TempDir()
	second := t.TempDir()
	alpha := expertDefinition{ID: "alpha", Triggers: []string{"部署"}}
	beta := expertDefinition{ID: "beta", Triggers: []string{"部署"}}
	writeExpertTestCatalog(t, first, alpha, beta)
	writeExpertTestCatalog(t, second, beta, alpha)
	for _, root := range []string{first, second} {
		selection, err := SelectExpert(root, "请部署服务")
		if err != nil || selection.State != "ambiguous" || selection.Reason != "tie" || selection.ExpertID != "" || selection.CandidateCount != 2 || len(selection.Candidates) != 2 || selection.Candidates[0].ExpertID != "alpha" || selection.Candidates[1].ExpertID != "beta" {
			t.Fatalf("catalog order biased selection: %#v err=%v", selection, err)
		}
	}
}

func TestSelectExpertBoundsCandidatesAndIgnoresBlankDuplicateTriggers(t *testing.T) {
	root := t.TempDir()
	experts := []expertDefinition{{ID: "expert-0", Triggers: []string{"", "部署", "部署"}}}
	for i := 1; i < 6; i++ {
		experts = append(experts, expertDefinition{ID: fmt.Sprintf("expert-%d", i), Triggers: []string{"部署"}})
	}
	writeExpertTestCatalog(t, root, experts...)
	selection, err := SelectExpert(root, "部署服务")
	if err != nil || selection.State != "ambiguous" || selection.CandidateCount != 6 || !selection.CandidatesTruncated || len(selection.Candidates) != expertMaxSelectionCandidates {
		t.Fatalf("candidate shortlist is not bounded: %#v err=%v", selection, err)
	}
	for _, candidate := range selection.Candidates {
		if candidate.MatchCount != 1 {
			t.Fatalf("blank or duplicate trigger biased score: %#v", selection)
		}
	}
}

func TestPrepareExpertHandoffRejectsAbstainedSelection(t *testing.T) {
	root := t.TempDir()
	writeExpertTestCatalog(t, root)
	if _, err := PrepareExpertHandoff(root, root, "请解释这个概念", expertTestWorker(), "task-1", "graph-1", "node-1", "attempt-1"); err == nil || !strings.Contains(err.Error(), "requires a selected expert") {
		t.Fatalf("handoff accepted no-match selection: %v", err)
	}
	writeExpertTestCatalog(t, root, expertDefinition{ID: "alpha", Triggers: []string{"部署"}}, expertDefinition{ID: "beta", Triggers: []string{"部署"}})
	if _, err := PrepareExpertHandoff(root, root, "请部署服务", expertTestWorker(), "task-1", "graph-1", "node-1", "attempt-1"); err == nil || !strings.Contains(err.Error(), "requires a selected expert") {
		t.Fatalf("handoff accepted ambiguous selection: %v", err)
	}
}

func TestSelectExpertEnforcesQueryAndCatalogBounds(t *testing.T) {
	if _, err := SelectExpert(expertTestRoot(t), strings.Repeat("x", expertMaxQueryBytes+1)); err == nil || !strings.Contains(err.Error(), "query exceeds") {
		t.Fatalf("oversized query accepted: %v", err)
	}
	root := t.TempDir()
	dir := filepath.Join(root, "capabilities", "experts")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), []byte(strings.Repeat(" ", int(expertMaxCatalogBytes)+1)), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := SelectExpert(root, "测试"); err == nil || !strings.Contains(err.Error(), "catalog exceeds") {
		t.Fatalf("oversized catalog accepted: %v", err)
	}
}

func TestPrepareExpertHandoffBindsCallableCapabilitiesAndIdentity(t *testing.T) {
	root := expertTestRoot(t)
	contract, err := PrepareExpertHandoff(root, root, "测试失败，请修复代码", expertTestWorker(), "task-1", "graph-1", "node-1", "attempt-1")
	if err != nil {
		t.Fatal(err)
	}
	if contract.ExpertID != "code-repair" || contract.ContractSHA256 == "" || len(contract.Bindings) != 2 {
		t.Fatalf("incomplete contract: %#v", contract)
	}
	for _, binding := range contract.Bindings {
		if rank(binding.Status) < rank("callable") || binding.ManifestSHA256 == "" || binding.PrimarySkill == "" {
			t.Fatalf("invalid binding: %#v", binding)
		}
	}
	contract.AttemptID = "attempt-2"
	if _, err := DispatchExpertHandoff(contract, DispatchOptions{OutputDir: t.TempDir(), DryRun: true}); err == nil || !strings.Contains(err.Error(), "hash mismatch") {
		t.Fatalf("tampered contract accepted: %v", err)
	}
}

func TestVerifyExpertReceiptRejectsStaleBindingBeforeRecording(t *testing.T) {
	root := expertTestRoot(t)
	contract, err := PrepareExpertHandoff(root, root, "测试失败，请修复代码", expertTestWorker(), "task-1", "graph-1", "node-1", "attempt-1")
	if err != nil {
		t.Fatal(err)
	}
	receipt := ExpertExecutionReceipt{SchemaVersion: 1, ExpertID: contract.ExpertID, ExpertVersion: contract.ExpertVersion, ContractSHA256: contract.ContractSHA256, WorkspaceSHA256: contract.WorkspaceSHA256, NativeAgentID: "native-agent-1", TaskInstanceID: "task-1", GraphVersion: "graph-1", ExecutionNodeID: "node-1", AttemptID: "old-attempt"}
	if _, err := VerifyAndRecordExpertReceipt(contract, receipt, t.TempDir(), t.TempDir()); err == nil || !strings.Contains(err.Error(), "stale") {
		t.Fatalf("stale receipt accepted: %v", err)
	}
}

func TestForgedNativeIdentityCannotCompleteExecutionGraph(t *testing.T) {
	root := expertTestRoot(t)
	workspace := t.TempDir()
	contract, err := PrepareExpertHandoff(root, workspace, "测试失败，请修复代码", expertTestWorker(), "task-1", "graph-1", "node-1", "attempt-1")
	if err != nil {
		t.Fatal(err)
	}
	result := expertEvidence(t, workspace, "result.txt", "result")
	evidence := expertEvidence(t, workspace, "verify.txt", "verified")
	workerReceipt := validReceipt(contract.Worker)
	workerReceipt.ResultHandle = "wuji-result://sha256/" + result.SHA256
	receipt := ExpertExecutionReceipt{SchemaVersion: 1, ExpertID: contract.ExpertID, ExpertVersion: contract.ExpertVersion, ContractSHA256: contract.ContractSHA256, WorkspaceSHA256: contract.WorkspaceSHA256, NativeAgentID: "forged-native-id", TaskInstanceID: contract.TaskInstanceID, GraphVersion: contract.GraphVersion, ExecutionNodeID: contract.ExecutionNodeID, AttemptID: contract.AttemptID, WorkerReceipt: workerReceipt, Result: result, Evidence: []ExpertEvidenceFile{evidence}}
	graphStore := filepath.Join(t.TempDir(), "graph")
	verification, err := VerifyAndRecordExpertReceipt(contract, receipt, graphStore, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if !verification.Consistent || verification.HostExecutionVerified || verification.GraphMutationAllowed {
		t.Fatalf("forged receipt gained execution authority: %#v", verification)
	}
	if _, err := os.Stat(graphStore); !os.IsNotExist(err) {
		t.Fatalf("consistency check mutated graph store: %v", err)
	}
}

func TestExpertEvidenceBoundsAndSymlinkContainment(t *testing.T) {
	workspace := t.TempDir()
	outside := expertEvidence(t, t.TempDir(), "outside.txt", "x")
	link := filepath.Join(workspace, "link.txt")
	if err := os.Symlink(outside.Path, link); err == nil {
		outside.Path = link
		if _, err := verifyExpertFile(workspace, outside); err == nil || !strings.Contains(err.Error(), "escapes") {
			t.Fatalf("symlink escape accepted: %v", err)
		}
	}
	oversized := ExpertEvidenceFile{Path: filepath.Join(workspace, "unused"), Bytes: expertMaxEvidenceFileBytes + 1}
	if _, err := verifyExpertFile(workspace, oversized); err == nil || !strings.Contains(err.Error(), "byte limit") {
		t.Fatalf("oversized evidence accepted: %v", err)
	}
}

func TestDispatchRejectsStaleLiveManifest(t *testing.T) {
	sourceRoot := expertTestRoot(t)
	tempRoot := t.TempDir()
	for _, id := range []string{"experts", "code", "code-review"} {
		dir := filepath.Join(tempRoot, "capabilities", id)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(filepath.Join(sourceRoot, "capabilities", id, "manifest.json"))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "manifest.json"), data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	contract, err := PrepareExpertHandoff(tempRoot, tempRoot, "测试失败，请修复代码", expertTestWorker(), "task-1", "graph-1", "node-1", "attempt-1")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(tempRoot, "capabilities", "code", "manifest.json")
	data, _ := os.ReadFile(path)
	if err := os.WriteFile(path, append(data, ' '), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := DispatchExpertHandoff(contract, DispatchOptions{OutputDir: t.TempDir(), DryRun: true}); err == nil || !strings.Contains(err.Error(), "changed") {
		t.Fatalf("stale manifest accepted: %v", err)
	}
}

func TestDispatchExpertHandoffVerifiesSourceBearingWorker(t *testing.T) {
	sourceRoot := expertTestRoot(t)
	root := t.TempDir()
	for _, id := range []string{"experts", "code", "code-review"} {
		dir := filepath.Join(root, "capabilities", id)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(filepath.Join(sourceRoot, "capabilities", id, "manifest.json"))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "manifest.json"), data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	skillDir := filepath.Join(root, "capabilities", "code-review", "skills", "wuji-code-review-suite")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	skillPath := filepath.Join(skillDir, "SKILL.md")
	skillData, err := os.ReadFile(filepath.Join(sourceRoot, "capabilities", "code-review", "skills", "wuji-code-review-suite", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(skillPath, skillData, 0o600); err != nil {
		t.Fatal(err)
	}
	manifests, err := LoadManifests(root)
	if err != nil {
		t.Fatal(err)
	}
	manifest, ok := FindManifest(manifests, "code-review")
	if !ok {
		t.Fatal("code-review manifest not found")
	}
	contracts, err := BuildSourceExecutionContracts(manifest, []MountedSource{{
		ID: "wuji-code-review-suite-unified", Entrypoint: "SKILL.md", ActivationReason: "expert-regression",
	}})
	if err != nil {
		t.Fatal(err)
	}
	worker := expertTestWorker()
	worker.SourceExecution = contracts
	worker.SourceExecutionBytes = contracts[0].EntrypointBytes
	contract, err := PrepareExpertHandoff(root, root, "测试失败，请修复代码", worker, "task-source", "graph-1", "node-1", "attempt-1")
	if err != nil {
		t.Fatal(err)
	}
	result, err := DispatchExpertHandoff(contract, DispatchOptions{OutputDir: t.TempDir(), DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.SourceContracts) != 1 || result.SourceContracts[0].EntrypointSHA256 != contracts[0].EntrypointSHA256 {
		t.Fatalf("source contract was not verified: %#v", result.SourceContracts)
	}

	if err := os.WriteFile(skillPath, append(skillData, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := DispatchExpertHandoff(contract, DispatchOptions{OutputDir: t.TempDir(), DryRun: true}); err == nil || !strings.Contains(err.Error(), "changed after routing") {
		t.Fatalf("changed source accepted: %v", err)
	}
}
