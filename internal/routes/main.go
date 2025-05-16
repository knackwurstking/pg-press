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

type Global struct {
	Title            string
	SubTitle         string
	Version          string
	ServerPathPrefix string
}

func (g Global) HasSubTitle() bool {
	return g.SubTitle != ""
}

type MetalSheetsPage struct {
	Global
}

func (msp MetalSheetsPage) SearchDataList() []Option {
	// TODO: Generate list from database

	return []Option{
		"120x60 G01",
		"120x60 G01",
		"120x60 G01",
	} // NOTE: Data for testing
}

type Option string
