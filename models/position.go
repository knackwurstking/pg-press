package models

import (
	"fmt"
	"slices"
	"strings"
)

const (
	PositionTop         = Position("top")
	PositionTopCassette = Position("cassette top")
	PositionBottom      = Position("bottom")
)

type Position string

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

func IsValidPosition(p *Position) bool {
	if p == nil {
		return false
	}

	return slices.Contains([]Position{PositionTop, PositionBottom, PositionTopCassette}, *p)
}

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

// ValidateUniquePositions checks that there's only one tool per position
// Returns an error if duplicates are found
func ValidateUniquePositions(tools []*Tool) bool {
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
		return false
	}

	return true
}
