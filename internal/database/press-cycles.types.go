package database

import "time"

const (
	MinPressNumber = 0
	MaxPressNumber = 5
)

type PressNumber int8

// IsValid checks if the press number is within the valid range (0-5)
func (pn PressNumber) IsValid() bool {
	return pn >= MinPressNumber && pn <= MaxPressNumber
}

type PressCycle struct {
	ID            int64       `json:"id"`
	PressNumber   PressNumber `json:"press_number"`
	ToolID        int64       `json:"tool_id"`
	FromDate      time.Time   `json:"from_date"`
	ToDate        *time.Time  `json:"to_date"`
	TotalCycles   int64       `json:"total_cycles"`
	PartialCycles int64       `json:"partial_cycles"`
	PerformedBy   int64       `json:"performed_by"`
}
