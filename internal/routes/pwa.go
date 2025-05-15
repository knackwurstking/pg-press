// Routes:
//
//   - GET: "/service-worker.js"
//   - GET: "/manifest.json"
package routes

import (
	"log/slog"

	"github.com/labstack/echo/v4"
)

func RegisterPWA(e *echo.Echo, o *Options) {
	e.GET(o.Data.ServerPathPrefix+"/service-worker.js", func(c echo.Context) error {
		err := serveTemplate(c, o.Templates, o.Data, "service-worker.js")
		if err != nil {
			slog.Error(err.Error())
		}
		return err
	})

	e.GET(o.Data.ServerPathPrefix+"/manifest.json", func(c echo.Context) error {
		err := serveTemplate(c, o.Templates, o.Data, "manifest.json")
		if err != nil {
			slog.Error(err.Error())
		}
		return err
	})
}
