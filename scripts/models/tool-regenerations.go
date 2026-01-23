package models

// CREATE TABLE IF NOT EXISTS tool_regenerations (
// 		id INTEGER PRIMARY KEY AUTOINCREMENT,
// 		tool_id INTEGER NOT NULL,
// 		cycle_id INTEGER NOT NULL,
// 		reason TEXT,
// 		performed_by INTEGER NOT NULL
// 	);

type ToolRegeneration struct {
	ID          int64  `json:"id"`
	ToolID      int64  `json:"tool_id"`
	CycleID     int64  `json:"cycle_id"`
	Reason      string `json:"reason"`
	PerformedBy *int64 `json:"performed_by"`
}
