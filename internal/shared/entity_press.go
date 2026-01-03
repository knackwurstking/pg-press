package shared

import (
	"fmt"
	"slices"

	"github.com/knackwurstking/pg-press/internal/errors"
)

// Press represents a press machine with its associated tools and cassettes
//
// Notes:
//   - This is a new type which does not exist in the original models package
//   - The upper cassette is handled by the tool type, the press does not care about it
type Press struct {
	ID           PressNumber `json:"id"`            // Press number, required
	SlotUp       EntityID    `json:"slot_up"`       // Upper tool entity ID, required
	SlotDown     EntityID    `json:"slot_down"`     // Lower tool entity ID, required
	CyclesOffset int64       `json:"cycles_offset"` // Press cycles since last regeneration, optional
	Type         MachineType `json:"type"`          // Type of press, e.g., "SACMI", "SITI"
}

func (p *Press) Validate() *errors.ValidationError {
	if !slices.Contains([]MachineType{MachineTypeSACMI, MachineTypeSITI}, p.Type) {
		return errors.NewValidationError("invalid press type: %s", p.Type)
	}

	return nil
}

func (p *Press) Clone() *Press {
	return &Press{
		ID:           p.ID,
		SlotUp:       p.SlotUp,
		SlotDown:     p.SlotDown,
		CyclesOffset: p.CyclesOffset,
		Type:         p.Type,
	}
}

func (p *Press) String() string {
	return fmt.Sprintf(
		"Press{ID:%d, SlotUp:%d, SlotDown:%d, CyclesOffset:%d, Type:%s}",
		p.ID, p.SlotUp, p.SlotDown, p.CyclesOffset, p.Type,
	)
}
