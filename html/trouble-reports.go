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
		t, err := template.ParseFS(templates,
			"templates/layout.html",
			"templates/trouble-reports.html",
		)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		if err = t.Execute(c.Response(), nil); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		return nil
	})

	// HTMX: Dialog Edit

	e.GET(options.ServerPathPrefix+"/trouble-reports/dialog-edit", func(c echo.Context) error {
		return trDialogEditGET(c, options.DB, nil)
	})

	// FormValues:
	//   - title: string
	//   - content: multiline-string
	e.POST(options.ServerPathPrefix+"/trouble-reports/dialog-edit", func(c echo.Context) error {
		return trDialogEditPOST(c, options.DB)
	})

	// QueryParam:
	//   - id: int
	//
	// FormValue:
	//   - title: string
	//   - content: multiline-string
	e.PUT(options.ServerPathPrefix+"/trouble-reports/dialog-edit", func(c echo.Context) error {
		return trDialogEditPUT(c, options.DB)
	})

	// HTMX: Data

	e.GET(options.ServerPathPrefix+"/trouble-reports/data", func(c echo.Context) error {
		return trDataGET(c, options.DB)
	})

	e.DELETE(options.ServerPathPrefix+"/trouble-reports/data", func(c echo.Context) error {
		return trDataDELETE(c, options.DB)
	})
}

type TRDialogEdit struct {
	ID                int
	Submitted         bool // Submitted set to true will close the dialog
	Title             string
	Content           string
	LinkedAttachments []*pgvis.Attachment
	InvalidTitle      bool
	InvalidContent    bool
}

// trDialogEditGET
//
// QueryParam:
//
//	cancel: "true"
//	id: int
func trDialogEditGET(c echo.Context, db *pgvis.DB, pageData *TRDialogEdit) *echo.HTTPError {
	if cancel := c.QueryParam("cancel"); cancel == "true" {
		pageData.Submitted = true
	}

	if pageData == nil {
		pageData = &TRDialogEdit{Submitted: false}
	}

	if !pageData.Submitted {
		if id, err := strconv.Atoi(c.QueryParam("id")); err == nil {
			if id > 0 {
				log.Debugf("Get database entry with id %d", id)

				pageData.ID = id

				tr, err := db.TroubleReports.Get(int64(id))
				if err != nil {
					return echo.NewHTTPError(
						http.StatusBadRequest,
						fmt.Errorf("no data: %s", err.Error()),
					)
				}
				pageData.Title = tr.Title
				pageData.Content = tr.Content
				pageData.LinkedAttachments = tr.LinkedAttachments
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

	if err = t.Execute(c.Response(), pageData); err != nil {
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			fmt.Errorf("template executing failed: %s", err.Error()),
		)
	}

	return nil
}

func trDialogEditPOST(c echo.Context, db *pgvis.DB) *echo.HTTPError {
	pageData := &TRDialogEdit{Submitted: true}

	user, ok := c.Get("user").(*pgvis.User)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "cannot get the user from the echos context")
	}

	title, content, err := trGetTitleAndContent(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "query data: %s", err.Error())
	}

	pageData.InvalidTitle = title == ""
	pageData.InvalidContent = content == ""

	if pageData.InvalidTitle || pageData.InvalidContent {
		log.Debugf("Add new database entry: title=%#v; content=%#v", title, content)

		modified := pgvis.NewModified[*pgvis.TroubleReport](user, nil)
		tr := pgvis.NewTroubleReport(modified, title, content)

		if err = db.TroubleReports.Add(tr); err != nil {
			return echo.NewHTTPError(
				http.StatusInternalServerError,
				fmt.Errorf("add data: %s", err.Error()),
			)
		}
	}

	return trDialogEditGET(c, db, pageData)
}

func trDialogEditPUT(c echo.Context, db *pgvis.DB) *echo.HTTPError {
	id, err := strconv.Atoi(c.QueryParam("id"))
	if err != nil || id <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("invalid or missing id"))
	}

	user, ok := c.Get("user").(*pgvis.User)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "cannot get the user from the echos context")
	}

	title, content, err := trGetTitleAndContent(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "query data: %s", err.Error())
	}

	if title == "" || content == "" {
		log.Debugf("Invalid input: title=%#v; content=%#v", title, content)

		// TODO: Invalid Input, set inputs to invalid and continue
		return echo.NewHTTPError(
			http.StatusBadRequest,
			fmt.Errorf("invalid input"),
		) // NOTE: Just a placeholder
	}

	log.Debugf("Update database entry with id %d: title=%#v; content=%#v", id, title, content)

	modified := pgvis.NewModified[*pgvis.TroubleReport](user, nil)
	trNew := pgvis.NewTroubleReport(modified, title, content)

	// Need to pass to old data to the modifieds origianal field
	if trOld, err := db.TroubleReports.Get(int64(id)); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	} else {
		trNew.Modified.Original = trOld
	}

	if err = db.TroubleReports.Update(int64(id), trNew); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return trDialogEditGET(c, db, &TRDialogEdit{Submitted: true})
}

func trDataGET(c echo.Context, db *pgvis.DB) *echo.HTTPError {
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

	if err = t.Execute(c.Response(), trs); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return nil
}

func trDataDELETE(c echo.Context, db *pgvis.DB) *echo.HTTPError {
	id, err := strconv.Atoi(c.QueryParam("id"))
	if err != nil {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			fmt.Errorf("query param \"id\": %s", err.Error()),
		)
	}
	if id <= 0 {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			fmt.Errorf("invalid \"id\": cannot be 0 or lower"),
		)
	}

	if err := db.TroubleReports.Remove(int64(id)); err != nil {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			fmt.Errorf("invalid \"id\" %d: not found", id),
		)
	}

	return trDataGET(c, db)
}

func trGetTitleAndContent(ctx echo.Context) (title, content string, err error) {
	title, err = url.QueryUnescape(ctx.FormValue("title"))
	if err != nil {
		return title, content, echo.NewHTTPError(
			http.StatusBadRequest,
			fmt.Errorf("query unescape \"title\" failed: %s", err.Error()),
		)
	}
	title = strings.Trim(title, " \n\r\t")

	content, err = url.QueryUnescape(ctx.FormValue("content"))
	if err != nil {
		return title, content, echo.NewHTTPError(
			http.StatusBadRequest,
			fmt.Errorf("query unescape \"content\" failed: %s", err.Error()),
		)
	}
	content = strings.Trim(content, " \n\r\t")

	return title, content, nil
}
