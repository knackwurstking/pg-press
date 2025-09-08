package tool

import (
	"fmt"

	"github.com/knackwurstking/pgpress/internal/database/dberror"
	"github.com/knackwurstking/pgpress/internal/database/models/mod"
	"github.com/knackwurstking/pgpress/internal/database/models/note"
	pressmodels "github.com/knackwurstking/pgpress/internal/database/models/press"
)

const (
	PositionTop         = Position("top")
	PositionTopCassette = Position("cassette top")
	PositionBottom      = Position("bottom")

	StatusActive       = Status("active")
	StatusAvailable    = Status("available")
	StatusRegenerating = Status("regenerating")

	ToolCycleWarning int64 = 800000  // Orange
	ToolCycleError   int64 = 1000000 // Red
)

type (
	Status   string
	Position string
)

type Format struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

func (f Format) String() string {
	return fmt.Sprintf("%dx%d", f.Width, f.Height)
}

type ToolMod struct {
	Position     Position                 `json:"position"`
	Format       Format                   `json:"format"`
	Type         string                   `json:"type"`
	Code         string                   `json:"code"`
	Regenerating bool                     `json:"regenerating"`
	Press        *pressmodels.PressNumber `json:"press"`
	LinkedNotes  []int64                  `json:"notes"`
}

// Tool represents a tool in the database.
// Max cycles: 800.000 (Orange) -> 1.000.000 (Red)
type Tool struct {
	ID           int64                    `json:"id"`
	Position     Position                 `json:"position"`
	Format       Format                   `json:"format"`
	Type         string                   `json:"type"` // Ex: FC, GTC, MASS
	Code         string                   `json:"code"` // Ex: G01, G02, ...
	Regenerating bool                     `json:"regenerating"`
	Press        *pressmodels.PressNumber `json:"press"` // Press number (0-5) when status is active
	LinkedNotes  []int64                  `json:"notes"` // Contains note ids from the "notes" table
	Mods         mod.Mods[ToolMod]        `json:"mods"`
}

func New(position Position) *Tool {
	return &Tool{
		Format:       Format{},
		Position:     position,
		Type:         "",
		Code:         "",
		Regenerating: false,
		Press:        nil,
		LinkedNotes:  make([]int64, 0),
		Mods:         mod.NewMods[ToolMod](),
	}
}

func (t *Tool) Status() Status {
	if t.Regenerating {
		return StatusRegenerating
	}
	if t.Press != nil {
		return StatusActive
	}
	return StatusAvailable
}

func (t *Tool) String() string {
	var base string
	switch t.Position {
	case PositionTop:
		base = fmt.Sprintf("%s %s (%s, Oberteil)", t.Format, t.Code, t.Type)
	case PositionTopCassette:
		base = fmt.Sprintf("%s %s (%s, Kassette Oberteil)", t.Format, t.Code, t.Type)
	case PositionBottom:
		base = fmt.Sprintf("%s %s (%s, Unterteil)", t.Format, t.Code, t.Type)
	default:
		base = fmt.Sprintf("%s %s (%s)", t.Format, t.Code, t.Type)
	}

	// Add press information if tool is active
	if t.Status() == StatusActive && t.Press != nil {
		base = fmt.Sprintf("%s - Presse %d", base, *t.Press)
	}

	return base
}

// SetPress sets the press for the tool with validation (0-5)
func (t *Tool) SetPress(pressNumber *pressmodels.PressNumber) error {
	if pressNumber == nil {
		t.Press = nil
		return nil
	}

	if !pressmodels.IsValidPressNumber(pressNumber) {
		return dberror.NewValidationError("press", "invalid press number", pressNumber)
	}

	t.Press = pressNumber

	return nil
}

// ClearPress removes the press assignment
func (t *Tool) ClearPress() {
	t.Press = nil
}

// IsActive checks if the tool is active on a press
func (t *Tool) IsActive() bool {
	return t.Status() == StatusActive && t.Press != nil
}

// GetPressString returns a formatted string of the press assignment
func (t *Tool) GetPress() pressmodels.PressNumber {
	if t.Press == nil {
		return -1
	}
	return *t.Press
}

// GetPressString returns a formatted string of the press assignment
func (t *Tool) GetPressString() string {
	if t.Press == nil {
		return "Nicht zugewiesen"
	}
	return fmt.Sprintf("Presse %d", *t.Press)
}

// ToolWithNotes represents a tool with its related notes loaded.
type ToolWithNotes struct {
	*Tool
	LoadedNotes []*note.Note `json:"loaded_notes"`
}
