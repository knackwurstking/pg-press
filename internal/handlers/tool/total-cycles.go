package tool

import (
	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/labstack/echo/v4"
)

func GetToolTotalCycles(c echo.Context) *echo.HTTPError {
	id, merr := shared.ParseParamInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	totalCycles, merr := db.GetTotalToolCycles(shared.EntityID(id))
	if merr != nil {
		return merr.Echo()
	}

	t := TotalCycles(totalCycles, shared.ParseQueryBool(c, "input"))
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "TotalCycles")
	}
	return nil
}
