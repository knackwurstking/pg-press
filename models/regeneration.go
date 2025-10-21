package models

// Regeneration represents a tool regeneration event
type Regeneration struct {
	ID          int64  `json:"id"`
	ToolID      int64  `json:"tool_id"`
	CycleID     int64  `json:"cycle_id"`
	Reason      string `json:"reason"`
	PerformedBy *int64 `json:"performed_by"`
}

func NewRegeneration(toolID int64, cycleID int64, reason string, performedBy *int64) *Regeneration {
	return &Regeneration{
		ToolID:      toolID,
		CycleID:     cycleID,
		Reason:      reason,
		PerformedBy: performedBy,
	}
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
