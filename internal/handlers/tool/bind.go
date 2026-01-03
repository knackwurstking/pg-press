package tool

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/labstack/echo/v4"
)

func ToolBinding(c echo.Context) *echo.HTTPError {
	id, merr := shared.ParseParamInt64(c, "id")
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

func renderBindingSection(c echo.Context, tool *shared.Tool) *echo.HTTPError {
	user, merr := shared.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	bindableCassettes, merr := db.ListTools()
	if merr != nil {
		return merr.Echo()
	}

	// Filter bindable cassettes
	i := 0
	for _, t := range bindableCassettes {
		if t.IsCassette() && t.Width == tool.Width && t.Height == tool.Width {
			bindableCassettes[i] = t
			i++
		}
	}
	bindableCassettes = bindableCassettes[:i]

	// Render the template
	t := BindingSection(BindingSectionProps{
		Tool:                tool,
		CassettesForBinding: bindableCassettes,
		IsAdmin:             user.IsAdmin(),
	})
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "BindingSection")
	}
	return nil
}
