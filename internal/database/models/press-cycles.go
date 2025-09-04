package models

import "time"

const (
	MinPressNumber = 0
	MaxPressNumber = 5
)

type PressNumber int8

// IsValid checks if the press number is within the valid range (0-5)
func IsValidPressNumber(n *PressNumber) bool {
	if n == nil {
		return false
	}

	return *n >= MinPressNumber && *n <= MaxPressNumber
}

type PressCycle struct {
	ID          int64       `json:"id"`
	PressNumber PressNumber `json:"press_number"` // PressNumber is optional
	ToolID      int64       `json:"tool_id"`
	Date        time.Time   `json:"date"`
	TotalCycles int64       `json:"total_cycles"`
	PerformedBy int64       `json:"performed_by"`
}

func NewPressCycle(toolID int64, press PressNumber, totalCycles, user int64) *PressCycle {
	return &PressCycle{
		ToolID:      toolID,
		PressNumber: press,
		Date:        time.Now(),
		TotalCycles: totalCycles,
		PerformedBy: user,
	}
}
