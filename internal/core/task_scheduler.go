package core

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

// TaskStatus is the lifecycle state controlled by TaskScheduler.
type TaskStatus string

const (
	TaskPlanned     TaskStatus = "planned"
	TaskRunning     TaskStatus = "running"
	TaskSucceeded   TaskStatus = "succeeded"
	TaskIncomplete  TaskStatus = "incomplete"
	TaskFailed      TaskStatus = "failed"
	TaskCancelled   TaskStatus = "cancelled"
	TaskInvalidated TaskStatus = "invalidated"
	TaskBlocked     TaskStatus = "blocked"
)

type TaskFailureKind string

const (
	TaskFailureNone                     TaskFailureKind = ""
	TaskFailureProviderBeforeGeneration TaskFailureKind = "provider-before-generation"
	TaskFailureGeneration               TaskFailureKind = "generation"
	TaskFailureContract                 TaskFailureKind = "contract"
	TaskFailureVerification             TaskFailureKind = "verification"
)

type TaskDecisionKind string

const (
	TaskDecisionNone       TaskDecisionKind = "none"
	TaskDecisionContinue   TaskDecisionKind = "continue"
	TaskDecisionRedispatch TaskDecisionKind = "redispatch"
	TaskDecisionReplan     TaskDecisionKind = "replan"
	TaskDecisionIgnored    TaskDecisionKind = "ignored"
	// TaskDecisionExhausted is terminal: the caller must replan instead of
	// silently creating another execution lease.
	TaskDecisionExhausted TaskDecisionKind = "exhausted"
)

const (
	taskSchedulerMaxNodes                         = 1024
	taskSchedulerMaxAttemptsPerNode        uint64 = 3
	taskSchedulerMaxRedispatchesPerLineage        = 2
)

// TaskNode describes work only. The scheduler never runs it.
type TaskNode struct {
	ID        string
	DependsOn []string
}

// TaskRun is the lease a caller may hand to an executor. Attempt rejects late
// results after a node has been retried, cancelled, or invalidated.
type TaskRun struct {
	NodeID  string
	Attempt uint64
}

type TaskResult struct {
	Run         TaskRun
	Status      TaskStatus
	FailureKind TaskFailureKind
}

type TaskDecision struct {
	Kind       TaskDecisionKind
	NodeID     string
	Redispatch *TaskNode
}

type scheduledTask struct {
	node    TaskNode
	status  TaskStatus
	attempt uint64
	lineage string
}

// TaskScheduler deterministically maintains a DAG. All state changes are
// serialized; callers may execute every run returned by DispatchReady in parallel.
type TaskScheduler struct {
	mu                    sync.Mutex
	nodes                 map[string]*scheduledTask
	redispatchesByLineage map[string]uint64
}

func NewTaskScheduler(nodes []TaskNode) (*TaskScheduler, error) {
	if len(nodes) > taskSchedulerMaxNodes {
		return nil, fmt.Errorf("task graph node limit exceeded: %d", taskSchedulerMaxNodes)
	}
	s := &TaskScheduler{nodes: make(map[string]*scheduledTask, len(nodes)), redispatchesByLineage: make(map[string]uint64)}
	for _, node := range nodes {
		node = normalizeTaskNode(node)
		if node.ID == "" {
			return nil, fmt.Errorf("task node id is required")
		}
		if _, exists := s.nodes[node.ID]; exists {
			return nil, fmt.Errorf("duplicate task node %q", node.ID)
		}
		s.nodes[node.ID] = &scheduledTask{node: node, status: TaskPlanned, lineage: node.ID}
	}
	for _, task := range s.nodes {
		for _, dependency := range task.node.DependsOn {
			if dependency == task.node.ID {
				return nil, fmt.Errorf("task node %q depends on itself", dependency)
			}
			if _, exists := s.nodes[dependency]; !exists {
				return nil, fmt.Errorf("task node %q depends on missing node %q", task.node.ID, dependency)
			}
		}
	}
	if hasTaskCycle(s.nodes) {
		return nil, fmt.Errorf("task graph contains a cycle")
	}
	return s, nil
}

// DispatchReady atomically leases every currently ready node in stable ID order.
func (s *TaskScheduler) DispatchReady() []TaskRun {
	s.mu.Lock()
	defer s.mu.Unlock()
	ids := s.readyIDs()
	runs := make([]TaskRun, 0, len(ids))
	for _, id := range ids {
		task := s.nodes[id]
		task.status = TaskRunning
		task.attempt++
		runs = append(runs, TaskRun{NodeID: id, Attempt: task.attempt})
	}
	return runs
}

// Resolve records an executor result and returns only a scheduling decision.
// A result for an old lease is ignored, so it cannot revive downstream work.
func (s *TaskScheduler) Resolve(result TaskResult) (TaskDecision, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	task, exists := s.nodes[result.Run.NodeID]
	if !exists {
		return TaskDecision{}, fmt.Errorf("task node %q is unknown", result.Run.NodeID)
	}
	if task.status != TaskRunning || task.attempt != result.Run.Attempt {
		return TaskDecision{Kind: TaskDecisionIgnored, NodeID: result.Run.NodeID}, nil
	}
	switch result.Status {
	case TaskSucceeded:
		task.status = TaskSucceeded
		return TaskDecision{Kind: TaskDecisionNone, NodeID: task.node.ID}, nil
	case TaskIncomplete:
		if task.attempt >= taskSchedulerMaxAttemptsPerNode {
			return s.exhaust(task)
		}
		task.status = TaskIncomplete
		return TaskDecision{Kind: TaskDecisionContinue, NodeID: task.node.ID}, nil
	case TaskFailed:
		switch result.FailureKind {
		case TaskFailureProviderBeforeGeneration:
			return s.redispatch(task)
		case TaskFailureContract, TaskFailureVerification:
			task.status = TaskFailed
			s.blockDescendants(task.node.ID)
			return TaskDecision{Kind: TaskDecisionReplan, NodeID: task.node.ID}, nil
		default:
			task.status = TaskFailed
			s.blockDescendants(task.node.ID)
			return TaskDecision{Kind: TaskDecisionNone, NodeID: task.node.ID}, nil
		}
	default:
		return TaskDecision{}, fmt.Errorf("result status %q is not resolvable", result.Status)
	}
}

func (s *TaskScheduler) Cancel(id string) error     { return s.terminate(id, TaskCancelled) }
func (s *TaskScheduler) Invalidate(id string) error { return s.terminate(id, TaskInvalidated) }

func (s *TaskScheduler) terminate(id string, status TaskStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	task, exists := s.nodes[id]
	if !exists {
		return fmt.Errorf("task node %q is unknown", id)
	}
	if status == TaskCancelled && terminalTaskStatus(task.status) {
		return nil
	}
	task.status = status
	s.propagate(id, status)
	return nil
}

func (s *TaskScheduler) Status(id string) (TaskStatus, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	task, exists := s.nodes[id]
	if !exists {
		return "", false
	}
	return task.status, true
}

func (s *TaskScheduler) redispatch(task *scheduledTask) (TaskDecision, error) {
	if s.redispatchesByLineage[task.lineage] >= taskSchedulerMaxRedispatchesPerLineage || len(s.nodes) >= taskSchedulerMaxNodes {
		return s.exhaust(task)
	}
	task.status = TaskFailed
	base := task.node.ID + "-redispatch-"
	n := 1
	for {
		candidate := fmt.Sprintf("%s%d", base, n)
		if _, exists := s.nodes[candidate]; !exists {
			redispatch := TaskNode{ID: candidate, DependsOn: append([]string(nil), task.node.DependsOn...)}
			s.nodes[candidate] = &scheduledTask{node: redispatch, status: TaskPlanned, lineage: task.lineage}
			s.redispatchesByLineage[task.lineage]++
			for _, child := range s.nodes {
				for i, dependency := range child.node.DependsOn {
					if dependency == task.node.ID {
						child.node.DependsOn[i] = candidate
					}
				}
			}
			return TaskDecision{Kind: TaskDecisionRedispatch, NodeID: task.node.ID, Redispatch: copyTaskNode(redispatch)}, nil
		}
		n++
	}
}

func (s *TaskScheduler) exhaust(task *scheduledTask) (TaskDecision, error) {
	task.status = TaskFailed
	s.blockDescendants(task.node.ID)
	return TaskDecision{Kind: TaskDecisionExhausted, NodeID: task.node.ID}, nil
}

func (s *TaskScheduler) readyIDs() []string {
	ids := make([]string, 0)
	for id, task := range s.nodes {
		if task.status != TaskPlanned && task.status != TaskIncomplete {
			continue
		}
		ready := true
		for _, dependency := range task.node.DependsOn {
			if s.nodes[dependency].status != TaskSucceeded {
				ready = false
				break
			}
		}
		if ready {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	return ids
}

func (s *TaskScheduler) blockDescendants(id string) { s.propagate(id, TaskBlocked) }

func (s *TaskScheduler) propagate(id string, status TaskStatus) {
	queue := []string{id}
	visited := map[string]bool{id: true}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, task := range s.nodes {
			if visited[task.node.ID] || task.status == TaskSucceeded || task.status == TaskCancelled || task.status == TaskInvalidated {
				continue
			}
			for _, dependency := range task.node.DependsOn {
				if dependency == current {
					task.status = status
					visited[task.node.ID] = true
					queue = append(queue, task.node.ID)
					break
				}
			}
		}
	}
}

func normalizeTaskNode(node TaskNode) TaskNode {
	node.ID = strings.TrimSpace(node.ID)
	node.DependsOn = append([]string(nil), node.DependsOn...)
	for i := range node.DependsOn {
		node.DependsOn[i] = strings.TrimSpace(node.DependsOn[i])
	}
	sort.Strings(node.DependsOn)
	return node
}
func copyTaskNode(node TaskNode) *TaskNode {
	node.DependsOn = append([]string(nil), node.DependsOn...)
	return &node
}
func terminalTaskStatus(status TaskStatus) bool {
	return status == TaskSucceeded || status == TaskCancelled || status == TaskInvalidated || status == TaskBlocked
}
func hasTaskCycle(nodes map[string]*scheduledTask) bool {
	visiting, visited := map[string]bool{}, map[string]bool{}
	var visit func(string) bool
	visit = func(id string) bool {
		if visiting[id] {
			return true
		}
		if visited[id] {
			return false
		}
		visiting[id] = true
		for _, d := range nodes[id].node.DependsOn {
			if visit(d) {
				return true
			}
		}
		delete(visiting, id)
		visited[id] = true
		return false
	}
	for id := range nodes {
		if visit(id) {
			return true
		}
	}
	return false
}
