package tool

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/tool/templates"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/utils"

	"github.com/labstack/echo/v4"
)

func ToolBinding(c echo.Context) *echo.HTTPError {
	id, merr := utils.GetParamInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	toolID := shared.EntityID(id)

	// Get target_id from form value
	vTargetID := c.FormValue("target_id")
	if vTargetID == "" {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			fmt.Sprintf("failed to parse target_id: %#v", vTargetID),
		)
	}
	id, err := strconv.ParseInt(vTargetID, 10, 64)
	if err != nil {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"invalid target_id",
		)
	}
	cassetteID := shared.EntityID(id)

	// Bind tool to target, this will get an error if target already has a binding
	merr = db.BindTool(toolID, cassetteID)
	if merr != nil {
		return merr.Echo()
	}

	tool, merr := db.GetTool(toolID)
	if merr != nil {
		return merr.Echo()
	}
	return renderBindingSection(c, tool)
}

func ToolUnBinding(c echo.Context) *echo.HTTPError {
	id, merr := utils.GetParamInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	toolID := shared.EntityID(id)
	merr = db.UnbindTool(toolID)
	if merr != nil {
		return merr.Echo()
	}
	tool, merr := db.GetTool(toolID)
	if merr != nil {
		return merr.Echo()
	}

	return renderBindingSection(c, tool)
}

func renderBindingSection(c echo.Context, tool *shared.Tool) *echo.HTTPError {
	user, merr := utils.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	bindableCassettes, eerr := listBindableCassettes(tool)
	if eerr != nil {
		return eerr
	}

	// Render the template
	t := templates.BindingSection(templates.BindingSectionProps{
		Tool:                tool,
		CassettesForBinding: bindableCassettes,
		IsAdmin:             user.IsAdmin(),
	})
	if err := t.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.NewRenderError(err, "BindingSection")
	}
	return nil
}
