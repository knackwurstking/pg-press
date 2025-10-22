package models

import (
	"time"
)

type Cycle struct {
	ID            int64       `json:"id"`
	PressNumber   PressNumber `json:"press_number"`
	ToolID        int64       `json:"tool_id"`
	ToolPosition  Position    `json:"tool_position"`
	Date          time.Time   `json:"date"`
	TotalCycles   int64       `json:"total_cycles"`
	PartialCycles int64       `json:"partial_cycles"`
	PerformedBy   int64       `json:"performed_by"`
}

func NewCycle(press PressNumber, toolID int64, toolPosition Position, totalCycles, userID int64) *Cycle {
	return &Cycle{
		PressNumber:  press,
		ToolID:       toolID,
		ToolPosition: toolPosition,
		Date:         time.Now(),
		TotalCycles:  totalCycles,
		PerformedBy:  userID,
	}
}

func NewCycleWithID(id int64, press PressNumber, toolID int64, toolPosition Position, totalCycles, userID int64, date time.Time) *Cycle {
	return &Cycle{
		ID:           id,
		PressNumber:  press,
		ToolID:       toolID,
		ToolPosition: toolPosition,
		Date:         date,
		TotalCycles:  totalCycles,
		PerformedBy:  userID,
	}
}

func FilterCyclesByToolPosition(toolPosition Position, cycles ...*Cycle) []*Cycle {
	filteredCycles := make([]*Cycle, 0, len(cycles))
	for _, cycle := range cycles {
		if cycle.ToolPosition == toolPosition {
			filteredCycles = append(filteredCycles, cycle)
		}
	}
	return filteredCycles
}

func (c *Cycle) Validate() error {
	// TODO: Validate press number, position, TotalCycles and ToolID
	return nil
}
