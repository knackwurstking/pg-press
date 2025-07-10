package html

import (
	"html/template"
	"net/http"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/labstack/echo/v4"
)

func ServeTroubleReports(e *echo.Echo, options Options) {
	e.GET(options.ServerPathPrefix+"/trouble-reports", func(c echo.Context) error {
		return handleTroubleReportsPage(c)
	})

	// TODO: Add "GET /trouble-reports/dialog-edit" (dialog id == "dialogEdit")

	// TODO: Add "POST /trouble-reports/dialog-edit", Form data containing a new trouble report
}

type TroubleReportsPageData struct {
	PageData

	Reports []*pgvis.TroubleReport
}

func NewTroubleReportsPageData() TroubleReportsPageData {
	return TroubleReportsPageData{
		PageData: NewPageData(),
		Reports:  make([]*pgvis.TroubleReport, 0),
	}
}

func handleTroubleReportsPage(ctx echo.Context) *echo.HTTPError {
	pageData := NewTroubleReportsPageData()

	// TODO: Get data from the database here

	t, err := template.ParseFS(templates,
		"templates/layout.html",
		"templates/trouble-reports.html",
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err = t.Execute(ctx.Response(), pageData); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return nil
}
