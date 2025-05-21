package routes

import (
	"io/fs"
)

type Options struct {
	Global
	Templates fs.FS
}

func (o *Options) MetalSheets(tableSearch string) MetalSheets {
	g := o.Global
	g.SubTitle = "Metal Sheets"

	return MetalSheets{
		Global:      g,
		TableSearch: tableSearch,
	}
}
