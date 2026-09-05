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
	requirementGraphSchemaVersion = 1
	requirementGraphMaxNodes      = 256
	requirementGraphMaxSources    = 32
	requirementGraphMaxSummary    = 4096
	requirementGraphMaxFields     = 16
	requirementGraphMaxFieldBytes = 512
)

type RequirementGraphNode struct {
	ID             string   `json:"id"`
	VersionID      string   `json:"version_id"`
	Kind           string   `json:"kind"`
	Revision       int      `json:"revision"`
	Status         string   `json:"status"`
	Summary        string   `json:"summary"`
	Goal           string   `json:"goal,omitempty"`
	Wants          []string `json:"wants,omitempty"`
	Avoids         []string `json:"avoids,omitempty"`
	Constraints    []string `json:"constraints,omitempty"`
	Preferences    []string `json:"preferences,omitempty"`
	Priority       string   `json:"priority,omitempty"`
	Acceptance     []string `json:"acceptance,omitempty"`
	Decisions      []string `json:"decisions,omitempty"`
	OpenQuestions  []string `json:"open_questions,omitempty"`
	Conflicts      []string `json:"conflicts,omitempty"`
	Sources        []string `json:"sources"`
	SourceMessages []string `json:"source_messages,omitempty"`
	DependsOn      []string `json:"depends_on,omitempty"`
	Supersedes     string   `json:"supersedes,omitempty"`
}

type RequirementGraph struct {
	SchemaVersion int                    `json:"schema_version"`
	Nodes         []RequirementGraphNode `json:"nodes"`
}

type RequirementInput struct {
	ID             string
	Summary        string
	Goal           string
	Wants          []string
	Avoids         []string
	Constraints    []string
	Preferences    []string
	Priority       string
	Acceptance     []string
	Decisions      []string
	OpenQuestions  []string
	Conflicts      []string
	Sources        []string
	SourceMessages []string
	DependsOn      []string
}

type DecisionInput struct {
	ID             string
	Summary        string
	Goal           string
	Wants          []string
	Avoids         []string
	Constraints    []string
	Preferences    []string
	Priority       string
	Acceptance     []string
	Decisions      []string
	OpenQuestions  []string
	Conflicts      []string
	Sources        []string
	SourceMessages []string
	Requirements   []string
	Status         string
}

type RequirementGraphProjection struct {
	SchemaVersion int                    `json:"schema_version"`
	Handle        string                 `json:"handle"`
	PayloadSHA256 string                 `json:"payload_sha256"`
	PayloadBytes  int                    `json:"payload_bytes"`
	Target        string                 `json:"target"`
	Nodes         []RequirementGraphNode `json:"nodes"`
	Payload       string                 `json:"payload"`
	ArtifactPath  string                 `json:"artifact_path,omitempty"`
}

type requirementGraphPayload struct {
	SchemaVersion int                    `json:"schema_version"`
	Target        string                 `json:"target"`
	Nodes         []RequirementGraphNode `json:"nodes"`
}

func DefaultRequirementGraphStore() string {
	if value := strings.TrimSpace(os.Getenv("WUJI_REQUIREMENT_GRAPH_DIR")); value != "" {
		return filepath.Clean(value)
	}
	return filepath.Join(".wuji", "requirement-graph")
}

func UpsertRequirement(store string, input RequirementInput) (RequirementGraphNode, error) {
	if err := validateRequirementInput(input); err != nil {
		return RequirementGraphNode{}, err
	}
	var node RequirementGraphNode
	err := withKnowledgeStoreLock(store, func() error {
		graph, err := loadRequirementGraph(store)
		if err != nil {
			return err
		}
		current := latestGraphNode(graph.Nodes, input.ID, "requirement")
		normalized := requirementNodeFields(input)
		if current != nil && current.Status == "active" && sameRequirementFields(*current, normalized) {
			node = *current
			return nil
		}
		revision := 1
		supersedes := ""
		if current != nil {
			revision = current.Revision + 1
			supersedes = current.VersionID
			for index := range graph.Nodes {
				if graph.Nodes[index].VersionID == current.VersionID {
					graph.Nodes[index].Status = "superseded"
				}
			}
		}
		node = normalized
		node.ID, node.VersionID, node.Kind, node.Revision = input.ID, graphVersionID(input.ID, revision), "requirement", revision
		node.Status, node.Supersedes = "active", supersedes
		graph.Nodes = append(graph.Nodes, node)
		if len(graph.Nodes) > requirementGraphMaxNodes {
			return fmt.Errorf("requirement graph exceeds %d nodes", requirementGraphMaxNodes)
		}
		if err := writeRequirementGraph(store, graph); err != nil {
			return err
		}
		return AuditEventRecord(auditStoreFor(store), AuditEvent{EventType: "requirement-recorded", Actor: "aji", Authority: "aji-merge", Target: node.VersionID, InputRevision: node.Supersedes, ResultHandle: "wuji-requirement://" + node.VersionID})
	})
	if err != nil {
		return RequirementGraphNode{}, err
	}
	return node, nil
}

func RecordDecision(store string, input DecisionInput) (RequirementGraphNode, error) {
	if err := validateDecisionInput(input); err != nil {
		return RequirementGraphNode{}, err
	}
	var node RequirementGraphNode
	err := withKnowledgeStoreLock(store, func() error {
		graph, err := loadRequirementGraph(store)
		if err != nil {
			return err
		}
		dependencies, err := resolveRequirementDependencies(graph.Nodes, input.Requirements)
		if err != nil {
			return err
		}
		current := latestGraphNode(graph.Nodes, input.ID, "decision")
		status := strings.ToLower(strings.TrimSpace(input.Status))
		if status == "" {
			status = "proposed"
		}
		normalized := decisionNodeFields(input)
		normalized.DependsOn = dependencies
		if current != nil && current.Status == status && sameRequirementFields(*current, normalized) {
			node = *current
			return nil
		}
		revision := 1
		supersedes := ""
		if current != nil {
			revision = current.Revision + 1
			supersedes = current.VersionID
			for index := range graph.Nodes {
				if graph.Nodes[index].VersionID == current.VersionID {
					graph.Nodes[index].Status = "superseded"
				}
			}
		}
		node = normalized
		node.ID, node.VersionID, node.Kind, node.Revision = input.ID, graphVersionID(input.ID, revision), "decision", revision
		node.Status, node.Supersedes = status, supersedes
		graph.Nodes = append(graph.Nodes, node)
		if len(graph.Nodes) > requirementGraphMaxNodes {
			return fmt.Errorf("requirement graph exceeds %d nodes", requirementGraphMaxNodes)
		}
		if err := writeRequirementGraph(store, graph); err != nil {
			return err
		}
		return AuditEventRecord(auditStoreFor(store), AuditEvent{EventType: "decision-recorded", Actor: "aji", Authority: "aji-merge", Target: node.VersionID, InputRevision: strings.Join(node.DependsOn, ","), ResultHandle: "wuji-decision://" + node.VersionID})
	})
	if err != nil {
		return RequirementGraphNode{}, err
	}
	return node, nil
}

func ProjectRequirementGraph(store, target string, maxBytes int) (RequirementGraphProjection, error) {
	if maxBytes < 256 || maxBytes > 4096 {
		return RequirementGraphProjection{}, fmt.Errorf("projection max bytes must be between 256 and 4096")
	}
	graph, err := loadRequirementGraph(store)
	if err != nil {
		return RequirementGraphProjection{}, err
	}
	targetNode := resolveGraphTarget(graph.Nodes, target)
	if targetNode == nil {
		return RequirementGraphProjection{}, fmt.Errorf("requirement graph target %q is not found", target)
	}
	nodesByVersion := make(map[string]RequirementGraphNode, len(graph.Nodes))
	for _, node := range graph.Nodes {
		nodesByVersion[node.VersionID] = node
	}
	nodes := []RequirementGraphNode{*targetNode}
	for _, dependency := range targetNode.DependsOn {
		node, ok := nodesByVersion[dependency]
		if !ok {
			return RequirementGraphProjection{}, fmt.Errorf("target %q references missing dependency %q", targetNode.VersionID, dependency)
		}
		nodes = append(nodes, node)
	}
	payloadNodes := make([]RequirementGraphNode, 0, len(nodes))
	for _, node := range nodes {
		candidate := append(payloadNodes, node)
		payload, err := json.Marshal(requirementGraphPayload{SchemaVersion: requirementGraphSchemaVersion, Target: targetNode.VersionID, Nodes: candidate})
		if err != nil {
			return RequirementGraphProjection{}, err
		}
		if len(payload) > maxBytes {
			if len(payloadNodes) == 0 {
				return RequirementGraphProjection{}, fmt.Errorf("target %q exceeds the %d-byte projection budget", targetNode.VersionID, maxBytes)
			}
			break
		}
		payloadNodes = candidate
	}
	payload, err := json.Marshal(requirementGraphPayload{SchemaVersion: requirementGraphSchemaVersion, Target: targetNode.VersionID, Nodes: payloadNodes})
	if err != nil {
		return RequirementGraphProjection{}, err
	}
	digest := sha256.Sum256(payload)
	encoded := hex.EncodeToString(digest[:])
	return RequirementGraphProjection{
		SchemaVersion: requirementGraphSchemaVersion, Handle: "wuji-requirement-graph://sha256/" + encoded,
		PayloadSHA256: encoded, PayloadBytes: len(payload), Target: targetNode.VersionID, Nodes: payloadNodes, Payload: string(payload),
	}, nil
}

func WriteRequirementGraphProjection(projection RequirementGraphProjection, artifactDir string) (string, error) {
	if err := validateRequirementGraphProjection(projection); err != nil {
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

func LoadRequirementGraphProjection(path string) (RequirementGraphProjection, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return RequirementGraphProjection{}, err
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	var projection RequirementGraphProjection
	if err := decoder.Decode(&projection); err != nil {
		return RequirementGraphProjection{}, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return RequirementGraphProjection{}, fmt.Errorf("multiple JSON values are not allowed")
		}
		return RequirementGraphProjection{}, err
	}
	if err := validateRequirementGraphProjection(projection); err != nil {
		return RequirementGraphProjection{}, err
	}
	projection.ArtifactPath = path
	return projection, nil
}

func validateRequirementInput(input RequirementInput) error {
	if !componentIDPattern.MatchString(strings.TrimSpace(input.ID)) {
		return fmt.Errorf("requirement id is invalid")
	}
	if err := validateGraphContent(input.Summary, input.Sources); err != nil {
		return err
	}
	return validateRequirementFields(input.Goal, input.Wants, input.Avoids, input.Constraints, input.Preferences, input.Priority, input.Acceptance, input.Decisions, input.OpenQuestions, input.Conflicts, input.SourceMessages, input.DependsOn)
}

func validateDecisionInput(input DecisionInput) error {
	if !componentIDPattern.MatchString(strings.TrimSpace(input.ID)) {
		return fmt.Errorf("decision id is invalid")
	}
	if err := validateGraphContent(input.Summary, input.Sources); err != nil {
		return err
	}
	if err := validateRequirementFields(input.Goal, input.Wants, input.Avoids, input.Constraints, input.Preferences, input.Priority, input.Acceptance, input.Decisions, input.OpenQuestions, input.Conflicts, input.SourceMessages, input.Requirements); err != nil {
		return err
	}
	switch strings.ToLower(strings.TrimSpace(input.Status)) {
	case "", "proposed", "accepted", "rejected":
	default:
		return fmt.Errorf("decision status is invalid")
	}
	if len(input.Requirements) == 0 || len(input.Requirements) > requirementGraphMaxSources {
		return fmt.Errorf("decision requires between 1 and %d requirement references", requirementGraphMaxSources)
	}
	for _, requirement := range input.Requirements {
		if strings.TrimSpace(requirement) == "" || len(requirement) > 160 {
			return fmt.Errorf("decision requirement reference is invalid")
		}
	}
	return nil
}

func requirementNodeFields(input RequirementInput) RequirementGraphNode {
	return RequirementGraphNode{
		Summary: strings.TrimSpace(input.Summary), Goal: strings.TrimSpace(input.Goal),
		Wants: normalizedGraphList(input.Wants), Avoids: normalizedGraphList(input.Avoids),
		Constraints: normalizedGraphList(input.Constraints), Preferences: normalizedGraphList(input.Preferences),
		Priority: strings.ToLower(strings.TrimSpace(input.Priority)), Acceptance: normalizedGraphList(input.Acceptance),
		Decisions: normalizedGraphList(input.Decisions), OpenQuestions: normalizedGraphList(input.OpenQuestions),
		Conflicts: normalizedGraphList(input.Conflicts), Sources: normalizedGraphSources(input.Sources),
		SourceMessages: normalizedGraphList(input.SourceMessages), DependsOn: normalizedGraphList(input.DependsOn),
	}
}

func decisionNodeFields(input DecisionInput) RequirementGraphNode {
	return RequirementGraphNode{
		Summary: strings.TrimSpace(input.Summary), Goal: strings.TrimSpace(input.Goal),
		Wants: normalizedGraphList(input.Wants), Avoids: normalizedGraphList(input.Avoids),
		Constraints: normalizedGraphList(input.Constraints), Preferences: normalizedGraphList(input.Preferences),
		Priority: strings.ToLower(strings.TrimSpace(input.Priority)), Acceptance: normalizedGraphList(input.Acceptance),
		Decisions: normalizedGraphList(input.Decisions), OpenQuestions: normalizedGraphList(input.OpenQuestions),
		Conflicts: normalizedGraphList(input.Conflicts), Sources: normalizedGraphSources(input.Sources),
		SourceMessages: normalizedGraphList(input.SourceMessages),
	}
}

func sameRequirementFields(current, next RequirementGraphNode) bool {
	return current.Summary == next.Summary && current.Goal == next.Goal && current.Priority == next.Priority &&
		sameStringSlice(current.Wants, next.Wants) && sameStringSlice(current.Avoids, next.Avoids) &&
		sameStringSlice(current.Constraints, next.Constraints) && sameStringSlice(current.Preferences, next.Preferences) &&
		sameStringSlice(current.Acceptance, next.Acceptance) && sameStringSlice(current.Decisions, next.Decisions) &&
		sameStringSlice(current.OpenQuestions, next.OpenQuestions) && sameStringSlice(current.Conflicts, next.Conflicts) &&
		sameStringSlice(current.Sources, next.Sources) && sameStringSlice(current.SourceMessages, next.SourceMessages) &&
		sameStringSlice(current.DependsOn, next.DependsOn)
}

func validateRequirementFields(goal string, wants, avoids, constraints, preferences []string, priority string, acceptance, decisions, openQuestions, conflicts, sourceMessages, dependencies []string) error {
	if len([]byte(strings.TrimSpace(goal))) > requirementGraphMaxFieldBytes {
		return fmt.Errorf("goal must not exceed %d bytes", requirementGraphMaxFieldBytes)
	}
	if strings.TrimSpace(priority) != "" && len([]byte(strings.TrimSpace(priority))) > 32 {
		return fmt.Errorf("priority must not exceed 32 bytes")
	}
	for name, values := range map[string][]string{
		"wants": wants, "avoids": avoids, "constraints": constraints, "preferences": preferences,
		"acceptance": acceptance, "decisions": decisions, "open questions": openQuestions,
		"conflicts": conflicts, "source messages": sourceMessages, "dependencies": dependencies,
	} {
		if len(values) > requirementGraphMaxFields {
			return fmt.Errorf("%s cannot contain more than %d entries", name, requirementGraphMaxFields)
		}
		for _, value := range values {
			value = strings.TrimSpace(value)
			if value == "" || len([]byte(value)) > requirementGraphMaxFieldBytes || strings.ContainsAny(value, "\r\n\t") {
				return fmt.Errorf("%s entry is invalid", name)
			}
			for _, pattern := range knowledgeSecretPatterns {
				if pattern.MatchString(value) {
					return fmt.Errorf("%s must not contain a secret", name)
				}
			}
		}
	}
	return nil
}

func normalizedGraphList(values []string) []string {
	result := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" && !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	sort.Strings(result)
	return result
}

func validateGraphContent(summary string, sources []string) error {
	summary = strings.TrimSpace(summary)
	if summary == "" || len([]byte(summary)) > requirementGraphMaxSummary {
		return fmt.Errorf("summary is required and must not exceed %d bytes", requirementGraphMaxSummary)
	}
	for _, pattern := range knowledgeSecretPatterns {
		if pattern.MatchString(summary) {
			return fmt.Errorf("summary must not contain a secret")
		}
	}
	if len(sources) == 0 || len(sources) > requirementGraphMaxSources {
		return fmt.Errorf("graph node requires between 1 and %d source references", requirementGraphMaxSources)
	}
	for _, source := range sources {
		source = strings.TrimSpace(source)
		if source == "" || len(source) > 512 || strings.ContainsAny(source, "\r\n\t") || !strings.Contains(source, ":") {
			return fmt.Errorf("source reference is invalid")
		}
		for _, pattern := range knowledgeSecretPatterns {
			if pattern.MatchString(source) {
				return fmt.Errorf("source reference must not contain a secret")
			}
		}
	}
	return nil
}

func loadRequirementGraph(store string) (RequirementGraph, error) {
	path := requirementGraphPath(store)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return RequirementGraph{SchemaVersion: requirementGraphSchemaVersion, Nodes: []RequirementGraphNode{}}, nil
	}
	if err != nil {
		return RequirementGraph{}, err
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	var graph RequirementGraph
	if err := decoder.Decode(&graph); err != nil {
		return RequirementGraph{}, fmt.Errorf("decode requirement graph: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return RequirementGraph{}, fmt.Errorf("decode requirement graph: multiple JSON values are not allowed")
		}
		return RequirementGraph{}, err
	}
	if err := validateRequirementGraph(graph); err != nil {
		return RequirementGraph{}, err
	}
	return graph, nil
}

func writeRequirementGraph(store string, graph RequirementGraph) error {
	if err := validateRequirementGraph(graph); err != nil {
		return err
	}
	path := requirementGraphPath(store)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(graph, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	temporary, err := os.CreateTemp(filepath.Dir(path), ".requirement-graph-*.json")
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

func validateRequirementGraph(graph RequirementGraph) error {
	if graph.SchemaVersion != requirementGraphSchemaVersion {
		return fmt.Errorf("unsupported requirement graph schema version")
	}
	if len(graph.Nodes) > requirementGraphMaxNodes {
		return fmt.Errorf("requirement graph exceeds %d nodes", requirementGraphMaxNodes)
	}
	versions := make(map[string]bool, len(graph.Nodes))
	for _, node := range graph.Nodes {
		if !componentIDPattern.MatchString(node.ID) || node.Revision < 1 || node.VersionID != graphVersionID(node.ID, node.Revision) || versions[node.VersionID] {
			return fmt.Errorf("requirement graph node identity is invalid")
		}
		versions[node.VersionID] = true
		switch node.Kind {
		case "requirement":
			if node.Status != "active" && node.Status != "superseded" {
				return fmt.Errorf("requirement graph requirement status is invalid")
			}
		case "decision":
			if node.Status != "proposed" && node.Status != "accepted" && node.Status != "rejected" && node.Status != "superseded" {
				return fmt.Errorf("requirement graph decision status is invalid")
			}
		default:
			return fmt.Errorf("requirement graph node kind is invalid")
		}
		if err := validateGraphContent(node.Summary, node.Sources); err != nil {
			return err
		}
	}
	for _, node := range graph.Nodes {
		for _, dependency := range node.DependsOn {
			if !versions[dependency] {
				return fmt.Errorf("requirement graph dependency %q is missing", dependency)
			}
		}
		if node.Supersedes != "" && !versions[node.Supersedes] {
			return fmt.Errorf("requirement graph supersedes reference %q is missing", node.Supersedes)
		}
	}
	return nil
}

func validateRequirementGraphProjection(projection RequirementGraphProjection) error {
	if projection.SchemaVersion != requirementGraphSchemaVersion || projection.PayloadBytes != len([]byte(projection.Payload)) || !taskCircuitHashPattern.MatchString(projection.PayloadSHA256) {
		return fmt.Errorf("requirement graph projection metadata is invalid")
	}
	digest := sha256.Sum256([]byte(projection.Payload))
	if !strings.EqualFold(hex.EncodeToString(digest[:]), projection.PayloadSHA256) || projection.Handle != "wuji-requirement-graph://sha256/"+strings.ToLower(projection.PayloadSHA256) {
		return fmt.Errorf("requirement graph projection hash does not match its payload")
	}
	var payload requirementGraphPayload
	if err := json.Unmarshal([]byte(projection.Payload), &payload); err != nil {
		return fmt.Errorf("requirement graph projection payload is invalid: %w", err)
	}
	if payload.SchemaVersion != projection.SchemaVersion || payload.Target != projection.Target {
		return fmt.Errorf("requirement graph projection payload metadata does not match")
	}
	payloadNodes, err := json.Marshal(payload.Nodes)
	if err != nil {
		return err
	}
	projectionNodes, err := json.Marshal(projection.Nodes)
	if err != nil {
		return err
	}
	if string(payloadNodes) != string(projectionNodes) {
		return fmt.Errorf("requirement graph projection nodes do not match its payload")
	}
	return nil
}

func resolveRequirementDependencies(nodes []RequirementGraphNode, references []string) ([]string, error) {
	dependencies := make([]string, 0, len(references))
	seen := map[string]bool{}
	for _, reference := range references {
		node := resolveGraphTarget(nodes, reference)
		if node == nil || node.Kind != "requirement" || node.Status != "active" {
			return nil, fmt.Errorf("requirement reference %q is not an active requirement", reference)
		}
		if !seen[node.VersionID] {
			dependencies = append(dependencies, node.VersionID)
			seen[node.VersionID] = true
		}
	}
	sort.Strings(dependencies)
	return dependencies, nil
}

func resolveGraphTarget(nodes []RequirementGraphNode, target string) *RequirementGraphNode {
	target = strings.TrimSpace(target)
	for index := range nodes {
		if nodes[index].VersionID == target {
			return &nodes[index]
		}
	}
	var latest *RequirementGraphNode
	for index := range nodes {
		node := &nodes[index]
		if node.ID != target || node.Status == "superseded" {
			continue
		}
		if latest == nil || node.Revision > latest.Revision {
			latest = node
		}
	}
	return latest
}

func latestGraphNode(nodes []RequirementGraphNode, id, kind string) *RequirementGraphNode {
	var latest *RequirementGraphNode
	for index := range nodes {
		node := &nodes[index]
		if node.ID == id && node.Kind == kind && (latest == nil || node.Revision > latest.Revision) {
			latest = node
		}
	}
	return latest
}

func normalizedGraphSources(sources []string) []string {
	set := map[string]bool{}
	for _, source := range sources {
		if source = strings.TrimSpace(source); source != "" {
			set[source] = true
		}
	}
	result := make([]string, 0, len(set))
	for source := range set {
		result = append(result, source)
	}
	sort.Strings(result)
	return result
}

func graphVersionID(id string, revision int) string {
	return fmt.Sprintf("%s@%d", id, revision)
}

func requirementGraphPath(store string) string {
	return filepath.Join(filepath.Clean(store), "v1", "graph.json")
}
