package shared

import (
	"fmt"

	"github.com/knackwurstking/pg-press/internal/errors"
)

type Slot int

const (
	SlotUnknown Slot = iota
	SlotUpper
	SlotLower
	SlotUpperCassette
)

const (
	ToolCyclesWarning int64 = 800000
	ToolCyclesError   int64 = 1000000
)

func (s Slot) German() string {
	switch s {
	case SlotUpper:
		return "Oberteil"
	case SlotLower:
		return "Unterteil"
	case SlotUpperCassette:
		return "Kassette"
	default:
		return "Unbekannt"
	}
}

type BaseTool struct {
	ID               EntityID `json:"id"`
	Width            int      `json:"width"`         // Width defines the tile width this tool can press
	Height           int      `json:"height"`        // Height defines the tile height this tool can press
	Position         Slot     `json:"position"`      // Position indicates the position of the tool in the press (e.g., 1 for upper, 2 for lower)
	Type             string   `json:"type"`          // Type represents the tool type, e.g., "MASS", "FC", "GTC", etc.
	Code             string   `json:"code"`          // Code is the unique tool code/identifier, "G01", "12345", etc.
	CyclesOffset     int64    `json:"cycles_offset"` // CyclesOffset is an offset added to the cycles count
	Cycles           int64    `json:"cycles"`        // Cycles indicates how many cycles this tool has done
	LastRegeneration EntityID `json:"last_regeneration,omitempty"`
	Regenerating     bool     `json:"regenerating"` // A regeneration resets the cycles counter, including the offset, back to zero
	IsDead           bool     `json:"is_dead"`      // IsDead indicates if the tool is dead/destroyed
}

func (bt *BaseTool) Validate() *errors.ValidationError {
	if bt.Width < 0 {
		return errors.NewValidationError("Tool width cannot be negative")
	}
	if bt.Height < 0 {
		return errors.NewValidationError("Tool height cannot be negative")
	}
	if bt.Type == "" {
		return errors.NewValidationError("Tool type is required")
	}
	if bt.Code == "" {
		return errors.NewValidationError("Tool code is required")
	}
	if bt.Cycles < 0 {
		return errors.NewValidationError("Tool cycles cannot be negative")
	}
	return nil
}

func (bt *BaseTool) Clone() BaseTool {
	return BaseTool{
		ID:               bt.ID,
		Width:            bt.Width,
		Height:           bt.Height,
		Position:         bt.Position,
		Type:             bt.Type,
		Code:             bt.Code,
		CyclesOffset:     bt.CyclesOffset,
		Cycles:           bt.Cycles,
		LastRegeneration: bt.LastRegeneration,
		Regenerating:     bt.Regenerating,
		IsDead:           bt.IsDead,
	}
}

func (bt *BaseTool) String() string {
	return fmt.Sprintf(
		"BaseTool[ID=%d, Width=%d, Height=%d, Position=%d, Type=%s, Code=%s, "+
			"CyclesOffset=%d, Cycles=%d, LastRegeneration=%d, Regenerating=%t, "+
			"IsDead=%t]",
		bt.ID,
		bt.Width,
		bt.Height,
		bt.Position,
		bt.Type,
		bt.Code,
		bt.CyclesOffset,
		bt.Cycles,
		bt.LastRegeneration,
		bt.Regenerating,
		bt.IsDead,
	)
}

// Tool represents a tool used in a press machine,
// there are upper and lower tools. Each tool can have its own regeneration history.
type Tool struct {
	BaseTool
	Cassette EntityID `json:"cassette"` // Cassette indicates the cassette ID this tool belongs to (if any)
}

func (t *Tool) German() string {
	return fmt.Sprintf("%dx%d %s %s", t.Width, t.Height, t.Type, t.Code)
}

// Validate checks if the tool data is valid
func (t *Tool) Validate() *errors.ValidationError {
	if verr := t.BaseTool.Validate(); verr != nil {
		return verr
	}

	if t.Cassette < 0 {
		return errors.NewValidationError("Tool cassette ID cannot be negative")
	}

	if t.Position != SlotUpper && t.Position != SlotLower {
		return errors.NewValidationError("Tool position must be either upper or lower")
	}

	return nil
}

// Clone creates a copy of the tool
func (t *Tool) Clone() *Tool {
	return &Tool{
		BaseTool: t.BaseTool.Clone(),
		Cassette: t.Cassette,
	}
}

func (t *Tool) String() string {
	return fmt.Sprintf(
		"Tool[BaseTool=%s, Cassette=%s]",
		t.BaseTool.String(),
		t.Cassette,
	)
}

type Cassette struct {
	BaseTool
	MinThickness float32 `json:"min_thickness"` // required
	MaxThickness float32 `json:"max_thickness"` // required
}

func (c *Cassette) German() string {
	return fmt.Sprintf("%dx%d %s %s %.1fmm-%.1fmm",
		c.Width, c.Height,
		c.Type,
		c.Code,
		c.MinThickness,
		c.MaxThickness,
	)
}

func (c *Cassette) Validate() *errors.ValidationError {
	if verr := c.BaseTool.Validate(); verr != nil {
		return verr
	}

	if c.MinThickness < 0 {
		return errors.NewValidationError("Cassette min_thickness cannot be negative")
	}
	if c.MaxThickness < 0 {
		return errors.NewValidationError("Cassette max_thickness cannot be negative")
	}
	if c.MaxThickness < c.MinThickness {
		return errors.NewValidationError("Cassette max_thickness cannot be less than min_thickness")
	}

	if c.Position != SlotUpperCassette {
		return errors.NewValidationError("Cassette position must be upper cassette")
	}

	return nil
}

func (c *Cassette) Clone() *Cassette {
	return &Cassette{
		BaseTool:     c.BaseTool.Clone(),
		MinThickness: c.MinThickness,
		MaxThickness: c.MaxThickness,
	}
}

func (c *Cassette) String() string {
	return fmt.Sprintf(
		"Cassette[BaseTool=%s, MinThickness=%.1f, MaxThickness=%.1f]",
		c.BaseTool.String(),
		c.MinThickness,
		c.MaxThickness,
	)
}

var _ Entity[*Tool] = (*Tool)(nil)
var _ Entity[*Cassette] = (*Cassette)(nil)
