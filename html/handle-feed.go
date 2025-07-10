package html

import (
	"html/template"
	"net/http"

	"github.com/labstack/echo/v4"
)

func handleFeed(c echo.Context) error {
	pageData := PageData{}

	t, err := template.ParseFS(routes,
		"routes/layout.html",
		"routes/feed.html",
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
