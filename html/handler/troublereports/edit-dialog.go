package troublereports

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/html/handler"
	"github.com/knackwurstking/pg-vis/pgvis"
)

type EditDialogPageData struct {
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
func GETDialogEdit(templates fs.FS, c echo.Context, db *pgvis.DB, pageData *EditDialogPageData) *echo.HTTPError {
	if pageData == nil {
		pageData = &EditDialogPageData{Submitted: false}
	}

	if cancel := c.QueryParam("cancel"); cancel == "true" {
		pageData.Submitted = true
	}

	if !pageData.Submitted && (!pageData.InvalidTitle && !pageData.InvalidContent) {
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

func POSTDialogEdit(templates fs.FS, c echo.Context, db *pgvis.DB) *echo.HTTPError {
	dialogEditData := &EditDialogPageData{Submitted: true}

	user, herr := handler.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	title, content, err := getTitleAndContent(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "query data: %s", err.Error())
	}

	dialogEditData.Title = title
	dialogEditData.Content = content
	dialogEditData.InvalidTitle = title == ""
	dialogEditData.InvalidContent = content == ""

	if !dialogEditData.InvalidTitle && !dialogEditData.InvalidContent {
		log.Debugf("Add new database entry: title=%#v; content=%#v", title, content)

		modified := pgvis.NewModified[*pgvis.TroubleReport](user, nil)
		tr := pgvis.NewTroubleReport(modified, title, content)

		if err = db.TroubleReports.Add(tr); err != nil {
			return echo.NewHTTPError(
				http.StatusInternalServerError,
				fmt.Errorf("add data: %s", err.Error()),
			)
		}
	} else {
		dialogEditData.Submitted = false
	}

	return GETDialogEdit(templates, c, db, dialogEditData)
}

func PUTDialogEdit(templates fs.FS, c echo.Context, db *pgvis.DB) *echo.HTTPError {
	id, err := strconv.Atoi(c.QueryParam("id"))
	if err != nil || id <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("invalid or missing id"))
	}

	user, herr := handler.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	title, content, err := getTitleAndContent(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "query data: %s", err.Error())
	}

	dialogEditData := &EditDialogPageData{
		Submitted:      true,
		ID:             id,
		Title:          title,
		Content:        content,
		InvalidTitle:   title == "",
		InvalidContent: content == "",
	}

	if !dialogEditData.InvalidTitle && !dialogEditData.InvalidContent {
		log.Debugf("Update database entry with id %d: title=%#v; content=%#v", id, title, content)

		modified := pgvis.NewModified[*pgvis.TroubleReport](user, nil)
		trNew := pgvis.NewTroubleReport(modified, title, content)

		// Need to pass to old data to the modified original field
		if trOld, err := db.TroubleReports.Get(int64(dialogEditData.ID)); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		} else {
			modified.Original = trOld
		}

		if err = db.TroubleReports.Update(int64(dialogEditData.ID), trNew); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	} else {
		dialogEditData.Submitted = false
	}

	return GETDialogEdit(templates, c, db, dialogEditData)
}

func getTitleAndContent(ctx echo.Context) (title, content string, err error) {
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
