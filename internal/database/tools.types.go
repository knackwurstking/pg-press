package database

import "fmt"

const (
	PositionTop         = Position("top")
	PositionTopCassette = Position("top cassette")
	PositionBottom      = Position("bottom")

	ToolStatusActive       = ToolStatus("active")
	ToolStatusAvailable    = ToolStatus("available")
	ToolStatusRegenerating = ToolStatus("regenerating")
)

type ToolStatus string

type Position string

type ToolFormat struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

func (tf ToolFormat) String() string {
	return fmt.Sprintf("%dx%d", tf.Width, tf.Height)
}

// Tool represents a tool in the database.
//
// TODO: Max cycles: 800.000 (Orange) -> 1.000.000 (Red)
type Tool struct {
	ID           int64         `json:"id"`
	Position     Position      `json:"position"`
	Format       ToolFormat    `json:"format"`
	Type         string        `json:"type"` // Ex: FC, GTC, MASS
	Code         string        `json:"code"` // Ex: G01, G02, ...
	Regenerating bool          `json:"regenerating"`
	Press        *PressNumber  `json:"press"` // Press number (0-5) when status is active
	LinkedNotes  []int64       `json:"notes"` // Contains note ids from the "notes" table
	Mods         Mods[ToolMod] `json:"mods"`
}

func NewTool(position Position) *Tool {
	return &Tool{
		Format:       ToolFormat{},
		Position:     position,
		Type:         "",
		Code:         "",
		Regenerating: false,
		Press:        nil,
		LinkedNotes:  make([]int64, 0),
		Mods:         make([]*Modified[ToolMod], 0),
	}
}

func (t *Tool) Status() ToolStatus {
	if t.Regenerating {
		return ToolStatusRegenerating
	}
	if t.Press != nil {
		return ToolStatusActive
	}
	return ToolStatusAvailable
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
	if t.Status() == ToolStatusActive && t.Press != nil {
		base = fmt.Sprintf("%s - Presse %d", base, *t.Press)
	}

	return base
}

// SetPress sets the press for the tool with validation (0-5)
func (t *Tool) SetPress(pressNumber *PressNumber) error {
	if pressNumber == nil {
		t.Press = nil
		return nil
	}

	if !(*pressNumber).IsValid() {
		return NewValidationError("press", "invalid press number", pressNumber)
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
	return t.Status() == ToolStatusActive && t.Press != nil
}

// GetPressString returns a formatted string of the press assignment
func (t *Tool) GetPress() PressNumber {
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

type ToolMod struct {
	Position     Position     `json:"position"`
	Format       ToolFormat   `json:"format"`
	Type         string       `json:"type"`
	Code         string       `json:"code"`
	Regenerating bool         `json:"regenerating"`
	Press        *PressNumber `json:"press"`
	LinkedNotes  []int64      `json:"notes"`
}
