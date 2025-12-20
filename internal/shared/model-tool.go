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
// Slot type
// -----------------------------------------------------------------------------

type Slot int

const (
	SlotUnknown Slot = iota
	SlotUpper
	SlotLower
	SlotUpperCassette
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
	Cycles       int64    `json:"cycles"`        // Cycles indicates how many cycles this tool has done
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
			return fmt.Sprintf("%dx%d %s (%.1f-%.1f mm)", t.Width, t.Height, t.Type, t.MinThickness, t.MaxThickness)
		}
		return fmt.Sprintf("%dx%d %s %s (%.1f-%.1f mm)", t.Width, t.Height, t.Type, T.Code, t.MinThickness, t.MaxThickness)
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
		"Tool{ID:%d, Width:%d, Height:%d, Position:%d, Type:%s, Code:%s, "+,
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

// -----------------------------------------------------------------------------
// Interface compliance checks
// -----------------------------------------------------------------------------

var _ Entity[*Tool] = (*Tool)(nil)
