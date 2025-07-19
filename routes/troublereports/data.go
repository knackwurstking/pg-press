// Package troublereports provides HTTP route handlers for trouble report management.
package troublereports

import (
	"io/fs"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/knackwurstking/pg-vis/routes/shared"
	"github.com/knackwurstking/pg-vis/routes/utils"
)

const (
	AdminPrivilegesRequiredMessage = "administrator privileges required"
)

// TroubleReportsPageData contains all the reports and user information.
type TroubleReportsPageData struct {
	TroubleReports []*pgvis.TroubleReport `json:"trouble_reports"`
	User           *pgvis.User            `json:"user"`
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

	return utils.HandleTemplate(
		c,
		TroubleReportsPageData{
			TroubleReports: trs,
			User:           user,
		},
		templates,
		[]string{
			shared.TroubleReportsDataTemplatePath,
		},
	)
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
		return echo.NewHTTPError(
			http.StatusForbidden,
			AdminPrivilegesRequiredMessage,
		)
	}

	log.Infof("Administrator %s (Telegram ID: %d) is deleting trouble report %d",
		user.UserName, user.TelegramID, id)

	if err := db.TroubleReports.Remove(id); err != nil {
		return utils.HandlePgvisError(c, err)
	}

	return GETData(templates, c, db)
}
