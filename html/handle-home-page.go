package html

import (
	"html/template"
	"net/http"

	"github.com/labstack/echo/v4"
)

func handleHomePage(c echo.Context) error {
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
