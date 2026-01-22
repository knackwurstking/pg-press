package models

import "github.com/knackwurstking/pg-press/errors"

type RegenerationID int64

// Regeneration represents a tool regeneration event
type Regeneration struct {
	ID          RegenerationID `json:"id"`
	ToolID      ToolID         `json:"tool_id"`
	CycleID     CycleID        `json:"cycle_id"`
	Reason      string         `json:"reason"`
	PerformedBy *TelegramID    `json:"performed_by"`
}

func NewRegeneration(toolID ToolID, cycleID CycleID, reason string, performedBy *TelegramID) *Regeneration {
	return &Regeneration{
		ToolID:      toolID,
		CycleID:     cycleID,
		Reason:      reason,
		PerformedBy: performedBy,
	}
}

func (r *Regeneration) Validate() error {
	if r.ToolID <= 0 {
		return errors.NewValidationError("tool_id")
	}

	if r.CycleID <= 0 {
		return errors.NewValidationError("cycle_id")
	}

	return nil
}
