package routes

import (
	"io/fs"
)

type Options struct {
	Templates fs.FS
	Global    Global
}

func (o *Options) MetalSheets() MetalSheets {
	g := o.Global
	g.SubTitle = "Metal Sheets"

	return MetalSheets{
		Global: g,
	}
}
