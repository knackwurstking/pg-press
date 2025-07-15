package troublereports

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"strconv"

	"github.com/charmbracelet/log"
	"github.com/knackwurstking/pg-vis/html/handler"
	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/labstack/echo/v4"
)

type DataPageData struct {
	TroubleReports []*pgvis.TroubleReport
	IsAdmin        bool
}

func GETData(templates fs.FS, c echo.Context, db *pgvis.DB) *echo.HTTPError {
	user, herr := handler.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

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

	if err = t.Execute(c.Response(), DataPageData{
		TroubleReports: trs,
		IsAdmin: user.IsAdmin(),
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return nil
}

func DELETEData(templates fs.FS, c echo.Context, db *pgvis.DB) *echo.HTTPError {
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

	user, herr := handler.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	if user.IsAdmin() {
		log.Debugf(
			"User %d (%s) is deleting the trouble report %d",
			user.TelegramID, user.UserName, id,
		)

		if err := db.TroubleReports.Remove(int64(id)); err != nil {
			return echo.NewHTTPError(
				http.StatusBadRequest,
				fmt.Errorf("invalid \"id\" %d: not found", id),
			)
		}
	} else {
		// TODO: Voting system for deletion
		log.Warnf(
			"User %d (%s) not allowed for deletion. "+
				"Voting system not implemented for now.",
			user.TelegramID, user.UserName,
		)
	}

	return GETData(templates, c, db)
}
