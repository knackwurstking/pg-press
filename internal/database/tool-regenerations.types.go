package database

import "time"

// ToolRegeneration represents a tool regeneration event
type ToolRegeneration struct {
	ID                   int64     `json:"id"`
	ToolID               int64     `json:"tool_id"`
	RegeneratedAt        time.Time `json:"regenerated_at"`
	CyclesAtRegeneration int64     `json:"cycles_at_regeneration"`
	Reason               string    `json:"reason"`
	PerformedBy          *int64    `json:"performed_by"`
}

func NewToolRegeneration(toolID int64, regeneratedAt time.Time, cyclesAtRegeneration int64, reason string, performedBy *int64) *ToolRegeneration {
	return &ToolRegeneration{
		ToolID:               toolID,
		RegeneratedAt:        regeneratedAt,
		CyclesAtRegeneration: cyclesAtRegeneration,
		Reason:               reason,
		PerformedBy:          performedBy,
	}
}
