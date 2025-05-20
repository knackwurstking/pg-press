package routes

import (
	"io/fs"
	"strings"
)

type Options struct {
	Templates fs.FS
	Global    Global
}

func (o *Options) MetalSheets(tableSearch string) MetalSheets {
	g := o.Global
	g.SubTitle = "Metal Sheets"

	return MetalSheets{
		Global:      g,
		TableSearch: strings.Trim(tableSearch, " "),
	}
}
