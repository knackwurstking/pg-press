package models

import "slices"

// IsValid checks if the press number is within the valid range (0-5)
func IsValidPressNumber(n *PressNumber) bool {
	if n == nil {
		return false
	}

	return slices.Contains([]PressNumber{0, 2, 3, 4, 5}, *n)
}

func IsValidPosition(p *Position) bool {
	if p == nil {
		return false
	}

	return slices.Contains([]Position{PositionTop, PositionBottom, PositionTopCassette}, *p)
}
