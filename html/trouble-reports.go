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

	e.GET(options.ServerPathPrefix+"/trouble-reports/dialog-edit", func(c echo.Context) error {
		return handleTroubleReportsDialogEditGET(c, options)
	})

	e.POST(options.ServerPathPrefix+"/trouble-reports/dialog-edit", func(c echo.Context) error {
		return handleTroubleReportsDialogEditPOST(c, options)
	})
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

type TroubleReportsDialogEditTemplateData struct {
	Submitted   bool     // Submitted set to true will close the dialog
	AriaInvalid []string // AriaInvalid containing input names for all invalid input elements
}

func handleTroubleReportsDialogEditGET(ctx echo.Context, options Options) *echo.HTTPError {
	// TODO: Send the dialog, taking query vars: "id" (optional), if not set, a new entry will be added to the database on submit

	return nil
}

func handleTroubleReportsDialogEditPOST(ctx echo.Context, options Options) *echo.HTTPError {
	// TODO: Add new database entry and close dialog on success (just replace with span again) Reading form variables: "title", "content"

	return nil
}
