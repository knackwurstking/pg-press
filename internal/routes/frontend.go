package routes

import (
	"log/slog"

	"github.com/labstack/echo/v4"
)

func RegisterPages(e *echo.Echo, o *Options) {
	e.GET(o.Data.ServerPathPrefix+"/", func(c echo.Context) error {
		err := serveTemplate(
			c, o.Templates, o.Data,
			"main.go.html",
			"layout/base.go.html",
			"content/home.go.html",
		)
		if err != nil {
			slog.Error(err.Error())
		}
		return err
	})
}
