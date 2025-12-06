package metalsheets

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/knackwurstking/pg-press/env"
	"github.com/knackwurstking/pg-press/models"
	"github.com/knackwurstking/pg-press/services"
	"github.com/knackwurstking/pg-press/utils"
	ui "github.com/knackwurstking/ui/ui-templ"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	registry *services.Registry
}

func NewHandler(r *services.Registry) *Handler {
	return &Handler{
		registry: r,
	}
}

func (h *Handler) RegisterRoutes(e *echo.Echo, path string) {
	ui.RegisterEchoRoutes(e, env.ServerPathPrefix, []*ui.EchoRoute{
		// DELETE route for removing a metal sheet
		ui.NewEchoRoute(http.MethodDelete, path+"/delete", h.HTMXDeleteMetalSheet),
	})
}

// DeleteMetalSheet handles the deletion of metal sheets
func (h *Handler) HTMXDeleteMetalSheet(c echo.Context) error {
	slog.Info("Remove a metal sheet entry")

	// Get current user for feed creation
	user, merr := utils.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	// Extract metal sheet ID from query parameters
	metalSheetIDQuery, merr := utils.ParseQueryInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	metalSheetID := models.MetalSheetID(metalSheetIDQuery)

	// Fetch the existing metal sheet before deletion for feed creation
	existingSheet, merr := h.registry.MetalSheets.Get(metalSheetID)
	if merr != nil {
		return merr.Echo()
	}

	// Fetch the associated tool for feed creation
	tool, merr := h.registry.Tools.Get(existingSheet.ToolID)
	if merr != nil {
		return merr.Echo()
	}

	// Delete the metal sheet from database
	merr = h.registry.MetalSheets.Delete(metalSheetID)
	if merr != nil {
		return merr.Echo()
	}

	// Create feed entry for the deleted metal sheet
	h.createFeed(user, tool, existingSheet, "Blech gelöscht")

	utils.SetHXTrigger(c, env.HXGlobalTrigger)

	return nil
}

// createFeed creates a feed entry for metal sheet operations
func (h *Handler) createFeed(user *models.User, tool *models.Tool, metalSheet *models.MetalSheet, title string) {
	// Build base feed content with tool and metal sheet info
	content := fmt.Sprintf("Werkzeug: %s\nStärke: %.1f mm\nBlech: %.1f mm\nTyp: %s",
		tool.String(), metalSheet.TileHeight, metalSheet.Value, metalSheet.Identifier.String())

	// Add additional fields for bottom position tools
	if tool.Position == models.PositionBottom {
		content += fmt.Sprintf("\nMarke: %d mm\nStf.: %.1f\nStf. Max: %.1f",
			metalSheet.MarkeHeight, metalSheet.STF, metalSheet.STFMax)
	}

	// Create and save the feed entry
	if _, err := h.registry.Feeds.AddSimple(title, content, user.TelegramID); err != nil {
		slog.Warn("Failed to create feed", "error", err)
	}
}
