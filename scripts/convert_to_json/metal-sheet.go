package main

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
