package routes

import "regexp"

type MetalSheets struct {
	Global
	TableSearch string
}

func (msp MetalSheets) SearchDataList() []string {
	// TODO: Get data from database

	return []string{
		"120x60 G06",
	} // NOTE: Data for testing
}

func (msp MetalSheets) Tables() []MetalSheetTable {
	// TODO: Get data from database

	tables := []MetalSheetTable{
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
	}

	if msp.TableSearch == "" {
		return tables
	}

	filtered := []MetalSheetTable{}

	r := regexp.MustCompile(msp.TableSearch)
	for _, t := range tables {
		if r.MatchString(t.DataSearch) {
			filtered = append(filtered, t)
		}
	}

	return filtered
}
