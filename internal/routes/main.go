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
		{Value: "120x60 G01", Title: "120x60 G01"},
		{Value: "120x60 G02", Title: "120x60 G01"},
		{Value: "120x60 G03", Title: "120x60 G01"},
	} // NOTE: Data for testing
}

type Option struct {
	Value string
	Title string
}
