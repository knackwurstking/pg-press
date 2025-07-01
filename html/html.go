package html

import (
	"embed"
	"html/template"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/knackwurstking/pg-vis/pkg/pgvis"
	"github.com/labstack/echo/v4"
)

//go:embed routes
var routes embed.FS

//go:embed static
var static embed.FS

type Options struct {
	ServerPathPrefix string
	DB               *pgvis.DB
}

func Serve(e *echo.Echo, options Options) {
	e.StaticFS(options.ServerPathPrefix+"/", echo.MustSubFS(static, "static"))

	e.GET(options.ServerPathPrefix+"/", func(c echo.Context) error {
		t, err := template.ParseFS(routes,
			"routes/layout.html",
			"routes/page.html",
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
		v, err := c.FormParams()
		apiKey := v.Get("api-key")

		if apiKey != "" {
			log.Debugf("Form: Api Key: %#v", apiKey)

			// TODO: Check api key and get user

			// TODO: If user exists for this api key, redirect to start page "/"
		}

		t, err := template.ParseFS(routes,
			"routes/layout.html",
			"routes/signup/page.html",
		)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		err = t.Execute(c.Response(), SignUpData{
			ApiKey: apiKey,
		})
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		return nil
	})

	e.GET(options.ServerPathPrefix+"/feed", func(c echo.Context) error {
		t, err := template.ParseFS(routes,
			"routes/layout.html",
			"routes/feed/page.html",
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

type SignUpData struct {
	ApiKey string
}
