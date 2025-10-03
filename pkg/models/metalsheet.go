// TODO: Need to add a identifier field for the metal sheet to detect if stf_max is from type "SACMI" or "SITI"
package models

import (
	"fmt"
	"sort"
)

type MetalSheetList []*MetalSheet

func (msl *MetalSheetList) Sort() MetalSheetList {
	// Create a copy of the slice
	result := make(MetalSheetList, len(*msl))
	copy(result, *msl)

	sort.Slice(result, func(i, j int) bool {
		// First sort by TileHeight (low to high)
		if result[i].TileHeight != result[j].TileHeight {
			return result[i].TileHeight < result[j].TileHeight
		}
		// If TileHeight is equal, sort by Value (low to high)
		return result[i].Value < result[j].Value
	})
	return result
}

// MetalSheet represents a metal sheet in the database
type MetalSheet struct {
	ID          int64   `json:"id"`
	TileHeight  float64 `json:"tile_height"`  // Tile height in mm
	Value       float64 `json:"value"`        // Value
	MarkeHeight int     `json:"marke_height"` // Marke height
	STF         float64 `json:"stf"`          // STF value
	STFMax      float64 `json:"stf_max"`      // STF max value
	ToolID      int64   `json:"tool_id"`      // Currently assigned tool
	LinkedNotes []int64 `json:"notes"`        // Contains note ids from the "notes" table
}

// New creates a new MetalSheet with default values
func NewMetalSheet(u *User, toolID int64) *MetalSheet {
	sheet := &MetalSheet{
		TileHeight:  0,
		Value:       0,
		MarkeHeight: 0,
		STF:         0,
		STFMax:      0,
		ToolID:      toolID,
		LinkedNotes: make([]int64, 0),
	}

	return sheet
}

// String returns a string representation of the metal sheet
func (ms *MetalSheet) String() string {
	return fmt.Sprintf("Blech #%d (TH: %.1f, V: %.1f, MH: %d, STF: %.1f/%.1f)",
		ms.ID, ms.TileHeight, ms.Value, ms.MarkeHeight, ms.STF, ms.STFMax)
}

// MetalSheetWithNotes represents a metal sheet with its related notes loaded
type MetalSheetWithNotes struct {
	*MetalSheet
	LoadedNotes []*Note `json:"loaded_notes"`
	Tool        *Tool   `json:"tool,omitempty"` // The tool currently using this sheet
}
