package routes

import (
	"io/fs"
)

type Global struct {
	Title            string
	SubTitle         string
	Version          string
	ServerPathPrefix string
}

func (g Global) HasSubTitle() bool {
	return g.SubTitle != ""
}

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
