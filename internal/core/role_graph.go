package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

const roleTaskGraphSchemaVersion = 1

// RoleTaskGraph is the compact execution map owned by one Legion role.
// A role graph is deliberately separate from the conversation transcript:
// child roles receive graph identity plus content-addressed context handles,
// not a replay of the complete parent conversation.
type RoleTaskGraph struct {
	SchemaVersion  string         `json:"schema_version"`
	GraphID        string         `json:"graph_id"`
	RoleID         string         `json:"role_id"`
	RoleKind       string         `json:"role_kind"`
	Scope          string         `json:"scope"`
	GraphVersion   string         `json:"graph_version"`
	ParentGraphID  string         `json:"parent_graph_id,omitempty"`
	RootNode       string         `json:"root_node"`
	Nodes          []RoleTaskNode `json:"nodes"`
	ParallelGroups [][]string     `json:"parallel_groups,omitempty"`
	ContextHandles []string       `json:"context_handles,omitempty"`
	MoE            string         `json:"moe"`
	PonyTail       string         `json:"ponytail"`
	SHA256         string         `json:"sha256"`
}

type RoleTaskNode struct {
	ID             string   `json:"id"`
	RoleID         string   `json:"role_id"`
	Kind           string   `json:"kind"`
	DependsOn      []string `json:"depends_on,omitempty"`
	Objective      string   `json:"objective"`
	Acceptance     []string `json:"acceptance,omitempty"`
	State          string   `json:"state"`
	ContextHandles []string `json:"context_handles,omitempty"`
	ExecutionRef   string   `json:"execution_ref,omitempty"`
	ConflictKey    string   `json:"conflict_key,omitempty"`
	Writes         bool     `json:"writes,omitempty"`
}

// ExpertTaskGraph remains as a source-compatible name for older callers. An
// expert graph is now just the same graph contract every other role uses.
type ExpertTaskGraph = RoleTaskGraph
type ExpertTaskGraphNode = RoleTaskNode

func makeRoleTaskGraph(roleID, roleKind, scope, graphVersion, parentGraphID, root string, nodes []RoleTaskNode, parallelGroups [][]string, contextHandles []string) (RoleTaskGraph, error) {
	roleID = strings.TrimSpace(roleID)
	roleKind = strings.TrimSpace(roleKind)
	if roleID == "" || roleKind == "" || strings.TrimSpace(root) == "" {
		return RoleTaskGraph{}, fmt.Errorf("role graph requires role, kind, and root node")
	}
	graph := RoleTaskGraph{
		SchemaVersion:  fmt.Sprintf("wuji-role-graph-v%d", roleTaskGraphSchemaVersion),
		GraphID:        roleGraphID(roleID, scope, graphVersion),
		RoleID:         roleID,
		RoleKind:       roleKind,
		Scope:          strings.TrimSpace(scope),
		GraphVersion:   strings.TrimSpace(graphVersion),
		ParentGraphID:  strings.TrimSpace(parentGraphID),
		RootNode:       root,
		Nodes:          cloneRoleNodes(nodes),
		ParallelGroups: cloneParallelGroups(parallelGroups),
		ContextHandles: uniqueStrings(contextHandles),
		MoE:            "sparse-role-moe",
		PonyTail:       ponyTailDoctrine,
	}
	if err := validateRoleTaskGraphShape(graph); err != nil {
		return RoleTaskGraph{}, err
	}
	graph.SHA256 = hashRoleTaskGraph(graph)
	return graph, nil
}

func roleGraphID(roleID, scope, graphVersion string) string {
	payload := strings.Join([]string{"wuji-role-graph-v1", roleID, scope, graphVersion}, "\n")
	return "wuji-graph://sha256/" + sha256Hex([]byte(payload))[:24]
}

func hashRoleTaskGraph(graph RoleTaskGraph) string {
	graph.SHA256 = ""
	data, _ := json.Marshal(graph)
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}

func validateRoleTaskGraph(graph RoleTaskGraph) error {
	if err := validateRoleTaskGraphShape(graph); err != nil {
		return err
	}
	if strings.TrimSpace(graph.SHA256) == "" || hashRoleTaskGraph(graph) != graph.SHA256 {
		return fmt.Errorf("role graph %s hash mismatch", graph.GraphID)
	}
	return nil
}

func validateRoleTaskGraphShape(graph RoleTaskGraph) error {
	if graph.SchemaVersion != fmt.Sprintf("wuji-role-graph-v%d", roleTaskGraphSchemaVersion) ||
		strings.TrimSpace(graph.GraphID) == "" ||
		strings.TrimSpace(graph.RoleID) == "" ||
		strings.TrimSpace(graph.RoleKind) == "" ||
		strings.TrimSpace(graph.RootNode) == "" ||
		graph.MoE != "sparse-role-moe" ||
		graph.PonyTail != ponyTailDoctrine {
		return fmt.Errorf("role graph has an invalid identity or doctrine")
	}
	nodes := make(map[string]RoleTaskNode, len(graph.Nodes))
	for _, node := range graph.Nodes {
		if strings.TrimSpace(node.ID) == "" || strings.TrimSpace(node.RoleID) == "" ||
			strings.TrimSpace(node.Kind) == "" || strings.TrimSpace(node.Objective) == "" ||
			!validRoleTaskNodeState(node.State) {
			return fmt.Errorf("role graph %s contains an invalid node %q (role=%q kind=%q objective=%q state=%q)", graph.GraphID, node.ID, node.RoleID, node.Kind, node.Objective, node.State)
		}
		if _, exists := nodes[node.ID]; exists {
			return fmt.Errorf("role graph %s contains duplicate node %s", graph.GraphID, node.ID)
		}
		nodes[node.ID] = node
	}
	if _, ok := nodes[graph.RootNode]; !ok {
		return fmt.Errorf("role graph %s root node %s is missing", graph.GraphID, graph.RootNode)
	}
	for _, node := range graph.Nodes {
		for _, dependency := range node.DependsOn {
			if dependency == node.ID {
				return fmt.Errorf("role graph %s node %s depends on itself", graph.GraphID, node.ID)
			}
			if _, ok := nodes[dependency]; !ok {
				return fmt.Errorf("role graph %s node %s depends on missing %s", graph.GraphID, node.ID, dependency)
			}
		}
	}
	if hasRoleGraphCycle(graph.Nodes) {
		return fmt.Errorf("role graph %s contains a dependency cycle", graph.GraphID)
	}
	for _, group := range graph.ParallelGroups {
		seen := map[string]bool{}
		for _, nodeID := range group {
			node, ok := nodes[nodeID]
			if !ok {
				return fmt.Errorf("role graph %s parallel group references missing %s", graph.GraphID, nodeID)
			}
			if seen[nodeID] {
				return fmt.Errorf("role graph %s parallel group repeats %s", graph.GraphID, nodeID)
			}
			seen[nodeID] = true
			for _, otherID := range group {
				if otherID == nodeID {
					continue
				}
				other := nodes[otherID]
				if containsString(node.DependsOn, otherID) || containsString(other.DependsOn, nodeID) {
					return fmt.Errorf("role graph %s puts dependent nodes in one parallel group", graph.GraphID)
				}
				if node.Writes && other.Writes && node.ConflictKey != "" && node.ConflictKey == other.ConflictKey {
					return fmt.Errorf("role graph %s puts conflicting writers in one parallel group", graph.GraphID)
				}
			}
		}
	}
	return nil
}

func validRoleTaskNodeState(state string) bool {
	switch state {
	case "planned", "ready", "active", "blocked", "verified", "complete", "failed":
		return true
	default:
		return false
	}
}

func hasRoleGraphCycle(nodes []RoleTaskNode) bool {
	byID := make(map[string]RoleTaskNode, len(nodes))
	for _, node := range nodes {
		byID[node.ID] = node
	}
	visiting := map[string]bool{}
	visited := map[string]bool{}
	var visit func(string) bool
	visit = func(id string) bool {
		if visiting[id] {
			return true
		}
		if visited[id] {
			return false
		}
		visiting[id] = true
		for _, dependency := range byID[id].DependsOn {
			if visit(dependency) {
				return true
			}
		}
		delete(visiting, id)
		visited[id] = true
		return false
	}
	for _, node := range nodes {
		if visit(node.ID) {
			return true
		}
	}
	return false
}

func cloneRoleNodes(nodes []RoleTaskNode) []RoleTaskNode {
	cloned := make([]RoleTaskNode, len(nodes))
	for index, node := range nodes {
		cloned[index] = node
		cloned[index].DependsOn = append([]string(nil), node.DependsOn...)
		cloned[index].Acceptance = append([]string(nil), node.Acceptance...)
		cloned[index].ContextHandles = append([]string(nil), node.ContextHandles...)
	}
	return cloned
}

func cloneParallelGroups(groups [][]string) [][]string {
	cloned := make([][]string, len(groups))
	for index, group := range groups {
		cloned[index] = append([]string(nil), group...)
		sort.Strings(cloned[index])
	}
	return cloned
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]bool, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func graphContextHandles(context DelegationContext) []string {
	if context.Handle == "" {
		return nil
	}
	return []string{context.Handle}
}

func roleGraphNode(id, roleID, kind, objective string, dependsOn []string, executionRef string, writes bool) RoleTaskNode {
	if strings.TrimSpace(objective) == "" {
		objective = "execute the assigned bounded role task"
	}
	conflictKey := ""
	if writes {
		conflictKey = executionRef
	}
	return RoleTaskNode{
		ID:           id,
		RoleID:       roleID,
		Kind:         kind,
		DependsOn:    append([]string(nil), dependsOn...),
		Objective:    objective,
		Acceptance:   []string{"return bounded evidence", "do not claim completion without host receipt"},
		State:        "planned",
		ExecutionRef: executionRef,
		ConflictKey:  conflictKey,
		Writes:       writes,
	}
}

func annotateWorkerRoleGraph(worker *WorkerTask, roleID, roleKind string, parent RoleTaskGraph, child RoleTaskGraph) error {
	if worker == nil {
		return fmt.Errorf("worker is required")
	}
	if parent.GraphVersion == "" || parent.GraphVersion != child.GraphVersion {
		return fmt.Errorf("worker %s graph parent or version binding is inconsistent", worker.ID)
	}
	if err := validateRoleTaskGraph(parent); err != nil {
		return err
	}
	if err := validateRoleTaskGraph(child); err != nil {
		return err
	}
	var contract workerContract
	if err := json.Unmarshal([]byte(worker.TaskContract), &contract); err != nil {
		return err
	}
	contract.RoleID = roleID
	contract.ParentGraphID = parent.GraphID
	contract.TaskGraphID = child.GraphID
	contract.TaskGraphSHA256 = child.SHA256
	data, err := json.Marshal(contract)
	if err != nil {
		return err
	}
	if len(data) > maxTaskContractBytes {
		return fmt.Errorf("worker %s role graph contract exceeds %d bytes", worker.ID, maxTaskContractBytes)
	}
	worker.TaskContract = string(data)
	worker.TaskContractSHA256 = sha256Hex(data)
	worker.AllocatedTaskContractBytes = len(data)
	worker.RoleID = roleID
	worker.RoleKind = roleKind
	worker.ParentGraphID = parent.GraphID
	worker.TaskGraphID = child.GraphID
	worker.TaskGraphSHA256 = child.SHA256
	worker.PonyTail = ponyTailDoctrine
	return nil
}

func makeRoleGraphSet(query, capability string, secondaryCapabilities []string, direct bool, context DelegationContext, expertRoute *ExpertRouteDecision, secondaryRoutes map[string]ExpertRouteDecision, preflightWorkers, workers, officerWorkers []WorkerTask) (RoleTaskGraph, map[string]RoleTaskGraph, []WorkerTask, []WorkerTask, []WorkerTask, error) {
	roleGraphs := map[string]RoleTaskGraph{}
	handles := graphContextHandles(context)
	graphVersion := routeVersion
	rootID := "aji-intake"
	rootNodes := []RoleTaskNode{
		roleGraphNode(rootID, "aji", "intake", "interpret the user request and choose the minimum correct path", nil, "", false),
	}
	if direct {
		reportID := "aji-report"
		rootNodes = append(rootNodes, roleGraphNode(reportID, "aji", "report", "answer the user directly without creating unnecessary delegation", []string{rootID}, "", false))
		root, err := makeRoleTaskGraph("aji", "aji", query, graphVersion, "", rootID, rootNodes, nil, handles)
		if err != nil {
			return RoleTaskGraph{}, nil, nil, nil, nil, err
		}
		roleGraphs["aji"] = root
		return root, roleGraphs, preflightWorkers, workers, officerWorkers, nil
	}

	staffID := "general-staff"
	commanderID := capability
	staffNode := roleGraphNode(staffID, "general-staff", "schedule", "compress the request into bounded role work and reconcile child evidence", []string{rootID}, "", false)
	commanderNode := roleGraphNode("commander-"+commanderID, commanderID, "commander", "coordinate the selected domain expert team", []string{staffID}, "", false)
	rootNodes = append(rootNodes, staffNode, commanderNode)
	graphSecondary := make([]string, 0, len(secondaryCapabilities))
	for _, secondary := range secondaryCapabilities {
		if secondary == "" || secondary == capability || secondary == "search" || secondary == responsePolicyCapabilityID {
			continue
		}
		if !containsString(graphSecondary, secondary) {
			graphSecondary = append(graphSecondary, secondary)
			secondaryNode := roleGraphNode("commander-"+secondary, secondary, "commander", "coordinate the selected secondary domain team", []string{staffID}, "", false)
			rootNodes = append(rootNodes, secondaryNode)
		}
	}
	sort.Strings(graphSecondary)
	rootGraphID := roleGraphID("aji", query, graphVersion)
	staffGraph, err := makeRoleTaskGraph("general-staff", "staff", query, graphVersion, rootGraphID, "staff-intake", []RoleTaskNode{
		roleGraphNode("staff-intake", "general-staff", "intake", "receive the compressed Aji task contract", nil, "", false),
		roleGraphNode("staff-plan", "general-staff", "plan", "choose the smallest necessary role set and dependency order", []string{"staff-intake"}, "", false),
		roleGraphNode("staff-reconcile", "general-staff", "reconcile", "reconcile child receipts and leave completion to evidence", []string{"staff-plan"}, "", false),
	}, nil, handles)
	if err != nil {
		return RoleTaskGraph{}, nil, nil, nil, nil, err
	}
	roleGraphs["general-staff"] = staffGraph

	if expertRoute != nil && expertRoute.Selection.State == "selected" && len(expertRoute.BoundWorkers) > 0 {
		expertID := expertRoute.Selection.ExpertID
		expertNodeID := "expert-" + expertID
		expertNode := roleGraphNode(expertNodeID, expertID, "expert", "execute the selected professional sub-task", []string{"commander-" + commanderID}, "", false)
		rootNodes = append(rootNodes, expertNode)
		expertGraph, err := makeExpertRoleGraph(expertID, query, graphVersion, roleGraphID("commander:"+commanderID, query, graphVersion), handles)
		if err != nil {
			return RoleTaskGraph{}, nil, nil, nil, nil, err
		}
		roleGraphs["expert:"+expertID] = expertGraph
		expertRoute.TaskGraph = expertGraph
		expertRoute.ExpertGraphs = map[string]ExpertTaskGraph{expertID: expertGraph}
	}

	commanderGraph, err := makeCommanderRoleGraph(commanderID, query, graphVersion, rootGraphID, expertRoute, handles)
	if err != nil {
		return RoleTaskGraph{}, nil, nil, nil, nil, err
	}
	roleGraphs["commander:"+commanderID] = commanderGraph
	for _, secondary := range graphSecondary {
		route, hasRoute := secondaryRoutes[secondary]
		var routePtr *ExpertRouteDecision
		if hasRoute {
			routeCopy := route
			routePtr = &routeCopy
			if routeCopy.Selection.State == "selected" && len(routeCopy.BoundWorkers) > 0 {
				expertID := routeCopy.Selection.ExpertID
				expertNodeID := "expert-" + secondary + "-" + expertID
				expertNode := roleGraphNode(expertNodeID, expertID, "expert", "execute the selected secondary professional sub-task", []string{"commander-" + secondary}, "", false)
				rootNodes = append(rootNodes, expertNode)
				expertGraph, graphErr := makeExpertRoleGraph(expertID, query, graphVersion, roleGraphID("commander:"+secondary, query, graphVersion), handles)
				if graphErr != nil {
					return RoleTaskGraph{}, nil, nil, nil, nil, graphErr
				}
				roleGraphs["expert:"+secondary+":"+expertID] = expertGraph
				routeCopy.TaskGraph = expertGraph
				routeCopy.ExpertGraphs = map[string]ExpertTaskGraph{expertID: expertGraph}
			}
			secondaryRoutes[secondary] = routeCopy
		}
		secondaryGraph, graphErr := makeCommanderRoleGraph(secondary, query, graphVersion, rootGraphID, routePtr, handles)
		if graphErr != nil {
			return RoleTaskGraph{}, nil, nil, nil, nil, graphErr
		}
		roleGraphs["commander:"+secondary] = secondaryGraph
	}
	for _, worker := range preflightWorkers {
		node := roleGraphNode("preflight-"+worker.ID, "general-staff", "preflight", worker.Purpose, []string{staffID}, worker.ID, worker.Writes)
		node.State = "planned"
		rootNodes = append(rootNodes, node)
	}
	for _, worker := range workers {
		workerCapability := workerCapabilityForRole(worker.ID, capability, graphSecondary)
		deps := []string{"commander-" + workerCapability}
		if workerCapability == capability && expertRoute != nil && expertRoute.Selection.State == "selected" && len(expertRoute.BoundWorkers) > 0 {
			deps = append(deps, "expert-"+expertRoute.Selection.ExpertID)
		}
		if secondaryRoute, ok := secondaryRoutes[workerCapability]; ok && secondaryRoute.Selection.State == "selected" && len(secondaryRoute.BoundWorkers) > 0 {
			deps = append(deps, "expert-"+workerCapability+"-"+secondaryRoute.Selection.ExpertID)
		}
		if workerCapability != capability && secondaryNeedsPrimary(query, capability, workerCapability) {
			for _, primaryWorker := range workers {
				if workerCapabilityForRole(primaryWorker.ID, capability, graphSecondary) == capability {
					deps = append(deps, "worker-"+primaryWorker.ID)
				}
			}
		}
		if len(preflightWorkers) > 0 {
			deps = append(deps, "preflight-"+preflightWorkers[0].ID)
		}
		roleID := worker.RoleID
		if roleID == "" {
			roleID = "expert:" + worker.ID
		}
		rootNodes = append(rootNodes, roleGraphNode("worker-"+worker.ID, roleID, "worker", worker.Purpose, deps, worker.ID, worker.Writes))
	}
	for _, worker := range officerWorkers {
		officerNode := roleGraphNode("officer-"+worker.ID, "officer:"+worker.ID, "officer", worker.Purpose, []string{staffID}, worker.ID, worker.Writes)
		officerNode.State = "planned"
		rootNodes = append(rootNodes, officerNode)
	}
	verificationDeps := []string{"commander-" + commanderID}
	for _, secondary := range graphSecondary {
		verificationDeps = append(verificationDeps, "commander-"+secondary)
	}
	for _, worker := range workers {
		verificationDeps = append(verificationDeps, "worker-"+worker.ID)
	}
	for _, worker := range officerWorkers {
		verificationDeps = append(verificationDeps, "officer-"+worker.ID)
	}
	rootNodes = append(rootNodes,
		roleGraphNode("verification", "verification", "verify", "check execution evidence, scope, and acceptance", verificationDeps, "", false),
		roleGraphNode("aji-report", "aji", "report", "summarize verified results and remaining uncertainty for the user", []string{"verification"}, "", false),
	)
	parallel := [][]string{}
	workerGroup := []string{}
	for _, worker := range workers {
		workerGroup = append(workerGroup, "worker-"+worker.ID)
	}
	parallel = append(parallel, independentRoleGroups(rootNodes, workerGroup)...)
	officerGroup := make([]string, 0, len(officerWorkers))
	for _, worker := range officerWorkers {
		officerGroup = append(officerGroup, "officer-"+worker.ID)
	}
	parallel = append(parallel, independentRoleGroups(rootNodes, officerGroup)...)
	root, err := makeRoleTaskGraph("aji", "aji", query, graphVersion, "", rootID, rootNodes, parallel, handles)
	if err != nil {
		return RoleTaskGraph{}, nil, nil, nil, nil, err
	}
	roleGraphs["aji"] = root
	verificationGraph, err := makeRoleTaskGraph("verification", "verification", query, graphVersion, root.GraphID, "verification-intake", []RoleTaskNode{
		roleGraphNode("verification-intake", "verification", "intake", "receive child execution receipts", nil, "", false),
		roleGraphNode("verification-check", "verification", "check", "check hashes, scope, acceptance, and required evidence", []string{"verification-intake"}, "", false),
		roleGraphNode("verification-report", "verification", "report", "return verification evidence and unresolved gaps", []string{"verification-check"}, "", false),
	}, nil, handles)
	if err != nil {
		return RoleTaskGraph{}, nil, nil, nil, nil, err
	}
	roleGraphs["verification"] = verificationGraph
	reportGraph, err := makeRoleTaskGraph("aji-report", "report", query, graphVersion, root.GraphID, "report-intake", []RoleTaskNode{
		roleGraphNode("report-intake", "aji-report", "intake", "receive compressed verified evidence", nil, "", false),
		roleGraphNode("report-write", "aji-report", "write", "tell the user what is verified, unknown, and still pending", []string{"report-intake"}, "", false),
	}, nil, handles)
	if err != nil {
		return RoleTaskGraph{}, nil, nil, nil, nil, err
	}
	roleGraphs["aji-report"] = reportGraph

	for index := range preflightWorkers {
		child, err := makeWorkerRoleGraph(preflightWorkers[index], graphVersion, "general-staff", staffID, handles)
		if err != nil {
			return RoleTaskGraph{}, nil, nil, nil, nil, err
		}
		roleGraphs["worker:"+preflightWorkers[index].ID] = child
		if err := annotateWorkerRoleGraph(&preflightWorkers[index], "general-staff", "preflight-worker", roleGraphs["general-staff"], child); err != nil {
			return RoleTaskGraph{}, nil, nil, nil, nil, err
		}
	}
	for index := range workers {
		roleID := workers[index].RoleID
		if roleID == "" {
			roleID = "expert:" + workers[index].ID
		}
		child, err := makeWorkerRoleGraph(workers[index], graphVersion, roleID, root.GraphID, handles)
		if err != nil {
			return RoleTaskGraph{}, nil, nil, nil, nil, err
		}
		roleGraphs["worker:"+workers[index].ID] = child
		if err := annotateWorkerRoleGraph(&workers[index], roleID, "execution-worker", root, child); err != nil {
			return RoleTaskGraph{}, nil, nil, nil, nil, err
		}
	}
	for index := range officerWorkers {
		child, err := makeWorkerRoleGraph(officerWorkers[index], graphVersion, "officer:"+officerWorkers[index].ID, root.GraphID, handles)
		if err != nil {
			return RoleTaskGraph{}, nil, nil, nil, nil, err
		}
		roleGraphs["worker:"+officerWorkers[index].ID] = child
		if err := annotateWorkerRoleGraph(&officerWorkers[index], "officer:"+officerWorkers[index].ID, "officer-worker", root, child); err != nil {
			return RoleTaskGraph{}, nil, nil, nil, nil, err
		}
	}
	return root, roleGraphs, preflightWorkers, workers, officerWorkers, nil
}

func workerCapabilityForRole(workerID, primary string, secondary []string) string {
	for _, capability := range secondary {
		if strings.HasPrefix(workerID, "secondary-"+capability+"-") {
			return capability
		}
	}
	return primary
}

func secondaryNeedsPrimary(query, primary, secondary string) bool {
	if primary == secondary {
		return false
	}
	if (primary == "image" && secondary == "video") || (primary == "video" && secondary == "image") {
		return containsAny(query, "做成视频", "图生视频", "再做成视频", "into a video", "make a video", "image to video")
	}
	return false
}

func independentRoleGroups(nodes []RoleTaskNode, nodeIDs []string) [][]string {
	if len(nodeIDs) < 2 {
		return nil
	}
	byID := make(map[string]RoleTaskNode, len(nodes))
	for _, node := range nodes {
		byID[node.ID] = node
	}
	sort.Strings(nodeIDs)
	groups := make([][]string, 0, len(nodeIDs))
	for _, nodeID := range nodeIDs {
		node := byID[nodeID]
		placed := false
		for index := range groups {
			conflict := false
			for _, otherID := range groups[index] {
				other := byID[otherID]
				if containsString(node.DependsOn, otherID) || containsString(other.DependsOn, nodeID) ||
					(node.Writes && other.Writes && node.ConflictKey != "" && node.ConflictKey == other.ConflictKey) {
					conflict = true
					break
				}
			}
			if !conflict {
				groups[index] = append(groups[index], nodeID)
				placed = true
				break
			}
		}
		if !placed {
			groups = append(groups, []string{nodeID})
		}
	}
	return groups
}

func dependencyRoleGroups(nodes []RoleTaskNode, nodeIDs []string) [][]string {
	if len(nodeIDs) == 0 {
		return nil
	}
	byID := make(map[string]RoleTaskNode, len(nodes))
	pending := make(map[string]bool, len(nodeIDs))
	for _, node := range nodes {
		byID[node.ID] = node
	}
	for _, id := range nodeIDs {
		pending[id] = true
	}
	groups := make([][]string, 0, len(nodeIDs))
	completed := map[string]bool{}
	for len(pending) > 0 {
		ready := make([]string, 0, len(pending))
		for id := range pending {
			node := byID[id]
			ok := true
			for _, dependency := range node.DependsOn {
				if pending[dependency] || (byID[dependency].Kind == "worker" || byID[dependency].Kind == "officer" || byID[dependency].Kind == "preflight") && !completed[dependency] {
					ok = false
					break
				}
			}
			if ok {
				ready = append(ready, id)
			}
		}
		if len(ready) == 0 {
			return nil
		}
		sort.Strings(ready)
		groups = append(groups, ready)
		for _, id := range ready {
			delete(pending, id)
			completed[id] = true
		}
	}
	return groups
}

func makeCommanderRoleGraph(commanderID, query, graphVersion, parent string, expertRoute *ExpertRouteDecision, handles []string) (RoleTaskGraph, error) {
	graphRoleID := "commander:" + commanderID
	intake := "commander-intake"
	nodes := []RoleTaskNode{
		roleGraphNode(intake, graphRoleID, "intake", "receive the bounded domain assignment", nil, "", false),
	}
	parallel := [][]string{}
	if expertRoute != nil && expertRoute.Selection.State == "selected" && len(expertRoute.BoundWorkers) > 0 {
		expertID := expertRoute.Selection.ExpertID
		nodes = append(nodes,
			roleGraphNode("commander-select", graphRoleID, "select", "activate only the minimum necessary expert", []string{intake}, "", false),
			roleGraphNode("expert-"+expertID, graphRoleID, "expert", "activate the selected professional sub-task", []string{"commander-select"}, "", false),
			roleGraphNode("commander-accept", graphRoleID, "accept", "accept expert evidence against the domain mission", []string{"commander-select", "expert-" + expertID}, "", false),
		)
	} else {
		nodes = append(nodes, roleGraphNode("commander-accept", graphRoleID, "accept", "retain the original Aji route when no expert is uniquely selected", []string{intake}, "", false))
	}
	graph, err := makeRoleTaskGraph(graphRoleID, "commander", query, graphVersion, parent, intake, nodes, parallel, handles)
	if expertRoute != nil {
		expertRoute.TaskGraph = graph
	}
	return graph, err
}

func makeExpertRoleGraph(expertID, query, graphVersion, parent string, handles []string) (RoleTaskGraph, error) {
	graphRoleID := "expert:" + expertID
	intake := "expert-intake"
	nodes := []RoleTaskNode{
		roleGraphNode(intake, graphRoleID, "intake", "receive the commander assignment", nil, "", false),
		roleGraphNode("expert-execute", graphRoleID, "execute", "perform the smallest professional sub-task", []string{intake}, "", false),
		roleGraphNode("expert-verify", graphRoleID, "verify", "check the sub-task result and report evidence", []string{"expert-execute"}, "", false),
	}
	return makeRoleTaskGraph(graphRoleID, "expert", query, graphVersion, parent, intake, nodes, nil, handles)
}

func makeWorkerRoleGraph(worker WorkerTask, graphVersion, roleID, parent string, handles []string) (RoleTaskGraph, error) {
	graphRoleID := "worker:" + worker.ID
	root := "worker-intake"
	nodes := []RoleTaskNode{
		roleGraphNode(root, graphRoleID, "intake", "receive the complete bounded worker contract", nil, worker.ID, worker.Writes),
		roleGraphNode("worker-execute", graphRoleID, "execute", worker.Purpose, []string{root}, worker.ID, worker.Writes),
		roleGraphNode("worker-verify", graphRoleID, "verify", "return execution evidence and do not claim final completion", []string{"worker-execute"}, worker.ID, false),
	}
	return makeRoleTaskGraph(graphRoleID, "worker", worker.Purpose, graphVersion, parent, root, nodes, nil, handles)
}

func refreshWorkerGraphContract(worker *WorkerTask) error {
	if worker == nil || worker.TaskContract == "" {
		return nil
	}
	var contract workerContract
	if err := json.Unmarshal([]byte(worker.TaskContract), &contract); err != nil {
		return err
	}
	contract.RoleID = worker.RoleID
	contract.ParentGraphID = worker.ParentGraphID
	contract.TaskGraphID = worker.TaskGraphID
	contract.TaskGraphSHA256 = worker.TaskGraphSHA256
	data, err := json.Marshal(contract)
	if err != nil {
		return err
	}
	if len(data) > maxTaskContractBytes {
		return fmt.Errorf("worker %s role graph contract exceeds %d bytes", worker.ID, maxTaskContractBytes)
	}
	worker.TaskContract = string(data)
	worker.TaskContractSHA256 = sha256Hex(data)
	worker.AllocatedTaskContractBytes = len(data)
	return nil
}
