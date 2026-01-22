package models

import "time"

// OverlappingToolInstance represents one instance of a tool on a specific press
type OverlappingToolInstance struct {
	PressNumber PressNumber `json:"press_number"`
	Position    Position    `json:"position"`
	StartDate   time.Time   `json:"start_date"`
	EndDate     time.Time   `json:"end_date"`
}

// OverlappingTool represents a tool that appears on multiple presses simultaneously
type OverlappingTool struct {
	ToolID    ToolID                     `json:"tool_id"`
	ToolCode  string                     `json:"tool_code"`
	Overlaps  []*OverlappingToolInstance `json:"overlaps"`
	StartDate time.Time                  `json:"start_date"`
	EndDate   time.Time                  `json:"end_date"`
}
