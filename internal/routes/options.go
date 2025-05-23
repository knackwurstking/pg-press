package routes

import (
	"io/fs"
)

type Global struct {
	Title            string
	SubTitle         string
	Version          string
	ServerPathPrefix string
	WSConnect        string
}

func (g Global) Array(v ...any) []any {
	return v
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
	g.WSConnect = "/htmx/metal-sheets"

	return MetalSheets{
		Global:      g,
		TableSearch: tableSearch,
	}
}
