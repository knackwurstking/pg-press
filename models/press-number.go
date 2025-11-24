package models

import "slices"

const (
	// PressNumberUnknown is used for tools i can no longer bind to a valid press
	PressNumberUnknown PressNumber = -1
)

type PressNumber int8

func IsValidPressNumber(n *PressNumber) bool {
	if n == nil {
		return false
	}

	return slices.Contains([]PressNumber{0, 2, 3, 4, 5}, *n)
}
