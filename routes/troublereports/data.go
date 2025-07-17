package troublereports

import (
	"html/template"
	"io/fs"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/knackwurstking/pg-vis/routes/utils"
)

type DataPageData struct {
	TroubleReports []*pgvis.TroubleReport
	IsAdmin        bool
}

func GETData(templates fs.FS, c echo.Context, db *pgvis.DB) *echo.HTTPError {
	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	trs, err := db.TroubleReports.List()
	if err != nil {
		return utils.HandlePgvisError(c, err)
	}

	t, err := template.ParseFS(templates, "templates/trouble-reports/data.html")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err = t.Execute(c.Response(), DataPageData{
		TroubleReports: trs,
		IsAdmin:        user.IsAdmin(),
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return nil
}

func DELETEData(templates fs.FS, c echo.Context, db *pgvis.DB) *echo.HTTPError {
	id, herr := utils.ParseRequiredIDQuery(c, "id")
	if herr != nil {
		return herr
	}

	// Check if user has admin privileges and get user info
	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	if !user.IsAdmin() {
		// TODO: Voting system for deletion
		log.Warnf("Non-admin user attempted deletion. Voting system not implemented for now.")
		return echo.NewHTTPError(http.StatusForbidden, "administrator privileges required")
	}

	log.Debugf(
		"User %d (%s) is deleting the trouble report %d",
		user.TelegramID, user.UserName, id,
	)

	if err := db.TroubleReports.Remove(id); err != nil {
		return utils.HandlePgvisError(c, err)
	}

	return GETData(templates, c, db)
}
