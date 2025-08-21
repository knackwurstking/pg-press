package database

import "fmt"

// MetalSheetStatus represents the current status of a metal sheet
type MetalSheetStatus string

const (
	MetalSheetStatusAvailable   MetalSheetStatus = "available"
	MetalSheetStatusInUse       MetalSheetStatus = "in_use"
	MetalSheetStatusMaintenance MetalSheetStatus = "maintenance"
	MetalSheetStatusReserved    MetalSheetStatus = "reserved"
	MetalSheetStatusDamaged     MetalSheetStatus = "damaged"
)

// MetalSheet represents a metal sheet in the database
type MetalSheet struct {
	ID          int64               `json:"id"`
	TileHeight  float64             `json:"tile_height"`  // Tile height in mm
	Value       float64             `json:"value"`        // Value
	MarkeHeight int                 `json:"marke_height"` // Marke height
	STF         float64             `json:"stf"`          // STF value
	ToolID      *int64              `json:"tool_id"`      // Currently assigned tool (nullable)
	LinkedNotes []int64             `json:"notes"`        // Contains note ids from the "notes" table
	Mods        Mods[MetalSheetMod] `json:"mods"`
}

// NewMetalSheet creates a new MetalSheet with default values
func NewMetalSheet(m ...*Modified[MetalSheetMod]) *MetalSheet {
	return &MetalSheet{
		LinkedNotes: make([]int64, 0),
		Mods:        m,
	}
}

// String returns a string representation of the metal sheet
func (ms *MetalSheet) String() string {
	return fmt.Sprintf("Blech #%d (TH: %.1f, V: %.1f, MH: %d, STF: %.1f)",
		ms.ID, ms.TileHeight, ms.Value, ms.MarkeHeight, ms.STF)
}

// IsAssigned checks if the metal sheet is assigned to a tool
func (ms *MetalSheet) IsAssigned() bool {
	return ms.ToolID != nil
}

// MetalSheetMod represents modifications to a metal sheet
type MetalSheetMod struct {
	TileHeight  float64 `json:"tile_height"`
	Value       float64 `json:"value"`
	MarkeHeight int     `json:"marke_height"`
	STF         float64 `json:"stf"`
	ToolID      *int64  `json:"tool_id"`
	LinkedNotes []int64 `json:"notes"`
}

// MetalSheetWithNotes represents a metal sheet with its related notes loaded
type MetalSheetWithNotes struct {
	*MetalSheet
	LoadedNotes []*Note `json:"loaded_notes"`
	Tool        *Tool   `json:"tool,omitempty"` // The tool currently using this sheet
}
