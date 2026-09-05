package core

import "testing"

func TestGeneralStaffStatePersistsIncrementalUpdate(t *testing.T) {
	store := t.TempDir()
	initial := GeneralStaffSnapshot{TaskInstanceID: "task-1", SessionKey: "session-1", RequirementVersion: "requirements@1", GraphVersion: "graph@1"}
	created, err := CreateGeneralStaffState(store, initial)
	if err != nil {
		t.Fatal(err)
	}
	if created.Current.SessionKey != "session-1" {
		t.Fatalf("unexpected created state: %#v", created)
	}

	updated, err := UpdateGeneralStaffState(store, GeneralStaffUpdate{Snapshot: GeneralStaffSnapshot{
		TaskInstanceID: "task-1", SessionKey: "session-1", RequirementVersion: "requirements@2", GraphVersion: "graph@2",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Current.SessionKey != "session-1" || updated.Current.RequirementVersion != "requirements@2" || len(updated.Replaced) != 0 {
		t.Fatalf("incremental update replaced the staff instance: %#v", updated)
	}
	loaded, err := ReadGeneralStaffState(store)
	if err != nil || loaded.Current.GraphVersion != "graph@2" {
		t.Fatalf("persisted state mismatch: %#v err=%v", loaded, err)
	}
}

func TestGeneralStaffStateArchivesOnlyOnVetoOrIdentityChange(t *testing.T) {
	store := t.TempDir()
	_, err := CreateGeneralStaffState(store, GeneralStaffSnapshot{TaskInstanceID: "task-1", SessionKey: "session-1", RequirementVersion: "requirements@1", GraphVersion: "graph@1"})
	if err != nil {
		t.Fatal(err)
	}
	updated, err := UpdateGeneralStaffState(store, GeneralStaffUpdate{Veto: true, Snapshot: GeneralStaffSnapshot{
		TaskInstanceID: "task-1", SessionKey: "session-2", RequirementVersion: "requirements@2", GraphVersion: "graph@2",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Current.SessionKey != "session-2" || len(updated.Replaced) != 1 || updated.Replaced[0].SessionKey != "session-1" {
		t.Fatalf("veto did not replace and archive staff: %#v", updated)
	}
}

func TestGeneralStaffStateRejectsSessionChangeForIncrementalUpdate(t *testing.T) {
	store := t.TempDir()
	_, err := CreateGeneralStaffState(store, GeneralStaffSnapshot{TaskInstanceID: "task-1", SessionKey: "session-1", RequirementVersion: "requirements@1", GraphVersion: "graph@1"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = UpdateGeneralStaffState(store, GeneralStaffUpdate{Snapshot: GeneralStaffSnapshot{
		TaskInstanceID: "task-1", SessionKey: "session-2", RequirementVersion: "requirements@2", GraphVersion: "graph@2",
	}})
	if err == nil {
		t.Fatal("incremental update accepted a new session key")
	}
}
