package shared

import (
	"fmt"

	"github.com/knackwurstking/pg-press/internal/errors"
)

type PressRegeneration struct {
	ID          EntityID    `json:"id"`           // ID is the unique identifier for the PressRegeneration entity
	PressNumber PressNumber `json:"press_number"` // PressNumber indicates which press has regenerated
	Start       UnixMilli   `json:"start"`        // Start timestamp in milliseconds
	Stop        UnixMilli   `json:"stop"`         // Stop timestamp in milliseconds
}

func (pr *PressRegeneration) Validate() *errors.ValidationError {
	if pr.PressNumber < 0 {
		return errors.NewValidationError("press_number must be non-negative")
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
	}
}

func (pr *PressRegeneration) String() string {
	return fmt.Sprintf(
		"PressRegeneration{ID:%d, PressNumber:%d, Start:%d, Stop:%d}",
		pr.ID, pr.PressNumber, pr.Start, pr.Stop,
	)
}
