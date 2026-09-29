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
const expertMaxCatalogEntries = 128
const expertMaxEvidenceFiles = 16
const expertMaxEvidenceFileBytes int64 = 4 << 20
const expertMaxEvidenceTotalBytes int64 = 16 << 20

type expertDefinition struct {
	ID                string         `json:"id"`
	Name              string         `json:"name"`
	Commander         string         `json:"commander"`
	RoleType          string         `json:"role_type"`
	DecisionMode      string         `json:"decision_mode"`
	TemplateBasis     []string       `json:"template_basis,omitempty"`
	Capabilities      []string       `json:"capabilities"`
	SourceIDs         []string       `json:"source_ids,omitempty"`
	SkillSources      []string       `json:"skill_sources,omitempty"`
	Mission           string         `json:"mission"`
	Inputs            []string       `json:"inputs"`
	Constraints       []string       `json:"constraints"`
	ToolPolicy        []string       `json:"tool_policy"`
	Outputs           []string       `json:"outputs"`
	Boundaries        []string       `json:"boundaries"`
	Escalation        []string       `json:"escalation"`
	Triggers          []string       `json:"triggers"`
	AntiTriggers      []string       `json:"anti_triggers"`
	PromptCompiler    string         `json:"prompt_compiler"`
	Workflow          []string       `json:"workflow"`
	Verify            []string       `json:"verify"`
	Acceptance        []string       `json:"acceptance"`
	PonyTailRules     []string       `json:"ponytail_rules"`
	WhiteHatChecks    []string       `json:"white_hat_checks"`
	FailureStates     []string       `json:"failure_states"`
	PromotionEvidence []string       `json:"promotion_evidence"`
	Methods           []expertMethod `json:"methods,omitempty"`
}

// A method is a small distilled instruction owned by one expert, not a second
// Skill router or permission to execute its source package.
type expertMethod struct {
	ID           string   `json:"id"`
	Sources      []string `json:"sources"`
	Signals      []string `json:"signals"`
	Instruction  string   `json:"instruction"`
	Verification string   `json:"verification"`
}

type expertAppliedMethod struct {
	ID           string `json:"id"`
	Instruction  string `json:"instruction"`
	Verification string `json:"verification"`
}

type expertTeamStage struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Mode        string   `json:"mode"`
	ExpertIDs   []string `json:"expert_ids"`
	DependsOn   []string `json:"depends_on,omitempty"`
	Deliverable string   `json:"deliverable"`
	Gates       []string `json:"gates"`
}

// expertCommander is the domain commander and the expert team itself. It is
// deliberately not another routing layer: one commander owns one bounded
// domain, its mission, and the experts that execute its stable sub-tasks.
type expertCommander struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	RoleType        string            `json:"role_type,omitempty"`
	DecisionMode    string            `json:"decision_mode,omitempty"`
	TemplateBasis   []string          `json:"template_basis,omitempty"`
	Capabilities    []string          `json:"capabilities"`
	Mission         string            `json:"mission"`
	TeamPrompt      string            `json:"team_prompt,omitempty"`
	LeaderExpertID  string            `json:"leader_expert_id,omitempty"`
	Orchestrator    string            `json:"orchestrator,omitempty"`
	Knowledge       []string          `json:"knowledge"`
	Workflow        []string          `json:"workflow"`
	Stages          []expertTeamStage `json:"stages,omitempty"`
	ParallelPolicy  string            `json:"parallel_policy,omitempty"`
	Outputs         []string          `json:"outputs"`
	Acceptance      []string          `json:"acceptance"`
	Boundaries      []string          `json:"boundaries"`
	PonyTailRules   []string          `json:"ponytail_rules,omitempty"`
	WhiteHatChecks  []string          `json:"white_hat_checks,omitempty"`
	TemplateSources []string          `json:"template_sources,omitempty"`
	ExpertIDs       []string          `json:"expert_ids"`
}

type expertCatalog struct {
	Commanders []expertCommander  `json:"commanders"`
	Experts    []expertDefinition `json:"experts"`
}

type ExpertWorkflowBinding struct {
	Capability     string `json:"capability"`
	Status         string `json:"status"`
	PrimarySkill   string `json:"primary_skill"`
	ManifestSHA256 string `json:"manifest_sha256"`
}

type ExpertSelection struct {
	State               string                     `json:"state"`
	Commander           string                     `json:"commander,omitempty"`
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

// ExpertRouteDecision is a staff-only, capability-scoped commander decision.
// It describes a prepared worker contract, not a native expert execution.
type ExpertRouteDecision struct {
	Commander       string                     `json:"commander"`
	CommanderName   string                     `json:"commander_name"`
	Capability      string                     `json:"capability"`
	Mission         string                     `json:"mission"`
	Workflow        []string                   `json:"workflow"`
	Acceptance      []string                   `json:"acceptance"`
	Experts         []string                   `json:"experts"`
	TeamPrompt      string                     `json:"team_prompt,omitempty"`
	LeaderExpertID  string                     `json:"leader_expert_id,omitempty"`
	Stages          []ExpertTeamStage          `json:"stages,omitempty"`
	ParallelPolicy  string                     `json:"parallel_policy,omitempty"`
	TemplateSources []string                   `json:"template_sources,omitempty"`
	MoE             string                     `json:"moe"`
	TaskGraph       ExpertTaskGraph            `json:"task_graph"`
	ExpertGraphs    map[string]ExpertTaskGraph `json:"expert_graphs,omitempty"`
	Selection       ExpertSelection            `json:"selection"`
	BoundWorkers    []string                   `json:"bound_workers,omitempty"`
	ExecutionStatus string                     `json:"execution_status"`
}

// ExpertTeamStage is the public route representation of a commander/teacher
// team's bounded SOP. A stage is either parallel or serial; downstream stages
// cannot run until every dependency gate has returned.
type ExpertTeamStage struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Mode        string   `json:"mode"`
	ExpertIDs   []string `json:"expert_ids"`
	DependsOn   []string `json:"depends_on,omitempty"`
	Deliverable string   `json:"deliverable"`
	Gates       []string `json:"gates"`
}

// ExpertSelectionCandidate is a compact, bounded summary of a top-scoring expert.
type ExpertSelectionCandidate struct {
	ExpertID   string `json:"expert_id"`
	MatchCount int    `json:"match_count"`
}

type ExpertHandoffContract struct {
	SchemaVersion   int                     `json:"schema_version"`
	Commander       string                  `json:"commander"`
	ExpertID        string                  `json:"expert_id"`
	MoE             string                  `json:"moe"`
	TaskGraph       ExpertTaskGraph         `json:"task_graph"`
	RoleID          string                  `json:"role_id"`
	RoleKind        string                  `json:"role_kind"`
	ParentGraphID   string                  `json:"parent_graph_id,omitempty"`
	TaskGraphID     string                  `json:"task_graph_id"`
	TaskGraphSHA256 string                  `json:"task_graph_sha256"`
	PonyTail        string                  `json:"ponytail"`
	ExpertVersion   string                  `json:"expert_version"`
	Workspace       string                  `json:"workspace"`
	WorkspaceSHA256 string                  `json:"workspace_sha256"`
	RepositoryRoot  string                  `json:"repository_root"`
	Worker          WorkerTask              `json:"worker"`
	Workflow        []string                `json:"workflow"`
	Verification    []string                `json:"verification"`
	Acceptance      []string                `json:"acceptance"`
	Constraints     []string                `json:"constraints"`
	ToolPolicy      []string                `json:"tool_policy"`
	PonyTailRules   []string                `json:"ponytail_rules"`
	WhiteHatChecks  []string                `json:"white_hat_checks"`
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
	return selectExpert(root, query, "")
}

func SelectExpertForCapability(root, query, capability string) (ExpertSelection, error) {
	if !capabilityIDPattern.MatchString(capability) {
		return ExpertSelection{}, fmt.Errorf("invalid commander capability %q", capability)
	}
	return selectExpert(root, query, capability)
}

func selectExpert(root, query, capability string) (ExpertSelection, error) {
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
	commanderID := ""
	if capability != "" {
		commander, ok := catalog.commanderForCapability(capability)
		if !ok {
			selection.State = "none"
			selection.Reason = "no-commander"
			selection.FallbackHint = "return-to-aji-original-route"
			return selection, nil
		}
		commanderID = commander.ID
		selection.Commander = commander.ID
	}
	top := make([]expertSelectionMatch, 0, len(catalog.Experts))
	bestScore := 0
	bestSpecificity := 0
	anyVeto := false
	for _, expert := range catalog.Experts {
		if commanderID != "" && expert.Commander != commanderID {
			continue
		}
		if capability != "" && !containsString(expert.Capabilities, capability) {
			continue
		}
		veto := uniqueMatchingSignals(q, expert.AntiTriggers)
		if len(veto) > 0 {
			anyVeto = true
			continue
		}
		matched := uniqueMatchingSignals(q, expert.Triggers)
		if len(matched) == 0 {
			for _, method := range expert.Methods {
				matched = append(matched, uniqueMatchingSignals(q, method.Signals)...)
			}
		}
		if len(matched) == 0 {
			continue
		}
		specificity := 0
		for _, signal := range matched {
			if length := len([]rune(signal)); length > specificity {
				specificity = length
			}
		}
		if len(matched) > bestScore || (len(matched) == bestScore && specificity > bestSpecificity) {
			bestScore = len(matched)
			bestSpecificity = specificity
			top = top[:0]
		}
		if len(matched) == bestScore && specificity == bestSpecificity {
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
	selection.Commander = top[0].expert.Commander
	selection.Matched = append([]string(nil), top[0].matched...)
	return selection, nil
}

func expertCommanderForCapability(root, capability string) (expertCommander, bool, error) {
	catalog, _, err := loadExpertCatalog(root)
	if err != nil {
		return expertCommander{}, false, err
	}
	commander, ok := catalog.commanderForCapability(capability)
	return commander, ok, nil
}

func publicTeamStages(stages []expertTeamStage) []ExpertTeamStage {
	if len(stages) == 0 {
		return nil
	}
	result := make([]ExpertTeamStage, 0, len(stages))
	for _, stage := range stages {
		result = append(result, ExpertTeamStage{
			ID:          stage.ID,
			Name:        stage.Name,
			Mode:        stage.Mode,
			ExpertIDs:   append([]string(nil), stage.ExpertIDs...),
			DependsOn:   append([]string(nil), stage.DependsOn...),
			Deliverable: stage.Deliverable,
			Gates:       append([]string(nil), stage.Gates...),
		})
	}
	return result
}

// bindExpertWorker places the actual professional instructions inside the
// hashed task contract that the native-host prompt consumes.
func bindExpertWorker(worker *WorkerTask, expert expertDefinition, commander expertCommander, capability, catalogSHA256 string) error {
	var task workerContract
	if err := json.Unmarshal([]byte(worker.TaskContract), &task); err != nil {
		return err
	}
	task.Expert = &expertTaskDirective{
		Commander: commander.ID, Capability: capability, ID: expert.ID, CatalogSHA256: catalogSHA256, PromptCompiler: expert.PromptCompiler,
		Inputs:         append([]string(nil), expert.Inputs...),
		Outputs:        append([]string(nil), expert.Outputs...),
		Workflow:       append([]string(nil), expert.Workflow...),
		Verification:   append([]string(nil), expert.Verify...),
		Acceptance:     append([]string(nil), expert.Acceptance...),
		Constraints:    append([]string(nil), expert.Constraints...),
		ToolPolicy:     append([]string(nil), expert.ToolPolicy...),
		PonyTailRules:  append([]string(nil), expert.PonyTailRules...),
		WhiteHatChecks: append([]string(nil), expert.WhiteHatChecks...),
		TeamMission:    commander.Mission,
		TeamAcceptance: append([]string(nil), commander.Acceptance...),
		Methods:        selectedExpertMethods(expert, task.Objective),
	}
	if len(task.Expert.Acceptance) == 0 {
		task.Expert.Acceptance = []string{"return bounded evidence"}
	}
	if len(task.Expert.Constraints) == 0 {
		task.Expert.Constraints = []string{"stay within assigned scope"}
	}
	if len(task.Expert.ToolPolicy) == 0 {
		task.Expert.ToolPolicy = []string{"use only authorized tools"}
	}
	if len(task.Expert.PonyTailRules) == 0 {
		task.Expert.PonyTailRules = []string{"use the minimum correct path"}
	}
	if len(task.Expert.WhiteHatChecks) == 0 {
		task.Expert.WhiteHatChecks = []string{"do not claim unverified completion"}
	}
	data, err := json.Marshal(task)
	if err != nil {
		return err
	}
	if len(data) > maxTaskContractBytes {
		return fmt.Errorf("expert %s task contract exceeds %d bytes", expert.ID, maxTaskContractBytes)
	}
	worker.TaskContract = string(data)
	worker.TaskContractSHA256 = sha256Hex(data)
	worker.AllocatedTaskContractBytes = len(data)
	return nil
}

func selectedExpertMethods(expert expertDefinition, query string) []expertAppliedMethod {
	query = strings.ToLower(query)
	var selected []expertAppliedMethod
	for _, method := range expert.Methods {
		if len(uniqueMatchingSignals(query, method.Signals)) == 0 {
			continue
		}
		selected = append(selected, expertAppliedMethod{
			ID: method.ID, Instruction: method.Instruction, Verification: method.Verification,
		})
	}
	// A bounded task never loads a whole expert's methods. Additional matching
	// methods remain cold; the specialist can request a narrower follow-up.
	if len(selected) > 2 {
		selected = selected[:2]
	}
	return selected
}

func equalExpertMethods(a, b []expertAppliedMethod) bool {
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

func compactExpertRules(rules, fallback []string) []string {
	if len(rules) == 0 {
		return append([]string(nil), fallback...)
	}
	result := make([]string, 0, len(rules))
	for _, rule := range rules[:minInt(len(rules), 1)] {
		rule = strings.TrimSpace(rule)
		if rule != "" {
			if len([]byte(rule)) > 96 {
				rule = string([]byte(rule)[:96])
			}
			result = append(result, rule)
		}
	}
	if len(result) == 0 {
		return append([]string(nil), fallback...)
	}
	return result
}

type expertSelectionMatch struct {
	expert  expertDefinition
	matched []string
}

// PrepareExpertHandoff validates callable workflow bindings and returns an immutable native-host contract.
func PrepareExpertHandoff(root, workspace, query string, worker WorkerTask, taskInstanceID, graphVersion, executionNodeID, attemptID string) (ExpertHandoffContract, error) {
	var prefix struct {
		Capability string `json:"capability"`
	}
	if err := json.Unmarshal([]byte(worker.StableCapabilityPrefix), &prefix); err != nil {
		return ExpertHandoffContract{}, fmt.Errorf("expert worker capability prefix is invalid: %w", err)
	}
	selection, err := SelectExpertForCapability(root, query, prefix.Capability)
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
	commander, ok := catalog.commanderForCapability(prefix.Capability)
	if !ok || selected.Commander != commander.ID || selection.Commander != commander.ID {
		return ExpertHandoffContract{}, fmt.Errorf("expert %s is outside the selected commander", selected.ID)
	}
	if worker.ID == "" || worker.SessionKey == "" {
		return ExpertHandoffContract{}, fmt.Errorf("expert handoff requires worker and session identity")
	}
	if err := validateWorkerExecutionPolicy(worker); err != nil {
		return ExpertHandoffContract{}, err
	}
	if err := validateExpertPonytail(worker); err != nil {
		return ExpertHandoffContract{}, err
	}
	if !containsString(selected.Capabilities, prefix.Capability) {
		return ExpertHandoffContract{}, fmt.Errorf("expert %s is outside the worker commander capability", selected.ID)
	}
	var task workerContract
	if err := json.Unmarshal([]byte(worker.TaskContract), &task); err != nil {
		return ExpertHandoffContract{}, err
	}
	if task.Expert != nil && (task.Expert.ID != selected.ID || task.Expert.Commander != commander.ID || task.Expert.Capability != prefix.Capability) {
		return ExpertHandoffContract{}, fmt.Errorf("expert handoff conflicts with selected worker directive")
	}
	for _, sourceID := range selected.SourceIDs {
		found := false
		for _, source := range worker.SourceExecution {
			if source.SourceID == sourceID && source.Capability == prefix.Capability {
				found = true
				break
			}
		}
		if !found {
			return ExpertHandoffContract{}, fmt.Errorf("expert %s requires selected source %s", selected.ID, sourceID)
		}
	}
	if err := bindExpertWorker(&worker, selected, commander, prefix.Capability, selection.CatalogSHA256); err != nil {
		return ExpertHandoffContract{}, err
	}
	if err := json.Unmarshal([]byte(worker.TaskContract), &task); err != nil {
		return ExpertHandoffContract{}, fmt.Errorf("expert task contract refresh failed: %w", err)
	}
	expertGraph, err := makeExpertRoleGraph(selected.ID, query, graphVersion, worker.ParentGraphID, worker.ContextHandles)
	if err != nil {
		return ExpertHandoffContract{}, err
	}
	if worker.TaskGraphID == "" {
		child, graphErr := makeWorkerRoleGraph(worker, graphVersion, "expert:"+selected.ID, expertGraph.GraphID, worker.ContextHandles)
		if graphErr != nil {
			return ExpertHandoffContract{}, graphErr
		}
		parent, graphErr := makeRoleTaskGraph("expert:"+selected.ID, "expert", query, graphVersion, worker.ParentGraphID, "expert-intake", []RoleTaskNode{
			roleGraphNode("expert-intake", "expert:"+selected.ID, "intake", "receive commander assignment", nil, "", false),
			roleGraphNode("expert-execute", "expert:"+selected.ID, "execute", "perform bounded specialist work", []string{"expert-intake"}, "", false),
		}, nil, worker.ContextHandles)
		if graphErr != nil {
			return ExpertHandoffContract{}, graphErr
		}
		if graphErr := annotateWorkerRoleGraph(&worker, "expert:"+selected.ID, "expert-worker", parent, child); graphErr != nil {
			return ExpertHandoffContract{}, graphErr
		}
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
	contract := ExpertHandoffContract{
		SchemaVersion: expertBridgeSchemaVersion, Commander: commander.ID, ExpertID: selected.ID,
		MoE: expertGraph.MoE, TaskGraph: expertGraph, RoleID: expertGraph.RoleID, RoleKind: expertGraph.RoleKind,
		ParentGraphID: worker.ParentGraphID, TaskGraphID: expertGraph.GraphID, TaskGraphSHA256: expertGraph.SHA256,
		PonyTail: ponyTailDoctrine, ExpertVersion: selection.CatalogSHA256, Workspace: absWorkspace,
		WorkspaceSHA256: hex.EncodeToString(workspaceDigest[:]), RepositoryRoot: absRoot, Worker: worker,
		Workflow: append([]string(nil), selected.Workflow...), Verification: append([]string(nil), selected.Verify...),
		Acceptance:     append([]string(nil), task.Expert.Acceptance...),
		Constraints:    append([]string(nil), task.Expert.Constraints...),
		ToolPolicy:     append([]string(nil), task.Expert.ToolPolicy...),
		PonyTailRules:  append([]string(nil), task.Expert.PonyTailRules...),
		WhiteHatChecks: append([]string(nil), task.Expert.WhiteHatChecks...),
		Bindings:       bindings, TaskInstanceID: taskInstanceID, GraphVersion: graphVersion,
		ExecutionNodeID: executionNodeID, AttemptID: attemptID,
	}
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
	var catalogFields map[string]json.RawMessage
	if err := json.Unmarshal(data, &catalogFields); err != nil {
		return catalog, nil, err
	}
	if _, exists := catalogFields["extensions"]; exists {
		return catalog, nil, fmt.Errorf("expert catalog extensions are not supported; use one fused manifest")
	}
	if len(catalog.Commanders) == 0 || len(catalog.Commanders) > expertMaxCatalogEntries {
		return catalog, nil, fmt.Errorf("expert catalog requires 1..%d commanders", expertMaxCatalogEntries)
	}
	if len(catalog.Experts) == 0 || len(catalog.Experts) > expertMaxCatalogEntries {
		return catalog, nil, fmt.Errorf("expert catalog requires 1..%d contracts", expertMaxCatalogEntries)
	}
	commanderByID := make(map[string]expertCommander, len(catalog.Commanders))
	claimedCapabilities := make(map[string]string)
	for _, commander := range catalog.Commanders {
		if !capabilityIDPattern.MatchString(commander.ID) || commanderByID[commander.ID].ID != "" ||
			commander.RoleType != "expert-team" || commander.DecisionMode != "sparse-moe" ||
			strings.TrimSpace(commander.Name) == "" || len(commander.Capabilities) == 0 ||
			strings.TrimSpace(commander.Mission) == "" || len(commander.Knowledge) == 0 ||
			len(commander.Workflow) == 0 || len(commander.Outputs) == 0 ||
			len(commander.Acceptance) == 0 || len(commander.Boundaries) == 0 ||
			len(commander.ExpertIDs) == 0 {
			return catalog, nil, fmt.Errorf("invalid or duplicate expert commander %q", commander.ID)
		}
		commanderByID[commander.ID] = commander
		seenCapabilities := map[string]bool{}
		for _, capability := range commander.Capabilities {
			if !capabilityIDPattern.MatchString(capability) || seenCapabilities[capability] {
				return catalog, nil, fmt.Errorf("commander %s has invalid or duplicate capability %q", commander.ID, capability)
			}
			if previous, exists := claimedCapabilities[capability]; exists {
				return catalog, nil, fmt.Errorf("capability %s belongs to commanders %s and %s", capability, previous, commander.ID)
			}
			seenCapabilities[capability] = true
			claimedCapabilities[capability] = commander.ID
		}
		seenExperts := map[string]bool{}
		for _, expertID := range commander.ExpertIDs {
			if !capabilityIDPattern.MatchString(expertID) || seenExperts[expertID] {
				return catalog, nil, fmt.Errorf("commander %s has invalid or duplicate expert %q", commander.ID, expertID)
			}
			seenExperts[expertID] = true
		}
		if commander.LeaderExpertID != "" && !seenExperts[commander.LeaderExpertID] {
			return catalog, nil, fmt.Errorf("commander %s leader %s is outside the team", commander.ID, commander.LeaderExpertID)
		}
		if len(commander.Stages) > 0 &&
			(commander.RoleType != "expert-team" || commander.DecisionMode != "sparse-moe" ||
				strings.TrimSpace(commander.TeamPrompt) == "" ||
				len(commander.PonyTailRules) == 0 || len(commander.WhiteHatChecks) == 0) {
			return catalog, nil, fmt.Errorf("commander %s has an incomplete team contract", commander.ID)
		}
		stages := make(map[string]expertTeamStage, len(commander.Stages))
		for _, stage := range commander.Stages {
			if !capabilityIDPattern.MatchString(stage.ID) || stages[stage.ID].ID != "" ||
				strings.TrimSpace(stage.Name) == "" || strings.TrimSpace(stage.Deliverable) == "" ||
				len(stage.Gates) == 0 || (stage.Mode != "serial" && stage.Mode != "parallel") {
				return catalog, nil, fmt.Errorf("commander %s has invalid or duplicate stage %q", commander.ID, stage.ID)
			}
			stages[stage.ID] = stage
			stageExperts := make(map[string]bool, len(stage.ExpertIDs))
			for _, id := range stage.ExpertIDs {
				if !seenExperts[id] || stageExperts[id] {
					return catalog, nil, fmt.Errorf("commander %s stage %s references an absent or repeated expert %q", commander.ID, stage.ID, id)
				}
				stageExperts[id] = true
			}
		}
		stageNodes := make([]RoleTaskNode, 0, len(commander.Stages))
		for _, stage := range commander.Stages {
			for _, dependency := range stage.DependsOn {
				if _, ok := stages[dependency]; !ok || dependency == stage.ID {
					return catalog, nil, fmt.Errorf("commander %s stage %s has invalid dependency %q", commander.ID, stage.ID, dependency)
				}
			}
			stageNodes = append(stageNodes, RoleTaskNode{ID: stage.ID, DependsOn: stage.DependsOn})
		}
		if hasRoleGraphCycle(stageNodes) {
			return catalog, nil, fmt.Errorf("commander %s stage dependencies contain a cycle", commander.ID)
		}
	}
	seen := make(map[string]bool, len(catalog.Experts))
	methodOwners := map[string]string{}
	for _, expert := range catalog.Experts {
		commander, ok := commanderByID[expert.Commander]
		if !capabilityIDPattern.MatchString(expert.ID) || seen[expert.ID] || !ok ||
			expert.RoleType != "specialist" || expert.DecisionMode != "bounded-moe" ||
			strings.TrimSpace(expert.Name) == "" || len(expert.Capabilities) == 0 ||
			strings.TrimSpace(expert.Mission) == "" || len(expert.Inputs) == 0 ||
			len(expert.Outputs) == 0 || len(expert.Boundaries) == 0 ||
			len(expert.Escalation) == 0 || len(expert.Triggers) == 0 ||
			strings.TrimSpace(expert.PromptCompiler) == "" || len(expert.Workflow) == 0 || len(expert.Verify) == 0 ||
			len(expert.Constraints) == 0 || len(expert.ToolPolicy) == 0 || len(expert.Acceptance) == 0 ||
			len(expert.PonyTailRules) == 0 || len(expert.WhiteHatChecks) == 0 {
			return catalog, nil, fmt.Errorf("invalid or duplicate expert contract %q", expert.ID)
		}
		seen[expert.ID] = true
		if !containsString(commander.ExpertIDs, expert.ID) {
			return catalog, nil, fmt.Errorf("expert %s is not listed by commander %s", expert.ID, commander.ID)
		}
		for _, capability := range expert.Capabilities {
			if !capabilityIDPattern.MatchString(capability) {
				return catalog, nil, fmt.Errorf("expert %s has invalid capability %q", expert.ID, capability)
			}
			if !containsString(commander.Capabilities, capability) {
				return catalog, nil, fmt.Errorf("expert %s capability %s is outside commander %s", expert.ID, capability, commander.ID)
			}
		}
		seenSources := map[string]bool{}
		for _, sourceID := range expert.SourceIDs {
			if !componentIDPattern.MatchString(sourceID) || seenSources[sourceID] {
				return catalog, nil, fmt.Errorf("expert %s has invalid or duplicate source %q", expert.ID, sourceID)
			}
			seenSources[sourceID] = true
		}
		for _, method := range expert.Methods {
			if !capabilityIDPattern.MatchString(method.ID) || len(method.Sources) == 0 ||
				len(method.Signals) == 0 || strings.TrimSpace(method.Instruction) == "" ||
				strings.TrimSpace(method.Verification) == "" || len(method.Instruction) > 512 {
				return catalog, nil, fmt.Errorf("expert %s has invalid distilled method %q", expert.ID, method.ID)
			}
			if owner, exists := methodOwners[method.ID]; exists {
				return catalog, nil, fmt.Errorf("distilled method %s belongs to both %s and %s", method.ID, owner, expert.ID)
			}
			methodOwners[method.ID] = expert.ID
		}
	}
	for _, commander := range catalog.Commanders {
		for _, expertID := range commander.ExpertIDs {
			found := false
			for _, expert := range catalog.Experts {
				if expert.ID == expertID {
					if expert.Commander != commander.ID {
						return catalog, nil, fmt.Errorf("expert %s is assigned to multiple commanders", expertID)
					}
					found = true
					break
				}
			}
			if !found {
				return catalog, nil, fmt.Errorf("commander %s references missing expert %s", commander.ID, expertID)
			}
		}
	}
	return catalog, data, nil
}

func (catalog expertCatalog) commanderForCapability(capability string) (expertCommander, bool) {
	for _, commander := range catalog.Commanders {
		if containsString(commander.Capabilities, capability) {
			return commander, true
		}
	}
	return expertCommander{}, false
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
	if contract.MoE != "sparse-role-moe" ||
		contract.RoleID != "expert:"+contract.ExpertID ||
		contract.RoleKind != "expert" ||
		contract.TaskGraphID != contract.TaskGraph.GraphID ||
		contract.TaskGraphSHA256 != contract.TaskGraph.SHA256 ||
		contract.PonyTail != ponyTailDoctrine {
		return fmt.Errorf("expert role graph identity or doctrine is inconsistent")
	}
	if err := validateRoleTaskGraph(contract.TaskGraph); err != nil {
		return err
	}
	if err := validateWorkerExecutionPolicy(contract.Worker); err != nil {
		return err
	}
	var task workerContract
	var prefix struct {
		Capability string `json:"capability"`
	}
	if json.Unmarshal([]byte(contract.Worker.TaskContract), &task) != nil ||
		json.Unmarshal([]byte(contract.Worker.StableCapabilityPrefix), &prefix) != nil ||
		task.Expert == nil || task.Expert.ID != contract.ExpertID ||
		task.Expert.Commander != contract.Commander ||
		task.Expert.Capability != prefix.Capability ||
		task.Expert.CatalogSHA256 != contract.ExpertVersion ||
		!equalStringSlices(task.Expert.Workflow, contract.Workflow) ||
		!equalStringSlices(task.Expert.Verification, contract.Verification) ||
		(!isLegacyExpertDirective(task.Expert) && len(task.Expert.Acceptance) > 0 && !equalStringSlices(task.Expert.Acceptance, contract.Acceptance)) ||
		(!isLegacyExpertDirective(task.Expert) && len(task.Expert.Constraints) > 0 && !equalStringSlices(task.Expert.Constraints, contract.Constraints)) ||
		(!isLegacyExpertDirective(task.Expert) && len(task.Expert.ToolPolicy) > 0 && !equalStringSlices(task.Expert.ToolPolicy, contract.ToolPolicy)) ||
		(!isLegacyExpertDirective(task.Expert) && len(task.Expert.PonyTailRules) > 0 && !equalStringSlices(task.Expert.PonyTailRules, contract.PonyTailRules)) ||
		(!isLegacyExpertDirective(task.Expert) && len(task.Expert.WhiteHatChecks) > 0 && !equalStringSlices(task.Expert.WhiteHatChecks, contract.WhiteHatChecks)) ||
		strings.TrimSpace(task.Expert.TeamMission) == "" ||
		len(task.Expert.TeamAcceptance) == 0 ||
		strings.TrimSpace(task.Expert.PromptCompiler) == "" ||
		contract.Worker.TaskContractSHA256 != sha256Hex([]byte(contract.Worker.TaskContract)) ||
		contract.Worker.AllocatedTaskContractBytes != len([]byte(contract.Worker.TaskContract)) {
		return fmt.Errorf("expert task directive is missing or inconsistent")
	}
	return validateExpertPonytail(contract.Worker)
}

func isLegacyExpertDirective(d *expertTaskDirective) bool {
	return d == nil || (len(d.Acceptance) == 0 && len(d.Constraints) == 0 &&
		len(d.ToolPolicy) == 0 && len(d.PonyTailRules) == 0 && len(d.WhiteHatChecks) == 0)
}

// The ordinary route's dispatch must not accept an expert directive whose
// catalog changed since routing. The trusted commander manifest supplies the
// catalog root; caller-provided task text never chooses that path.
func verifyWorkerExpertBinding(worker WorkerTask, manifests []Manifest) error {
	if strings.TrimSpace(worker.TaskContract) == "" {
		return nil // legacy non-expert compatibility contract
	}
	var task workerContract
	if err := json.Unmarshal([]byte(worker.TaskContract), &task); err != nil {
		return fmt.Errorf("invalid worker task contract: %w", err)
	}
	if task.Expert == nil {
		return nil
	}
	if worker.TaskContractSHA256 != sha256Hex([]byte(worker.TaskContract)) ||
		worker.AllocatedTaskContractBytes != len([]byte(worker.TaskContract)) {
		return fmt.Errorf("expert task contract hash or byte count mismatch")
	}
	for _, manifest := range manifests {
		if task.Expert.Capability != manifest.ID {
			continue
		}
		if manifest.Root == "" {
			break
		}
		catalog, data, err := loadExpertCatalog(manifest.Root)
		if err != nil {
			return err
		}
		digest := sha256.Sum256(data)
		if task.Expert.CatalogSHA256 != hex.EncodeToString(digest[:]) {
			return fmt.Errorf("expert catalog changed after routing")
		}
		commander, ok := catalog.commanderByID(task.Expert.Commander)
		if !ok || !containsString(commander.Capabilities, manifest.ID) || !containsString(commander.ExpertIDs, task.Expert.ID) {
			return fmt.Errorf("expert is outside trusted commander")
		}
		for _, expert := range catalog.Experts {
			if expert.ID != task.Expert.ID {
				continue
			}
			if expert.Commander != commander.ID || !containsString(expert.Capabilities, manifest.ID) {
				return fmt.Errorf("expert is outside trusted commander")
			}
			if task.Expert.PromptCompiler != expert.PromptCompiler ||
				!equalStringSlices(task.Expert.Inputs, expert.Inputs) ||
				!equalStringSlices(task.Expert.Outputs, expert.Outputs) ||
				!equalStringSlices(task.Expert.Workflow, expert.Workflow) ||
				!equalStringSlices(task.Expert.Verification, expert.Verify) ||
				!equalStringSlices(task.Expert.Acceptance, expert.Acceptance) ||
				!equalStringSlices(task.Expert.Constraints, expert.Constraints) ||
				!equalStringSlices(task.Expert.ToolPolicy, expert.ToolPolicy) ||
				!equalStringSlices(task.Expert.PonyTailRules, expert.PonyTailRules) ||
				!equalStringSlices(task.Expert.WhiteHatChecks, expert.WhiteHatChecks) ||
				!equalExpertMethods(task.Expert.Methods, selectedExpertMethods(expert, task.Objective)) ||
				task.Expert.TeamMission != commander.Mission ||
				!equalStringSlices(task.Expert.TeamAcceptance, commander.Acceptance) {
				return fmt.Errorf("expert directive differs from trusted catalog")
			}
			for _, sourceID := range expert.SourceIDs {
				found := false
				for _, source := range worker.SourceExecution {
					if source.SourceID == sourceID && source.Capability == manifest.ID {
						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf("expert required source %s is missing", sourceID)
				}
			}
			return nil
		}
		return fmt.Errorf("expert is absent from trusted catalog")
	}
	return fmt.Errorf("expert commander has no trusted manifest root")
}

func (catalog expertCatalog) commanderByID(id string) (expertCommander, bool) {
	for _, commander := range catalog.Commanders {
		if commander.ID == id {
			return commander, true
		}
	}
	return expertCommander{}, false
}

func validateExpertPonytail(worker WorkerTask) error {
	if worker.StableCapabilityPrefix == "" || worker.StablePrefixSHA256 != sha256Hex([]byte(worker.StableCapabilityPrefix)) ||
		worker.TaskContract == "" || worker.TaskContractSHA256 != sha256Hex([]byte(worker.TaskContract)) {
		return fmt.Errorf("expert handoff requires intact PonyTail worker payload hashes")
	}
	var prefix struct {
		ImplementationDoctrine string `json:"implementation_doctrine"`
	}
	var task workerContract
	if json.Unmarshal([]byte(worker.StableCapabilityPrefix), &prefix) != nil ||
		prefix.ImplementationDoctrine != ponyTailDoctrine ||
		json.Unmarshal([]byte(worker.TaskContract), &task) != nil ||
		task.Schema != "wuji-worker-contract-v2" || task.Branch != worker.ID {
		return fmt.Errorf("expert handoff requires a PonyTail worker contract")
	}
	for _, rule := range workerProtocol("", "", "", "") {
		if !containsString(worker.Protocol, rule) || !containsString(task.Protocol, rule) {
			return fmt.Errorf("expert handoff omits universal PonyTail rule %q", rule)
		}
	}
	return nil
}

func validateLiveExpertBindings(contract ExpertHandoffContract) error {
	catalog, catalogData, err := loadExpertCatalog(contract.RepositoryRoot)
	if err != nil {
		return err
	}
	d := sha256.Sum256(catalogData)
	if hex.EncodeToString(d[:]) != contract.ExpertVersion {
		return fmt.Errorf("expert catalog changed after contract preparation")
	}
	var task workerContract
	if err := json.Unmarshal([]byte(contract.Worker.TaskContract), &task); err != nil || task.Expert == nil {
		return fmt.Errorf("expert task contract is invalid")
	}
	found := false
	for _, expert := range catalog.Experts {
		if expert.ID != contract.ExpertID {
			continue
		}
		found = true
		if !equalExpertMethods(task.Expert.Methods, selectedExpertMethods(expert, task.Objective)) {
			return fmt.Errorf("expert distilled methods differ from trusted catalog")
		}
		break
	}
	if !found {
		return fmt.Errorf("expert is absent from trusted catalog")
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
