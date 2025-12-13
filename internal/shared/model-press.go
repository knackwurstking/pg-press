package shared

import (
	"fmt"

	"github.com/knackwurstking/pg-press/internal/errors"
)

// Press represents a press machine with its associated tools and cassettes
//
// Notes:
//   - This is a new type which does not exist in the original models package
//   - The upper cassette is handled by the tool type, the press does not care about it
type Press struct {
	ID               PressNumber `json:"id"`                          // Press number, required
	SlotUp           EntityID    `json:"slot_up"`                     // Upper tool entity ID, required
	SlotDown         EntityID    `json:"slot_down"`                   // Lower tool entity ID, required
	LastRegeneration EntityID    `json:"last_regeneration,omitempty"` // Tools last regeneration (entity) ID, optional
	StartCycles      int64       `json:"start_cycles"`                // Press cycles since last regeneration, optional
	Cycles           int64       `json:"cycles"`                      // Current press cycles, required
	Type             PressType   `json:"type"`                        // Type of press, e.g., "SACMI", "SITI"
}

func (p *Press) Validate() *errors.ValidationError {
	if p.SlotUp <= 0 {
		return errors.NewValidationError("upper tool id cannot be lower or equal 0")
	}
	if p.SlotDown <= 0 {
		return errors.NewValidationError("lower tool id cannot be lower or equal 0")
	}

	if p.Cycles < 0 {
		return errors.NewValidationError("cycles have to be positive or zero")
	}

	return nil
}

func (p *Press) Clone() *Press {
	return &Press{
		ID:               p.ID,
		SlotUp:           p.SlotUp,
		SlotDown:         p.SlotDown,
		LastRegeneration: p.LastRegeneration,
		StartCycles:      p.StartCycles,
		Cycles:           p.Cycles,
		Type:             p.Type,
	}
}

func (p *Press) String() string {
	return fmt.Sprintf(
		"Press[ID=%d, SlotUp=%d, SlotDown=%d, LastRegeneration=%d, StartCycles=%d, Cycles=%d, Type=%s]",
		p.ID, p.SlotUp, p.SlotDown, p.LastRegeneration, p.StartCycles, p.Cycles, p.Type,
	)
}

var _ Entity[*Press] = (*Press)(nil)
