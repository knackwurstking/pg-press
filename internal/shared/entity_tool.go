package shared

import (
	"fmt"

	"github.com/knackwurstking/pg-press/internal/errors"
)

const (
	ToolCyclesWarning int64 = 800000
	ToolCyclesError   int64 = 1000000
)

// -----------------------------------------------------------------------------
// Tool
// -----------------------------------------------------------------------------

// Tool represents a tool used in a press machine,
type Tool struct {
	ID           EntityID `json:"id"`
	Width        int      `json:"width"`         // Width defines the tile width this tool can press
	Height       int      `json:"height"`        // Height defines the tile height this tool can press
	Position     Slot     `json:"position"`      // Position indicates the position of the tool in the press (e.g., 1 for upper, 2 for lower)
	Type         string   `json:"type"`          // Type represents the tool type, e.g., "MASS", "FC", "GTC", etc.
	Code         string   `json:"code"`          // Code is the unique tool code/identifier, "G01", "12345", etc.
	CyclesOffset int64    `json:"cycles_offset"` // CyclesOffset is an offset added to the cycles count
	Cycles       int64    `json:"cycles"`        // Cycles indicates how many cycles this tool has done (injected)
	IsDead       bool     `json:"is_dead"`       // IsDead indicates if the tool is dead/destroyed
	Cassette     EntityID `json:"cassette"`      // Cassette indicates the cassette ID this tool belongs to (if any)
	MinThickness float32  `json:"min_thickness"`
	MaxThickness float32  `json:"max_thickness"`
}

func (t *Tool) IsCassette() bool {
	return t.Position == SlotUpperCassette
}

func (t *Tool) IsTool() bool {
	return !t.IsCassette()
}

func (t *Tool) IsUpperTool() bool {
	return t.Position == SlotUpper
}

func (t *Tool) IsLowerTool() bool {
	return t.Position == SlotLower
}

func (t *Tool) German() string {
	if t.IsCassette() {
		if t.Code == "" {
			return fmt.Sprintf("%dx%d %s (%.1f-%.1fmm)", t.Width, t.Height, t.Type, t.MinThickness, t.MaxThickness)
		}
		return fmt.Sprintf("%dx%d %s %s (%.1f-%.1fmm)", t.Width, t.Height, t.Type, t.Code, t.MinThickness, t.MaxThickness)
	}
	return fmt.Sprintf("%dx%d %s %s", t.Width, t.Height, t.Type, t.Code)
}

// Clone creates a copy of the tool
func (t *Tool) Clone() *Tool {
	return &Tool{
		ID:           t.ID,
		Width:        t.Width,
		Height:       t.Height,
		Position:     t.Position,
		Type:         t.Type,
		Code:         t.Code,
		CyclesOffset: t.CyclesOffset,
		Cycles:       t.Cycles,
		IsDead:       t.IsDead,
		Cassette:     t.Cassette,
		MinThickness: t.MinThickness,
		MaxThickness: t.MaxThickness,
	}
}

func (t *Tool) String() string {
	return fmt.Sprintf(
		"Tool{ID:%d, Width:%d, Height:%d, Position:%d, Type:%s, Code:%s, "+
			"CyclesOffset:%d, Cycles:%d, IsDead:%t, Cassette:%d, MinThickness:%.1f, MaxThickness:%.1f}",
		t.ID,
		t.Width,
		t.Height,
		t.Position,
		t.Type,
		t.Code,
		t.CyclesOffset,
		t.Cycles,
		t.IsDead,
		t.Cassette,
		t.MinThickness,
		t.MaxThickness,
	)
}

// Validate checks if the tool data is valid
func (t *Tool) Validate() *errors.ValidationError {
	// Validate position to be upper/lower or cassette
	switch t.Position {
	case SlotUpper, SlotLower, SlotUpperCassette:
	default:
		return errors.NewValidationError("invalid position: %d", t.Position)
	}

	// Width and Height must be positive, zero is allowed for reason of special (placeholder) tools
	if t.Width < 0 {
		return errors.NewValidationError("width must be positive: %d", t.Width)
	}
	if t.Height < 0 {
		return errors.NewValidationError("height must be positive: %d", t.Height)
	}

	// Type and Code must be set
	if t.Type == "" {
		return errors.NewValidationError("type is required")
	}
	if !t.IsCassette() && t.Code == "" {
		return errors.NewValidationError("code is required for all tools not being cassettes")
	}

	// For cassettes, MinThickness must be less than MaxThickness
	if t.IsCassette() {
		if t.MinThickness <= 0 {
			return errors.NewValidationError("min thickness must be positive: %.1f", t.MinThickness)
		}
		if t.MaxThickness <= 0 {
			return errors.NewValidationError("max thickness must be positive: %.1f", t.MaxThickness)
		}
		if t.MinThickness >= t.MaxThickness {
			return errors.NewValidationError("min thickness %.1f must be less than max thickness %.1f", t.MinThickness, t.MaxThickness)
		}
	}

	return nil
}
