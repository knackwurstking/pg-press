package html

import (
	"embed"

	"github.com/labstack/echo/v4"
)

//go:embed pages
var pages embed.FS

//go:embed static
var static embed.FS

type Options struct {
	ServerPathPrefix string
}

func Serve(e *echo.Echo, options Options) {
	e.StaticFS(options.ServerPathPrefix+"/", echo.MustSubFS(static, "static"))

	e.GET(options.ServerPathPrefix+"/", func(c echo.Context) error {
		// TODO: ...

		return nil
	})

	e.GET(options.ServerPathPrefix+"/signup", func(c echo.Context) error {
		// TODO: ...

		return nil
	})

	e.GET(options.ServerPathPrefix+"/feed", func(c echo.Context) error {
		// TODO: ...

		return nil
	})
}
