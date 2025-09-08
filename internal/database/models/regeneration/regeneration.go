package regeneration

// ToolRegeneration represents a tool regeneration event
type ToolRegeneration struct {
	ID          int64  `json:"id"`
	ToolID      int64  `json:"tool_id"`
	CycleID     int64  `json:"cycle_id"`
	Reason      string `json:"reason"`
	PerformedBy *int64 `json:"performed_by"`
}

func NewToolRegeneration(toolID int64, cycleID int64, reason string, performedBy *int64) *ToolRegeneration {
	return &ToolRegeneration{
		ToolID:      toolID,
		CycleID:     cycleID,
		Reason:      reason,
		PerformedBy: performedBy,
	}
}
