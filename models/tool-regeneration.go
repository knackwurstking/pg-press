package models

import (
	"fmt"
)

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
		return fmt.Errorf("tool_id")
	}

	if r.CycleID <= 0 {
		return fmt.Errorf("cycle_id")
	}

	return nil
}
