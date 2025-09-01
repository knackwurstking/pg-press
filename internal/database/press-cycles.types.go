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
	ID          int64       `json:"id"`
	PressNumber PressNumber `json:"press_number"`
	ToolID      int64       `json:"tool_id"`
	Date        time.Time   `json:"date"`
	TotalCycles int64       `json:"total_cycles"`
	PerformedBy int64       `json:"performed_by"`
}

// TODO: I need to make the press argument optional, because i will allow editing tools not active
func NewPressCycle(toolID int64, press PressNumber, totalCycles, user int64) *PressCycle {
	return &PressCycle{
		ToolID:      toolID,
		PressNumber: press,
		Date:        time.Now(),
		TotalCycles: totalCycles,
		PerformedBy: user,
	}
}
