package html

import (
	"html/template"
	"net/http"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/labstack/echo/v4"
)

type TroubleReportsPageData struct {
	PageData

	Reports []*pgvis.TroubleReport
}

func NewTroubleReportsPageData() TroubleReportsPageData {
	return TroubleReportsPageData{
		PageData: NewPageData(),
		Reports: make([]*pgvis.TroubleReport, 0),
	}
}

func handleTroubleReports(ctx echo.Context) *echo.HTTPError {
	pageData := NewTroubleReportsPageData()

	// TODO: Get data from the database here

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
