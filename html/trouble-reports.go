package html

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/charmbracelet/log"
	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/labstack/echo/v4"
)

func ServeTroubleReports(e *echo.Echo, options Options) {
	e.GET(options.ServerPathPrefix+"/trouble-reports", func(c echo.Context) error {
		return handleTroubleReportsPage(c)
	})

	e.GET(options.ServerPathPrefix+"/trouble-reports/dialog-edit", func(c echo.Context) error {
		return handleTroubleReportsDialogEditGET(false, c) // TODO: Pass options
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
	ID        int
	Submitted bool              // Submitted set to true will close the dialog
	Inputs    map[string]string // Inputs containing all data to render
}

func handleTroubleReportsDialogEditGET(submitted bool, ctx echo.Context) *echo.HTTPError {
	data := TroubleReportsDialogEditTemplateData{
		Submitted: submitted,
		Inputs:    map[string]string{},
	}

	if !data.Submitted {
		if id, err := strconv.Atoi(ctx.QueryParam("id")); err == nil {
			if id > 0 {
				// TODO: Get data from the database if ID is bigger 0
				data.ID = id
			}
		}
	}

	t, err := template.ParseFS(templates, "templates/trouble-reports/dialog-edit.html")
	if err != nil {
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			fmt.Errorf("template parsing failed: %s", err.Error()),
		)
	}

	if err = t.Execute(ctx.Response(), data); err != nil {
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			fmt.Errorf("template executing failed: %s", err.Error()),
		)
	}

	return nil
}

func handleTroubleReportsDialogEditPOST(ctx echo.Context, options Options) *echo.HTTPError {
	title, err := url.QueryUnescape(ctx.QueryParam("title"))
	if err != nil {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			fmt.Errorf("query unescape \"title\" failed: %s", err.Error()),
		)
	}

	content, err := url.QueryUnescape(ctx.QueryParam("content"))
	if err != nil {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			fmt.Errorf("query unescape \"content\" failed: %s", err.Error()),
		)
	}

	user, ok := ctx.Get("user").(*pgvis.User)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "cannot get the user from the echos context")
	}

	_ = pgvis.NewTroubleReport(
		&pgvis.Modified[*pgvis.TroubleReport]{
			User:     user,
			Time:     time.Now().UnixMilli(),
			Original: nil,
		},
		title,
		content,
	)

	if id, err := strconv.Atoi(ctx.QueryParam("id")); err != nil || id <= 0 {
		// TODO: Add data `data` to database (new entry)
	} else {
		// TODO: Get old data from the database before write the new one, add this to the modified.DataBefore
		//tr.Modified.DataBefore

		// TODO: Update data with ID in the database (existing data)
	}

	log.Warn(
		"@TODO: Storing title and content in trouble-reports database: title=%#v; content=%#v",
		title, content,
	)

	return handleTroubleReportsDialogEditGET(true, ctx)
}
