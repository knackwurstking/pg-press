package main

import "time"

// CREATE TABLE press_cycles (
// 		id INTEGER PRIMARY KEY AUTOINCREMENT,
// 		press_number INTEGER NOT NULL,
// 		tool_id INTEGER NOT NULL,
// 		tool_position TEXT NOT NULL,
// 		total_cycles INTEGER NOT NULL DEFAULT 0,
// 		date DATETIME NOT NULL,
// 		performed_by INTEGER NOT NULL
// 	);

const (
	PositionTop         = Position("top")
	PositionTopCassette = Position("cassette top")
	PositionBottom      = Position("bottom")
)

type Position string

type Cycle struct {
	ID            int64     `json:"id"`
	PressNumber   int8      `json:"press_number"`
	ToolID        int64     `json:"tool_id"`
	ToolPosition  Position  `json:"tool_position"`
	Date          time.Time `json:"date"`
	TotalCycles   int64     `json:"total_cycles"`
	PartialCycles int64     `json:"partial_cycles"`
	PerformedBy   int64     `json:"performed_by"`
}
