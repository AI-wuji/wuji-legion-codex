package core

import "testing"

func TestGeneralStaffLifecycleAndBoundaries(t *testing.T) {
	staff, err := NewGeneralStaff(GeneralStaffSnapshot{
		TaskInstanceID: "task-1", SessionKey: "session-1", RequirementVersion: "requirements@1", GraphVersion: "graph@1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if staff.Lifecycle != GeneralStaffPlanning || staff.ArtifactExecution || staff.ArtifactWrite {
		t.Fatalf("new staff violated its planning or authority boundary: %#v", staff)
	}
	for _, next := range []GeneralStaffLifecycle{GeneralStaffDispatching, GeneralStaffWaiting, GeneralStaffReconciling, GeneralStaffReviewing, GeneralStaffCompleted} {
		staff, err = staff.Advance(next)
		if err != nil {
			t.Fatalf("advance to %q: %v", next, err)
		}
	}
	if _, err := staff.Advance(GeneralStaffPlanning); err == nil {
		t.Fatal("terminal staff accepted a transition")
	}
}

func TestGeneralStaffIncrementalUpdateReusesSessionAndReplans(t *testing.T) {
	staff, err := NewGeneralStaff(GeneralStaffSnapshot{TaskInstanceID: "task-1", SessionKey: "session-1", RequirementVersion: "requirements@1", GraphVersion: "graph@1"})
	if err != nil {
		t.Fatal(err)
	}
	staff, err = staff.Advance(GeneralStaffDispatching)
	if err != nil {
		t.Fatal(err)
	}
	result, err := ApplyGeneralStaffUpdate(staff, GeneralStaffUpdate{Snapshot: GeneralStaffSnapshot{TaskInstanceID: "task-1", SessionKey: "session-1", RequirementVersion: "requirements@2", GraphVersion: "graph@2"}})
	if err != nil {
		t.Fatal(err)
	}
	if result.Replaced != nil || result.Current.SessionKey != "session-1" || result.Current.Lifecycle != GeneralStaffPlanning || result.Current.RequirementVersion != "requirements@2" || result.Current.GraphVersion != "graph@2" {
		t.Fatalf("incremental update did not preserve identity and replan: %#v", result)
	}
	_, err = ApplyGeneralStaffUpdate(result.Current, GeneralStaffUpdate{Snapshot: GeneralStaffSnapshot{TaskInstanceID: "task-1", SessionKey: "session-2", RequirementVersion: "requirements@3", GraphVersion: "graph@3"}})
	if err == nil {
		t.Fatal("incremental update accepted a new session key")
	}
}

func TestGeneralStaffVetoOrIdentityChangeReplacesInstance(t *testing.T) {
	staff, err := NewGeneralStaff(GeneralStaffSnapshot{TaskInstanceID: "task-1", SessionKey: "session-1", RequirementVersion: "requirements@1", GraphVersion: "graph@1"})
	if err != nil {
		t.Fatal(err)
	}
	for _, update := range []GeneralStaffUpdate{
		{Veto: true, Snapshot: GeneralStaffSnapshot{TaskInstanceID: "task-1", SessionKey: "session-2", RequirementVersion: "requirements@2", GraphVersion: "graph@2"}},
		{Snapshot: GeneralStaffSnapshot{TaskInstanceID: "task-2", SessionKey: "session-3", RequirementVersion: "requirements@1", GraphVersion: "graph@1"}},
	} {
		result, err := ApplyGeneralStaffUpdate(staff, update)
		if err != nil {
			t.Fatal(err)
		}
		if result.Replaced == nil || result.Replaced.Lifecycle != GeneralStaffReplaced || result.Current.Lifecycle != GeneralStaffPlanning || result.Current.SessionKey == result.Replaced.SessionKey {
			t.Fatalf("replacement was not explicit and fresh: %#v", result)
		}
		staff = result.Current
	}
}

func TestGeneralStaffRejectsArtifactAuthority(t *testing.T) {
	staff, err := NewGeneralStaff(GeneralStaffSnapshot{TaskInstanceID: "task-1", SessionKey: "session-1", RequirementVersion: "requirements@1", GraphVersion: "graph@1"})
	if err != nil {
		t.Fatal(err)
	}
	staff.ArtifactWrite = true
	if err := staff.Validate(); err == nil {
		t.Fatal("artifact write authority was accepted")
	}
}
