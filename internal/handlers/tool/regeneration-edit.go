package tool

import (
	"net/http"

	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/tool/templates"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/urlb"

	"github.com/labstack/echo/v4"
)

func RegenerationEditable(c echo.Context) *echo.HTTPError {
	id, merr := urlb.ParseParamInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	tool, merr := db.GetTool(shared.EntityID(id))
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
	id, merr := urlb.ParseParamInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	tool, merr := db.GetTool(shared.EntityID(id))
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
	user, merr := urlb.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	id, merr := urlb.ParseParamInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	tool, merr := db.GetTool(shared.EntityID(id))
	if merr != nil {
		return merr.WrapEcho("getting tool by ID %d failed", id)
	}

	statusStr := c.FormValue("status")
	if statusStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "status is required")
	}

	log.Debug("Regeneration status change requested: tool ID %d, status %s", tool.ID, statusStr)

	// Handle regeneration start/stop/abort only
	switch statusStr {
	case "regenerating":
		merr = db.StopToolRegeneration(tool.ID)
		if merr != nil {
			return merr.WrapEcho("stopping regeneration for tool ID %d failed", tool.ID)
		}

	case "active":
		merr = db.StartToolRegeneration(tool.ID)
		if merr != nil {
			return merr.WrapEcho("starting regeneration for tool ID %d failed", tool.ID)
		}

	case "abort":
		merr := db.AbortToolRegeneration(tool.ID)
		if merr != nil {
			return merr.WrapEcho("aborting regeneration for tool ID %d failed", tool.ID)
		}

	default:
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"invalid status: must be 'regenerating', 'active', or 'abort'",
		)
	}

	// Get updated tool and render status display
	tool, merr = db.GetTool(tool.ID)
	if merr != nil {
		return merr.WrapEcho("getting tool by ID %d failed", tool.ID)
	}

	// Render the updated status component
	eerr := renderRegenerationEdit(c, tool, false, user)
	if eerr != nil {
		return eerr
	}

	return renderCyclesSection(c, tool)
}

func renderRegenerationEdit(c echo.Context, tool *shared.Tool, editable bool, user *shared.User) *echo.HTTPError {
	if user == nil {
		var merr *errors.HTTPError
		user, merr = urlb.GetUserFromContext(c)
		if merr != nil {
			return merr.Wrap("getting user from context failed").Echo()
		}
	}

	regenerations, merr := db.ListToolRegenerationsByTool(tool.ID)
	if merr != nil && !merr.IsNotFoundError() {
		return merr.Wrap("listing regenerations for tool ID %d failed", tool.ID).Echo()
	}
	isRegenerating := false
	for _, r := range regenerations {
		if r.Stop == 0 {
			isRegenerating = true
			break
		}
	}

	pressNumber, merr := db.GetPressNumberForTool(tool.ID)
	if merr != nil {
		return merr.Wrap("getting press number for tool ID %d failed", tool.ID).Echo()
	}

	t := templates.RegenerationEdit(templates.RegenerationEditProps{
		Tool:              tool,
		ActivePressNumber: pressNumber,
		IsRegenerating:    isRegenerating,
		Editable:          editable,
		UserHasPermission: user.IsAdmin(),
	})
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "RegenerationEdit")
	}
	return nil
}
