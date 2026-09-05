package core

import (
	"fmt"
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
		StaffReconciliationRequired: true,
		AjiReportRequired:           true,
		CompletionBoundary:          "only the Desktop host may create native execution nodes and submit their receipts; this CLI prepares contracts only. Deterministic General Staff tracks stages and reconciles receipts without accepting completion; execution and independent verification evidence determine completion; Aji returns the final report only",
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
		stage, err := dispatchParallel("workers", result.ExecutionRoute.Workers, options.Dispatch, options.MaxParallel)
		if err != nil {
			return result, err
		}
		result.Stages = append(result.Stages, stage)
		failedDispatches(stage.Results, &result)
	}
	if len(result.FailedWorkers) == 0 && len(result.ExecutionRoute.OfficerWorkers) > 0 {
		stage, err := dispatchParallel("officers", result.ExecutionRoute.OfficerWorkers, options.Dispatch, options.MaxParallel)
		if err != nil {
			return result, err
		}
		result.Stages = append(result.Stages, stage)
		failedDispatches(stage.Results, &result)
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
