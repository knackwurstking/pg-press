package shared

import "github.com/knackwurstking/pg-press/errors"

type Cycle struct {
	ID          EntityID    `json:"id"`           // ID is the unique identifier for the Cycle entity
	PressNumber PressNumber `json:"press_number"` // PressNumber indicates which press machine performed the cycles
	Cycles      int64       `json:"cycles"`       // Cycles indicates the number of (partial) cycles
	Start       int64       `json:"start"`        // Start timestamp in milliseconds
	Stop        int64       `json:"stop"`         // Stop timestamp in milliseconds
}

func (c *Cycle) Validate() *errors.ValidationError {
	if c.PressNumber < 0 {
		return errors.NewValidationError("press_number must be non-negative")
	}

	if c.Cycles < 0 {
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

	return nil
}

func (c *Cycle) Clone() *Cycle {
	return &Cycle{
		ID:          c.ID,
		PressNumber: c.PressNumber,
		Cycles:      c.Cycles,
		Start:       c.Start,
		Stop:        c.Stop,
	}
}

var _ Entity[*Cycle] = (*Cycle)(nil)
