package routes

import "io/fs"

type Options struct {
	Templates fs.FS
	Global    Global
}

func (o *Options) MetalSheetsPage() MetalSheetsPage {
	g := o.Global
	g.SubTitle = "Metal Sheets"

	return MetalSheetsPage{
		Global: g,
	}
}
