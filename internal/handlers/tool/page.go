package tool

import (
	"net/http"

	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"

	"github.com/labstack/echo/v4"
)

func GetToolPage(c echo.Context) *echo.HTTPError {
	user, merr := shared.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	id, merr := shared.ParseParamInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}

	var tool shared.ModelTool
	tool, merr = DB.Tool.Tool.GetByID(shared.EntityID(id))
	if merr != nil {
		if merr.Code == http.StatusNotFound {
			tool, merr = DB.Tool.Cassette.GetByID(shared.EntityID(id))
			if merr != nil {
				return merr.Echo()
			}
		} else {
			return merr.Echo()
		}
	}

	t := Page(&PageProps{
		Tool: tool,
		User: user,
	})
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "Tool Page")
	}

	return nil
}
