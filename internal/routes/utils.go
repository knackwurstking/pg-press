package routes

import (
	"html/template"
	"io/fs"

	"github.com/labstack/echo/v4"
)

func serveTemplate(c echo.Context, f fs.FS, data any, pattern ...string) error {
	t, err := template.ParseFS(f, pattern...)
	if err != nil {
		return err
	}

	resp := c.Response()
	resp.Header().Add("Content-Type", "text/html; charset=utf-8")

	return t.Execute(resp, data)
}
