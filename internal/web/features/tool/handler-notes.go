package tool

import (
	"github.com/knackwurstking/pgpress/internal/web/features/tool/templates"
	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/labstack/echo/v4"
)

func (h *Handler) HTMXGetToolNotes(c echo.Context) error {
	toolID, err := h.ParseInt64Query(c, "tool_id")
	if err != nil {
		return h.RenderBadRequest(c, "failed to parse tool_id: "+err.Error())
	}

	h.Log.Debug("Fetching notes for tool %d", toolID)

	// Get the tool
	tool, err := h.DB.Tools.Get(toolID)
	if err != nil {
		return h.HandleError(c, err, "failed to get tool")
	}

	// Get notes for this tool
	notes, err := h.DB.Notes.GetByTool(toolID)
	if err != nil {
		return h.HandleError(c, err, "failed to get notes for tool")
	}

	// Create ToolWithNotes for template compatibility
	toolWithNotes := &models.ToolWithNotes{
		Tool:        tool,
		LoadedNotes: notes,
	}

	notesSection := templates.SectionNotes(toolWithNotes)

	if err := notesSection.Render(c.Request().Context(), c.Response()); err != nil {
		return h.HandleError(c, err, "failed to render tool notes section")
	}

	return nil
}
