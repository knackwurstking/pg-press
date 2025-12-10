package shared

import "github.com/knackwurstking/pg-press/errors"

// Press represents a press machine with its associated tools and cassettes
//
// Notes:
//   - This is a new type which does not exist in the original models package
//   - The upper cassette is handled by the tool type, the press does not care about it
type Press struct {
	ID               PressNumber
	UpperTool        EntityID // Upper tool entity ID, required
	LowerTool        EntityID // Lower tool entity ID, required
	LastRegeneration EntityID // Regeneration entity ID, cycles will be zeroed on regeneration
	StartCycles      int64    // Press cycles at last regeneration, optional, default 0
	Cycles           int64    // Current total press cycles since last regeneration
}

func (p *Press) Validate() *errors.ValidationError {
	if p.UpperTool <= 0 {
		return errors.NewValidationError("upper tool id cannot be lower or equal 0")
	}
	if p.LowerTool <= 0 {
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
		UpperTool:        p.UpperTool,
		LowerTool:        p.LowerTool,
		LastRegeneration: p.LastRegeneration,
		Cycles:           p.Cycles,
	}
}

var _ Entity[*Press] = (*Press)(nil)
