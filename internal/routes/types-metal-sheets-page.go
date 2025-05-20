package routes

import "slices"

type MetalSheetsPage struct {
	Global
}

func (msp MetalSheetsPage) SearchDataList() []string {
	// TODO: Get data from database

	return []string{
		"120x60 G06",
	} // NOTE: Data for testing
}

func (msp MetalSheetsPage) Tables() []Table {
	// TODO: Get data from database

	return []Table{
		{
			DataSearch: "120x60 G06",
			Head: []string{
				"Stärke",
				"Marke (Höhe)",
				"Blech Stempel",
				"Bleck Marke",
				"Stf. P0",
				"Stf. P2-4",
			},
			Body: [][]any{
				{
					NewTableCell_Float64(6),
					NewTableCell_Int(50),

					NewTableCell_Float64(4),
					NewTableCell_Float64(13),

					NewTableCell_SacmiThickness(SacmiThickness{Current: -1, Max: -1}),
					NewTableCell_Float64(5.0),
				},
			},
			HiddenCells: []int{4},
		},
	} // NOTE: Data for testing
}

type Table struct {
	DataSearch  string
	Head        []string
	Body        [][]any
	HiddenCells []int
}

func (t Table) IsHidden(index int) bool {
	return slices.Contains(t.HiddenCells, index)
}

type TableCell[T string | int | float64 | SacmiThickness] struct {
	Value T

	valueType string
}

func NewTableCell_Int(value int) TableCell[int] {
	return TableCell[int]{
		valueType: "int",
		Value:     value,
	}
}

func NewTableCell_Float64(value float64) TableCell[float64] {
	return TableCell[float64]{
		valueType: "float64",
		Value:     value,
	}
}

func NewTableCell_SacmiThickness(value SacmiThickness) TableCell[SacmiThickness] {
	return TableCell[SacmiThickness]{
		valueType: "SacmiThickness",
		Value:     value,
	}
}

func (tc TableCell[T]) IsInt() bool {
	return tc.valueType == "int"
}

func (tc TableCell[T]) IsFloat64() bool {
	return tc.valueType == "float64"
}

func (tc TableCell[T]) IsSacmiThickness() bool {
	return tc.valueType == "SacmiThickness"
}

// SacmiThickness is used as value for a TableCell
//
// TODO: Maybe add "Min"
type SacmiThickness struct {
	Current float64
	Max     float64
}
