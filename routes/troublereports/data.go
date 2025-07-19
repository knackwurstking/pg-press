// Package troublereports provides HTTP route handlers for trouble report management.
package troublereports

import (
	"html/template"
	"io/fs"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/knackwurstking/pg-vis/routes/shared"
	"github.com/knackwurstking/pg-vis/routes/utils"
)

// DataPageData contains the data structure for trouble reports templates.
type DataPageData struct {
	TroubleReports []*pgvis.TroubleReport `json:"trouble_reports"`
	IsAdmin        bool                   `json:"is_admin"`
}

// GETData handles GET requests to retrieve and render trouble reports data.
func GETData(templates fs.FS, c echo.Context, db *pgvis.DB) *echo.HTTPError {
	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	trs, err := db.TroubleReports.List()
	if err != nil {
		return utils.HandlePgvisError(c, err)
	}

	t, err := template.ParseFS(templates, shared.TroubleReportsDataTemplatePath)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			pgvis.WrapError(err, "failed to load page template"))
	}

	pageData := DataPageData{
		TroubleReports: trs,
		IsAdmin:        user.IsAdmin(),
	}

	if err = t.Execute(c.Response(), pageData); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			pgvis.WrapError(err, "failed to render page"))
	}

	return nil
}

// DELETEData handles DELETE requests to remove trouble reports.
func DELETEData(templates fs.FS, c echo.Context, db *pgvis.DB) *echo.HTTPError {
	id, herr := utils.ParseRequiredIDQuery(c, "id")
	if herr != nil {
		return herr
	}

	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	if !user.IsAdmin() {
		log.Warnf("Non-admin user %s (ID: %d) attempted to delete trouble report %d",
			user.UserName, user.TelegramID, id)
		return echo.NewHTTPError(http.StatusForbidden,
			"administrator privileges required for deletion")
	}

	log.Infof("Administrator %s (Telegram ID: %d) is deleting trouble report %d",
		user.UserName, user.TelegramID, id)

	if err := db.TroubleReports.Remove(id); err != nil {
		return utils.HandlePgvisError(c, err)
	}

	return GETData(templates, c, db)
}
