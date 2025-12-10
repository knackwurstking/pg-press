package shared

import "github.com/knackwurstking/pg-press/errors"

type PressRegeneration struct {
	ID          EntityID    `json:"id"`           // ID is the unique identifier for the PressRegeneration entity
	PressNumber PressNumber `json:"press_number"` // PressNumber indicates which press has regenerated
	Start       int64       `json:"start"`        // Start timestamp in milliseconds
	Stop        int64       `json:"stop"`         // Stop timestamp in milliseconds
	Cycles      int64       `json:"cycles"`       // Cycles indicates the number cyles done before regeneration
}

func (pr *PressRegeneration) Validate() *errors.ValidationError {
	if pr.PressNumber < 0 {
		return errors.NewValidationError("press_number must be non-negative")
	}

	if pr.Cycles < 0 {
		return errors.NewValidationError("cycles must be non-negative")
	}

	if pr.Start < 0 {
		return errors.NewValidationError("start must be non-negative")
	}
	if pr.Stop < 0 {
		return errors.NewValidationError("stop must be non-negative")
	}
	if pr.Stop < pr.Start {
		return errors.NewValidationError("stop must be greater than or equal to start")
	}

	return nil
}

func (pr *PressRegeneration) Clone() *PressRegeneration {
	return &PressRegeneration{
		ID:          pr.ID,
		PressNumber: pr.PressNumber,
		Start:       pr.Start,
		Stop:        pr.Stop,
		Cycles:      pr.Cycles,
	}
}

var _ Entity[*PressRegeneration] = (*PressRegeneration)(nil)
