package models

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/knackwurstking/pg-press/errors"
)

const (
	ToolCycleWarning int64 = 800000  // Orange
	ToolCycleError   int64 = 1000000 // Red

	StatusActive       = Status("active")
	StatusAvailable    = Status("available")
	StatusRegenerating = Status("regenerating")
	StatusDead         = Status("dead")

	PositionTop         = Position("top")
	PositionTopCassette = Position("cassette top")
	PositionBottom      = Position("bottom")
)

type (
	ToolID      int64
	Status      string
	PressNumber int8
	Position    string
)

func (p Position) GermanString() string {
	switch p {
	case PositionTop:
		return "Oberteil"
	case PositionTopCassette:
		return "Kassette"
	case PositionBottom:
		return "Unterteil"
	default:
		return "unknown"
	}
}

type Format struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

func (f Format) String() string {
	if f.Width == 0 && f.Height == 0 {
		return ""
	}

	return fmt.Sprintf("%dx%d", f.Width, f.Height)
}

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

func (t *Tool) Validate() error {
	if !IsValidPosition(&t.Position) {
		return fmt.Errorf("position cannot be empty")
	}

	if t.Code == "" {
		return fmt.Errorf("code cannot be empty")
	}

	return nil
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
		return errors.NewValidationError("invalid press number")
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

// GetPositionOrder returns the sort order for a position (lower number = higher priority)
func GetPositionOrder(position Position) int {
	switch position {
	case PositionTop:
		return 1
	case PositionTopCassette:
		return 2
	case PositionBottom:
		return 3
	default:
		return 999 // Unknown positions go to the end
	}
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

// ValidateUniquePositions checks that there's only one tool per position
// Returns an error if duplicates are found
func ValidateUniquePositions(tools []*Tool) error {
	positionCount := make(map[Position][]ToolID)

	for _, tool := range tools {
		positionCount[tool.Position] = append(positionCount[tool.Position], tool.ID)
	}

	var duplicates []string
	for position, toolIDs := range positionCount {
		if len(toolIDs) > 1 {
			duplicates = append(duplicates, fmt.Sprintf("%s (%s): Tools %v",
				position, position.GermanString(), toolIDs))
		}
	}

	if len(duplicates) > 0 {
		return fmt.Errorf("duplicate tool positions found: %s", strings.Join(duplicates, ", "))
	}

	return nil
}

// OverlappingToolInstance represents one instance of a tool on a specific press
type OverlappingToolInstance struct {
	PressNumber PressNumber `json:"press_number"`
	Position    Position    `json:"position"`
	StartDate   time.Time   `json:"start_date"`
	EndDate     time.Time   `json:"end_date"`
}

// OverlappingTool represents a tool that appears on multiple presses simultaneously
type OverlappingTool struct {
	ToolID    ToolID                     `json:"tool_id"`
	ToolCode  string                     `json:"tool_code"`
	Overlaps  []*OverlappingToolInstance `json:"overlaps"`
	StartDate time.Time                  `json:"start_date"`
	EndDate   time.Time                  `json:"end_date"`
}
