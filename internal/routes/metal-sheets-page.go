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

func (msp MetalSheetsPage) Tables() []MetalSheetTable {
	// TODO: Get data from database

	return []MetalSheetTable{
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
					NewMetalSheetTableCell_Float64(6),
					NewMetalSheetTableCell_Int(50),

					NewMetalSheetTableCell_Float64(4),
					NewMetalSheetTableCell_Float64(13),

					NewMetalSheetTableCell_SacmiThickness(SacmiThickness{Current: -1, Max: -1}),
					NewMetalSheetTableCell_Float64(5.0),
				},
			},
			HiddenCells: []int{4},
		},
	} // NOTE: Data for testing
}
