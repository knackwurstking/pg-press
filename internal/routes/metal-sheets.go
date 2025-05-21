package routes

import (
	"pg-vis/internal/types"
	"regexp"
)

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

func (msp MetalSheets) Tables() []types.MetalSheetTable {
	// TODO: Get data from database

	tables := []types.MetalSheetTable{
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
					types.NewMetalSheetTableCell_Float64(6),
					types.NewMetalSheetTableCell_Int(50),

					types.NewMetalSheetTableCell_Float64(4),
					types.NewMetalSheetTableCell_Float64(13),

					types.NewMetalSheetTableCell_SacmiThickness(
						types.SacmiThickness{Current: -1, Max: -1},
					),
					types.NewMetalSheetTableCell_Float64(5.0),
				},
			},
			HiddenCells: []int{4},
		},
	}

	if msp.TableSearch == "" {
		return tables
	}

	filtered := []types.MetalSheetTable{}

	r := regexp.MustCompile(msp.TableSearch)
	for _, t := range tables {
		if r.MatchString(t.DataSearch) {
			filtered = append(filtered, t)
		}
	}

	return filtered
}
