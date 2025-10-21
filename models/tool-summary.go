package models

import "time"

// ToolSummary represents a summary of a tool's usage during a specific period
type ToolSummary struct {
	ToolID            int64     `json:"tool_id"`
	ToolCode          string    `json:"tool_code"`
	Position          Position  `json:"position"`
	StartDate         time.Time `json:"start_date"`
	EndDate           time.Time `json:"end_date"`
	MaxCycles         int64     `json:"max_cycles"`
	TotalPartial      int64     `json:"total_partial"`
	IsFirstAppearance bool      `json:"is_first_appearance"`
}
