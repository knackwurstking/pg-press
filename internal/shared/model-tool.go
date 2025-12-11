package shared

import (
	"slices"

	"github.com/knackwurstking/pg-press/internal/errors"
)

// Tool represents a tool used in a press machine,
// there are upper and lower tools. Each tool can have its own regeneration history.
// And the upper tool type has an optional cassette slot.
type Tool struct {
	ID               EntityID `json:"id"`
	Width            int      `json:"width"`         // Width defines the tile width this tool can press
	Height           int      `json:"height"`        // Height defines the tile height this tool can press
	Type             string   `json:"type"`          // Type represents the tool type, e.g., "MASS", "FC", "GTC", etc.
	Code             string   `json:"code"`          // Code is the unique tool code/identifier, "G01", "12345", etc.
	CyclesOffset     int64    `json:"cycles_offset"` // CyclesOffset is an offset added to the cycles count
	Cycles           int64    `json:"cycles"`        // Cycles indicates how many cycles this tool has done
	LastRegeneration EntityID `json:"last_regeneration,omitempty"`
	Regenerating     bool     `json:"regenerating"` // A regeneration resets the cycles counter, including the offset, back to zero
	IsDead           bool     `json:"is_dead"`      // IsDead indicates if the tool is dead/destroyed
	Slot             Slot     `json:"slot"`         // SlotType indicates the cassette slot type
}

// Validate checks if the tool data is valid
func (t *Tool) Validate() *errors.ValidationError {
	// verify width and height not negative
	if t.Width < 0 {
		return errors.NewValidationError("Tool width cannot be negative")
	}
	if t.Height < 0 {
		return errors.NewValidationError("Tool height cannot be negative")
	}
	// verify type and code are not empty
	if t.Type == "" {
		return errors.NewValidationError("Tool type is required")
	}
	if t.Code == "" {
		return errors.NewValidationError("Tool code is required")
	}
	if !slices.Contains([]Slot{SlotUp, SlotDown}, t.Slot) {
		return errors.NewValidationError("Tool slot type is invalid: %v", t.Slot)
	}
	return nil
}

// Clone creates a copy of the tool
func (t *Tool) Clone() *Tool {
	return &Tool{
		ID:               t.ID,
		Width:            t.Width,
		Height:           t.Height,
		Type:             t.Type,
		Code:             t.Code,
		Slot:             t.Slot,
		CyclesOffset:     t.CyclesOffset,
		Cycles:           t.Cycles,
		Regenerating:     t.Regenerating,
		LastRegeneration: t.LastRegeneration,
		IsDead:           t.IsDead,
	}
}

var _ Entity[*Tool] = (*Tool)(nil)
