package core

import (
	"os"
	"sync"
	"testing"
	"time"
)

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

func TestNativeTaskClaimIsAtomicAndTaskWide(t *testing.T) {
	store := t.TempDir()
	policy := TaskCircuitPolicy{ID: "native-v1", MaxNoProgress: 2, MaxAttempts: 3, DeadlineSeconds: 60, LeaseSeconds: 30}
	inputs := []TaskAttemptInput{{TaskID: "task", StrategyID: "a", AttemptID: "one"}, {TaskID: "task", StrategyID: "b", AttemptID: "two"}}
	results := make([]NativeTaskClaimResult, 2)
	errs := make([]error, 2)
	var wg sync.WaitGroup
	for i := range inputs {
		wg.Add(1)
		go func(i int) { defer wg.Done(); results[i], errs[i] = ClaimNativeTask(store, policy, inputs[i]) }(i)
	}
	wg.Wait()
	allowed := 0
	for i := range results {
		if errs[i] != nil {
			t.Fatal(errs[i])
		}
		if results[i].Allowed {
			allowed++
		} else if results[i].Reason != "active-lease" {
			t.Fatalf("unexpected loser: %#v", results[i])
		}
	}
	if allowed != 1 {
		t.Fatalf("got %d successful claims", allowed)
	}
}

func TestNativeTaskAttemptsDuplicatesAndPolicyAreFailClosed(t *testing.T) {
	store := t.TempDir()
	policy := TaskCircuitPolicy{ID: "native-v1", MaxNoProgress: 4, MaxAttempts: 2, DeadlineSeconds: 60, LeaseSeconds: 30}
	first := TaskAttemptInput{TaskID: "task", StrategyID: "a", AttemptID: "one"}
	claim, err := ClaimNativeTask(store, policy, first)
	if err != nil || !claim.Allowed {
		t.Fatalf("claim: %#v %v", claim, err)
	}
	bad := first
	bad.Outcome = "failure"
	bad.TransientFailure = true
	bad.LeaseID = "stale"
	blocked, err := RecordTaskAttempt(store, policy, bad)
	if err != nil || blocked.Reason != "lease-mismatch" {
		t.Fatalf("stale finish: %#v %v", blocked, err)
	}
	bad.LeaseID = claim.LeaseID
	if _, err = RecordTaskAttempt(store, policy, bad); err != nil {
		t.Fatal(err)
	}
	dup, err := ClaimNativeTask(store, policy, first)
	if err != nil || dup.Reason != "duplicate-attempt" {
		t.Fatalf("duplicate: %#v %v", dup, err)
	}
	second := TaskAttemptInput{TaskID: "task", StrategyID: "b", AttemptID: "two"}
	claim, err = ClaimNativeTask(store, policy, second)
	if err != nil || !claim.Allowed {
		t.Fatalf("second: %#v %v", claim, err)
	}
	second.Outcome = "failure"
	second.TransientFailure = true
	second.LeaseID = claim.LeaseID
	if _, err = RecordTaskAttempt(store, policy, second); err != nil {
		t.Fatal(err)
	}
	exhausted, err := ClaimNativeTask(store, policy, TaskAttemptInput{TaskID: "task", StrategyID: "c", AttemptID: "three"})
	if err != nil || exhausted.Reason != "attempt-limit" {
		t.Fatalf("limit: %#v %v", exhausted, err)
	}
	weaker := policy
	weaker.MaxAttempts = 3
	if _, err = ClaimNativeTask(store, weaker, TaskAttemptInput{TaskID: "task", StrategyID: "c", AttemptID: "four"}); err == nil {
		t.Fatal("policy mutation reset persisted budget")
	}
}

func TestNativeTaskNoProgressIsTaskWideAndProgressResets(t *testing.T) {
	store := t.TempDir()
	policy := TaskCircuitPolicy{ID: "native-v1", MaxNoProgress: 2, MaxAttempts: 4, DeadlineSeconds: 60, LeaseSeconds: 30}
	claim, err := ClaimNativeTask(store, policy, TaskAttemptInput{TaskID: "task", StrategyID: "a", AttemptID: "one"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = RecordTaskAttempt(store, policy, TaskAttemptInput{TaskID: "task", StrategyID: "a", AttemptID: "one", Outcome: "no-progress", LeaseID: claim.LeaseID}); err != nil {
		t.Fatal(err)
	}
	claim, err = ClaimNativeTask(store, policy, TaskAttemptInput{TaskID: "task", StrategyID: "b", AttemptID: "two"})
	if err != nil || !claim.Allowed || claim.State.NoProgress != 1 {
		t.Fatalf("cross-strategy claim: %#v %v", claim, err)
	}
	if _, err = RecordTaskAttempt(store, policy, TaskAttemptInput{TaskID: "task", StrategyID: "b", AttemptID: "two", Outcome: "progress", LeaseID: claim.LeaseID}); err != nil {
		t.Fatal(err)
	}
	claim, err = ClaimNativeTask(store, policy, TaskAttemptInput{TaskID: "task", StrategyID: "c", AttemptID: "three"})
	if err != nil || !claim.Allowed || claim.State.NoProgress != 0 {
		t.Fatalf("progress reset: %#v %v", claim, err)
	}
	if _, err = RecordTaskAttempt(store, policy, TaskAttemptInput{TaskID: "task", StrategyID: "c", AttemptID: "three", Outcome: "no-progress", LeaseID: claim.LeaseID}); err != nil {
		t.Fatal(err)
	}
	claim, err = ClaimNativeTask(store, policy, TaskAttemptInput{TaskID: "task", StrategyID: "d", AttemptID: "four"})
	if err != nil || !claim.Allowed {
		t.Fatalf("fourth claim: %#v %v", claim, err)
	}
	recorded, err := RecordTaskAttempt(store, policy, TaskAttemptInput{TaskID: "task", StrategyID: "d", AttemptID: "four", Outcome: "no-progress", LeaseID: claim.LeaseID})
	if err != nil || !recorded.State.CircuitOpen || recorded.Reason != "outcome-recorded" {
		t.Fatalf("task-wide stop: %#v %v", recorded, err)
	}
	blocked, err := ClaimNativeTask(store, policy, TaskAttemptInput{TaskID: "task", StrategyID: "e", AttemptID: "five"})
	if err != nil || blocked.Reason != "no-progress-limit" {
		t.Fatalf("cross-strategy stop bypassed: %#v %v", blocked, err)
	}
}

func TestNativeTaskPinsPolicyAndRejectsLegacyGate(t *testing.T) {
	store := t.TempDir()
	policy := TaskCircuitPolicy{ID: "native-v1", MaxNoProgress: 2, MaxAttempts: 2, DeadlineSeconds: 60, LeaseSeconds: 30}
	if _, err := ClaimNativeTask(store, policy, TaskAttemptInput{TaskID: "task", StrategyID: "a", AttemptID: "one"}); err != nil {
		t.Fatal(err)
	}
	changedID := policy
	changedID.ID = "other"
	if _, err := ClaimNativeTask(store, changedID, TaskAttemptInput{TaskID: "task", StrategyID: "a", AttemptID: "two"}); err == nil {
		t.Fatal("policy id bypass was accepted")
	}
	legacy := TaskCircuitPolicy{ID: policy.ID, MaxNoProgress: policy.MaxNoProgress}
	if _, err := CheckTaskCircuit(store, legacy, TaskAttemptInput{TaskID: "task", StrategyID: "a", AttemptID: "two"}); err == nil {
		t.Fatal("legacy gate silently downgraded guarded task")
	}
	for _, id := range []string{policy.ID, "renamed"} {
		legacy.ID = id
		if _, err := RecordTaskAttempt(store, legacy, TaskAttemptInput{TaskID: "task", StrategyID: "a", AttemptID: "one", Outcome: "success"}); err == nil {
			t.Fatal("legacy record bypassed the pinned policy and required lease")
		}
	}
	gate, err := CheckTaskCircuit(store, policy, TaskAttemptInput{TaskID: "task", StrategyID: "a", AttemptID: "two"})
	if err != nil || gate.Reason != "native-claim-required" {
		t.Fatalf("guarded gate: %#v %v", gate, err)
	}
}

func TestNativeTaskReadsUniqueLegacyStateWithoutReset(t *testing.T) {
	store := t.TempDir()
	policy := TaskCircuitPolicy{ID: "native-v1", MaxNoProgress: 2, MaxAttempts: 2, DeadlineSeconds: 60, LeaseSeconds: 30}
	input := TaskAttemptInput{TaskID: "task", StrategyID: "a", AttemptID: "one"}
	claim, err := ClaimNativeTask(store, policy, input)
	if err != nil {
		t.Fatal(err)
	}
	current := nativeTaskPath(store, input.TaskID)
	legacy := legacyNativeTaskPath(store, input.TaskID, policy.ID)
	if err := os.Rename(current, legacy); err != nil {
		t.Fatal(err)
	}
	reloaded, err := ClaimNativeTask(store, policy, TaskAttemptInput{TaskID: "task", StrategyID: "b", AttemptID: "two"})
	if err != nil || reloaded.Reason != "active-lease" || reloaded.State.ActiveLease.ID != claim.LeaseID {
		t.Fatalf("legacy state reset: %#v %v", reloaded, err)
	}
	changed := policy
	changed.ID = "other"
	if _, err := ClaimNativeTask(store, changed, TaskAttemptInput{TaskID: "task", StrategyID: "b", AttemptID: "two"}); err == nil {
		t.Fatal("legacy policy change was accepted")
	}
}

func TestNativeTaskExpiredLeaseAndDeadlineFailClosed(t *testing.T) {
	store := t.TempDir()
	policy := TaskCircuitPolicy{ID: "lease", MaxNoProgress: 2, MaxAttempts: 2, DeadlineSeconds: 2, LeaseSeconds: 1}
	claim, err := ClaimNativeTask(store, policy, TaskAttemptInput{TaskID: "lease-task", StrategyID: "a", AttemptID: "one"})
	if err != nil || !claim.Allowed {
		t.Fatal(err)
	}
	time.Sleep(1100 * time.Millisecond)
	abandoned, err := ClaimNativeTask(store, policy, TaskAttemptInput{TaskID: "lease-task", StrategyID: "b", AttemptID: "two"})
	if err != nil || abandoned.Reason != "lease-abandoned" {
		t.Fatalf("abandoned: %#v %v", abandoned, err)
	}
	deadlinePolicy := TaskCircuitPolicy{ID: "deadline", MaxNoProgress: 2, MaxAttempts: 2, DeadlineSeconds: 1, LeaseSeconds: 1}
	deadlineClaim, err := ClaimNativeTask(store, deadlinePolicy, TaskAttemptInput{TaskID: "deadline-task", StrategyID: "a", AttemptID: "one"})
	if err != nil {
		t.Fatal(err)
	}
	finish := TaskAttemptInput{TaskID: "deadline-task", StrategyID: "a", AttemptID: "one", Outcome: "success", LeaseID: deadlineClaim.LeaseID}
	time.Sleep(1100 * time.Millisecond)
	recorded, err := RecordTaskAttempt(store, deadlinePolicy, finish)
	if err != nil || recorded.Reason != "deadline-expired" {
		t.Fatalf("deadline finish: %#v %v", recorded, err)
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
