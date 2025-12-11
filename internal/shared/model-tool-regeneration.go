package shared

import "github.com/knackwurstking/pg-press/internal/errors"

type ToolRegeneration struct {
	ID     EntityID `json:"id"`      // ID is the unique identifier for the ToolRegeneration entity
	ToolID int64    `json:"tool_id"` // ToolID indicates which tool has regenerated
	Start  int64    `json:"start"`   // Start timestamp in milliseconds
	Stop   int64    `json:"stop"`    // Stop timestamp in milliseconds
	Cycles int64    `json:"cycles"`  // Cycles indicates the number cyles done before regeneration
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

	return nil
}

func (tr *ToolRegeneration) Clone() *ToolRegeneration {
	return &ToolRegeneration{
		ID:     tr.ID,
		ToolID: tr.ToolID,
		Start:  tr.Start,
		Stop:   tr.Stop,
		Cycles: tr.Cycles,
	}
}

var _ Entity[*ToolRegeneration] = (*ToolRegeneration)(nil)
