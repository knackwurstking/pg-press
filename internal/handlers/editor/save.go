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
		if merr != nil {
			return merr.Echo()
		}
		tr.Title = title
		tr.Content = content
		tr.UseMarkdown = useMarkdown

		// TODO: Handle linked attachments, uploads will be handled inside a
		//       separate handler, Just set the linked attachments list here.
		//       The page needs to be changed for this new attachment system.
		//tr.LinkedAttachments =

		if merr = db.AddTroubleReport(tr); merr != nil {
			return merr.Echo()
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
