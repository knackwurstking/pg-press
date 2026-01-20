package shared

import (
	"fmt"
	"slices"

	"github.com/knackwurstking/pg-press/internal/errors"
)

// Press represents a press machine with its associated tools and cassettes.
//
// The Press struct models manufacturing equipment that has two tool slots (upper and lower)
// and tracks production cycles. It is designed to work with the overall system's tool management
// and cycle tracking features.
//
// Notes:
//   - This is a new type which does not exist in the original models package
//   - The upper cassette is handled by the tool type, the press does not care about it
type Press struct {
	// ID represents the unique identifier for this press machine (e.g., 1, 2, 3)
	ID PressNumber `json:"id"`

	// SlotUp is the EntityID of the upper tool in this press's slot
	SlotUp EntityID `json:"slot_up"`

	// SlotDown is the EntityID of the lower tool in this press's slot
	SlotDown EntityID `json:"slot_down"`

	// CyclesOffset represents the number of cycles since last regeneration
	CyclesOffset int64 `json:"cycles_offset"`

	// Type specifies the type of press (e.g., "SACMI", "SITI")
	Type MachineType `json:"type"`
}

// Validate checks if the Press struct contains valid data.
//
// It ensures that:
//   - The press type is one of the supported types (SACMI or SITI)
//
// Returns:
//   - *errors.ValidationError: Validation error if type is invalid, nil otherwise
func (p *Press) Validate() *errors.ValidationError {
	if !slices.Contains([]MachineType{MachineTypeSACMI, MachineTypeSITI}, p.Type) {
		return errors.NewValidationError("press type must be either 'SACMI' or 'SITI'")
	}

	return nil
}

// Clone creates a copy of the Press struct.
//
// Returns:
//   - *Press: A new Press struct with identical values to the receiver
func (p *Press) Clone() *Press {
	return &Press{
		ID:           p.ID,
		SlotUp:       p.SlotUp,
		SlotDown:     p.SlotDown,
		CyclesOffset: p.CyclesOffset,
		Type:         p.Type,
	}
}

// String returns a string representation of the Press struct.
//
// This is useful for logging and debugging purposes to quickly visualize
// the press configuration.
//
// Returns:
//   - string: Formatted string showing all Press fields
func (p *Press) String() string {
	return fmt.Sprintf(
		"Press{ID:%d, SlotUp:%d, SlotDown:%d, CyclesOffset:%d, Type:%s}",
		p.ID, p.SlotUp, p.SlotDown, p.CyclesOffset, p.Type,
	)
}
