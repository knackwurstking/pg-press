package models

import (
	"fmt"
	"time"
)

type CycleID int64

type Cycle struct {
	ID            CycleID     `json:"id"`
	PressNumber   PressNumber `json:"press_number"`
	ToolID        ToolID      `json:"tool_id"`
	ToolPosition  Position    `json:"tool_position"`
	Date          time.Time   `json:"date"`
	TotalCycles   int64       `json:"total_cycles"`
	PartialCycles int64       `json:"partial_cycles"`
	PerformedBy   TelegramID  `json:"performed_by"`
}

func NewCycle(press PressNumber, toolID ToolID, toolPosition Position, totalCycles int64, userID TelegramID) *Cycle {
	return &Cycle{
		PressNumber:  press,
		ToolID:       toolID,
		ToolPosition: toolPosition,
		Date:         time.Now(),
		TotalCycles:  totalCycles,
		PerformedBy:  userID,
	}
}

func NewCycleWithID(id CycleID, press PressNumber, toolID ToolID, toolPosition Position, totalCycles int64, userID TelegramID, date time.Time) *Cycle {
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

func (c *Cycle) Validate() error {
	if !IsValidPressNumber(&c.PressNumber) {
		return fmt.Errorf("invalid press number")
	}
	if c.ToolID <= 0 {
		return fmt.Errorf("invalid tool ID: %d", c.ToolID)
	}
	if !IsValidPosition(&c.ToolPosition) {
		return fmt.Errorf("invalid tool position")
	}
	if c.TotalCycles < 0 {
		return fmt.Errorf("total cycles must be non-negative")
	}
	return nil
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
