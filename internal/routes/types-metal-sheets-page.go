package routes

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
		{ // TODO: Add some data for testing here
			Head: []string{
				"Stärke",
				"Marke (Höhe)",
				"Blech Stempel",
				"Bleck Marke",
				"Stf. P5",
				"Stf. P2-4",
				"Stf. P0",
			},
			Body: [][]any{
				{
					TableCell[string]{Type: "float", Value: "6"},
					TableCell[string]{Type: "int", Value: "50"},
					TableCell[string]{Type: "string", Value: "1-2"},
					TableCell[string]{Type: "string", Value: "8+5"},
					TableCell[SacmiThickness]{Type: "sacmi", Value: SacmiThickness{}},
					TableCell[string]{Type: "string", Value: "5.0"},
					TableCell[SacmiThickness]{Type: "sacmi", Value: SacmiThickness{}}, // TODO: Use a special type here (Min/Max)
				},
			},
			HiddenCells: []int{4, 7},
		},
	} // NOTE: Data for testing
}

type Table struct {
	Head        []string
	Body        [][]any
	HiddenCells []int
}

type TableCell[T string | SacmiThickness] struct {
	Type  string
	Value T
}

// SacmiThickness is used as value for a TableCell
type SacmiThickness struct {
	Min float64
	Max float64
}
