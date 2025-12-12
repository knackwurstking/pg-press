// TODO: Implement a method, to detect if this tool is from type cassette (top)
package shared

import (
	"fmt"

	"github.com/knackwurstking/pg-press/internal/errors"
)

// Tool represents a tool used in a press machine,
// there are upper and lower tools. Each tool can have its own regeneration history.
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
	Cassette         EntityID `json:"cassette"`     // Cassette indicates the cassette ID this tool belongs to (if any)
}

// Validate checks if the tool data is valid
func (t *Tool) Validate() *errors.ValidationError {
	if t.Width < 0 {
		return errors.NewValidationError("Tool width cannot be negative")
	}
	if t.Height < 0 {
		return errors.NewValidationError("Tool height cannot be negative")
	}
	if t.Type == "" {
		return errors.NewValidationError("Tool type is required")
	}
	if t.Code == "" {
		return errors.NewValidationError("Tool code is required")
	}
	if t.Cycles < 0 {
		return errors.NewValidationError("Tool cycles cannot be negative")
	}
	if t.Cassette < 0 {
		return errors.NewValidationError("Tool cassette ID cannot be negative")
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
		Cassette:         t.Cassette,
		CyclesOffset:     t.CyclesOffset,
		Cycles:           t.Cycles,
		Regenerating:     t.Regenerating,
		LastRegeneration: t.LastRegeneration,
		IsDead:           t.IsDead,
	}
}

func (t *Tool) String() string {
	return fmt.Sprintf("Tool[ID=%s, Type=%s, Code=%s]", t.ID.String(), t.Type, t.Code)
}

var _ Entity[*Tool] = (*Tool)(nil)
