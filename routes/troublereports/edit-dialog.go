package troublereports

import (
	"html/template"
	"io/fs"
	"net/http"
	"net/url"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/knackwurstking/pg-vis/routes/utils"
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
		// Try to get ID from query parameters (optional)
		if idStr := c.QueryParam("id"); idStr != "" {
			if id, herr := utils.ParseRequiredIDQuery(c, "id"); herr == nil {
				log.Debugf("Get database entry with id %d", id)

				pageData.ID = int(id)

				tr, err := db.TroubleReports.Get(id)
				if err != nil {
					return utils.HandlePgvisError(c, err)
				}
				pageData.Title = tr.Title
				pageData.Content = tr.Content
				pageData.LinkedAttachments = tr.LinkedAttachments
			}
		}
	}

	t, err := template.ParseFS(templates, "templates/trouble-reports/dialog-edit.html")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err = t.Execute(c.Response(), pageData); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return nil
}

func POSTDialogEdit(templates fs.FS, c echo.Context, db *pgvis.DB) *echo.HTTPError {
	dialogEditData := &EditDialogPageData{Submitted: true}

	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	title, content, herr := getTitleAndContent(c)
	if herr != nil {
		return herr
	}

	dialogEditData.Title = title
	dialogEditData.Content = content
	dialogEditData.InvalidTitle = title == ""
	dialogEditData.InvalidContent = content == ""

	if !dialogEditData.InvalidTitle && !dialogEditData.InvalidContent {
		log.Debugf("Add new database entry: title=%#v; content=%#v", title, content)

		modified := pgvis.NewModified[*pgvis.TroubleReport](user, nil)
		tr := pgvis.NewTroubleReport(modified, title, content)

		if err := db.TroubleReports.Add(tr); err != nil {
			return utils.HandlePgvisError(c, err)
		}
	} else {
		dialogEditData.Submitted = false
	}

	return GETDialogEdit(templates, c, db, dialogEditData)
}

func PUTDialogEdit(templates fs.FS, c echo.Context, db *pgvis.DB) *echo.HTTPError {
	id, herr := utils.ParseRequiredIDQuery(c, "id")
	if herr != nil {
		return herr
	}

	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	title, content, herr := getTitleAndContent(c)
	if herr != nil {
		return herr
	}

	dialogEditData := &EditDialogPageData{
		Submitted:      true,
		ID:             int(id),
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
		if trOld, err := db.TroubleReports.Get(id); err != nil {
			return utils.HandlePgvisError(c, err)
		} else {
			modified.Original = trOld
		}

		if err := db.TroubleReports.Update(id, trNew); err != nil {
			return utils.HandlePgvisError(c, err)
		}
	} else {
		dialogEditData.Submitted = false
	}

	return GETDialogEdit(templates, c, db, dialogEditData)
}

func getTitleAndContent(ctx echo.Context) (title, content string, httpErr *echo.HTTPError) {
	var err error

	title, err = url.QueryUnescape(ctx.FormValue("title"))
	if err != nil {
		return "", "", echo.NewHTTPError(http.StatusBadRequest, "invalid title encoding")
	}
	title = utils.SanitizeInput(title)

	content, err = url.QueryUnescape(ctx.FormValue("content"))
	if err != nil {
		return "", "", echo.NewHTTPError(http.StatusBadRequest, "invalid content encoding")
	}
	content = utils.SanitizeInput(content)

	// Validate title length
	if httpErr := utils.ValidateStringLength(title, "title", 1, 500); httpErr != nil {
		return title, content, httpErr
	}

	// Validate content length
	if httpErr := utils.ValidateStringLength(content, "content", 1, 50000); httpErr != nil {
		return title, content, httpErr
	}

	return title, content, nil
}
