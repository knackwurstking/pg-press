package editor

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/urlb"
	"github.com/knackwurstking/pg-press/internal/utils"
	"github.com/labstack/echo/v4"
)

func Save(c echo.Context) *echo.HTTPError {
	var (
		editorType = c.FormValue("type")
		idParam    = c.FormValue("id")
	)

	log.Info("Save editor content with type %s and ID %s", editorType, idParam)

	// Parse form data
	var (
		title       = strings.TrimSpace(c.FormValue("title"))
		content     = strings.TrimSpace(c.FormValue("content"))
		useMarkdown = c.FormValue("use_markdown") == "on"
	)

	if editorType == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "editor type is required")
	}

	if title == "" || content == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "title and content are required")
	}

	// TODO: Get linked attachments from form
	// Form values: url.Values{
	// 		"attachments":[]string{""},
	// 		"content":[]string{"content..."},
	// 		"existing_attachments_removal":[]string{""},
	// 		"id":[]string{"0"},
	// 		"return_url":[]string{"/trouble-reports"},
	// 		"title":[]string{"test 1"},
	// 		"type":[]string{"troublereport"},
	// }
	v, _ := c.FormParams()
	log.Debug("Form values: %#v", v) // Need to check things first
	//var (
	//	vAttachments                = c.FormValue("attachments")
	//	vExistingAttachmentsRemoval = c.FormValue("existing_attachments_removal")
	//)

	var id int64
	if idParam != "" {
		var err error
		if id, err = strconv.ParseInt(idParam, 10, 64); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid ID parameter")
		}
	}

	switch editorType {
	case "troublereport":
		tr, merr := db.GetTroubleReport(shared.EntityID(id))
		if merr != nil && !merr.IsNotFoundError() {
			return merr.Echo()
		}
		if tr == nil {
			tr = &shared.TroubleReport{
				Title:       title,
				Content:     content,
				UseMarkdown: useMarkdown,
			}
		} else {
			tr.Title = title
			tr.Content = content
			tr.UseMarkdown = useMarkdown
		}

		// TODO: Handle linked attachments, uploads will be handled inside a
		//       separate handler, Just set the linked attachments list here.
		//       The page needs to be changed for this new attachment system.
		//tr.LinkedAttachments = append(tr.LinkedAttachments, ...)

		if merr != nil && merr.IsNotFoundError() {
			if merr = db.AddTroubleReport(tr); merr != nil {
				return merr.Echo()
			}
		} else {
			if merr = db.UpdateTroubleReport(tr); merr != nil {
				return merr.Echo()
			}
		}
	}

	// Redirect back to return URL or appropriate page
	returnURL := c.FormValue("return_url")
	if returnURL != "" {
		url := templ.SafeURL(returnURL)
		if merr := utils.RedirectTo(c, url); merr != nil {
			return merr.WrapEcho("redirect to %#v", url)
		}
		return nil
	}

	// Default redirects based on type
	switch editorType {
	case "troublereport":
		url := urlb.TroubleReports()
		if merr := utils.RedirectTo(c, url); merr != nil {
			return merr.WrapEcho("redirect to %#v", url)
		}
		return nil

	default:
		url := urlb.Home()
		if merr := utils.RedirectTo(c, url); merr != nil {
			return merr.WrapEcho("redirect to %#v", url)
		}
		return nil
	}
}
