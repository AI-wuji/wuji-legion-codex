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

const expertBridgeSchemaVersion = 1
const expertMaxCatalogBytes int64 = 1 << 20
const expertMaxQueryBytes = 16 << 10
const expertMaxSelectionCandidates = 5
const expertMaxEvidenceFiles = 16
const expertMaxEvidenceFileBytes int64 = 4 << 20
const expertMaxEvidenceTotalBytes int64 = 16 << 20

type expertDefinition struct {
	ID             string   `json:"id"`
	Capabilities   []string `json:"capabilities"`
	Triggers       []string `json:"triggers"`
	AntiTriggers   []string `json:"anti_triggers"`
	PromptCompiler string   `json:"prompt_compiler"`
	Workflow       []string `json:"workflow"`
	Verify         []string `json:"verify"`
}

type expertCatalog struct {
	Experts []expertDefinition `json:"experts"`
}

type ExpertWorkflowBinding struct {
	Capability     string `json:"capability"`
	Status         string `json:"status"`
	PrimarySkill   string `json:"primary_skill"`
	ManifestSHA256 string `json:"manifest_sha256"`
}

type ExpertSelection struct {
	State               string                     `json:"state"`
	ExpertID            string                     `json:"expert_id"`
	Matched             []string                   `json:"matched_triggers"`
	RejectedBy          []string                   `json:"rejected_by,omitempty"`
	Reason              string                     `json:"reason"`
	FallbackHint        string                     `json:"fallback_hint,omitempty"`
	Candidates          []ExpertSelectionCandidate `json:"candidates"`
	CandidateCount      int                        `json:"candidate_count"`
	CandidatesTruncated bool                       `json:"candidates_truncated"`
	CatalogSHA256       string                     `json:"catalog_sha256"`
}

// ExpertSelectionCandidate is a compact, bounded summary of a top-scoring expert.
type ExpertSelectionCandidate struct {
	ExpertID   string `json:"expert_id"`
	MatchCount int    `json:"match_count"`
}

type ExpertHandoffContract struct {
	SchemaVersion   int                     `json:"schema_version"`
	ExpertID        string                  `json:"expert_id"`
	ExpertVersion   string                  `json:"expert_version"`
	Workspace       string                  `json:"workspace"`
	WorkspaceSHA256 string                  `json:"workspace_sha256"`
	RepositoryRoot  string                  `json:"repository_root"`
	Worker          WorkerTask              `json:"worker"`
	Workflow        []string                `json:"workflow"`
	Verification    []string                `json:"verification"`
	Bindings        []ExpertWorkflowBinding `json:"bindings"`
	TaskInstanceID  string                  `json:"task_instance_id"`
	GraphVersion    string                  `json:"graph_version"`
	ExecutionNodeID string                  `json:"execution_node_id"`
	AttemptID       string                  `json:"attempt_id"`
	ContractSHA256  string                  `json:"contract_sha256"`
}

type ExpertEvidenceFile struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Bytes  int64  `json:"bytes"`
}

type ExpertExecutionReceipt struct {
	SchemaVersion   int                    `json:"schema_version"`
	ExpertID        string                 `json:"expert_id"`
	ExpertVersion   string                 `json:"expert_version"`
	ContractSHA256  string                 `json:"contract_sha256"`
	WorkspaceSHA256 string                 `json:"workspace_sha256"`
	NativeAgentID   string                 `json:"native_agent_id"`
	TaskInstanceID  string                 `json:"task_instance_id"`
	GraphVersion    string                 `json:"graph_version"`
	ExecutionNodeID string                 `json:"execution_node_id"`
	AttemptID       string                 `json:"attempt_id"`
	WorkerReceipt   WorkerExecutionReceipt `json:"worker_receipt"`
	Result          ExpertEvidenceFile     `json:"result"`
	Evidence        []ExpertEvidenceFile   `json:"evidence"`
}

type ExpertReceiptVerification struct {
	Consistent            bool     `json:"consistent"`
	HostExecutionVerified bool     `json:"host_execution_verified"`
	GraphMutationAllowed  bool     `json:"graph_mutation_allowed"`
	ResultHandle          string   `json:"result_handle"`
	EvidenceHandles       []string `json:"evidence_handles"`
	Reason                string   `json:"reason"`
}

// SelectExpert returns a bounded decision. A lexical miss or anti-trigger is not evidence
// that no skill is needed, so those states explicitly return control to the original route.
func SelectExpert(root, query string) (ExpertSelection, error) {
	catalog, data, err := loadExpertCatalog(root)
	if err != nil {
		return ExpertSelection{}, err
	}
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return ExpertSelection{}, fmt.Errorf("expert selection query is required")
	}
	if len(q) > expertMaxQueryBytes {
		return ExpertSelection{}, fmt.Errorf("expert selection query exceeds %d bytes", expertMaxQueryBytes)
	}
	digest := sha256.Sum256(data)
	selection := ExpertSelection{CatalogSHA256: hex.EncodeToString(digest[:]), Candidates: []ExpertSelectionCandidate{}}
	top := make([]expertSelectionMatch, 0, len(catalog.Experts))
	bestScore := 0
	anyVeto := false
	for _, expert := range catalog.Experts {
		veto := uniqueMatchingSignals(q, expert.AntiTriggers)
		if len(veto) > 0 {
			anyVeto = true
			continue
		}
		matched := uniqueMatchingSignals(q, expert.Triggers)
		if len(matched) == 0 {
			continue
		}
		if len(matched) > bestScore {
			bestScore = len(matched)
			top = top[:0]
		}
		if len(matched) == bestScore {
			top = append(top, expertSelectionMatch{expert: expert, matched: matched})
		}
	}
	if len(top) == 0 {
		selection.State = "none"
		selection.FallbackHint = "return-to-aji-original-route"
		if anyVeto {
			selection.Reason = "anti-trigger"
		} else {
			selection.Reason = "no-match"
		}
		return selection, nil
	}
	sort.Slice(top, func(i, j int) bool { return top[i].expert.ID < top[j].expert.ID })
	selection.CandidateCount = len(top)
	selection.CandidatesTruncated = len(top) > expertMaxSelectionCandidates
	for i, candidate := range top {
		if i == expertMaxSelectionCandidates {
			break
		}
		selection.Candidates = append(selection.Candidates, ExpertSelectionCandidate{ExpertID: candidate.expert.ID, MatchCount: len(candidate.matched)})
	}
	if len(top) > 1 {
		selection.State = "ambiguous"
		selection.Reason = "tie"
		selection.FallbackHint = "return-to-aji-original-route"
		return selection, nil
	}
	selection.State = "selected"
	selection.Reason = "unique-match"
	selection.ExpertID = top[0].expert.ID
	selection.Matched = append([]string(nil), top[0].matched...)
	return selection, nil
}

type expertSelectionMatch struct {
	expert  expertDefinition
	matched []string
}

// PrepareExpertHandoff validates callable workflow bindings and returns an immutable native-host contract.
func PrepareExpertHandoff(root, workspace, query string, worker WorkerTask, taskInstanceID, graphVersion, executionNodeID, attemptID string) (ExpertHandoffContract, error) {
	selection, err := SelectExpert(root, query)
	if err != nil {
		return ExpertHandoffContract{}, err
	}
	if selection.State != "selected" || selection.ExpertID == "" {
		return ExpertHandoffContract{}, fmt.Errorf("expert handoff requires a selected expert; selection is %s (%s)", selection.State, selection.Reason)
	}
	catalog, _, err := loadExpertCatalog(root)
	if err != nil {
		return ExpertHandoffContract{}, err
	}
	var selected expertDefinition
	for _, candidate := range catalog.Experts {
		if candidate.ID == selection.ExpertID {
			selected = candidate
			break
		}
	}
	if worker.ID == "" || worker.SessionKey == "" {
		return ExpertHandoffContract{}, fmt.Errorf("expert handoff requires worker and session identity")
	}
	if err := validateWorkerExecutionPolicy(worker); err != nil {
		return ExpertHandoffContract{}, err
	}
	if taskInstanceID == "" || graphVersion == "" || executionNodeID == "" || attemptID == "" {
		return ExpertHandoffContract{}, fmt.Errorf("expert handoff requires task, graph, execution-node and attempt identity")
	}
	absWorkspace, err := filepath.Abs(workspace)
	if err != nil {
		return ExpertHandoffContract{}, err
	}
	info, err := os.Stat(absWorkspace)
	if err != nil || !info.IsDir() {
		return ExpertHandoffContract{}, fmt.Errorf("expert workspace is not a directory")
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return ExpertHandoffContract{}, err
	}
	bindings := make([]ExpertWorkflowBinding, 0, len(selected.Capabilities))
	for _, id := range selected.Capabilities {
		manifestPath := filepath.Join(root, "capabilities", id, "manifest.json")
		manifestData, err := os.ReadFile(manifestPath)
		if err != nil {
			return ExpertHandoffContract{}, err
		}
		var manifest Manifest
		if err := json.Unmarshal(manifestData, &manifest); err != nil {
			return ExpertHandoffContract{}, err
		}
		manifest.Root = root
		if err := ValidateManifest(manifest); err != nil {
			return ExpertHandoffContract{}, fmt.Errorf("capability %s: %w", id, err)
		}
		if rank(manifest.Status) < rank("callable") {
			return ExpertHandoffContract{}, fmt.Errorf("expert %s requires callable capability %s", selected.ID, id)
		}
		d := sha256.Sum256(manifestData)
		bindings = append(bindings, ExpertWorkflowBinding{Capability: id, Status: manifest.Status, PrimarySkill: manifest.PrimarySkill, ManifestSHA256: hex.EncodeToString(d[:])})
	}
	workspaceDigest := sha256.Sum256([]byte(filepath.Clean(absWorkspace)))
	contract := ExpertHandoffContract{SchemaVersion: expertBridgeSchemaVersion, ExpertID: selected.ID, ExpertVersion: selection.CatalogSHA256, Workspace: absWorkspace, WorkspaceSHA256: hex.EncodeToString(workspaceDigest[:]), RepositoryRoot: absRoot, Worker: worker, Workflow: append([]string(nil), selected.Workflow...), Verification: append([]string(nil), selected.Verify...), Bindings: bindings, TaskInstanceID: taskInstanceID, GraphVersion: graphVersion, ExecutionNodeID: executionNodeID, AttemptID: attemptID}
	contract.ContractSHA256, err = hashExpertContract(contract)
	return contract, err
}

func DispatchExpertHandoff(contract ExpertHandoffContract, options DispatchOptions) (DispatchResult, error) {
	if err := validateExpertContract(contract); err != nil {
		return DispatchResult{}, err
	}
	if err := validateLiveExpertBindings(contract); err != nil {
		return DispatchResult{}, err
	}
	manifests, err := LoadManifests(contract.RepositoryRoot)
	if err != nil {
		return DispatchResult{}, err
	}
	options.Workspace = contract.Workspace
	options.TrustedManifests = manifests
	return DispatchWorker(contract.Worker, options)
}

// VerifyAndRecordExpertReceipt is consistency-only. Caller-supplied identity and hashes
// are not independent native-host attestation and therefore never authorize graph mutation.
func VerifyAndRecordExpertReceipt(contract ExpertHandoffContract, receipt ExpertExecutionReceipt, executionStore, requirementStore string) (ExpertReceiptVerification, error) {
	_ = executionStore
	_ = requirementStore
	if err := validateExpertContract(contract); err != nil {
		return ExpertReceiptVerification{}, err
	}
	if err := validateLiveExpertBindings(contract); err != nil {
		return ExpertReceiptVerification{}, err
	}
	if receipt.SchemaVersion != expertBridgeSchemaVersion || receipt.ExpertID != contract.ExpertID || receipt.ExpertVersion != contract.ExpertVersion || receipt.ContractSHA256 != contract.ContractSHA256 || receipt.WorkspaceSHA256 != contract.WorkspaceSHA256 || receipt.TaskInstanceID != contract.TaskInstanceID || receipt.GraphVersion != contract.GraphVersion || receipt.ExecutionNodeID != contract.ExecutionNodeID || receipt.AttemptID != contract.AttemptID {
		return ExpertReceiptVerification{}, fmt.Errorf("expert receipt identity or runtime binding is stale")
	}
	if strings.TrimSpace(receipt.NativeAgentID) == "" {
		return ExpertReceiptVerification{}, fmt.Errorf("expert receipt lacks native agent identity")
	}
	if err := ValidateWorkerReceiptConsistency(contract.Worker, receipt.WorkerReceipt); err != nil {
		return ExpertReceiptVerification{}, err
	}
	if len(receipt.Evidence) == 0 || len(receipt.Evidence) > expertMaxEvidenceFiles {
		return ExpertReceiptVerification{}, fmt.Errorf("expert receipt evidence count is outside 1..%d", expertMaxEvidenceFiles)
	}
	resultHandle, err := verifyExpertFile(contract.Workspace, receipt.Result)
	if err != nil {
		return ExpertReceiptVerification{}, fmt.Errorf("result evidence: %w", err)
	}
	if resultHandle != receipt.WorkerReceipt.ResultHandle {
		return ExpertReceiptVerification{}, fmt.Errorf("result file does not match worker result handle")
	}
	verification := make([]string, 0, len(receipt.Evidence))
	totalBytes := receipt.Result.Bytes
	for _, file := range receipt.Evidence {
		totalBytes += file.Bytes
		if totalBytes > expertMaxEvidenceTotalBytes {
			return ExpertReceiptVerification{}, fmt.Errorf("expert receipt evidence exceeds total byte limit")
		}
		handle, err := verifyExpertFile(contract.Workspace, file)
		if err != nil {
			return ExpertReceiptVerification{}, fmt.Errorf("verification evidence: %w", err)
		}
		verification = append(verification, strings.Replace(handle, "wuji-result://", "wuji-evidence://", 1))
	}
	return ExpertReceiptVerification{Consistent: true, HostExecutionVerified: false, GraphMutationAllowed: false, ResultHandle: resultHandle, EvidenceHandles: verification, Reason: "receipt is caller-supplied consistency evidence; independent native-host attestation is unavailable"}, nil
}

func loadExpertCatalog(root string) (expertCatalog, []byte, error) {
	path := filepath.Join(root, "capabilities", "experts", "manifest.json")
	file, err := os.Open(path)
	if err != nil {
		return expertCatalog{}, nil, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return expertCatalog{}, nil, err
	}
	if info.Size() > expertMaxCatalogBytes {
		return expertCatalog{}, nil, fmt.Errorf("expert catalog exceeds %d bytes", expertMaxCatalogBytes)
	}
	data, err := io.ReadAll(io.LimitReader(file, expertMaxCatalogBytes+1))
	if err != nil {
		return expertCatalog{}, nil, err
	}
	if int64(len(data)) > expertMaxCatalogBytes {
		return expertCatalog{}, nil, fmt.Errorf("expert catalog exceeds %d bytes", expertMaxCatalogBytes)
	}
	var catalog expertCatalog
	if err := json.Unmarshal(data, &catalog); err != nil {
		return catalog, nil, err
	}
	if len(catalog.Experts) != 6 {
		return catalog, nil, fmt.Errorf("expert catalog must contain exactly six contracts")
	}
	return catalog, data, nil
}

func containsFold(query, signal string) bool {
	signal = strings.ToLower(strings.TrimSpace(signal))
	return signal != "" && strings.Contains(query, signal)
}

func uniqueMatchingSignals(query string, signals []string) []string {
	seen := make(map[string]struct{}, len(signals))
	matched := make([]string, 0, len(signals))
	for _, signal := range signals {
		normalized := strings.ToLower(strings.TrimSpace(signal))
		if normalized == "" || !containsFold(query, normalized) {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		matched = append(matched, signal)
	}
	return matched
}

func hashExpertContract(contract ExpertHandoffContract) (string, error) {
	contract.ContractSHA256 = ""
	data, err := json.Marshal(contract)
	if err != nil {
		return "", err
	}
	d := sha256.Sum256(data)
	return hex.EncodeToString(d[:]), nil
}

func validateExpertContract(contract ExpertHandoffContract) error {
	if contract.SchemaVersion != expertBridgeSchemaVersion {
		return fmt.Errorf("unsupported expert contract schema")
	}
	want, err := hashExpertContract(contract)
	if err != nil {
		return err
	}
	if want != contract.ContractSHA256 {
		return fmt.Errorf("expert contract hash mismatch")
	}
	return validateWorkerExecutionPolicy(contract.Worker)
}

func validateLiveExpertBindings(contract ExpertHandoffContract) error {
	_, catalogData, err := loadExpertCatalog(contract.RepositoryRoot)
	if err != nil {
		return err
	}
	d := sha256.Sum256(catalogData)
	if hex.EncodeToString(d[:]) != contract.ExpertVersion {
		return fmt.Errorf("expert catalog changed after contract preparation")
	}
	for _, binding := range contract.Bindings {
		data, err := os.ReadFile(filepath.Join(contract.RepositoryRoot, "capabilities", binding.Capability, "manifest.json"))
		if err != nil {
			return err
		}
		digest := sha256.Sum256(data)
		if hex.EncodeToString(digest[:]) != binding.ManifestSHA256 {
			return fmt.Errorf("capability %s manifest changed after contract preparation", binding.Capability)
		}
	}
	return nil
}

func verifyExpertFile(workspace string, evidence ExpertEvidenceFile) (string, error) {
	if evidence.Bytes < 0 || evidence.Bytes > expertMaxEvidenceFileBytes {
		return "", fmt.Errorf("evidence file exceeds byte limit")
	}
	abs, err := filepath.Abs(evidence.Path)
	if err != nil {
		return "", err
	}
	canonicalWorkspace, err := filepath.EvalSymlinks(workspace)
	if err != nil {
		return "", err
	}
	canonicalEvidence, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(canonicalWorkspace, canonicalEvidence)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("evidence path escapes workspace")
	}
	info, err := os.Stat(canonicalEvidence)
	if err != nil || !info.Mode().IsRegular() {
		return "", fmt.Errorf("evidence must be a regular file")
	}
	if info.Size() > expertMaxEvidenceFileBytes {
		return "", fmt.Errorf("evidence file exceeds byte limit")
	}
	data, err := os.ReadFile(canonicalEvidence)
	if err != nil {
		return "", err
	}
	d := sha256.Sum256(data)
	actual := hex.EncodeToString(d[:])
	if evidence.SHA256 != actual || evidence.Bytes != int64(len(data)) {
		return "", fmt.Errorf("evidence hash or size mismatch")
	}
	return "wuji-result://sha256/" + actual, nil
}
