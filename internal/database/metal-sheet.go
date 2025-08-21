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
	Material    string              `json:"material"`  // Ex: Steel, Aluminum, Stainless Steel
	Thickness   float64             `json:"thickness"` // Thickness in mm
	Width       float64             `json:"width"`     // Width in mm
	Height      float64             `json:"height"`    // Height in mm
	Position    Position            `json:"position"`  // Top or Bottom (reusing from tool.go)
	Status      MetalSheetStatus    `json:"status"`
	ToolID      *int64              `json:"tool_id"` // Currently assigned tool (nullable)
	LinkedNotes []int64             `json:"notes"`   // Contains note ids from the "notes" table
	Mods        Mods[MetalSheetMod] `json:"mods"`
}

// NewMetalSheet creates a new MetalSheet with default values
func NewMetalSheet(m ...*Modified[MetalSheetMod]) *MetalSheet {
	return &MetalSheet{
		Status:      MetalSheetStatusAvailable,
		LinkedNotes: make([]int64, 0),
		Mods:        m,
	}
}

// String returns a string representation of the metal sheet
func (ms *MetalSheet) String() string {
	dimensions := fmt.Sprintf("%.1fx%.1fx%.1f", ms.Width, ms.Height, ms.Thickness)
	positionStr := ""
	switch ms.Position {
	case PositionTop:
		positionStr = ", Oberteil"
	case PositionBottom:
		positionStr = ", Unterteil"
	}
	return fmt.Sprintf("%s %smm (%s%s)", ms.Material, dimensions, ms.Status, positionStr)
}

// IsAvailable checks if the metal sheet is available for use
func (ms *MetalSheet) IsAvailable() bool {
	return ms.Status == MetalSheetStatusAvailable
}

// MetalSheetMod represents modifications to a metal sheet
type MetalSheetMod struct {
	Material    string           `json:"material"`
	Thickness   float64          `json:"thickness"`
	Width       float64          `json:"width"`
	Height      float64          `json:"height"`
	Position    Position         `json:"position"`
	Status      MetalSheetStatus `json:"status"`
	ToolID      *int64           `json:"tool_id"`
	LinkedNotes []int64          `json:"notes"`
}

// MetalSheetWithNotes represents a metal sheet with its related notes loaded
type MetalSheetWithNotes struct {
	*MetalSheet
	LoadedNotes []*Note `json:"loaded_notes"`
	Tool        *Tool   `json:"tool,omitempty"` // The tool currently using this sheet
}
