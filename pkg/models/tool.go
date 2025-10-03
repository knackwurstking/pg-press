package models

import (
	"fmt"
	"slices"
	"sort"
	"strings"

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

func (p Position) GermanString() string {
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
	positionCount := make(map[Position][]int64)

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
