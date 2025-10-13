package tool

import (
	"github.com/knackwurstking/pgpress/internal/web/features/tool/templates"
	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/labstack/echo/v4"
)

func (h *Handler) HTMXGetToolMetalSheets(c echo.Context) error {
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	toolID, err := h.ParseInt64Query(c, "tool_id")
	if err != nil {
		return h.RenderBadRequest(c, "failed to parse tool_id: "+err.Error())
	}

	h.Log.Debug("Fetching metal sheets for tool %d", toolID)

	tool, err := h.DB.Tools.GetWithNotes(toolID)
	if err != nil {
		return h.HandleError(c, err, "failed to get tool")
	}

	// Fetch metal sheets assigned to this tool
	metalSheets, err := h.DB.MetalSheets.GetByToolID(toolID)
	if err != nil {
		// Log error but don't fail - metal sheets are supplementary data
		h.Log.Error("Failed to fetch metal sheets: %v", err)
		metalSheets = []*models.MetalSheet{}
	}

	metalSheetsSection := templates.MetalSheets(user, metalSheets, tool)

	if err := metalSheetsSection.Render(c.Request().Context(), c.Response()); err != nil {
		return h.HandleError(c, err, "failed to render tool metal sheets section")
	}

	return nil
}
