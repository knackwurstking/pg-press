package html

import (
	"html/template"
	"net/http"

	"github.com/labstack/echo/v4"
)

func ServeFeed(e *echo.Echo, options Options) {
	e.GET(options.ServerPathPrefix+"/feed", func(c echo.Context) error {
		return handleFeed(c)
	})
}

func handleFeed(c echo.Context) error {
	pageData := PageData{}

	t, err := template.ParseFS(templates,
		"templates/layout.html",
		"templates/feed.html",
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
