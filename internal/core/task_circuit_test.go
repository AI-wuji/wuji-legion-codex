package core

import "testing"

func TestTaskCircuitBlocksDuplicateAndNoProgressLimit(t *testing.T) {
	store := t.TempDir()
	policy := TaskCircuitPolicy{ID: "bounded-repair-v1", MaxNoProgress: 2}
	first := TaskAttemptInput{TaskID: "repair-routing", StrategyID: "direct-fix", AttemptID: "attempt-a", Outcome: "no-progress"}

	allowed, err := CheckTaskCircuit(store, policy, first)
	if err != nil || !allowed.Allowed || allowed.Reason != "new-attempt" {
		t.Fatalf("initial task attempt was not allowed: result=%#v err=%v", allowed, err)
	}
	recorded, err := RecordTaskAttempt(store, policy, first)
	if err != nil || recorded.State.NoProgress != 1 || recorded.State.CircuitOpen {
		t.Fatalf("first no-progress result was not persisted: result=%#v err=%v", recorded, err)
	}

	duplicate, err := CheckTaskCircuit(store, policy, first)
	if err != nil || duplicate.Allowed || duplicate.Reason != "duplicate-no-progress-attempt" {
		t.Fatalf("duplicate no-progress attempt escaped the gate: result=%#v err=%v", duplicate, err)
	}

	second := TaskAttemptInput{TaskID: "repair-routing", StrategyID: "direct-fix", AttemptID: "attempt-b", Outcome: "failure"}
	recorded, err = RecordTaskAttempt(store, policy, second)
	if err != nil || !recorded.State.CircuitOpen || recorded.State.CircuitReason != "no-progress-limit" {
		t.Fatalf("no-progress threshold did not open the circuit: result=%#v err=%v", recorded, err)
	}
	blocked, err := CheckTaskCircuit(store, policy, TaskAttemptInput{TaskID: "repair-routing", StrategyID: "direct-fix", AttemptID: "attempt-c"})
	if err != nil || blocked.Allowed || blocked.Reason != "no-progress-limit" {
		t.Fatalf("open circuit did not block a new attempt: result=%#v err=%v", blocked, err)
	}

	changedStrategy, err := CheckTaskCircuit(store, policy, TaskAttemptInput{TaskID: "repair-routing", StrategyID: "verified-fix", AttemptID: "attempt-a"})
	if err != nil || !changedStrategy.Allowed {
		t.Fatalf("new strategy should have its own circuit: result=%#v err=%v", changedStrategy, err)
	}
}

func TestTaskCircuitIgnoresTransientFailureAndResetsOnProgress(t *testing.T) {
	store := t.TempDir()
	policy := TaskCircuitPolicy{ID: "bounded-repair-v1", MaxNoProgress: 2}
	transient := TaskAttemptInput{TaskID: "repair-routing", StrategyID: "direct-fix", AttemptID: "network-timeout", Outcome: "failure", TransientFailure: true}
	recorded, err := RecordTaskAttempt(store, policy, transient)
	if err != nil || recorded.State.NoProgress != 0 || recorded.State.CircuitOpen {
		t.Fatalf("transient failure consumed the circuit budget: result=%#v err=%v", recorded, err)
	}
	allowed, err := CheckTaskCircuit(store, policy, TaskAttemptInput{TaskID: transient.TaskID, StrategyID: transient.StrategyID, AttemptID: transient.AttemptID})
	if err != nil || !allowed.Allowed {
		t.Fatalf("transient failure blocked a retry: result=%#v err=%v", allowed, err)
	}
	_, err = RecordTaskAttempt(store, policy, TaskAttemptInput{TaskID: transient.TaskID, StrategyID: transient.StrategyID, AttemptID: "no-progress", Outcome: "no-progress"})
	if err != nil {
		t.Fatal(err)
	}
	recorded, err = RecordTaskAttempt(store, policy, TaskAttemptInput{TaskID: transient.TaskID, StrategyID: transient.StrategyID, AttemptID: "progress", Outcome: "progress"})
	if err != nil || recorded.State.NoProgress != 0 || recorded.State.LastProgressAt == "" {
		t.Fatalf("progress did not reset the no-progress count: result=%#v err=%v", recorded, err)
	}
}
