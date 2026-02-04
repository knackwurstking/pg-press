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
	ID     EntityID    `json:"id"`
	Number PressNumber `json:"number"`

	// Type specifies the type of press (e.g., "SACMI", "SITI")
	Type MachineType `json:"type"`

	// Code is the press machine's code identifier
	Code string `json:"code"`

	// SlotUp is the EntityID of the upper tool in this press's slot
	SlotUp EntityID `json:"slot_up"`

	// SlotDown is the EntityID of the lower tool in this press's slot
	SlotDown EntityID `json:"slot_down"`

	// CyclesOffset represents the number of cycles since last regeneration
	CyclesOffset int64 `json:"cycles_offset"`
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
		Type:         p.Type,
		Code:         p.Code,
		SlotUp:       p.SlotUp,
		SlotDown:     p.SlotDown,
		CyclesOffset: p.CyclesOffset,
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
		"Press{ID:%d, Type:%s, Code:%s, SlotUp:%d, SlotDown:%d, CyclesOffset:%d}",
		p.ID, p.Type, p.Code, p.SlotUp, p.SlotDown, p.CyclesOffset,
	)
}

func (p *Press) German() string {
	return fmt.Sprintf("Presse %s (Type: %s, Code: %s)", p.Number.String(), p.Type, p.Code)
}
