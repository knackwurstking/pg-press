package metalsheet

import (
	"fmt"

	"github.com/knackwurstking/pgpress/internal/models/note"
	"github.com/knackwurstking/pgpress/internal/models/tool"
	"github.com/knackwurstking/pgpress/internal/models/user"
	"github.com/knackwurstking/pgpress/internal/modification"
)

// TODO: Add a `MetalSheetList` type with sorting functionality

// MetalSheet represents a metal sheet in the database
type MetalSheet struct {
	ID          int64                            `json:"id"`
	TileHeight  float64                          `json:"tile_height"`  // Tile height in mm
	Value       float64                          `json:"value"`        // Value
	MarkeHeight int                              `json:"marke_height"` // Marke height
	STF         float64                          `json:"stf"`          // STF value
	STFMax      float64                          `json:"stf_max"`      // STF max value
	ToolID      *int64                           `json:"tool_id"`      // Currently assigned tool (nullable)
	LinkedNotes []int64                          `json:"notes"`        // Contains note ids from the "notes" table
	Mods        modification.Mods[MetalSheetMod] `json:"mods"`
}

// New creates a new MetalSheet with default values
func New(u *user.User) *MetalSheet {
	sheet := &MetalSheet{
		TileHeight:  0,
		Value:       0,
		MarkeHeight: 0,
		STF:         0,
		STFMax:      0,
		ToolID:      nil,
		LinkedNotes: make([]int64, 0),
		Mods:        modification.NewMods[MetalSheetMod](),
	}

	// Create initial mod entry
	sheet.Mods = append(sheet.Mods, modification.NewMod(u, MetalSheetMod{
		TileHeight:  sheet.TileHeight,
		Value:       sheet.Value,
		MarkeHeight: sheet.MarkeHeight,
		STF:         sheet.STF,
		STFMax:      sheet.STFMax,
		ToolID:      sheet.ToolID,
		LinkedNotes: sheet.LinkedNotes,
	}))

	return sheet
}

// String returns a string representation of the metal sheet
func (ms *MetalSheet) String() string {
	return fmt.Sprintf("Blech #%d (TH: %.1f, V: %.1f, MH: %d, STF: %.1f/%.1f)",
		ms.ID, ms.TileHeight, ms.Value, ms.MarkeHeight, ms.STF, ms.STFMax)
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
	STFMax      float64 `json:"stf_max"`
	ToolID      *int64  `json:"tool_id"`
	LinkedNotes []int64 `json:"notes"`
}

// MetalSheetWithNotes represents a metal sheet with its related notes loaded
type MetalSheetWithNotes struct {
	*MetalSheet
	LoadedNotes []*note.Note `json:"loaded_notes"`
	Tool        *tool.Tool   `json:"tool,omitempty"` // The tool currently using this sheet
}
