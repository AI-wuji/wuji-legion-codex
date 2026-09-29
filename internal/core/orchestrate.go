package core

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

// OrchestrationOptions contains only host controls. Execution nodes may receive
// scoped artifact writes; staff reconciliation and Aji reporting remain outside
// this deterministic adapter.
type OrchestrationOptions struct {
	Dispatch    DispatchOptions
	MaxParallel int
}

type OrchestrationStage struct {
	Name    string           `json:"name"`
	Results []DispatchResult `json:"results"`
}

type OrchestrationResult struct {
	InitialRoute                RouteResult          `json:"initial_route"`
	ExecutionRoute              RouteResult          `json:"execution_route"`
	Stages                      []OrchestrationStage `json:"stages"`
	ResultHandles               []string             `json:"result_handles"`
	FailedWorkers               []string             `json:"failed_workers,omitempty"`
	StaffReconciliationRequired bool                 `json:"staff_reconciliation_required"`
	AjiReportRequired           bool                 `json:"aji_report_required"`
	// AjiMergeRequired is retained only for JSON compatibility. New consumers
	// must use StaffReconciliationRequired and AjiReportRequired.
	AjiMergeRequired   bool   `json:"aji_merge_required,omitempty"`
	CompletionBoundary string `json:"completion_boundary"`
}

// OrchestrateRoute prepares native-host contracts in dependency order. The Go
// CLI never presents an external codex exec process as worker execution. The
// General Staff is represented by deterministic state transitions around these
// stages. It is not a resident child and therefore never blocks dispatch as a
// separate model invocation.
func OrchestrateRoute(initial RouteResult, options OrchestrationOptions) (OrchestrationResult, error) {
	if options.MaxParallel <= 0 {
		options.MaxParallel = 3
	}
	result := OrchestrationResult{
		InitialRoute:                initial,
		ExecutionRoute:              initial,
		StaffReconciliationRequired: initial.GeneralStaffRequired,
		AjiReportRequired:           true,
		CompletionBoundary:          "only the Desktop host may create native execution nodes and submit their receipts; this CLI prepares contracts only. Deterministic General Staff tracks stages and reconciles receipts without accepting completion; execution and independent verification evidence determine completion; Aji returns the final report only",
	}

	if !initial.GeneralStaffRequired && len(initial.PreflightWorkers) == 0 && len(initial.Workers) == 0 && len(initial.OfficerWorkers) == 0 {
		result.CompletionBoundary = "Aji handles this bounded request directly; no General Staff state or native worker is required"
		return result, nil
	}

	if len(initial.PreflightWorkers) > 0 {
		stage := OrchestrationStage{Name: "preflight", Results: make([]DispatchResult, 0, len(initial.PreflightWorkers))}
		for _, worker := range initial.PreflightWorkers {
			dispatch, err := DispatchWorker(worker, options.Dispatch)
			if err != nil {
				return result, fmt.Errorf("dispatch preflight worker %s: %w", worker.ID, err)
			}
			stage.Results = append(stage.Results, dispatch)
		}
		result.Stages = append(result.Stages, stage)
		// Preparation is not a preflight result. The host must run this stage,
		// inspect native evidence, and issue a fresh route before any execution
		// contracts may be prepared.
		return result, nil
	}

	if len(result.ExecutionRoute.Workers) > 0 {
		batches, err := roleWorkerBatches(result.ExecutionRoute.RoleGraph, result.ExecutionRoute.Workers, "worker-")
		if err != nil {
			return result, err
		}
		for index, batch := range batches {
			stageName := "workers"
			if len(batches) > 1 {
				stageName = fmt.Sprintf("workers-group-%d", index+1)
			}
			stage, err := dispatchParallel(stageName, batch, options.Dispatch, options.MaxParallel)
			if err != nil {
				return result, err
			}
			result.Stages = append(result.Stages, stage)
			failedDispatches(stage.Results, &result)
			if len(result.FailedWorkers) > 0 {
				break
			}
		}
	}
	if len(result.FailedWorkers) == 0 && len(result.ExecutionRoute.OfficerWorkers) > 0 {
		batches, err := roleWorkerBatches(result.ExecutionRoute.RoleGraph, result.ExecutionRoute.OfficerWorkers, "officer-")
		if err != nil {
			return result, err
		}
		for index, batch := range batches {
			stageName := "officers"
			if len(batches) > 1 {
				stageName = fmt.Sprintf("officers-group-%d", index+1)
			}
			stage, err := dispatchParallel(stageName, batch, options.Dispatch, options.MaxParallel)
			if err != nil {
				return result, err
			}
			result.Stages = append(result.Stages, stage)
			failedDispatches(stage.Results, &result)
			if len(result.FailedWorkers) > 0 {
				break
			}
		}
	}
	for _, stage := range result.Stages {
		for _, dispatch := range stage.Results {
			for _, attempt := range dispatch.Attempts {
				if attempt.ResultHandle != "" {
					result.ResultHandles = append(result.ResultHandles, attempt.ResultHandle)
				}
			}
		}
	}
	return result, nil
}

// roleWorkerBatches turns the route graph into dependency-aware dispatch
// groups. Only worker nodes are executable here; commander, staff and
// verification nodes remain logical graph state. Independent workers may run
// together, while a worker depending on another worker waits for its group.
func roleWorkerBatches(graph RoleTaskGraph, workers []WorkerTask, nodePrefix string) ([][]WorkerTask, error) {
	if len(workers) == 0 {
		return nil, nil
	}
	if graph.GraphID == "" {
		return [][]WorkerTask{append([]WorkerTask(nil), workers...)}, nil
	}
	if err := validateRoleTaskGraph(graph); err != nil {
		return nil, err
	}
	nodes := make(map[string]RoleTaskNode, len(graph.Nodes))
	for _, node := range graph.Nodes {
		nodes[node.ID] = node
	}
	pending := make(map[string]WorkerTask, len(workers))
	for _, worker := range workers {
		pending[nodePrefix+worker.ID] = worker
	}
	if len(graph.ParallelGroups) > 0 {
		return roleWorkerBatchesFromDeclaredGroups(graph, workers, nodePrefix, nodes)
	}
	batches := make([][]WorkerTask, 0, len(workers))
	for len(pending) > 0 {
		readyIDs := make([]string, 0, len(pending))
		for nodeID := range pending {
			node, ok := nodes[nodeID]
			if !ok {
				return nil, fmt.Errorf("role graph %s has no node for worker %s", graph.GraphID, strings.TrimPrefix(nodeID, nodePrefix))
			}
			ready := true
			for _, dependency := range node.DependsOn {
				if _, waiting := pending[dependency]; waiting {
					ready = false
					break
				}
			}
			if ready {
				readyIDs = append(readyIDs, nodeID)
			}
		}
		if len(readyIDs) == 0 {
			return nil, fmt.Errorf("role graph %s cannot schedule worker dependencies", graph.GraphID)
		}
		sort.Strings(readyIDs)
		batch := make([]WorkerTask, 0, len(readyIDs))
		for _, nodeID := range readyIDs {
			batch = append(batch, pending[nodeID])
			delete(pending, nodeID)
		}
		batches = append(batches, batch)
	}
	return batches, nil
}

func roleWorkerBatchesFromDeclaredGroups(graph RoleTaskGraph, workers []WorkerTask, nodePrefix string, nodes map[string]RoleTaskNode) ([][]WorkerTask, error) {
	workerByNode := make(map[string]WorkerTask, len(workers))
	for _, worker := range workers {
		workerByNode[nodePrefix+worker.ID] = worker
	}
	assigned := map[string]bool{}
	batches := make([][]WorkerTask, 0, len(graph.ParallelGroups))
	completed := map[string]bool{}
	// Route construction satisfies only these deterministic planning
	// prerequisites. Do not mutate their graph states to "complete": no
	// native role has run, and executable dependencies still need evidence.
	for nodeID, node := range nodes {
		if node.Kind == "intake" || node.Kind == "schedule" || node.Kind == "plan" || node.Kind == "commander" || node.Kind == "select" || node.Kind == "expert" || node.Kind == "accept" {
			completed[nodeID] = true
		}
	}
	for _, group := range graph.ParallelGroups {
		batch := make([]WorkerTask, 0, len(group))
		for _, nodeID := range group {
			worker, ok := workerByNode[nodeID]
			if !ok || assigned[nodeID] {
				continue
			}
			node := nodes[nodeID]
			for _, dependency := range node.DependsOn {
				dep, exists := nodes[dependency]
				if !exists {
					return nil, fmt.Errorf("role graph %s worker %s depends on missing node %s", graph.GraphID, nodeID, dependency)
				}
				if dep.Kind == "worker" || dep.Kind == "officer" || dep.Kind == "preflight" {
					if !completed[dependency] {
						return nil, fmt.Errorf("role graph %s parallel group schedules %s before dependency %s", graph.GraphID, nodeID, dependency)
					}
				} else if !completed[dependency] && dep.State != "complete" && dep.State != "verified" {
					// Only the explicit deterministic planning kinds above can
					// be satisfied by route construction.
					return nil, fmt.Errorf("role graph %s dependency %s is not complete before %s", graph.GraphID, dependency, nodeID)
				}
			}
			batch = append(batch, worker)
			assigned[nodeID] = true
		}
		if len(batch) > 0 {
			batches = append(batches, batch)
			for _, nodeID := range group {
				if assigned[nodeID] {
					completed[nodeID] = true
				}
			}
		}
	}
	if len(assigned) != len(workerByNode) {
		return nil, fmt.Errorf("role graph %s parallel groups omit executable workers", graph.GraphID)
	}
	return batches, nil
}

func failedDispatches(dispatches []DispatchResult, result *OrchestrationResult) bool {
	failed := false
	for _, dispatch := range dispatches {
		if dispatch.Status != "native-host-dispatch-required" {
			failed = true
			result.FailedWorkers = append(result.FailedWorkers, dispatch.WorkerID)
		}
	}
	return failed
}

func dispatchParallel(name string, workers []WorkerTask, options DispatchOptions, limit int) (OrchestrationStage, error) {
	stage := OrchestrationStage{Name: name, Results: make([]DispatchResult, len(workers))}
	semaphore := make(chan struct{}, limit)
	errs := make(chan error, len(workers))
	var group sync.WaitGroup
	for index, worker := range workers {
		group.Add(1)
		go func(index int, worker WorkerTask) {
			defer group.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			dispatch, err := DispatchWorker(worker, options)
			if err != nil {
				errs <- fmt.Errorf("dispatch %s worker %s: %w", name, worker.ID, err)
				return
			}
			stage.Results[index] = dispatch
		}(index, worker)
	}
	group.Wait()
	close(errs)
	if err := <-errs; err != nil {
		return stage, err
	}
	return stage, nil
}
