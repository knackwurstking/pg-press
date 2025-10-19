package tool

import (
	"fmt"
	"strconv"

	"github.com/knackwurstking/pgpress/internal/web/features/tool/templates"
	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/labstack/echo/v4"
)

func (h *Handler) HTMXPatchToolBinding(c echo.Context) error {
	var isAdmin bool
	{ // Check for admin
		user, err := h.GetUserFromContext(c)
		if err != nil {
			return h.RenderBadRequest(c, err.Error())
		}
		isAdmin = user.IsAdmin()
	}

	// Get tool from param "/:id"
	toolID, err := h.ParseInt64Param(c, "id")
	if err != nil {
		return h.RenderBadRequest(c, "failed to parse tool_id: "+err.Error())
	}

	tool, err := h.DB.Tools.Get(toolID)
	if err != nil {
		return h.HandleError(c, err, "failed to get tool")
	}

	var targetID int64
	{ // Get target_id from form value
		targetIDString := c.FormValue("target_id")
		if targetIDString == "" {
			return h.RenderBadRequest(c, fmt.Sprintf(
				"failed to parse target_id: %+v", targetIDString))
		}

		targetID, err = strconv.ParseInt(targetIDString, 10, 64)
		if err != nil {
			return h.RenderBadRequest(c, "invalid target_id: "+err.Error())
		}
	}

	h.Log.Info("Updating tool binding: tool %d -> target %d",
		toolID, targetID)

	{ // Make sure to check for position first (target == top && toolID == cassette)
		var (
			cassette int64
			target   int64 // top position
		)

		if tool.Position == models.PositionTopCassette {
			cassette = tool.ID
			target = targetID
		} else {
			cassette = targetID // If this is not a cassette, the bind method will return an error
			target = tool.ID
		}

		// Bind tool to target, this will get an error if target already has a binding
		if err = h.DB.Tools.Bind(cassette, target); err != nil {
			return h.HandleError(c, err, "failed to bind tool")
		}
	}

	// Update tools binding, no need to re fetch this tools data from the database
	tool.Binding = &targetID

	// Get tools for binding
	toolsForBinding, err := h.getToolsForBinding(c, tool)
	if err != nil {
		return err
	}

	// Render the template
	bs := templates.BindingSection(models.NewResolvedTool(
		tool, h.getBindingTool(tool, toolsForBinding), nil,
	), toolsForBinding, isAdmin, nil)

	if err = bs.Render(c.Request().Context(), c.Response()); err != nil {
		return h.HandleError(c, err, "failed to render binding section")
	}

	return nil
}

func (h *Handler) HTMXPatchToolUnBinding(c echo.Context) error {
	var isAdmin bool
	{ // Check for admin
		user, err := h.GetUserFromContext(c)
		if err != nil {
			return h.RenderBadRequest(c, err.Error())
		}
		isAdmin = user.IsAdmin()
	}

	// Get tool from param "/:id"
	toolID, err := h.ParseInt64Param(c, "id")
	if err != nil {
		return h.RenderBadRequest(c, "failed to parse tool_id: "+err.Error())
	}

	if err := h.DB.Tools.UnBind(toolID); err != nil {
		return h.HandleError(c, err, "failed to unbind tool")
	}

	// Get tools for rendering the template
	tool, err := h.DB.Tools.Get(toolID)
	if err != nil {
		return h.RenderBadRequest(c, err.Error())
	}

	// Get tools for binding
	toolsForBinding, err := h.getToolsForBinding(c, tool)
	if err != nil {
		return err
	}

	// Render the template
	bs := templates.BindingSection(models.NewResolvedTool(
		tool, h.getBindingTool(tool, toolsForBinding), nil,
	), toolsForBinding, isAdmin, nil)

	if err = bs.Render(c.Request().Context(), c.Response()); err != nil {
		return h.HandleError(c, err, "failed to render binding section")
	}

	return nil
}

// NOTE: Used in handler-cycles for the cycles section
func (h *Handler) getToolsForBinding(c echo.Context, tool *models.Tool) ([]*models.Tool, error) {
	var filteredToolsForBinding []*models.Tool

	if tool.Position != models.PositionTopCassette && tool.Position != models.PositionTop {
		return filteredToolsForBinding, nil
	}

	// Get all tools
	tools, err := h.DB.Tools.List()
	if err != nil {
		return nil, h.HandleError(c, err, "failed to get tools")
	}
	// Filter tools for binding
	for _, t := range tools {
		if t.Format != tool.Format {
			continue
		}

		if tool.Position == models.PositionTop {
			if t.Position == models.PositionTopCassette {
				filteredToolsForBinding = append(filteredToolsForBinding, t)
			}

			continue
		}

		// Else: position top cassette
		if t.Position == models.PositionTop {
			filteredToolsForBinding = append(filteredToolsForBinding, t)
		}
	}

	return filteredToolsForBinding, nil
}

func (h *Handler) getBindingTool(tool *models.Tool, toolsForBinding []*models.Tool) *models.Tool {
	var bindingTool *models.Tool
	if tool.IsBound() {
		for _, t := range toolsForBinding {
			if t.ID == *tool.Binding {
				bindingTool = t
			}
		}
	}
	return bindingTool
}
