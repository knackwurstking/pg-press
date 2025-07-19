package troublereports

import (
	"html/template"
	"io/fs"
	"net/http"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/knackwurstking/pg-vis/routes/shared"
	"github.com/labstack/echo/v4"
)

type ModificationsPageData struct {
	User *pgvis.User
}

func GETModifications(templates fs.FS, c echo.Context, db *pgvis.DB) *echo.HTTPError {
	user, ok := c.Get("user").(*pgvis.User)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "user not found")
	}

	t, err := template.ParseFS(
		templates,
		shared.TroubleReportsModificationsTemplatePath,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			pgvis.WrapError(err, "failed to parse template"))
	}

	err = t.Execute(c.Response(), ModificationsPageData{
		User: user,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			pgvis.WrapError(err, "failed to execute template"))
	}

	return nil
}
