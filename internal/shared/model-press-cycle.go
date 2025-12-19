package shared

import (
	"fmt"

	"github.com/knackwurstking/pg-press/internal/errors"
)

type Cycle struct {
	ID            EntityID    `json:"id"`             // ID is the unique identifier for the Cycle entity
	ToolID        EntityID    `json:"tool_id"`        // ToolID is the identifier for the associated Tool entity
	PressNumber   PressNumber `json:"press_number"`   // PressNumber indicates which press machine performed the cycles
	PressCycles   int64       `json:"cycles"`         // PressCycles is the number of cycles completed during this time period
	PartialCycles int64       `json:"partial_cycles"` // PartialCycles are the completed cycles during this time period (calculated)
	Start         UnixMilli   `json:"start"`          // Start timestamp in milliseconds
	Stop          UnixMilli   `json:"stop"`           // Stop timestamp in milliseconds
}

func (c *Cycle) Validate() *errors.ValidationError {
	if c.PressNumber < 0 {
		return errors.NewValidationError("press_number must be non-negative")
	}

	if c.PressCycles < 0 {
		return errors.NewValidationError("cycles must be non-negative")
	}

	if c.Start < 0 {
		return errors.NewValidationError("start must be non-negative")
	}
	if c.Stop < 0 {
		return errors.NewValidationError("stop must be non-negative")
	}
	if c.Stop < c.Start {
		return errors.NewValidationError("stop must be greater than or equal to start")
	}

	if c.ToolID <= 0 {
		return errors.NewValidationError("tool_id must be positive")
	}

	return nil
}

func (c *Cycle) Clone() *Cycle {
	return &Cycle{
		ID:            c.ID,
		ToolID:        c.ToolID,
		PressNumber:   c.PressNumber,
		PressCycles:   c.PressCycles,
		PartialCycles: c.PartialCycles,
		Start:         c.Start,
		Stop:          c.Stop,
	}
}

func (c *Cycle) String() string {
	return fmt.Sprintf(
		"Cycle{ID:%d, ToolID:%d, PressNumber:%d, PressCycles:%d, PartialCycles:%d, Start:%d, Stop:%d}",
		c.ID, c.ToolID, c.PressNumber, c.PressCycles, c.PartialCycles, c.Start, c.Stop,
	)
}

var _ Entity[*Cycle] = (*Cycle)(nil)
