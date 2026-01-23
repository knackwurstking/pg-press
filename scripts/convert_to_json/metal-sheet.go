package main

// CREATE TABLE IF NOT EXISTS metal_sheets (
//		id INTEGER NOT NULL,
//		tile_height REAL NOT NULL,
//		value REAL NOT NULL,
//		marke_height INTEGER NOT NULL,
//		stf REAL NOT NULL,
//		stf_max REAL NOT NULL,
//		identifier TEXT NOT NULL,
//		tool_id INTEGER NOT NULL,
//		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
//		PRIMARY KEY("id" AUTOINCREMENT),
//		FOREIGN KEY("tool_id") REFERENCES "tools"("id") ON DELETE CASCADE
//	);

type MetalSheet struct {
	ID          int64   `json:"id"`
	TileHeight  float64 `json:"tile_height"`
	Value       float64 `json:"value"`
	MarkeHeight int     `json:"marke_height"`
	STF         float64 `json:"stf"`
	STFMax      float64 `json:"stf_max"`
	Identifier  string  `json:"identifier"`
	ToolID      int64   `json:"tool_id"`
}
