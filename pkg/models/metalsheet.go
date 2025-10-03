package models

import (
	"fmt"
	"sort"
)

// MachineType represents the type of machine (SACMI or SITI)
type MachineType string

// Machine type identifiers
const (
	MachineTypeSACMI MachineType = "SACMI"
	MachineTypeSITI  MachineType = "SITI"
)

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

// ParseMachineType parses a string into a MachineType with validation
func ParseMachineType(s string) (MachineType, error) {
	mt := MachineType(s)
	if !mt.IsValid() {
		return "", fmt.Errorf("invalid machine type: %s (must be %s or %s)", s, MachineTypeSACMI, MachineTypeSITI)
	}
	return mt, nil
}

// MustParseMachineType parses a string into a MachineType, panicking on invalid input
func MustParseMachineType(s string) MachineType {
	mt, err := ParseMachineType(s)
	if err != nil {
		panic(err)
	}
	return mt
}

// GetAllMachineTypes returns all valid machine types
func GetAllMachineTypes() []MachineType {
	return []MachineType{MachineTypeSACMI, MachineTypeSITI}
}

// GetAllMachineTypeStrings returns all valid machine types as strings
func GetAllMachineTypeStrings() []string {
	types := GetAllMachineTypes()
	result := make([]string, len(types))
	for i, mt := range types {
		result[i] = mt.String()
	}
	return result
}

// DisplayName returns a human-readable display name for the machine type
func (mt MachineType) DisplayName() string {
	switch mt {
	case MachineTypeSACMI:
		return "SACMI Machine"
	case MachineTypeSITI:
		return "SITI Machine"
	default:
		return string(mt)
	}
}

// MarshalJSON implements the json.Marshaler interface
func (mt MachineType) MarshalJSON() ([]byte, error) {
	return []byte(`"` + string(mt) + `"`), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (mt *MachineType) UnmarshalJSON(data []byte) error {
	// Remove quotes from JSON string
	if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
		return fmt.Errorf("invalid JSON string for MachineType: %s", string(data))
	}

	str := string(data[1 : len(data)-1])
	parsed, err := ParseMachineType(str)
	if err != nil {
		return err
	}

	*mt = parsed
	return nil
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
	ID          int64       `json:"id"`
	TileHeight  float64     `json:"tile_height"`  // Tile height in mm
	Value       float64     `json:"value"`        // Value
	MarkeHeight int         `json:"marke_height"` // Marke height
	STF         float64     `json:"stf"`          // STF value
	STFMax      float64     `json:"stf_max"`      // STF max value
	Identifier  MachineType `json:"identifier"`   // Machine type identifier ("SACMI" or "SITI")
	ToolID      int64       `json:"tool_id"`      // Currently assigned tool
	LinkedNotes []int64     `json:"notes"`        // Contains note ids from the "notes" table
}

// New creates a new MetalSheet with default values
func NewMetalSheet(u *User, toolID int64) *MetalSheet {
	sheet := &MetalSheet{
		TileHeight:  0,
		Value:       0,
		MarkeHeight: 0,
		STF:         0,
		STFMax:      0,
		Identifier:  MachineTypeSACMI, // Default to SACMI
		ToolID:      toolID,
		LinkedNotes: make([]int64, 0),
	}

	return sheet
}

// String returns a string representation of the metal sheet
func (ms *MetalSheet) String() string {
	return fmt.Sprintf("Blech #%d [%s] (TH: %.1f, V: %.1f, MH: %d, STF: %.1f/%.1f)",
		ms.ID, ms.Identifier, ms.TileHeight, ms.Value, ms.MarkeHeight, ms.STF, ms.STFMax)
}

// IsValidIdentifier checks if the identifier is a valid machine type
func (ms *MetalSheet) IsValidIdentifier() bool {
	return ms.Identifier.IsValid()
}

// IsSACMI returns true if the metal sheet is from a SACMI machine
func (ms *MetalSheet) IsSACMI() bool {
	return ms.Identifier.IsSACMI()
}

// IsSITI returns true if the metal sheet is from a SITI machine
func (ms *MetalSheet) IsSITI() bool {
	return ms.Identifier.IsSITI()
}

// SetMachineType sets the machine type identifier with validation
func (ms *MetalSheet) SetMachineType(machineType MachineType) error {
	if !machineType.IsValid() {
		return fmt.Errorf("invalid machine type: %s (must be %s or %s)",
			machineType, MachineTypeSACMI, MachineTypeSITI)
	}
	ms.Identifier = machineType
	return nil
}

// GetValidMachineTypes returns a slice of all valid machine types
// Deprecated: Use GetAllMachineTypes() instead
func GetValidMachineTypes() []MachineType {
	return GetAllMachineTypes()
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
func IsSACMIPress(pressNumber PressNumber) bool {
	return pressNumber == 0 || pressNumber == 5
}

// IsSITIPress returns true if the given press number uses SITI machines
// All presses except 0 and 5 use SITI machines
func IsSITIPress(pressNumber PressNumber) bool {
	return !IsSACMIPress(pressNumber)
}

// MetalSheetWithNotes represents a metal sheet with its related notes loaded
type MetalSheetWithNotes struct {
	*MetalSheet
	LoadedNotes []*Note `json:"loaded_notes"`
	Tool        *Tool   `json:"tool,omitempty"` // The tool currently using this sheet
}
