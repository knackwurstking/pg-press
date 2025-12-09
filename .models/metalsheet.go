package models

import (
	"fmt"
	"sort"
	"time"

	"github.com/knackwurstking/pg-press/errors"
)

// Machine type identifiers
const (
	MachineTypeSACMI MachineType = "SACMI"
	MachineTypeSITI  MachineType = "SITI"
)

type MetalSheetID int64

// MachineType represents the type of machine (SACMI or SITI)
type MachineType string

// String returns the string representation of the machine type
func (mt MachineType) String() string {
	return string(mt)
}

// IsValid checks if the machine type is valid
func (mt MachineType) IsValid() bool {
	return mt == MachineTypeSACMI || mt == MachineTypeSITI
}

// IsSACMI returns true if the machine type is SACMI
func (mt MachineType) IsSACMI() bool {
	return mt == MachineTypeSACMI
}

// IsSITI returns true if the machine type is SITI
func (mt MachineType) IsSITI() bool {
	return mt == MachineTypeSITI
}

// DisplayName returns a human-readable display name for the machine type
func (mt MachineType) DisplayName() string {
	switch mt {
	case MachineTypeSACMI:
		return "SACMI"
	case MachineTypeSITI:
		return "SITI"
	default:
		return string(mt)
	}
}

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
	ID          MetalSheetID `json:"id"`
	TileHeight  float64      `json:"tile_height"`  // Tile height in mm
	Value       float64      `json:"value"`        // Value
	MarkeHeight int          `json:"marke_height"` // Marke height
	STF         float64      `json:"stf"`          // STF value
	STFMax      float64      `json:"stf_max"`      // STF max value
	Identifier  MachineType  `json:"identifier"`   // Machine type identifier ("SACMI" or "SITI")
	ToolID      ToolID       `json:"tool_id"`      // Currently assigned tool
	UpdatedAt   time.Time    `json:"updated_at"`   // Last updated timestamp
}

// New creates a new MetalSheet with default values
func NewMetalSheet(u *User, toolID ToolID) *MetalSheet {
	sheet := &MetalSheet{
		TileHeight:  0,
		Value:       0,
		MarkeHeight: 0,
		STF:         0,
		STFMax:      0,
		Identifier:  MachineTypeSACMI, // Default to SACMI
		ToolID:      toolID,
	}

	return sheet
}

// String returns a string representation of the metal sheet
func (ms *MetalSheet) String() string {
	return fmt.Sprintf("Blech #%d [%s] (TH: %.1f, V: %.1f, MH: %d, STF: %.1f/%.1f)",
		ms.ID, ms.Identifier, ms.TileHeight, ms.Value, ms.MarkeHeight, ms.STF, ms.STFMax)
}

func (ms *MetalSheet) Validate() *errors.ValidationError {
	// Validate machine type identifier
	if !ms.Identifier.IsValid() {
		return errors.NewValidationError("invalid machine type identifier: %s", ms.Identifier)
	}

	// Validate tool ID (foreign key reference)
	if ms.ToolID <= 0 {
		return errors.NewValidationError("invalid tool ID: %d", ms.ToolID)
	}

	// Validate tile height
	if ms.TileHeight < 0 {
		return errors.NewValidationError("tile height cannot be negative: %.1f", ms.TileHeight)
	}

	// Validate value
	if ms.Value < 0 {
		return errors.NewValidationError("value cannot be negative: %.1f", ms.Value)
	}

	// Validate marke height
	if ms.MarkeHeight < 0 {
		return errors.NewValidationError("marke height cannot be negative: %d", ms.MarkeHeight)
	}

	// Validate STF
	if ms.STF < 0 {
		return errors.NewValidationError("STF cannot be negative: %.1f", ms.STF)
	}

	// Validate STF Max
	if ms.STFMax < 0 {
		return errors.NewValidationError("STF Max cannot be negative: %.1f", ms.STFMax)
	}

	// Validate STF relationship - STFMax should be >= STF
	if ms.STFMax < ms.STF {
		return errors.NewValidationError("STF Max (%.1f) cannot be less than STF (%.1f)", ms.STFMax, ms.STF)
	}

	return nil
}

// ParseMachineType parses a string into a MachineType with validation
func ParseMachineType(s string) (MachineType, error) {
	mt := MachineType(s)
	if !mt.IsValid() {
		return "", fmt.Errorf("invalid machine type: %s (must be %s or %s)", s, MachineTypeSACMI, MachineTypeSITI)
	}
	return mt, nil
}

// GetAllMachineTypes returns all valid machine types
func GetAllMachineTypes() []MachineType {
	return []MachineType{MachineTypeSACMI, MachineTypeSITI}
}

// GetMachineTypeForPress returns the machine type for a given press number
// Press 0 and 5 use SACMI machines, all others use SITI machines
func GetMachineTypeForPress(pressNumber PressNumber) MachineType {
	if pressNumber == 0 || pressNumber == 5 {
		return MachineTypeSACMI
	}
	return MachineTypeSITI
}

// IsSACMIPress returns true if the given press number uses SACMI machines
// Press 0 and 5 use SACMI machines
func IsSACMI(pressNumber PressNumber) bool {
	return pressNumber == 0 || pressNumber == 5
}

// IsSITIPress returns true if the given press number uses SITI machines
// All presses except 0 and 5 use SITI machines
func IsSITI(pressNumber PressNumber) bool {
	return !IsSACMI(pressNumber)
}
