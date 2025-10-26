package models

import "github.com/knackwurstking/pgpress/errors"

type RegenerationID int64

// Regeneration represents a tool regeneration event
type Regeneration struct {
	ID          RegenerationID `json:"id"`
	ToolID      ToolID         `json:"tool_id"`
	CycleID     CycleID        `json:"cycle_id"`
	Reason      string         `json:"reason"`
	PerformedBy *int64         `json:"performed_by"`
}

func NewRegeneration(toolID ToolID, cycleID CycleID, reason string, performedBy *int64) *Regeneration {
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

type ResolvedRegeneration struct {
	*Regeneration
	tool  *Tool
	cycle *Cycle
	user  *User
}

func NewResolvedRegeneration(r *Regeneration, tool *Tool, cycle *Cycle, user *User) *ResolvedRegeneration {
	return &ResolvedRegeneration{
		Regeneration: r,
		tool:         tool,
		cycle:        cycle,
		user:         user,
	}
}

func (rr *ResolvedRegeneration) GetTool() *Tool {
	return rr.tool
}

func (rr *ResolvedRegeneration) GetCycle() *Cycle {
	return rr.cycle
}

func (rr *ResolvedRegeneration) GetUser() *User {
	return rr.user
}
