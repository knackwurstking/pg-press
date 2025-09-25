package models

import (
	"fmt"
	"slices"

	"github.com/knackwurstking/pgpress/pkg/utils"
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
	Status      string
	Position    string
	PressNumber int8
)

func (p Position) String() string {
	switch p {
	case PositionTop:
		return "Oberteil"
	case PositionTopCassette:
		return "Oberteil Kassette"
	case PositionBottom:
		return "Unterteil"
	default:
		return "unknown"
	}
}

// IsValid checks if the press number is within the valid range (0-5)
func IsValidPressNumber(n *PressNumber) bool {
	if n == nil {
		return false
	}

	return slices.Contains([]PressNumber{0, 2, 3, 4, 5}, *n)
}

type Format struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

func (f Format) String() string {
	return fmt.Sprintf("%dx%d", f.Width, f.Height)
}

// Tool represents a tool in the database.
// Max cycles: 800.000 (Orange) -> 1.000.000 (Red)
type Tool struct {
	ID           int64        `json:"id"`
	Position     Position     `json:"position"`
	Format       Format       `json:"format"`
	Type         string       `json:"type"` // Ex: FC, GTC, MASS
	Code         string       `json:"code"` // Ex: G01, G02, ...
	Regenerating bool         `json:"regenerating"`
	Press        *PressNumber `json:"press"` // Press number (0-5) when status is active
	LinkedNotes  []int64      `json:"notes"` // Contains note ids from the "notes" table
}

func NewTool(position Position, format Format, code string, _type string) *Tool {
	return &Tool{
		Format:       format,
		Position:     position,
		Type:         _type,
		Code:         code,
		Regenerating: false,
		Press:        nil,
		LinkedNotes:  make([]int64, 0),
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
	return fmt.Sprintf("%s %s", t.Format, t.Code)
}

// SetPress sets the press for the tool with validation (0-5)
func (t *Tool) SetPress(pressNumber *PressNumber) error {
	if pressNumber == nil {
		t.Press = nil
		return nil
	}

	if !IsValidPressNumber(pressNumber) {
		return utils.NewValidationError("invalid press number")
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

// ToolWithNotes represents a tool with its related notes loaded.
type ToolWithNotes struct {
	*Tool
	LoadedNotes []*Note `json:"loaded_notes"`
}
