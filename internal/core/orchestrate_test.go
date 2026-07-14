package core

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestOrchestrateRoutePreparesOnlyThePreflightNativeHostContract(t *testing.T) {
	initial := Route("fix an SDK timeout bug", []Manifest{{
		ID: "code", Triggers: []string{"fix", "sdk", "bug"}, Status: "callable",
		Experts: []Expert{{ID: "implementation", Purpose: "implement", Independent: true, ModelClass: "sol"}},
	}})
	if len(initial.PreflightWorkers) != 1 {
		t.Fatalf("test route lacks preflight: %#v", initial)
	}
	result, err := OrchestrateRoute(initial, OrchestrationOptions{
		Dispatch: DispatchOptions{Workspace: t.TempDir(), OutputDir: t.TempDir(), Runner: func(_ context.Context, _ string, arguments []string) (CodexCommandResult, error) {
			t.Fatalf("orchestrate must not start external CLI: %#v", arguments)
			return CodexCommandResult{}, nil
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Stages) != 1 || result.Stages[0].Name != "preflight" || result.Stages[0].Results[0].Status != "native-host-dispatch-required" {
		t.Fatalf("preflight native-host contract was not prepared: %#v", result)
	}
	if !result.AjiMergeRequired || len(result.ResultHandles) != 0 {
		t.Fatalf("orchestrator failed to preserve its native merge boundary: %#v", result)
	}
}

func TestOrchestrateRouteBoundsIndependentParallelWorkers(t *testing.T) {
	route := Route("research the web", []Manifest{{ID: "search", Triggers: []string{"research"}, Status: "callable", Engines: []Engine{{ID: "web-research", Default: true}}}})
	var active, maximum int32
	result, err := OrchestrateRoute(route, OrchestrationOptions{
		MaxParallel: 2,
		Dispatch: DispatchOptions{Workspace: t.TempDir(), OutputDir: t.TempDir(), Runner: func(_ context.Context, _ string, arguments []string) (CodexCommandResult, error) {
			current := atomic.AddInt32(&active, 1)
			for {
				seen := atomic.LoadInt32(&maximum)
				if current <= seen || atomic.CompareAndSwapInt32(&maximum, seen, current) {
					break
				}
			}
			time.Sleep(30 * time.Millisecond)
			atomic.AddInt32(&active, -1)
			return CodexCommandResult{}, nil
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Stages) != 1 || len(result.Stages[0].Results) != 3 || maximum != 0 || len(result.FailedWorkers) != 0 {
		t.Fatalf("independent native-host contracts were not prepared without external execution: max=%d result=%#v", maximum, result)
	}
}

func TestEveryNonConversationalTaskGetsAWorkerButSimpleQuestionStaysDirect(t *testing.T) {
	if got := Route("分析这个方案是否可行", nil); len(got.Workers) != 1 || got.Workers[0].Model != "gpt-5.6-sol" {
		t.Fatalf("task did not receive an active Sol route: %#v", got)
	}
	if got := Route("无极军团是什么？", nil); len(got.Workers) != 0 || got.DelegationDecision.Reason != "simple-question-direct" {
		t.Fatalf("simple question should stay on Aji: %#v", got)
	}
}

func TestFailureAndCodeReviewUseIndependentDistilledBranches(t *testing.T) {
	failure := Route("debug timeout error", nil)
	if len(failure.Workers) != 2 || failure.Workers[0].ID != "root-cause" || failure.Workers[1].ID != "failure-evidence" || failure.Workers[0].Model != "gpt-5.6-sol" || failure.Workers[1].Model != "gpt-5.6-luna" {
		t.Fatalf("failure protocol was not made executable: %#v", failure.Workers)
	}
	if len(failure.Workers[0].TaskContract) == 0 || !containsString(workerProtocol("debug timeout error", "root-cause", "", failure.Workers[0].StableCapabilityPrefix), "reproduce or state why reproduction is unavailable") {
		t.Fatalf("failure contract omitted its protocol: %#v", failure.Workers[0])
	}
	review := Route("review this pull request", []Manifest{{ID: "code-review", Triggers: []string{"review", "pull request"}, Status: "callable"}})
	if len(review.Workers) != 2 || review.Workers[0].ID != "spec-conformance" || review.Workers[1].ID != "engineering-quality" || review.Workers[0].Model != "gpt-5.6-sol" || review.Workers[1].Model != "gpt-5.6-sol" {
		t.Fatalf("two-axis review was not made executable: %#v", review.Workers)
	}
}
