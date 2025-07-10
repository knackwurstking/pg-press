package html

import (
	"html/template"
	"net/http"

	"github.com/labstack/echo/v4"
)

func handleTroubleReports(ctx echo.Context) *echo.HTTPError {
	pageData := PageData{}

	t, err := template.ParseFS(routes,
		"routes/layout.html",
		"routes/trouble-reports.html",
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err = t.Execute(ctx.Response(), pageData); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return nil
}
