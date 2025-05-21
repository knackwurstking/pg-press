package types

import "slices"

type MetalSheetTable struct {
	DataSearch  string
	Head        []string
	Body        [][]any
	HiddenCells []int
}

func (t MetalSheetTable) IsHidden(index int) bool {
	return slices.Contains(t.HiddenCells, index)
}
