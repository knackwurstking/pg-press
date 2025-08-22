package database

import "time"

// ToolRegeneration represents a tool regeneration event
type ToolRegeneration struct {
	ID                   int64     `json:"id"`
	ToolID               int64     `json:"tool_id"`
	RegeneratedAt        time.Time `json:"regenerated_at"`
	CyclesAtRegeneration int64     `json:"cycles_at_regeneration"`
	Reason               string    `json:"reason"`
	PerformedBy          string    `json:"performed_by"`
	Notes                string    `json:"notes"`
	CreatedAt            time.Time `json:"created_at"`
}
