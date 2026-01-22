package models

import "slices"

const (
	// PressNumberUnknown is used for tools that can no longer be bound to a valid press.
	PressNumberUnknown PressNumber = -1
)

type PressNumber int8

// IsValidPressNumber checks if a given press number is valid.
// A valid press number must be one of: 0, 2, 3, 4, or 5.
// Returns false if the input is nil or not in the valid set.
func IsValidPressNumber(n *PressNumber) bool {
	if n == nil {
		return false
	}

	return slices.Contains([]PressNumber{PressNumberUnknown, 0, 2, 3, 4, 5}, *n)
}
