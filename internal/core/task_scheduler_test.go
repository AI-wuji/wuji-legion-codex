package core

import (
	"fmt"
	"reflect"
	"testing"
)

func TestTaskSchedulerDispatchesIndependentReadyNodesTogether(t *testing.T) {
	s := mustTaskScheduler(t, TaskNode{ID: "b"}, TaskNode{ID: "a"}, TaskNode{ID: "child", DependsOn: []string{"a"}})
	runs := s.DispatchReady()
	if got, want := runs, []TaskRun{{NodeID: "a", Attempt: 1}, {NodeID: "b", Attempt: 1}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("ready batch = %#v, want %#v", got, want)
	}
	if runs := s.DispatchReady(); len(runs) != 0 {
		t.Fatalf("already leased nodes dispatched again: %#v", runs)
	}
}

func TestTaskSchedulerWaitsForAllDependencies(t *testing.T) {
	s := mustTaskScheduler(t, TaskNode{ID: "a"}, TaskNode{ID: "b"}, TaskNode{ID: "child", DependsOn: []string{"a", "b"}})
	runs := s.DispatchReady()
	for _, run := range runs {
		if run.NodeID == "a" {
			if _, err := s.Resolve(TaskResult{Run: run, Status: TaskSucceeded}); err != nil {
				t.Fatal(err)
			}
		}
	}
	if runs := s.DispatchReady(); len(runs) != 0 {
		t.Fatalf("child ran before b completed: %#v", runs)
	}
	for _, run := range runs {
		if run.NodeID == "b" {
			if _, err := s.Resolve(TaskResult{Run: run, Status: TaskSucceeded}); err != nil {
				t.Fatal(err)
			}
		}
	}
	runs = s.DispatchReady()
	if got, want := runs, []TaskRun{{NodeID: "child", Attempt: 1}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("dependent run = %#v, want %#v", got, want)
	}
}

func TestTaskSchedulerFailureDecisions(t *testing.T) {
	t.Run("incomplete continues same node", func(t *testing.T) {
		s := mustTaskScheduler(t, TaskNode{ID: "work"})
		run := s.DispatchReady()[0]
		decision, err := s.Resolve(TaskResult{Run: run, Status: TaskIncomplete})
		if err != nil || decision.Kind != TaskDecisionContinue {
			t.Fatalf("decision=%#v err=%v", decision, err)
		}
		if got := s.DispatchReady()[0]; got.NodeID != "work" || got.Attempt != 2 {
			t.Fatalf("retry = %#v", got)
		}
	})
	t.Run("provider failure explicitly redispatches", func(t *testing.T) {
		s := mustTaskScheduler(t, TaskNode{ID: "work"}, TaskNode{ID: "child", DependsOn: []string{"work"}})
		run := s.DispatchReady()[0]
		decision, err := s.Resolve(TaskResult{Run: run, Status: TaskFailed, FailureKind: TaskFailureProviderBeforeGeneration})
		if err != nil || decision.Kind != TaskDecisionRedispatch || decision.Redispatch == nil {
			t.Fatalf("decision=%#v err=%v", decision, err)
		}
		if decision.Redispatch.ID != "work-redispatch-1" {
			t.Fatalf("unexpected redispatch %#v", decision.Redispatch)
		}
		run = s.DispatchReady()[0]
		if run.NodeID != "work-redispatch-1" {
			t.Fatalf("hidden fallback or wrong node: %#v", run)
		}
	})
	for _, kind := range []TaskFailureKind{TaskFailureContract, TaskFailureVerification} {
		t.Run(string(kind)+" requires replan", func(t *testing.T) {
			s := mustTaskScheduler(t, TaskNode{ID: "work"}, TaskNode{ID: "child", DependsOn: []string{"work"}})
			run := s.DispatchReady()[0]
			decision, err := s.Resolve(TaskResult{Run: run, Status: TaskFailed, FailureKind: kind})
			if err != nil || decision.Kind != TaskDecisionReplan {
				t.Fatalf("decision=%#v err=%v", decision, err)
			}
			if status, _ := s.Status("child"); status != TaskBlocked {
				t.Fatalf("child status = %s", status)
			}
		})
	}
}

func TestTaskSchedulerCancellationPropagatesAndRejectsLateResult(t *testing.T) {
	s := mustTaskScheduler(t, TaskNode{ID: "root"}, TaskNode{ID: "child", DependsOn: []string{"root"}}, TaskNode{ID: "leaf", DependsOn: []string{"child"}})
	run := s.DispatchReady()[0]
	if err := s.Cancel("root"); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"root", "child", "leaf"} {
		if status, _ := s.Status(id); status != TaskCancelled {
			t.Fatalf("%s status = %s", id, status)
		}
	}
	decision, err := s.Resolve(TaskResult{Run: run, Status: TaskSucceeded})
	if err != nil || decision.Kind != TaskDecisionIgnored {
		t.Fatalf("late decision=%#v err=%v", decision, err)
	}
	if runs := s.DispatchReady(); len(runs) != 0 {
		t.Fatalf("cancelled graph produced ready work: %#v", runs)
	}
}

func TestTaskSchedulerStopsIncompleteContinuationAtHardAttemptLimit(t *testing.T) {
	s := mustTaskScheduler(t, TaskNode{ID: "work"}, TaskNode{ID: "child", DependsOn: []string{"work"}})
	for attempt := uint64(1); attempt <= taskSchedulerMaxAttemptsPerNode; attempt++ {
		run := s.DispatchReady()[0]
		decision, err := s.Resolve(TaskResult{Run: run, Status: TaskIncomplete})
		if err != nil {
			t.Fatal(err)
		}
		if attempt < taskSchedulerMaxAttemptsPerNode && decision.Kind != TaskDecisionContinue {
			t.Fatalf("attempt %d decision=%#v", attempt, decision)
		}
		if attempt == taskSchedulerMaxAttemptsPerNode && decision.Kind != TaskDecisionExhausted {
			t.Fatalf("terminal decision=%#v", decision)
		}
	}
	if status, _ := s.Status("child"); status != TaskBlocked {
		t.Fatalf("dependent status = %s", status)
	}
	if runs := s.DispatchReady(); len(runs) != 0 {
		t.Fatalf("exhausted task resumed: %#v", runs)
	}
}

func TestTaskSchedulerCapsProviderRedispatchLineage(t *testing.T) {
	s := mustTaskScheduler(t, TaskNode{ID: "work"}, TaskNode{ID: "child", DependsOn: []string{"work"}})
	for n := uint64(0); n <= taskSchedulerMaxRedispatchesPerLineage; n++ {
		runs := s.DispatchReady()
		if len(runs) != 1 {
			t.Fatalf("redispatch %d runs=%#v", n, runs)
		}
		decision, err := s.Resolve(TaskResult{Run: runs[0], Status: TaskFailed, FailureKind: TaskFailureProviderBeforeGeneration})
		if err != nil {
			t.Fatal(err)
		}
		if n < taskSchedulerMaxRedispatchesPerLineage && decision.Kind != TaskDecisionRedispatch {
			t.Fatalf("redispatch %d decision=%#v", n, decision)
		}
		if n == taskSchedulerMaxRedispatchesPerLineage && decision.Kind != TaskDecisionExhausted {
			t.Fatalf("exhaustion decision=%#v", decision)
		}
	}
	if status, _ := s.Status("child"); status != TaskBlocked {
		t.Fatalf("dependent status = %s", status)
	}
	if len(s.nodes) != 1+int(taskSchedulerMaxRedispatchesPerLineage)+1 {
		t.Fatalf("lineage created unbounded nodes: %d", len(s.nodes))
	}
}

func TestTaskSchedulerRejectsOversizedGraph(t *testing.T) {
	nodes := make([]TaskNode, taskSchedulerMaxNodes+1)
	for i := range nodes {
		nodes[i].ID = fmt.Sprintf("node-%d", i)
	}
	if _, err := NewTaskScheduler(nodes); err == nil {
		t.Fatal("oversized graph was accepted")
	}
}

func TestTaskSchedulerCancellationTraversesSharedDescendantsOnce(t *testing.T) {
	nodes := []TaskNode{{ID: "root"}}
	previous := []string{"root"}
	for layer := 0; layer < 8; layer++ {
		current := []string{fmt.Sprintf("left-%d", layer), fmt.Sprintf("right-%d", layer)}
		for _, id := range current {
			nodes = append(nodes, TaskNode{ID: id, DependsOn: append([]string(nil), previous...)})
		}
		previous = current
	}
	nodes = append(nodes, TaskNode{ID: "leaf", DependsOn: previous})
	s := mustTaskScheduler(t, nodes...)
	if err := s.Cancel("root"); err != nil {
		t.Fatal(err)
	}
	for _, node := range nodes {
		if status, _ := s.Status(node.ID); status != TaskCancelled {
			t.Fatalf("%s status=%s", node.ID, status)
		}
	}
}

func mustTaskScheduler(t *testing.T, nodes ...TaskNode) *TaskScheduler {
	t.Helper()
	s, err := NewTaskScheduler(nodes)
	if err != nil {
		t.Fatal(err)
	}
	return s
}
