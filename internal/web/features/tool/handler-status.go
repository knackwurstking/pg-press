// This file contains handlers for the tool status and regenerations management.
package tool

import (
	"fmt"
	"strconv"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pgpress/internal/web/shared/components"
	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/labstack/echo/v4"
)

func (h *Handler) HTMXGetStatusEdit(c echo.Context) error {
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	toolID, err := h.ParseInt64Query(c, "id")
	if err != nil {
		return h.RenderBadRequest(c, "failed to parse tool ID: "+err.Error())
	}

	tool, err := h.DB.Tools.Get(toolID)
	if err != nil {
		return h.HandleError(c, err, "failed to get tool from database")
	}

	statusEdit := h.renderStatusComponent(tool, true, user)
	if err := statusEdit.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c, "failed to render tool status edit: "+err.Error())
	}

	return nil
}

func (h *Handler) HTMXGetStatusDisplay(c echo.Context) error {
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	toolID, err := h.ParseInt64Query(c, "id")
	if err != nil {
		return h.RenderBadRequest(c, "failed to parse tool ID: "+err.Error())
	}

	tool, err := h.DB.Tools.Get(toolID)
	if err != nil {
		return h.HandleError(c, err, "failed to get tool from database")
	}

	statusDisplay := h.renderStatusComponent(tool, false, user)
	if err := statusDisplay.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c, "failed to render tool status display: "+err.Error())
	}

	return nil
}

func (h *Handler) HTMXUpdateToolStatus(c echo.Context) error {
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	toolIDStr := c.FormValue("tool_id")
	if toolIDStr == "" {
		return h.RenderBadRequest(c, "tool_id is required")
	}

	toolID, err := strconv.ParseInt(toolIDStr, 10, 64)
	if err != nil {
		return h.RenderBadRequest(c, "invalid tool_id: "+err.Error())
	}

	statusStr := c.FormValue("status")
	if statusStr == "" {
		return h.RenderBadRequest(c, "status is required")
	}

	tool, err := h.DB.Tools.Get(toolID)
	if err != nil {
		return h.HandleError(c, err, "failed to get tool from database")
	}

	h.Log.Info("User %s updating status for tool %d from %s to %s", user.Name, toolID, tool.Status(), statusStr)

	// Handle regeneration start/stop/abort only
	// FIXME: Ok, this is bullshit, i mean the status does change, but no regeneration entry
	switch statusStr {
	case "regenerating":
		// Start regeneration
		_, err := h.DB.ToolRegenerations.StartToolRegeneration(toolID, "", user)
		if err != nil {
			return h.HandleError(c, err, "failed to start tool regeneration")
		}

		// Create feed entry
		title := "Werkzeug Regenerierung gestartet"
		content := fmt.Sprintf("Werkzeug: %s", tool.String())
		feed := models.NewFeed(title, content, user.TelegramID)
		if err := h.DB.Feeds.Add(feed); err != nil {
			h.Log.Error("Failed to create feed for regeneration start: %v", err)
		}

	case "active":
		if err := h.DB.ToolRegenerations.StopToolRegeneration(toolID, user); err != nil {
			return h.HandleError(c, err, "failed to stop tool regeneration")
		}

		// Create feed entry
		title := "Werkzeug Regenerierung beendet"
		content := fmt.Sprintf("Werkzeug: %s", tool.String())
		feed := models.NewFeed(title, content, user.TelegramID)
		if err := h.DB.Feeds.Add(feed); err != nil {
			h.Log.Error("Failed to create feed for regeneration stop: %v", err)
		}

	case "abort":
		// Abort regeneration (remove regeneration record and set status to false)
		if err := h.DB.ToolRegenerations.AbortToolRegeneration(toolID, user); err != nil {
			return h.HandleError(c, err, "failed to abort tool regeneration")
		}

		// Create feed entry
		title := "Werkzeug Regenerierung abgebrochen"
		content := fmt.Sprintf("Werkzeug: %s", tool.String())
		feed := models.NewFeed(title, content, user.TelegramID)
		if err := h.DB.Feeds.Add(feed); err != nil {
			h.Log.Error("Failed to create feed for regeneration abort: %v", err)
		}

	default:
		return h.RenderBadRequest(c, "invalid status: must be 'regenerating', 'active', or 'abort'")
	}

	// Get updated tool and render status display
	updatedTool, err := h.DB.Tools.Get(toolID)
	if err != nil {
		return h.HandleError(c, err, "failed to get updated tool from database")
	}

	statusDisplay := h.renderStatusComponent(updatedTool, false, user)
	if err := statusDisplay.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c, "failed to render updated tool status: "+err.Error())
	}

	return nil
}

func (h *Handler) renderStatusComponent(tool *models.Tool, editable bool, user *models.User) templ.Component {
	return components.ToolStatusEdit(&components.ToolStatusEditProps{
		Tool:              tool,
		Editable:          editable,
		UserHasPermission: user.IsAdmin(),
	})
}
