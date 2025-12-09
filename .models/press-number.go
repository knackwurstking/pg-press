package models

import "slices"

var (
	ValidPressNumbers = []PressNumber{0, 2, 3, 4, 5}
)

type PressNumber int8

// IsValidPressNumber checks if a given press number is valid.
// A valid press number must be one of: 0, 2, 3, 4, or 5.
// Returns false if the input is nil or not in the valid set.
func IsValidPressNumber(n *PressNumber) bool {
	if n == nil {
		return false
	}

	return slices.Contains(ValidPressNumbers, *n)
}
