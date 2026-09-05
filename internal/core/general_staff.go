package core

import (
	"fmt"
	"strings"
)

// GeneralStaffLifecycle is the lifecycle of one task-level staff instance.
type GeneralStaffLifecycle string

const (
	GeneralStaffPlanning    GeneralStaffLifecycle = "planning"
	GeneralStaffDispatching GeneralStaffLifecycle = "dispatching"
	GeneralStaffWaiting     GeneralStaffLifecycle = "waiting"
	GeneralStaffReconciling GeneralStaffLifecycle = "reconciling"
	GeneralStaffReviewing   GeneralStaffLifecycle = "reviewing"
	GeneralStaffCompleted   GeneralStaffLifecycle = "completed"
	GeneralStaffReplaced    GeneralStaffLifecycle = "replaced"
	GeneralStaffCancelled   GeneralStaffLifecycle = "cancelled"
)

// GeneralStaffSnapshot identifies the task material currently under review.
// Versions are opaque, caller-owned identifiers.
type GeneralStaffSnapshot struct {
	TaskInstanceID     string
	SessionKey         string
	RequirementVersion string
	GraphVersion       string
}

// GeneralStaff is a persistent task-level planning state. It is intentionally
// incapable of executing or writing artifacts; those operations belong to the
// assigned execution nodes.
type GeneralStaff struct {
	TaskInstanceID     string
	SessionKey         string
	RequirementVersion string
	GraphVersion       string
	Lifecycle          GeneralStaffLifecycle
	ArtifactExecution  bool
	ArtifactWrite      bool
}

// GeneralStaffUpdate applies either an incremental task update or a replacement.
// Veto discards the current plan even when the task identity remains unchanged.
type GeneralStaffUpdate struct {
	Snapshot GeneralStaffSnapshot
	Veto     bool
}

// GeneralStaffUpdateResult preserves the prior state when replacement occurs.
type GeneralStaffUpdateResult struct {
	Current  GeneralStaff
	Replaced *GeneralStaff
}

// NewGeneralStaff starts a staff instance in planning.
func NewGeneralStaff(snapshot GeneralStaffSnapshot) (GeneralStaff, error) {
	if err := validateGeneralStaffSnapshot(snapshot); err != nil {
		return GeneralStaff{}, err
	}
	return GeneralStaff{
		TaskInstanceID:     snapshot.TaskInstanceID,
		SessionKey:         snapshot.SessionKey,
		RequirementVersion: snapshot.RequirementVersion,
		GraphVersion:       snapshot.GraphVersion,
		Lifecycle:          GeneralStaffPlanning,
		ArtifactExecution:  false,
		ArtifactWrite:      false,
	}, nil
}

// Advance moves a live instance through its ordered lifecycle.
func (staff GeneralStaff) Advance(next GeneralStaffLifecycle) (GeneralStaff, error) {
	if err := staff.Validate(); err != nil {
		return GeneralStaff{}, err
	}
	if !canAdvanceGeneralStaff(staff.Lifecycle, next) {
		return GeneralStaff{}, fmt.Errorf("cannot advance general staff from %q to %q", staff.Lifecycle, next)
	}
	staff.Lifecycle = next
	return staff, nil
}

// ApplyGeneralStaffUpdate keeps the session key for an incremental update.
// A veto or changed task identity marks the previous instance as replaced and
// creates a fresh planning instance using the supplied replacement session key.
func ApplyGeneralStaffUpdate(staff GeneralStaff, update GeneralStaffUpdate) (GeneralStaffUpdateResult, error) {
	if err := staff.Validate(); err != nil {
		return GeneralStaffUpdateResult{}, err
	}
	if err := validateGeneralStaffSnapshot(update.Snapshot); err != nil {
		return GeneralStaffUpdateResult{}, err
	}

	replace := update.Veto || staff.TaskInstanceID != update.Snapshot.TaskInstanceID
	if !replace {
		if staff.SessionKey != update.Snapshot.SessionKey {
			return GeneralStaffUpdateResult{}, fmt.Errorf("incremental update must reuse the existing session key")
		}
		if isGeneralStaffTerminal(staff.Lifecycle) {
			return GeneralStaffUpdateResult{}, fmt.Errorf("cannot incrementally update terminal general staff")
		}
		staff.RequirementVersion = update.Snapshot.RequirementVersion
		staff.GraphVersion = update.Snapshot.GraphVersion
		staff.Lifecycle = GeneralStaffPlanning
		return GeneralStaffUpdateResult{Current: staff}, nil
	}

	if staff.SessionKey == update.Snapshot.SessionKey {
		return GeneralStaffUpdateResult{}, fmt.Errorf("replacement general staff requires a new session key")
	}
	previous := staff
	previous.Lifecycle = GeneralStaffReplaced
	current, err := NewGeneralStaff(update.Snapshot)
	if err != nil {
		return GeneralStaffUpdateResult{}, err
	}
	return GeneralStaffUpdateResult{Current: current, Replaced: &previous}, nil
}

// Validate confirms the staff is a valid non-executing planning record.
func (staff GeneralStaff) Validate() error {
	if err := validateGeneralStaffSnapshot(GeneralStaffSnapshot{
		TaskInstanceID:     staff.TaskInstanceID,
		SessionKey:         staff.SessionKey,
		RequirementVersion: staff.RequirementVersion,
		GraphVersion:       staff.GraphVersion,
	}); err != nil {
		return err
	}
	if !validGeneralStaffLifecycle(staff.Lifecycle) {
		return fmt.Errorf("invalid general staff lifecycle %q", staff.Lifecycle)
	}
	if staff.ArtifactExecution || staff.ArtifactWrite {
		return fmt.Errorf("general staff cannot execute or write artifacts")
	}
	return nil
}

func validateGeneralStaffSnapshot(snapshot GeneralStaffSnapshot) error {
	for name, value := range map[string]string{
		"task instance ID":    snapshot.TaskInstanceID,
		"session key":         snapshot.SessionKey,
		"requirement version": snapshot.RequirementVersion,
		"graph version":       snapshot.GraphVersion,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("general staff %s is required", name)
		}
	}
	return nil
}

func validGeneralStaffLifecycle(lifecycle GeneralStaffLifecycle) bool {
	switch lifecycle {
	case GeneralStaffPlanning, GeneralStaffDispatching, GeneralStaffWaiting, GeneralStaffReconciling, GeneralStaffReviewing, GeneralStaffCompleted, GeneralStaffReplaced, GeneralStaffCancelled:
		return true
	default:
		return false
	}
}

func isGeneralStaffTerminal(lifecycle GeneralStaffLifecycle) bool {
	return lifecycle == GeneralStaffCompleted || lifecycle == GeneralStaffReplaced || lifecycle == GeneralStaffCancelled
}

func canAdvanceGeneralStaff(current, next GeneralStaffLifecycle) bool {
	if isGeneralStaffTerminal(current) {
		return false
	}
	if next == GeneralStaffCancelled {
		return true
	}
	switch current {
	case GeneralStaffPlanning:
		return next == GeneralStaffDispatching
	case GeneralStaffDispatching:
		return next == GeneralStaffWaiting
	case GeneralStaffWaiting:
		return next == GeneralStaffReconciling
	case GeneralStaffReconciling:
		return next == GeneralStaffReviewing
	case GeneralStaffReviewing:
		return next == GeneralStaffCompleted
	default:
		return false
	}
}
