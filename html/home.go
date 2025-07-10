package html

import (
	"html/template"
	"net/http"

	"github.com/labstack/echo/v4"
)

func ServeHome(e *echo.Echo, options Options) {
	e.GET(options.ServerPathPrefix+"/", func(c echo.Context) error {
		return handleHomePage(c)
	})
}

func handleHomePage(c echo.Context) *echo.HTTPError {
	pageData := PageData{}

	t, err := template.ParseFS(templates,
		"templates/layout.html",
		"templates/home.html",
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	err = t.Execute(c.Response(), pageData)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return nil
}
