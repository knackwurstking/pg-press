package html

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/pgvis"
)

func ServeTroubleReports(e *echo.Echo, options Options) {
	e.GET(options.ServerPathPrefix+"/trouble-reports", func(c echo.Context) error {
		return handleTroubleReportsPage(c, options.DB)
	})

	e.GET(options.ServerPathPrefix+"/trouble-reports/dialog-edit", func(c echo.Context) error {
		return handleTroubleReportsDialogEditGET(false, c, options.DB)
	})

	e.POST(options.ServerPathPrefix+"/trouble-reports/dialog-edit", func(c echo.Context) error {
		return handleTroubleReportsDialogEditPOST(c, options.DB)
	})

	e.GET(options.ServerPathPrefix+"/trouble-reports/data", func(c echo.Context) error {
		return handleTroubleReportsDataGET(c, options.DB)
	})
}

func handleTroubleReportsPage(ctx echo.Context, db *pgvis.DB) *echo.HTTPError {
	t, err := template.ParseFS(templates,
		"templates/layout.html",
		"templates/trouble-reports.html",
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err = t.Execute(ctx.Response(), nil); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return nil
}

type TroubleReportsDialogEditTemplateData struct {
	ID        int
	Submitted bool              // Submitted set to true will close the dialog
	Inputs    map[string]string // Inputs containing all data to render
}

// handleTroubleReportsDialogEditGET
//
// QueryParam:
//
//	id: int
func handleTroubleReportsDialogEditGET(submitted bool, ctx echo.Context, db *pgvis.DB) *echo.HTTPError {
	data := TroubleReportsDialogEditTemplateData{
		Submitted: submitted,
		Inputs:    map[string]string{},
	}

	if !data.Submitted {
		if id, err := strconv.Atoi(ctx.QueryParam("id")); err == nil {
			if id > 0 {
				log.Debugf("Get database entry with id %d", id)
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

// handleTroubleReportsDialogEditPOST will add or update data
//
// QueryParam:
//   - cancel: "true"
//   - id: int
//
// FormValue:
//   - title: string
//   - content: multiline-string
func handleTroubleReportsDialogEditPOST(ctx echo.Context, db *pgvis.DB) *echo.HTTPError {
	if cancel := ctx.QueryParam("cancel"); cancel == "true" {
		return handleTroubleReportsDialogEditGET(true, ctx, db)
	}

	title, err := url.QueryUnescape(ctx.FormValue("title"))
	if err != nil {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			fmt.Errorf("query unescape \"title\" failed: %s", err.Error()),
		)
	}
	title = strings.Trim(title, " \n\r\t")

	content, err := url.QueryUnescape(ctx.FormValue("content"))
	if err != nil {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			fmt.Errorf("query unescape \"content\" failed: %s", err.Error()),
		)
	}
	content = strings.Trim(content, " \n\r\t")

	user, ok := ctx.Get("user").(*pgvis.User)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "cannot get the user from the echos context")
	}

	modified := pgvis.NewModified[*pgvis.TroubleReport](user, nil)
	tr := pgvis.NewTroubleReport(modified, title, content)

	if title == "" || content == "" {
		log.Debugf("Invalid input: title=%#v; content=%#v", title, content)

		// TODO: Invalid Input, set inputs to invalid and continue
		return echo.NewHTTPError(
			http.StatusBadRequest,
			fmt.Errorf("invalid input"),
		) // NOTE: Just a placeholder
	} else {
		if id, err := strconv.Atoi(ctx.QueryParam("id")); err != nil || id <= 0 {
			log.Debugf("Add new database entry: title=%#v; content=%#v", title, content)

			if err = db.TroubleReports.Add(tr); err != nil {
				return echo.NewHTTPError(
					http.StatusInternalServerError,
					fmt.Errorf("add data: %s", err.Error()),
				)
			}
		} else {
			log.Debugf("Update database entry with id %d: title=%#v; content=%#v", id, title, content)
			// TODO: Get old data from the database before write the new one, add this to the modified.DataBefore
			//tr.Modified.DataBefore

			// TODO: Update data with ID in the database (existing data)
		}
	}

	return handleTroubleReportsDialogEditGET(true, ctx, db)
}

func handleTroubleReportsDataGET(ctx echo.Context, db *pgvis.DB) *echo.HTTPError {
	trs, err := db.TroubleReports.List()
	if err != nil {
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			fmt.Errorf("list trouble-reports: %s", err.Error()),
		)
	}

	t, err := template.ParseFS(templates, "templates/trouble-reports/data.html")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err = t.Execute(ctx.Response(), trs); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return nil
}
