// TODO: Rename this from `Regeneration` to `ToolRegeneration`
package models

import "github.com/knackwurstking/pg-press/errors"

type ToolRegenerationID int64

// ToolRegeneration represents a tool regeneration event
type ToolRegeneration struct {
	ID          ToolRegenerationID `json:"id"`
	ToolID      ToolID             `json:"tool_id"`
	CycleID     CycleID            `json:"cycle_id"`
	Reason      string             `json:"reason"`
	PerformedBy *TelegramID        `json:"performed_by"`
}

func NewToolRegeneration(toolID ToolID, cycleID CycleID, reason string, performedBy *TelegramID) *ToolRegeneration {
	return &ToolRegeneration{
		ToolID:      toolID,
		CycleID:     cycleID,
		Reason:      reason,
		PerformedBy: performedBy,
	}
}

func (r *ToolRegeneration) Validate() error {
	if r.ToolID <= 0 {
		return errors.NewValidationError("tool_id")
	}

	if r.CycleID <= 0 {
		return errors.NewValidationError("cycle_id")
	}

	return nil
}
