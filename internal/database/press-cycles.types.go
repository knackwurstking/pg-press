package database

import "time"

const (
	MinPressNumber = 0
	MaxPressNumber = 5
)

type PressNumber int8

type PressCycle struct {
	ID            int64                      `json:"id"`
	PressNumber   PressNumber                `json:"press_number"`
	ToolID        int64                      `json:"tool_id"`
	FromDate      time.Time                  `json:"from_date"`
	ToDate        *time.Time                 `json:"to_date"`
	TotalCycles   int64                      `json:"total_cycles"`
	PartialCycles int64                      `json:"partial_cycles"`
	Mods          []*Modified[PressCycleMod] `json:"mods"`
}

// PressCycleMod represents the modifiable fields of a PressCycle for history tracking
type PressCycleMod struct {
	PressNumber   PressNumber `json:"press_number"`
	ToolID        int64       `json:"tool_id"`
	FromDate      time.Time   `json:"from_date"`
	ToDate        *time.Time  `json:"to_date"`
	TotalCycles   int64       `json:"total_cycles"`
	PartialCycles int64       `json:"partial_cycles"`
}
