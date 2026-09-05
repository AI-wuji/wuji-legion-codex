package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const taskCircuitSchemaVersion = 1

var taskCircuitAttemptPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{0,127}$`)
var taskCircuitHashPattern = regexp.MustCompile(`^[a-fA-F0-9]{64}$`)

type TaskCircuitPolicy struct {
	ID            string `json:"id"`
	MaxNoProgress int    `json:"max_no_progress"`
}

type TaskAttemptInput struct {
	TaskID           string `json:"task_id"`
	StrategyID       string `json:"strategy_id"`
	AttemptID        string `json:"attempt_id"`
	Outcome          string `json:"outcome,omitempty"`
	TransientFailure bool   `json:"transient_failure,omitempty"`
	EvidenceSHA256   string `json:"evidence_sha256,omitempty"`
}

type TaskAttempt struct {
	ID               string `json:"id"`
	Outcome          string `json:"outcome"`
	TransientFailure bool   `json:"transient_failure,omitempty"`
	EvidenceSHA256   string `json:"evidence_sha256,omitempty"`
	RecordedAt       string `json:"recorded_at"`
}

type TaskCircuitState struct {
	SchemaVersion  int               `json:"schema_version"`
	TaskID         string            `json:"task_id"`
	StrategyID     string            `json:"strategy_id"`
	Policy         TaskCircuitPolicy `json:"policy"`
	NoProgress     int               `json:"no_progress"`
	CircuitOpen    bool              `json:"circuit_open"`
	CircuitReason  string            `json:"circuit_reason,omitempty"`
	LastProgressAt string            `json:"last_progress_at,omitempty"`
	Attempts       []TaskAttempt     `json:"attempts,omitempty"`
}

type TaskCircuitResult struct {
	Allowed  bool             `json:"allowed"`
	Decision string           `json:"decision"`
	Reason   string           `json:"reason"`
	State    TaskCircuitState `json:"state"`
}

func DefaultTaskCircuitStore() string {
	if value := strings.TrimSpace(os.Getenv("WUJI_TASK_CIRCUIT_DIR")); value != "" {
		return filepath.Clean(value)
	}
	return filepath.Join(".wuji", "task-circuits")
}

func CheckTaskCircuit(store string, policy TaskCircuitPolicy, input TaskAttemptInput) (TaskCircuitResult, error) {
	if err := validateTaskCircuitPolicy(policy); err != nil {
		return TaskCircuitResult{}, err
	}
	if err := validateTaskAttempt(input, false); err != nil {
		return TaskCircuitResult{}, err
	}
	state, found, err := loadTaskCircuitState(store, policy, input)
	if err != nil {
		return TaskCircuitResult{}, err
	}
	if !found {
		state = newTaskCircuitState(policy, input)
	}
	if state.CircuitOpen {
		return TaskCircuitResult{Decision: "blocked", Reason: state.CircuitReason, State: state}, nil
	}
	for _, attempt := range state.Attempts {
		if attempt.ID == input.AttemptID && !attempt.TransientFailure && (attempt.Outcome == "no-progress" || attempt.Outcome == "failure") {
			return TaskCircuitResult{Decision: "blocked", Reason: "duplicate-no-progress-attempt", State: state}, nil
		}
	}
	return TaskCircuitResult{Allowed: true, Decision: "allowed", Reason: "new-attempt", State: state}, nil
}

func RecordTaskAttempt(store string, policy TaskCircuitPolicy, input TaskAttemptInput) (TaskCircuitResult, error) {
	if err := validateTaskCircuitPolicy(policy); err != nil {
		return TaskCircuitResult{}, err
	}
	if err := validateTaskAttempt(input, true); err != nil {
		return TaskCircuitResult{}, err
	}
	var result TaskCircuitResult
	err := withKnowledgeStoreLock(store, func() error {
		state, found, err := loadTaskCircuitState(store, policy, input)
		if err != nil {
			return err
		}
		if !found {
			state = newTaskCircuitState(policy, input)
		}
		if state.CircuitOpen {
			result = TaskCircuitResult{Decision: "blocked", Reason: state.CircuitReason, State: state}
			return nil
		}
		for _, attempt := range state.Attempts {
			if attempt.ID == input.AttemptID && !attempt.TransientFailure && (attempt.Outcome == "no-progress" || attempt.Outcome == "failure") {
				result = TaskCircuitResult{Decision: "blocked", Reason: "duplicate-no-progress-attempt", State: state}
				return nil
			}
		}

		state.Attempts = append(state.Attempts, TaskAttempt{
			ID: input.AttemptID, Outcome: input.Outcome, TransientFailure: input.TransientFailure,
			EvidenceSHA256: strings.ToLower(input.EvidenceSHA256), RecordedAt: time.Now().UTC().Format(time.RFC3339),
		})
		if len(state.Attempts) > 64 {
			state.Attempts = state.Attempts[len(state.Attempts)-64:]
		}
		switch input.Outcome {
		case "progress", "success":
			state.NoProgress = 0
			state.LastProgressAt = time.Now().UTC().Format(time.RFC3339)
		case "no-progress":
			state.NoProgress++
		case "failure":
			if !input.TransientFailure {
				state.NoProgress++
			}
		}
		if state.NoProgress >= policy.MaxNoProgress {
			state.CircuitOpen = true
			state.CircuitReason = "no-progress-limit"
		}
		if err := writeTaskCircuitState(store, policy, input, state); err != nil {
			return err
		}
		result = TaskCircuitResult{Allowed: true, Decision: "recorded", Reason: "outcome-recorded", State: state}
		return AuditEventRecord(auditStoreFor(store), AuditEvent{EventType: "task-attempt-recorded", Actor: "aji", Authority: "aji-merge", Target: input.TaskID + ":" + input.StrategyID, InputRevision: input.AttemptID, ResultHandle: "wuji-task://" + input.TaskID})
	})
	if err != nil {
		return TaskCircuitResult{}, err
	}
	return result, nil
}

func validateTaskCircuitPolicy(policy TaskCircuitPolicy) error {
	if !componentIDPattern.MatchString(strings.TrimSpace(policy.ID)) {
		return fmt.Errorf("policy id is invalid")
	}
	if policy.MaxNoProgress < 1 || policy.MaxNoProgress > 64 {
		return fmt.Errorf("max_no_progress must be between 1 and 64")
	}
	return nil
}

func validateTaskAttempt(input TaskAttemptInput, requireOutcome bool) error {
	for _, value := range []struct {
		name  string
		value string
	}{{"task id", input.TaskID}, {"strategy id", input.StrategyID}} {
		if !componentIDPattern.MatchString(strings.TrimSpace(value.value)) {
			return fmt.Errorf("%s is invalid", value.name)
		}
	}
	if !taskCircuitAttemptPattern.MatchString(strings.TrimSpace(input.AttemptID)) {
		return fmt.Errorf("attempt id is invalid")
	}
	if input.EvidenceSHA256 != "" && !taskCircuitHashPattern.MatchString(input.EvidenceSHA256) {
		return fmt.Errorf("evidence sha256 is invalid")
	}
	if requireOutcome {
		switch input.Outcome {
		case "progress", "success", "no-progress", "failure":
		default:
			return fmt.Errorf("outcome must be progress, success, no-progress, or failure")
		}
	}
	return nil
}

func newTaskCircuitState(policy TaskCircuitPolicy, input TaskAttemptInput) TaskCircuitState {
	return TaskCircuitState{SchemaVersion: taskCircuitSchemaVersion, TaskID: input.TaskID, StrategyID: input.StrategyID, Policy: policy}
}

func loadTaskCircuitState(store string, policy TaskCircuitPolicy, input TaskAttemptInput) (TaskCircuitState, bool, error) {
	path := taskCircuitPath(store, policy, input)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return TaskCircuitState{}, false, nil
	}
	if err != nil {
		return TaskCircuitState{}, false, err
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	var state TaskCircuitState
	if err := decoder.Decode(&state); err != nil {
		return TaskCircuitState{}, false, fmt.Errorf("decode task circuit state: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return TaskCircuitState{}, false, fmt.Errorf("decode task circuit state: multiple JSON values are not allowed")
		}
		return TaskCircuitState{}, false, err
	}
	if state.SchemaVersion != taskCircuitSchemaVersion || state.TaskID != input.TaskID || state.StrategyID != input.StrategyID || state.Policy != policy {
		return TaskCircuitState{}, false, fmt.Errorf("task circuit state does not match the requested task, strategy, and policy")
	}
	return state, true, nil
}

func writeTaskCircuitState(store string, policy TaskCircuitPolicy, input TaskAttemptInput, state TaskCircuitState) error {
	path := taskCircuitPath(store, policy, input)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	temporary, err := os.CreateTemp(filepath.Dir(path), ".task-circuit-*.json")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return err
	}
	if _, err := temporary.Write(data); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporaryPath, path)
}

func taskCircuitPath(store string, policy TaskCircuitPolicy, input TaskAttemptInput) string {
	key := input.TaskID + "\x00" + input.StrategyID + "\x00" + policy.ID
	digest := sha256.Sum256([]byte(key))
	return filepath.Join(filepath.Clean(store), "v1", "circuits", hex.EncodeToString(digest[:])+".json")
}
