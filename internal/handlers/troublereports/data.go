package troublereports

import (
	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/troublereports/templates"
	"github.com/knackwurstking/pg-press/internal/utils"

	"github.com/labstack/echo/v4"
)

func GetData(c echo.Context) *echo.HTTPError {
	user, merr := utils.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	troubleReports, merr := db.ListTroubleReports()
	if merr != nil {
		return merr.Echo()
	}

	t := templates.ListReports(user, troubleReports)
	if err := t.Render(c.Request().Context(), c.Response().Writer); err != nil {
		return errors.NewRenderError(err, "ListReports")
	}
	return nil
}
