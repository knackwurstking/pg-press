package shared

import (
	"fmt"

	"github.com/knackwurstking/pg-press/internal/errors"
)

// ToolRegeneration represents a regeneration event for a tool in a press machine
type ToolRegeneration struct {
	ID       EntityID  `json:"id"`        // ID is the unique identifier for the ToolRegeneration entity
	ToolID   EntityID  `json:"tool_id"`   // ToolID indicates which tool has regenerated
	Start    UnixMilli `json:"start"`     // Start timestamp in milliseconds
	Stop     UnixMilli `json:"stop"`      // Stop timestamp in milliseconds
	Cycles   int64     `json:"cycles"`    // Cycles indicates the number cycles done before regeneration
	Position Slot      `json:"position"`  // Position indicates the position of the tool that regenerated
}

func (tr *ToolRegeneration) Validate() *errors.ValidationError {
	if tr.ToolID < 0 {
		return errors.NewValidationError("tool_id must be non-negative")
	}

	if tr.Start < 0 {
		return errors.NewValidationError("start must be non-negative")
	}
	if tr.Stop < 0 {
		return errors.NewValidationError("stop must be non-negative")
	}
	if tr.Stop < tr.Start {
		return errors.NewValidationError("stop must be greater than or equal to start")
	}

	if tr.Position != SlotUpper && tr.Position != SlotLower && tr.Position != SlotUpperCassette {
		return errors.NewValidationError("position must be either upper, lower, or upper cassette")
	}

	return nil
}

func (tr *ToolRegeneration) Clone() *ToolRegeneration {
	return &ToolRegeneration{
		ID:       tr.ID,
		ToolID:   tr.ToolID,
		Start:    tr.Start,
		Stop:     tr.Stop,
		Cycles:   tr.Cycles,
		Position: tr.Position,
	}
}

func (tr *ToolRegeneration) String() string {
	return fmt.Sprintf(
		"ToolRegeneration[ID=%d, ToolID=%d, Cycles=%d, Start=%d, Stop=%d, Position=%d]",
		tr.ID, tr.ToolID, tr.Cycles, tr.Start, tr.Stop, tr.Position,
	)
}

var _ Entity[*ToolRegeneration] = (*ToolRegeneration)(nil)
