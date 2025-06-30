package html

import (
	"embed"
	"html/template"
	"net/http"

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
		t, err := template.ParseFS(pages,
			"pages/layout.html",
			"pages/page.html",
		)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		err = t.Execute(c.Response(), nil)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		return nil
	})

	e.GET(options.ServerPathPrefix+"/signup", func(c echo.Context) error {
		t, err := template.ParseFS(pages,
			"pages/layout.html",
			"pages/signup/page.html",
		)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		err = t.Execute(c.Response(), nil)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		return nil
	})

	e.GET(options.ServerPathPrefix+"/feed", func(c echo.Context) error {
		t, err := template.ParseFS(pages,
			"pages/layout.html",
			"pages/feed/page.html",
		)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		err = t.Execute(c.Response(), nil)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		return nil
	})
}
