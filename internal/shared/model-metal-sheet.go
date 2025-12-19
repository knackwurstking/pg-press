package shared

import (
	"fmt"

	"github.com/knackwurstking/pg-press/internal/errors"
)

const (
	MachineTypeSACMI MachineType = "SACMI"
	MachineTypeSITI  MachineType = "SITI"
)

type MachineType string

type BaseMetalSheet struct {
	ID         EntityID `json:"id"`
	ToolID     EntityID `json:"tool_id"`     // Currently assigned tool
	TileHeight float64  `json:"tile_height"` // Tile height in mm
	Value      float64  `json:"value"`       // Value
}

type LowerMetalSheet struct {
	BaseMetalSheet
	MarkeHeight int         `json:"marke_height"` // Marke height
	STF         float64     `json:"stf"`          // STF value
	STFMax      float64     `json:"stf_max"`      // STF max value
	Identifier  MachineType `json:"identifier"`   // Machine type identifier ("SACMI" or "SITI")
}

// Validate checks if the lower metal sheet has valid data
func (u *LowerMetalSheet) Validate() *errors.ValidationError {
	if u.MarkeHeight <= 0 {
		return errors.NewValidationError("marke height must be positive")
	}
	if u.STF <= 0 {
		return errors.NewValidationError("STF value must be positive")
	}
	if u.STFMax <= 0 {
		return errors.NewValidationError("STF max value must be positive")
	}
	if u.Identifier != MachineTypeSACMI && u.Identifier != MachineTypeSITI {
		return errors.NewValidationError("identifier must be either 'SACMI' or 'SITI'")
	}

	return nil
}

// Clone creates a copy of the lower metal sheet
func (u *LowerMetalSheet) Clone() *LowerMetalSheet {
	return &LowerMetalSheet{
		BaseMetalSheet: BaseMetalSheet{
			ID:         u.ID,
			ToolID:     u.ToolID,
			TileHeight: u.TileHeight,
			Value:      u.Value,
		},
		MarkeHeight: u.MarkeHeight,
		STF:         u.STF,
		STFMax:      u.STFMax,
		Identifier:  u.Identifier,
	}
}

// String returns a string representation of the lower metal sheet
func (u *LowerMetalSheet) String() string {
	return fmt.Sprintf(
		"LowerMetalSheet{ID:%s, ToolID:%s, TileHeight:%.2f, Value:%.2f, MarkeHeight:%d, STF:%.2f, STFMax:%.2f, Identifier:%s}",
		u.ID.String(),
		u.ToolID.String(),
		u.TileHeight,
		u.Value,
		u.MarkeHeight,
		u.STF,
		u.STFMax,
		u.Identifier,
	)
}

type UpperMetalSheet struct {
	BaseMetalSheet
}

// Validate checks if the upper metal sheet has valid data
func (u *UpperMetalSheet) Validate() *errors.ValidationError {
	return nil
}

// Clone creates a copy of the upper metal sheet
func (u *UpperMetalSheet) Clone() *UpperMetalSheet {
	return &UpperMetalSheet{
		BaseMetalSheet: BaseMetalSheet{
			ID:         u.ID,
			ToolID:     u.ToolID,
			TileHeight: u.TileHeight,
			Value:      u.Value,
		},
	}
}

// String returns a string representation of the upper metal sheet
func (u *UpperMetalSheet) String() string {
	return fmt.Sprintf(
		"UpperMetalSheet{ID:%s, ToolID:%s, TileHeight:%.2f, Value:%.2f}",
		u.ID.String(),
		u.ToolID.String(),
		u.TileHeight,
		u.Value,
	)
}

var _ Entity[*UpperMetalSheet] = (*UpperMetalSheet)(nil)
var _ Entity[*LowerMetalSheet] = (*LowerMetalSheet)(nil)
