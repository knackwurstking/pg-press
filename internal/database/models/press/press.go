package press

import (
	"slices"
	"time"
)

type PressNumber int8

// IsValid checks if the press number is within the valid range (0-5)
func IsValidPressNumber(n *PressNumber) bool {
	if n == nil {
		return false
	}

	return slices.Contains([]PressNumber{0, 2, 3, 4, 5}, *n)
}

type Cycle struct {
	ID            int64       `json:"id"`
	PressNumber   PressNumber `json:"press_number"`
	ToolID        int64       `json:"tool_id"`
	ToolPosition  string      `json:"tool_position"`
	Date          time.Time   `json:"date"`
	TotalCycles   int64       `json:"total_cycles"`
	PartialCycles int64       `json:"partial_cycles"`
	PerformedBy   int64       `json:"performed_by"`
}

func NewCycle(press PressNumber, toolID int64, toolPosition string, totalCycles, user int64) *Cycle {
	return &Cycle{
		PressNumber:  press,
		ToolID:       toolID,
		ToolPosition: toolPosition,
		Date:         time.Now(),
		TotalCycles:  totalCycles,
		PerformedBy:  user,
	}
}

func NewPressCycleWithID(id int64, press PressNumber, toolID int64, toolPosition string, totalCycles, user int64, date time.Time) *Cycle {
	return &Cycle{
		ID:           id,
		PressNumber:  press,
		ToolID:       toolID,
		ToolPosition: toolPosition,
		Date:         date,
		TotalCycles:  totalCycles,
		PerformedBy:  user,
	}
}

func FilterByTool(toolID int64, cycles ...*Cycle) []*Cycle {
	var filteredCycles []*Cycle

	for _, cycle := range cycles {
		if cycle.ToolID == toolID {
			filteredCycles = append(filteredCycles, cycle)
		}
	}

	return filteredCycles
}

func FilterByToolPosition(toolPosition string, cycles ...*Cycle) []*Cycle {
	var filteredCycles []*Cycle

	for _, cycle := range cycles {
		if cycle.ToolPosition == toolPosition {
			filteredCycles = append(filteredCycles, cycle)
		}
	}

	return filteredCycles
}
