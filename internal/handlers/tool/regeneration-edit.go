package tool

import (
	"net/http"

	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/helper"
	"github.com/knackwurstking/pg-press/internal/shared"

	"github.com/labstack/echo/v4"
)

func RegenerationEditable(c echo.Context) *echo.HTTPError {
	id, merr := shared.ParseParamInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	tool, merr := DB.Tool.Tool.GetByID(shared.EntityID(id))
	if merr != nil {
		return merr.Echo()
	}

	eerr := renderRegenerationEdit(c, tool, true, nil)
	if eerr != nil {
		return eerr
	}

	return nil
}

func RegenerationNonEditable(c echo.Context) *echo.HTTPError {
	id, merr := shared.ParseParamInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	tool, merr := DB.Tool.Tool.GetByID(shared.EntityID(id))
	if merr != nil {
		return merr.Echo()
	}

	eerr := renderRegenerationEdit(c, tool, false, nil)
	if eerr != nil {
		return eerr
	}

	return nil
}

func Regeneration(c echo.Context) *echo.HTTPError {
	user, merr := shared.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	id, merr := shared.ParseParamInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	tool, merr := DB.Tool.Tool.GetByID(shared.EntityID(id))
	if merr != nil {
		return merr.Echo()
	}

	statusStr := c.FormValue("status")
	if statusStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "status is required")
	}

	// Handle regeneration start/stop/abort only
	switch statusStr {
	case "regenerating":
		merr = helper.StartToolRegeneration(DB, tool.ID)
		if merr != nil {
			return merr.Echo()
		}

	case "active":
		merr = helper.StopToolRegeneration(DB, tool.ID)
		if merr != nil {
			return merr.Echo()
		}

	case "abort":
		merr := helper.AbortToolRegeneration(DB, tool.ID)
		if merr != nil {
			return merr.Echo()
		}

	default:
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"invalid status: must be 'regenerating', 'active', or 'abort'",
		)
	}

	// Get updated tool and render status display
	tool, merr = DB.Tool.Tool.GetByID(tool.ID)
	if merr != nil {
		return merr.Echo()
	}

	// Render the updated status component
	eerr := renderRegenerationEdit(c, tool, false, user)
	if eerr != nil {
		return eerr
	}

	return renderCyclesSection(c, tool)
}

func renderRegenerationEdit(c echo.Context, tool shared.ModelTool, editable bool, user *shared.User) *echo.HTTPError {
	if user == nil {
		var merr *errors.MasterError
		user, merr = shared.GetUserFromContext(c)
		if merr != nil {
			return merr.Echo()
		}
	}

	regenerations, merr := helper.GetRegenerationsForTool(DB, tool.GetID())
	if merr != nil && merr.Code != http.StatusNotFound {
		return merr.Echo()
	}
	isRegenerating := false
	for _, r := range regenerations {
		if r.Stop == 0 {
			isRegenerating = true
			break
		}
	}

	t := RegenerationEdit(RegenerationEditProps{
		Tool:              tool,
		ActivePressNumber: helper.GetPressNumberForTool(DB, tool.GetID()),
		IsRegenerating:    isRegenerating,
		Editable:          editable,
		UserHasPermission: user.IsAdmin(),
	})
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "ToolStatusEdit")
	}
	return nil
}
