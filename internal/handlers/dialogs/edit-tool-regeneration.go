package dialogs

import (
	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"

	"github.com/labstack/echo/v4"
)

func GetEditToolRegeneration(c echo.Context) *echo.HTTPError {
	id, merr := shared.ParseQueryInt64(c, "id")
	if merr != nil && merr.IsNotFoundError() {
		return merr.Echo()
	}

	if id > 0 {
		tr, merr := db.GetToolRegeneration(shared.EntityID(id))
		if merr != nil {
			return merr.Echo()
		}
		t := EditToolRegenerationDialog(tr)
		err := t.Render(c.Request().Context(), c.Response())
		if err != nil {
			return errors.NewRenderError(err, "EditToolRegenerationDialog")
		}
		return nil
	}

	id, merr = shared.ParseQueryInt64(c, "tool_id")
	if merr != nil {
		return merr.Echo()
	}
	t := NewToolRegenerationDialog(shared.EntityID(id))
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "EditToolRegenerationDialog")
	}
	return nil
}

func PostToolRegeneration(c echo.Context) *echo.HTTPError {
	// TODO: ...
}

func PutToolRegeneration(c echo.Context) *echo.HTTPError {
	// TODO: ...
}
