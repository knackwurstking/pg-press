package models

import (
	"fmt"
	"sort"
)

const (
	StatusActive       = Status("active")
	StatusAvailable    = Status("available")
	StatusRegenerating = Status("regenerating")
	StatusDead         = Status("dead")
)

type (
	ToolID int64
	Status string
)

// Tool represents a tool in the database.
// Max cycles: 800.000 (Orange) -> 1.000.000 (Red)
type Tool struct {
	ID           ToolID       `json:"id"`
	Position     Position     `json:"position"`
	Format       Format       `json:"format"`
	Type         string       `json:"type"` // Ex: FC, GTC, MASS
	Code         string       `json:"code"` // Ex: G01, G02, ...
	Regenerating bool         `json:"regenerating"`
	IsDead       bool         `json:"is_dead"`
	Press        *PressNumber `json:"press"` // Press number (0-5) when status is active
	Binding      *ToolID      `json:"binding"`
}

func NewTool(position Position, format Format, code string, _type string) *Tool {
	return &Tool{
		Format:       format,
		Position:     position,
		Type:         _type,
		Code:         code,
		Regenerating: false,
		Press:        nil,
	}
}

func (t *Tool) Validate() bool {
	if !IsValidPosition(&t.Position) {
		return false
	}

	if t.Code == "" {
		return false
	}

	return true
}

func (t *Tool) Status() Status {
	if t.IsDead {
		return StatusDead
	}

	if t.Regenerating {
		return StatusRegenerating
	}

	if t.Press != nil {
		return StatusActive
	}

	return StatusAvailable
}

func (t *Tool) String() string {
	if t.Type != "" {
		return fmt.Sprintf("%s %s %s", t.Format, t.Code, t.Type)
	}

	return fmt.Sprintf("%s %s", t.Format, t.Code)
}

// SetPress sets the press for the tool with validation (0-5)
func (t *Tool) SetPress(pressNumber *PressNumber) error {
	if pressNumber == nil {
		t.Press = nil
		return nil
	}

	if !IsValidPressNumber(pressNumber) {
		return fmt.Errorf("invalid press number")
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

func (t *Tool) IsBound() bool {
	return t.Binding != nil
}

func (t *Tool) IsBindable() bool {
	return t.Position == PositionTop || t.Position == PositionTopCassette
}

// SortToolsByPosition sorts tools by position: top, top cassette, bottom
func SortToolsByPosition(tools []*Tool) []*Tool {
	sorted := make([]*Tool, len(tools))
	copy(sorted, tools)

	sort.Slice(sorted, func(i, j int) bool {
		return GetPositionOrder(sorted[i].Position) < GetPositionOrder(sorted[j].Position)
	})

	return sorted
}
