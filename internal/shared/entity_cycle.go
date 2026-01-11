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
	Start         UnixMilli   `json:"start"`          // Start timestamp in milliseconds (injected)
	Stop          UnixMilli   `json:"stop"`           // Stop timestamp in milliseconds, should be the date were the press cycles got read
}

func (c *Cycle) Validate() *errors.ValidationError {
	if c.PressNumber < 0 {
		return errors.NewValidationError("press number must be 0 or greater")
	}

	if c.PressCycles < 0 {
		return errors.NewValidationError("number of cycles must be 0 or greater")
	}

	if c.Start < 0 {
		return errors.NewValidationError("start timestamp must be a positive number")
	}
	if c.Stop < 0 {
		return errors.NewValidationError("stop timestamp must be a positive number")
	}
	if c.Stop < c.Start {
		return errors.NewValidationError("stop date must be after or equal to start date")
	}

	if c.ToolID <= 0 {
		return errors.NewValidationError("tool ID must be specified")
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
