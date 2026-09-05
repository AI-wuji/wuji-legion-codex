package core

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const generalStaffStoreSchemaVersion = 1

type GeneralStaffState struct {
	SchemaVersion int            `json:"schema_version"`
	Current       GeneralStaff   `json:"current"`
	Replaced      []GeneralStaff `json:"replaced,omitempty"`
}

func DefaultGeneralStaffStore() string {
	if value := strings.TrimSpace(os.Getenv("WUJI_GENERAL_STAFF_DIR")); value != "" {
		return filepath.Clean(value)
	}
	return filepath.Join(".wuji", "general-staff")
}

func CreateGeneralStaffState(store string, snapshot GeneralStaffSnapshot) (GeneralStaffState, error) {
	staff, err := NewGeneralStaff(snapshot)
	if err != nil {
		return GeneralStaffState{}, err
	}
	var result GeneralStaffState
	err = withKnowledgeStoreLock(store, func() error {
		state, exists, err := loadGeneralStaffState(store)
		if err != nil {
			return err
		}
		if exists {
			if sameGeneralStaffSnapshot(state.Current, snapshot) {
				result = state
				return nil
			}
			return fmt.Errorf("general staff state already exists for task %q", state.Current.TaskInstanceID)
		}
		result = GeneralStaffState{SchemaVersion: generalStaffStoreSchemaVersion, Current: staff}
		return writeGeneralStaffState(store, result)
	})
	return result, err
}

func UpdateGeneralStaffState(store string, update GeneralStaffUpdate) (GeneralStaffState, error) {
	var result GeneralStaffState
	err := withKnowledgeStoreLock(store, func() error {
		state, exists, err := loadGeneralStaffState(store)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("general staff state does not exist")
		}
		updated, err := ApplyGeneralStaffUpdate(state.Current, update)
		if err != nil {
			return err
		}
		state.Current = updated.Current
		if updated.Replaced != nil {
			state.Replaced = append(state.Replaced, *updated.Replaced)
		}
		if len(state.Replaced) > 64 {
			state.Replaced = append([]GeneralStaff(nil), state.Replaced[len(state.Replaced)-64:]...)
		}
		result = state
		return writeGeneralStaffState(store, state)
	})
	return result, err
}

func ReadGeneralStaffState(store string) (GeneralStaffState, error) {
	state, exists, err := loadGeneralStaffState(store)
	if err != nil {
		return GeneralStaffState{}, err
	}
	if !exists {
		return GeneralStaffState{}, fmt.Errorf("general staff state does not exist")
	}
	return state, nil
}

func loadGeneralStaffState(store string) (GeneralStaffState, bool, error) {
	data, err := os.ReadFile(generalStaffStatePath(store))
	if os.IsNotExist(err) {
		return GeneralStaffState{}, false, nil
	}
	if err != nil {
		return GeneralStaffState{}, false, err
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	var state GeneralStaffState
	if err := decoder.Decode(&state); err != nil {
		return GeneralStaffState{}, false, fmt.Errorf("decode general staff state: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return GeneralStaffState{}, false, fmt.Errorf("decode general staff state: multiple JSON values are not allowed")
	}
	if err := validateGeneralStaffState(state); err != nil {
		return GeneralStaffState{}, false, err
	}
	return state, true, nil
}

func writeGeneralStaffState(store string, state GeneralStaffState) error {
	if err := validateGeneralStaffState(state); err != nil {
		return err
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	path := generalStaffStatePath(store)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return atomicWriteFile(path, append(data, '\n'), 0o600)
}

func validateGeneralStaffState(state GeneralStaffState) error {
	if state.SchemaVersion != generalStaffStoreSchemaVersion || len(state.Replaced) > 64 {
		return fmt.Errorf("general staff state metadata is invalid")
	}
	if err := state.Current.Validate(); err != nil {
		return err
	}
	for _, staff := range state.Replaced {
		if err := staff.Validate(); err != nil {
			return err
		}
		if staff.Lifecycle != GeneralStaffReplaced {
			return fmt.Errorf("archived general staff is not replaced")
		}
	}
	return nil
}

func sameGeneralStaffSnapshot(staff GeneralStaff, snapshot GeneralStaffSnapshot) bool {
	return staff.TaskInstanceID == strings.TrimSpace(snapshot.TaskInstanceID) &&
		staff.SessionKey == strings.TrimSpace(snapshot.SessionKey) &&
		staff.RequirementVersion == strings.TrimSpace(snapshot.RequirementVersion) &&
		staff.GraphVersion == strings.TrimSpace(snapshot.GraphVersion)
}

func generalStaffStatePath(store string) string {
	return filepath.Join(filepath.Clean(store), "v1", "state.json")
}
