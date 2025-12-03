package metalsheets

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/knackwurstking/pg-press/env"
	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
	"github.com/knackwurstking/pg-press/services"
	"github.com/knackwurstking/pg-press/utils"
	"github.com/knackwurstking/ui"
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
	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	// Extract metal sheet ID from query parameters
	metalSheetIDQuery, err := utils.ParseQueryInt64(c, "id")
	if err != nil {
		return errors.BadRequest(err, "get id from query")
	}
	metalSheetID := models.MetalSheetID(metalSheetIDQuery)

	// Fetch the existing metal sheet before deletion for feed creation
	existingSheet, err := h.registry.MetalSheets.Get(metalSheetID)
	if err != nil {
		return errors.Handler(err, "get existing metal sheet from database")
	}

	// Fetch the associated tool for feed creation
	tool, err := h.registry.Tools.Get(existingSheet.ToolID)
	if err != nil {
		return errors.Handler(err, "get tool from database")
	}

	// Delete the metal sheet from database
	if err := h.registry.MetalSheets.Delete(metalSheetID); err != nil {
		return errors.Handler(err, "delete metal sheet from database")
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
