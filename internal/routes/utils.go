package routes

import (
	"html/template"
	"io/fs"

	"github.com/labstack/echo/v4"
)

var (
	componentTemplates = []string{
		"components/metal-sheet-table.go.html",
	}
)

func serveTemplate(c echo.Context, f fs.FS, data any, patterns ...string) error {
	t, err := template.ParseFS(f, append(patterns, componentTemplates...)...)
	if err != nil {
		return err
	}

	resp := c.Response()
	resp.Header().Add("Content-Type", "text/html; charset=utf-8")

	return t.Execute(resp, data)
}
