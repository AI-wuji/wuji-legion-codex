package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	executionGraphSchemaVersion = 1
	executionGraphMaxNodes      = 256
	executionGraphMaxFields     = 32
	executionGraphMaxFieldBytes = 512
)

type executionGraphPayload struct {
	SchemaVersion int                  `json:"schema_version"`
	Target        string               `json:"target"`
	Nodes         []ExecutionGraphNode `json:"nodes"`
}

// VersionedExecutionNodeInput adds runtime identity without changing the
// durable execution-node schema used by existing callers.
type VersionedExecutionNodeInput struct {
	ExecutionNodeInput
	TaskInstanceID string
	GraphVersion   string
	AttemptID      string
}

type VersionedExecutionResultInput struct {
	ExecutionResultInput
	TaskInstanceID string
	GraphVersion   string
	AttemptID      string
}

type executionRuntimeBinding struct {
	TaskInstanceID string `json:"task_instance_id"`
	GraphVersion   string `json:"graph_version"`
	AttemptID      string `json:"attempt_id"`
}

type executionRuntimeGraph struct {
	SchemaVersion int                                `json:"schema_version"`
	Nodes         map[string]executionRuntimeBinding `json:"nodes"`
}

func DefaultExecutionGraphStore() string {
	if value := strings.TrimSpace(os.Getenv("WUJI_EXECUTION_GRAPH_DIR")); value != "" {
		return filepath.Clean(value)
	}
	return filepath.Join(".wuji", "execution-graph")
}

func RecordExecutionNode(store, requirementStore string, input ExecutionNodeInput) (ExecutionGraphNode, error) {
	if err := validateExecutionInput(input); err != nil {
		return ExecutionGraphNode{}, err
	}
	if strings.TrimSpace(requirementStore) == "" {
		requirementStore = DefaultRequirementGraphStore()
	}
	requirements, err := loadRequirementGraph(requirementStore)
	if err != nil {
		return ExecutionGraphNode{}, err
	}
	if err := validateRequirementRevisionReferences(requirements, input.RequirementRevisions); err != nil {
		return ExecutionGraphNode{}, err
	}
	var result ExecutionGraphNode
	err = withKnowledgeStoreLock(store, func() error {
		graph, err := loadExecutionGraph(store)
		if err != nil {
			return err
		}
		current := latestExecutionNode(graph.Nodes, input.ID)
		node := normalizedExecutionNode(input)
		if current != nil && current.Status != "superseded" && sameExecutionFields(*current, node) {
			result = *current
			return nil
		}
		if current != nil {
			node.Revision = current.Revision + 1
			node.Supersedes = current.VersionID
			for index := range graph.Nodes {
				if graph.Nodes[index].VersionID == current.VersionID {
					graph.Nodes[index].Status = "superseded"
				}
			}
		} else {
			node.Revision = 1
		}
		node.VersionID = executionVersionID(node.ID, node.Revision)
		graph.Nodes = append(graph.Nodes, node)
		refreshExecutionInvalidations(&graph, requirements)
		if len(graph.Nodes) > executionGraphMaxNodes {
			return fmt.Errorf("execution graph exceeds %d nodes", executionGraphMaxNodes)
		}
		if err := writeExecutionGraph(store, graph); err != nil {
			return err
		}
		result = node
		return AuditEventRecord(auditStoreFor(store), AuditEvent{EventType: "execution-node-recorded", Actor: input.Authority, Authority: input.Authority, Target: node.VersionID, InputRevision: strings.Join(node.RequirementRevisions, ","), ResultHandle: "wuji-execution://" + node.VersionID})
	})
	if err != nil {
		return ExecutionGraphNode{}, err
	}
	return result, nil
}

func RecordVersionedExecutionNode(store, requirementStore string, input VersionedExecutionNodeInput) (ExecutionGraphNode, error) {
	if err := validateExecutionRuntimeBinding(executionRuntimeBinding{TaskInstanceID: input.TaskInstanceID, GraphVersion: input.GraphVersion, AttemptID: input.AttemptID}); err != nil {
		return ExecutionGraphNode{}, err
	}
	node, err := RecordExecutionNode(store, requirementStore, input.ExecutionNodeInput)
	if err != nil {
		return ExecutionGraphNode{}, err
	}
	err = withKnowledgeStoreLock(store, func() error {
		runtime, err := loadExecutionRuntimeGraph(store)
		if err != nil {
			return err
		}
		runtime.Nodes[node.VersionID] = executionRuntimeBinding{TaskInstanceID: strings.TrimSpace(input.TaskInstanceID), GraphVersion: strings.TrimSpace(input.GraphVersion), AttemptID: strings.TrimSpace(input.AttemptID)}
		return writeExecutionRuntimeGraph(store, runtime)
	})
	if err != nil {
		return ExecutionGraphNode{}, err
	}
	return node, nil
}

func RecordExecutionResult(store, requirementStore string, input ExecutionResultInput) (ExecutionGraphNode, error) {
	return recordExecutionResult(store, requirementStore, input, nil)
}

func RecordVersionedExecutionResult(store, requirementStore string, input VersionedExecutionResultInput) (ExecutionGraphNode, error) {
	binding := executionRuntimeBinding{TaskInstanceID: input.TaskInstanceID, GraphVersion: input.GraphVersion, AttemptID: input.AttemptID}
	if err := validateExecutionRuntimeBinding(binding); err != nil {
		return ExecutionGraphNode{}, err
	}
	return recordExecutionResult(store, requirementStore, input.ExecutionResultInput, &binding)
}

func recordExecutionResult(store, requirementStore string, input ExecutionResultInput, binding *executionRuntimeBinding) (ExecutionGraphNode, error) {
	if !componentIDPattern.MatchString(strings.TrimSpace(input.ID)) {
		return ExecutionGraphNode{}, fmt.Errorf("execution id is invalid")
	}
	switch strings.ToLower(strings.TrimSpace(input.Status)) {
	case "planned", "running", "succeeded", "failed", "cancelled", "invalidated":
	default:
		return ExecutionGraphNode{}, fmt.Errorf("execution status is invalid")
	}
	if strings.TrimSpace(requirementStore) == "" {
		requirementStore = DefaultRequirementGraphStore()
	}
	requirements, err := loadRequirementGraph(requirementStore)
	if err != nil {
		return ExecutionGraphNode{}, err
	}
	var result ExecutionGraphNode
	err = withKnowledgeStoreLock(store, func() error {
		graph, err := loadExecutionGraph(store)
		if err != nil {
			return err
		}
		refreshExecutionInvalidations(&graph, requirements)
		current := latestExecutionNode(graph.Nodes, input.ID)
		if current == nil || current.Status == "superseded" || current.Status == "cancelled" || current.Status == "invalidated" {
			return fmt.Errorf("execution node %q is not active", input.ID)
		}
		if binding != nil {
			runtime, err := loadExecutionRuntimeGraph(store)
			if err != nil {
				return err
			}
			actual, ok := runtime.Nodes[current.VersionID]
			if !ok || actual != *binding {
				return fmt.Errorf("execution result runtime binding is stale")
			}
			if (current.Status == "succeeded" || current.Status == "failed") && !sameVersionedExecutionResult(*current, input) {
				return fmt.Errorf("execution terminal result is immutable for this runtime binding")
			}
		}
		for index := range graph.Nodes {
			if graph.Nodes[index].VersionID == current.VersionID {
				graph.Nodes[index].Status = strings.ToLower(strings.TrimSpace(input.Status))
				graph.Nodes[index].Failure = strings.TrimSpace(input.Failure)
				graph.Nodes[index].Recovery = strings.TrimSpace(input.Recovery)
				graph.Nodes[index].ArtifactHandles = normalizedGraphList(input.ArtifactHandles)
				graph.Nodes[index].VerificationHandles = normalizedGraphList(input.VerificationHandles)
				result = graph.Nodes[index]
			}
		}
		refreshExecutionInvalidations(&graph, requirements)
		if err := writeExecutionGraph(store, graph); err != nil {
			return err
		}
		// Keep the candidate snapshot in the same execution-store critical
		// section as the result. A later attempt cannot rebind this event.
		if binding != nil {
			_, _ = recordExecutionFeedbackCandidate(executionFeedbackStoreFor(store), result, *binding)
		}
		return AuditEventRecord(auditStoreFor(store), AuditEvent{EventType: "execution-result-recorded", Actor: "deterministic-executor", Authority: "deterministic-executor", Target: result.VersionID, InputRevision: strings.Join(result.RequirementRevisions, ","), ResultHandle: "wuji-execution://" + result.VersionID, EvidenceHandles: result.VerificationHandles})
	})
	if err != nil {
		return ExecutionGraphNode{}, err
	}
	return result, nil
}

func sameVersionedExecutionResult(node ExecutionGraphNode, input ExecutionResultInput) bool {
	return node.Status == strings.ToLower(strings.TrimSpace(input.Status)) &&
		node.Failure == strings.TrimSpace(input.Failure) &&
		node.Recovery == strings.TrimSpace(input.Recovery) &&
		sameStringSlice(node.ArtifactHandles, normalizedGraphList(input.ArtifactHandles)) &&
		sameStringSlice(node.VerificationHandles, normalizedGraphList(input.VerificationHandles))
}

func ProjectExecutionGraph(store, requirementStore, target string, maxBytes int) (ExecutionGraphProjection, error) {
	if maxBytes < 256 || maxBytes > 4096 {
		return ExecutionGraphProjection{}, fmt.Errorf("execution projection max bytes must be between 256 and 4096")
	}
	if strings.TrimSpace(requirementStore) == "" {
		requirementStore = DefaultRequirementGraphStore()
	}
	requirements, err := loadRequirementGraph(requirementStore)
	if err != nil {
		return ExecutionGraphProjection{}, err
	}
	graph, err := loadExecutionGraph(store)
	if err != nil {
		return ExecutionGraphProjection{}, err
	}
	before, _ := json.Marshal(graph)
	refreshExecutionInvalidations(&graph, requirements)
	after, _ := json.Marshal(graph)
	if string(before) != string(after) {
		if err := withKnowledgeStoreLock(store, func() error { return writeExecutionGraph(store, graph) }); err != nil {
			return ExecutionGraphProjection{}, err
		}
	}
	targetNode := resolveExecutionTarget(graph.Nodes, target)
	if targetNode == nil {
		return ExecutionGraphProjection{}, fmt.Errorf("execution graph target %q is not found", target)
	}
	byVersion := make(map[string]ExecutionGraphNode, len(graph.Nodes))
	for _, node := range graph.Nodes {
		byVersion[node.VersionID] = node
	}
	nodes := []ExecutionGraphNode{*targetNode}
	for _, dependency := range targetNode.DependsOn {
		node, ok := byVersion[dependency]
		if !ok {
			return ExecutionGraphProjection{}, fmt.Errorf("execution node %q references missing dependency %q", targetNode.VersionID, dependency)
		}
		nodes = append(nodes, node)
	}
	payloadNodes := make([]ExecutionGraphNode, 0, len(nodes))
	for _, node := range nodes {
		candidate := append(payloadNodes, node)
		payload, err := json.Marshal(executionGraphPayload{SchemaVersion: executionGraphSchemaVersion, Target: targetNode.VersionID, Nodes: candidate})
		if err != nil {
			return ExecutionGraphProjection{}, err
		}
		if len(payload) > maxBytes {
			if len(payloadNodes) == 0 {
				return ExecutionGraphProjection{}, fmt.Errorf("target %q exceeds the %d-byte projection budget", targetNode.VersionID, maxBytes)
			}
			break
		}
		payloadNodes = candidate
	}
	payload, err := json.Marshal(executionGraphPayload{SchemaVersion: executionGraphSchemaVersion, Target: targetNode.VersionID, Nodes: payloadNodes})
	if err != nil {
		return ExecutionGraphProjection{}, err
	}
	digest := sha256.Sum256(payload)
	hash := hex.EncodeToString(digest[:])
	return ExecutionGraphProjection{SchemaVersion: executionGraphSchemaVersion, Handle: "wuji-execution-graph://sha256/" + hash, PayloadSHA256: hash, PayloadBytes: len(payload), Target: targetNode.VersionID, Nodes: payloadNodes, Payload: string(payload)}, nil
}

func WriteExecutionGraphProjection(projection ExecutionGraphProjection, artifactDir string) (string, error) {
	if err := validateExecutionProjection(projection); err != nil {
		return "", err
	}
	if err := os.MkdirAll(artifactDir, 0o700); err != nil {
		return "", err
	}
	path := filepath.Join(filepath.Clean(artifactDir), projection.PayloadSHA256+".json")
	data, err := json.MarshalIndent(projection, "", "  ")
	if err != nil {
		return "", err
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return "", err
	}
	return path, nil
}

func LoadExecutionGraphProjection(path string) (ExecutionGraphProjection, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ExecutionGraphProjection{}, err
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	var projection ExecutionGraphProjection
	if err := decoder.Decode(&projection); err != nil {
		return ExecutionGraphProjection{}, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return ExecutionGraphProjection{}, fmt.Errorf("multiple JSON values are not allowed")
	}
	if err := validateExecutionProjection(projection); err != nil {
		return ExecutionGraphProjection{}, err
	}
	projection.ArtifactPath = path
	return projection, nil
}

func validateExecutionInput(input ExecutionNodeInput) error {
	if !componentIDPattern.MatchString(strings.TrimSpace(input.ID)) {
		return fmt.Errorf("execution id is invalid")
	}
	if strings.EqualFold(strings.TrimSpace(input.Authority), "aji") || (input.Authority != "staff" && input.Authority != "deterministic-executor") {
		return fmt.Errorf("execution authority must be staff or deterministic-executor")
	}
	if strings.TrimSpace(input.Goal) == "" || len([]byte(strings.TrimSpace(input.Goal))) > executionGraphMaxFieldBytes {
		return fmt.Errorf("execution goal is required and bounded")
	}
	if len(input.RequirementRevisions) == 0 || len(input.RequirementRevisions) > executionGraphMaxFields {
		return fmt.Errorf("execution requires between 1 and %d requirement revisions", executionGraphMaxFields)
	}
	if strings.TrimSpace(input.Model) == "" || strings.TrimSpace(input.ModelReason) == "" {
		return fmt.Errorf("execution model and model reason are required")
	}
	if input.TimeBudgetSeconds < 0 || input.CostBudgetMicrounits < 0 || input.MaxAttempts < 0 {
		return fmt.Errorf("execution budgets cannot be negative")
	}
	for name, values := range map[string][]string{"requirements": input.RequirementRevisions, "depends_on": input.DependsOn, "inputs": input.Inputs, "allowed_context": input.AllowedContext, "outputs": input.Outputs, "avoids": input.Avoids, "acceptance": input.Acceptance, "verification": input.Verification, "evidence_required": input.EvidenceRequired} {
		if len(values) > executionGraphMaxFields {
			return fmt.Errorf("%s cannot contain more than %d entries", name, executionGraphMaxFields)
		}
		for _, value := range values {
			if strings.TrimSpace(value) == "" || len([]byte(strings.TrimSpace(value))) > executionGraphMaxFieldBytes || strings.ContainsAny(value, "\r\n\t") {
				return fmt.Errorf("%s entry is invalid", name)
			}
		}
	}
	return nil
}

func normalizedExecutionNode(input ExecutionNodeInput) ExecutionGraphNode {
	return ExecutionGraphNode{ID: strings.TrimSpace(input.ID), Status: "planned", Authority: strings.TrimSpace(input.Authority), Goal: strings.TrimSpace(input.Goal), Avoids: normalizedGraphList(input.Avoids), RequirementRevisions: normalizedGraphList(input.RequirementRevisions), DependsOn: normalizedGraphList(input.DependsOn), Inputs: normalizedGraphList(input.Inputs), AllowedContext: normalizedGraphList(input.AllowedContext), Outputs: normalizedGraphList(input.Outputs), Model: strings.TrimSpace(input.Model), ModelReason: strings.TrimSpace(input.ModelReason), Acceptance: normalizedGraphList(input.Acceptance), Verification: normalizedGraphList(input.Verification), EvidenceRequired: normalizedGraphList(input.EvidenceRequired), TimeBudgetSeconds: input.TimeBudgetSeconds, CostBudgetMicrounits: input.CostBudgetMicrounits, MaxAttempts: input.MaxAttempts, NetworkBoundary: strings.TrimSpace(input.NetworkBoundary), WriteBoundary: strings.TrimSpace(input.WriteBoundary), BranchBoundary: strings.TrimSpace(input.BranchBoundary), Failure: strings.TrimSpace(input.Failure), Recovery: strings.TrimSpace(input.Recovery)}
}

func sameExecutionFields(a, b ExecutionGraphNode) bool {
	a.VersionID, a.Revision, a.Status, a.Supersedes = "", 0, "", ""
	b.VersionID, b.Revision, b.Status, b.Supersedes = "", 0, "", ""
	aa, _ := json.Marshal(a)
	bb, _ := json.Marshal(b)
	return string(aa) == string(bb)
}

func validateRequirementRevisionReferences(graph RequirementGraph, refs []string) error {
	byVersion := make(map[string]RequirementGraphNode, len(graph.Nodes))
	for _, node := range graph.Nodes {
		byVersion[node.VersionID] = node
	}
	for _, ref := range refs {
		node, ok := byVersion[strings.TrimSpace(ref)]
		if !ok || node.Kind != "requirement" || node.Status != "active" {
			return fmt.Errorf("requirement revision %q is not active", ref)
		}
	}
	return nil
}

func refreshExecutionInvalidations(graph *ExecutionGraph, requirements RequirementGraph) {
	active := map[string]string{}
	for _, node := range requirements.Nodes {
		if node.Kind == "requirement" && node.Status == "active" {
			active[node.ID] = node.VersionID
		}
	}
	invalid := map[string]bool{}
	for index := range graph.Nodes {
		node := &graph.Nodes[index]
		if node.Status == "superseded" {
			continue
		}
		if node.Status == "invalidated" {
			invalid[node.VersionID] = true
			continue
		}
		for _, ref := range node.RequirementRevisions {
			id := ref
			if at := strings.LastIndex(id, "@"); at > 0 {
				id = id[:at]
			}
			if active[id] != ref {
				node.Status = "invalidated"
				invalid[node.VersionID] = true
				break
			}
		}
	}
	for changed := true; changed; {
		changed = false
		for index := range graph.Nodes {
			node := &graph.Nodes[index]
			if node.Status == "superseded" || invalid[node.VersionID] {
				continue
			}
			for _, dep := range node.DependsOn {
				if invalid[dep] {
					node.Status = "invalidated"
					invalid[node.VersionID] = true
					changed = true
					break
				}
			}
		}
	}
}

func loadExecutionGraph(store string) (ExecutionGraph, error) {
	data, err := os.ReadFile(executionGraphPath(store))
	if os.IsNotExist(err) {
		return ExecutionGraph{SchemaVersion: executionGraphSchemaVersion, Nodes: []ExecutionGraphNode{}}, nil
	}
	if err != nil {
		return ExecutionGraph{}, err
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	var graph ExecutionGraph
	if err := decoder.Decode(&graph); err != nil {
		return ExecutionGraph{}, fmt.Errorf("decode execution graph: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return ExecutionGraph{}, fmt.Errorf("decode execution graph: multiple JSON values are not allowed")
	}
	if err := validateExecutionGraph(graph); err != nil {
		return ExecutionGraph{}, err
	}
	return graph, nil
}

func validateExecutionRuntimeBinding(binding executionRuntimeBinding) error {
	if !componentIDPattern.MatchString(strings.TrimSpace(binding.TaskInstanceID)) {
		return fmt.Errorf("execution task instance id is invalid")
	}
	for _, value := range []string{binding.GraphVersion, binding.AttemptID} {
		value = strings.TrimSpace(value)
		if value == "" || len([]byte(value)) > executionGraphMaxFieldBytes || strings.ContainsAny(value, "\r\n\t") {
			return fmt.Errorf("execution runtime binding is invalid")
		}
	}
	return nil
}

func loadExecutionRuntimeGraph(store string) (executionRuntimeGraph, error) {
	data, err := os.ReadFile(executionRuntimeGraphPath(store))
	if os.IsNotExist(err) {
		return executionRuntimeGraph{SchemaVersion: executionGraphSchemaVersion, Nodes: map[string]executionRuntimeBinding{}}, nil
	}
	if err != nil {
		return executionRuntimeGraph{}, err
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	var runtime executionRuntimeGraph
	if err := decoder.Decode(&runtime); err != nil {
		return executionRuntimeGraph{}, fmt.Errorf("decode execution runtime graph: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return executionRuntimeGraph{}, fmt.Errorf("decode execution runtime graph: multiple JSON values are not allowed")
	}
	if runtime.SchemaVersion != executionGraphSchemaVersion || len(runtime.Nodes) > executionGraphMaxNodes {
		return executionRuntimeGraph{}, fmt.Errorf("execution runtime graph metadata is invalid")
	}
	for version, binding := range runtime.Nodes {
		if strings.TrimSpace(version) == "" || validateExecutionRuntimeBinding(binding) != nil {
			return executionRuntimeGraph{}, fmt.Errorf("execution runtime graph entry is invalid")
		}
	}
	return runtime, nil
}

func writeExecutionRuntimeGraph(store string, runtime executionRuntimeGraph) error {
	if runtime.SchemaVersion != executionGraphSchemaVersion {
		return fmt.Errorf("execution runtime graph metadata is invalid")
	}
	path := executionRuntimeGraphPath(store)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(runtime, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o600)
}

func writeExecutionGraph(store string, graph ExecutionGraph) error {
	if err := validateExecutionGraph(graph); err != nil {
		return err
	}
	path := executionGraphPath(store)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(graph, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	temporary, err := os.CreateTemp(filepath.Dir(path), ".execution-graph-*.json")
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

func validateExecutionGraph(graph ExecutionGraph) error {
	if graph.SchemaVersion != executionGraphSchemaVersion || len(graph.Nodes) > executionGraphMaxNodes {
		return fmt.Errorf("execution graph metadata is invalid")
	}
	versions := map[string]bool{}
	for _, node := range graph.Nodes {
		if !componentIDPattern.MatchString(node.ID) || node.Revision < 1 || node.VersionID != executionVersionID(node.ID, node.Revision) || versions[node.VersionID] {
			return fmt.Errorf("execution node identity is invalid")
		}
		versions[node.VersionID] = true
		switch node.Status {
		case "planned", "running", "succeeded", "failed", "cancelled", "invalidated", "superseded":
		default:
			return fmt.Errorf("execution node status is invalid")
		}
		if err := validateExecutionInput(ExecutionNodeInput{ID: node.ID, Authority: node.Authority, Goal: node.Goal, Avoids: node.Avoids, RequirementRevisions: node.RequirementRevisions, DependsOn: node.DependsOn, Inputs: node.Inputs, AllowedContext: node.AllowedContext, Outputs: node.Outputs, Model: node.Model, ModelReason: node.ModelReason, Acceptance: node.Acceptance, Verification: node.Verification, EvidenceRequired: node.EvidenceRequired, TimeBudgetSeconds: node.TimeBudgetSeconds, CostBudgetMicrounits: node.CostBudgetMicrounits, MaxAttempts: node.MaxAttempts, NetworkBoundary: node.NetworkBoundary, WriteBoundary: node.WriteBoundary, BranchBoundary: node.BranchBoundary, Failure: node.Failure, Recovery: node.Recovery}); err != nil {
			return err
		}
	}
	for _, node := range graph.Nodes {
		for _, dependency := range node.DependsOn {
			if !versions[dependency] {
				return fmt.Errorf("execution dependency %q is missing", dependency)
			}
		}
		if node.Supersedes != "" && !versions[node.Supersedes] {
			return fmt.Errorf("execution supersedes reference %q is missing", node.Supersedes)
		}
	}
	return nil
}

func validateExecutionProjection(projection ExecutionGraphProjection) error {
	if projection.SchemaVersion != executionGraphSchemaVersion || projection.PayloadBytes != len([]byte(projection.Payload)) || !taskCircuitHashPattern.MatchString(projection.PayloadSHA256) {
		return fmt.Errorf("execution projection metadata is invalid")
	}
	digest := sha256.Sum256([]byte(projection.Payload))
	hash := hex.EncodeToString(digest[:])
	if !strings.EqualFold(hash, projection.PayloadSHA256) || projection.Handle != "wuji-execution-graph://sha256/"+strings.ToLower(projection.PayloadSHA256) {
		return fmt.Errorf("execution projection hash does not match its payload")
	}
	var payload executionGraphPayload
	if err := json.Unmarshal([]byte(projection.Payload), &payload); err != nil || payload.SchemaVersion != projection.SchemaVersion || payload.Target != projection.Target {
		return fmt.Errorf("execution projection payload is invalid")
	}
	a, _ := json.Marshal(payload.Nodes)
	b, _ := json.Marshal(projection.Nodes)
	if string(a) != string(b) {
		return fmt.Errorf("execution projection nodes do not match its payload")
	}
	return nil
}

func latestExecutionNode(nodes []ExecutionGraphNode, id string) *ExecutionGraphNode {
	var latest *ExecutionGraphNode
	for index := range nodes {
		node := &nodes[index]
		if node.ID == id && node.Status != "superseded" && (latest == nil || node.Revision > latest.Revision) {
			latest = node
		}
	}
	return latest
}

func resolveExecutionTarget(nodes []ExecutionGraphNode, target string) *ExecutionGraphNode {
	target = strings.TrimSpace(target)
	for index := range nodes {
		if nodes[index].VersionID == target {
			return &nodes[index]
		}
	}
	return latestExecutionNode(nodes, target)
}

func executionVersionID(id string, revision int) string {
	return fmt.Sprintf("%s@%d", id, revision)
}

func executionGraphPath(store string) string {
	return filepath.Join(filepath.Clean(store), "v1", "graph.json")
}

func executionRuntimeGraphPath(store string) string {
	return filepath.Join(filepath.Clean(store), "v1", "runtime.json")
}

func auditStoreFor(store string) string {
	store = filepath.Clean(store)
	if strings.EqualFold(filepath.Base(store), "audit") {
		return store
	}
	return filepath.Join(filepath.Dir(store), "audit")
}

func appendAuditEvent(store string, event AuditEvent) error {
	if strings.TrimSpace(event.EventType) == "" || strings.TrimSpace(event.Target) == "" {
		return fmt.Errorf("audit event type and target are required")
	}
	for _, value := range []string{event.EventType, event.Actor, event.Authority, event.Target, event.InputRevision, event.ResultHandle} {
		for _, pattern := range knowledgeSecretPatterns {
			if pattern.MatchString(value) {
				return fmt.Errorf("audit event must not contain a secret")
			}
		}
	}
	if event.EventID == "" {
		event.OccurredAt = time.Now().UTC().Format(time.RFC3339Nano)
		event.EventID = "event-" + auditEventDigest(event)
	} else if event.OccurredAt == "" {
		event.OccurredAt = time.Now().UTC().Format(time.RFC3339Nano)
	}
	event.RecordSHA256 = ""
	event.RecordSHA256 = auditEventDigest(event)
	path := filepath.Join(filepath.Clean(store), "v1", "events.jsonl")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	if existing, err := os.ReadFile(path); err == nil && strings.Contains(string(existing), `"event_id":"`+event.EventID+`"`) {
		return nil
	}
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(append(data, '\n'))
	return err
}

func auditEventDigest(event AuditEvent) string {
	data, _ := json.Marshal(event)
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}

func AuditEventRecord(store string, event AuditEvent) error {
	return withKnowledgeStoreLock(store, func() error { return appendAuditEvent(store, event) })
}
