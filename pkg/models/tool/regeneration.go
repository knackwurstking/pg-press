package tool

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
