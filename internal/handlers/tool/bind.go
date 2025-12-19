package tool

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/services/helper"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/labstack/echo/v4"
)

func HTMXPatchToolBinding(c echo.Context) *echo.HTTPError {
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

	// Get target_id from form value
	targetIDString := c.FormValue("target_id")
	if targetIDString == "" {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			fmt.Sprintf("failed to parse target_id: %#v", targetIDString),
		)
	}
	id, err := strconv.ParseInt(targetIDString, 10, 64)
	if err != nil {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"invalid target_id",
		)
	}
	cassette, merr := DB.Tool.Cassette.GetByID(shared.EntityID(id))
	if merr != nil {
		return merr.Echo()
	}

	// Bind tool to target, this will get an error if target already has a binding
	merr = helper.BindCassetteToTool(DB, tool.ID, cassette.ID)
	if merr != nil {
		return merr.Echo()
	}

	return renderBindingSection(c, tool, user)
}

func renderBindingSection(c echo.Context, tool *shared.Tool, user *shared.User) *echo.HTTPError {
	cassettesForBinding, merr := helper.ListAvailableCassettesForBinding(DB, tool.ID)
	if merr != nil {
		return merr.Echo()
	}

	// Render the template
	t := BindingSection(BindingSectionProps{
		Tool:                tool,
		CassettesForBinding: cassettesForBinding,
		IsAdmin:             user.IsAdmin(),
	})
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "BindingSection")
	}
	return nil
}
