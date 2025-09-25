package metalsheets

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/web/handlers"
	"github.com/knackwurstking/pgpress/internal/web/helpers"
	"github.com/knackwurstking/pgpress/internal/web/templates/dialogs"
	"github.com/knackwurstking/pgpress/pkg/models"

	"github.com/labstack/echo/v4"
)

type MetalSheets struct {
	*handlers.BaseHandler
}

func NewMetalSheets(db *database.DB) *MetalSheets {
	return &MetalSheets{
		BaseHandler: handlers.NewBaseHandler(db, logger.HTMXHandlerMetalSheets()),
	}
}

func (h *MetalSheets) RegisterRoutes(e *echo.Echo) {
	helpers.RegisterEchoRoutes(
		e,
		[]*helpers.EchoRoute{
			helpers.NewEchoRoute(http.MethodGet, "/htmx/metal-sheets/edit",
				h.GetEditDialog),
			helpers.NewEchoRoute(http.MethodPost, "/htmx/tools/metal-sheets/edit",
				h.PostCreateMetalSheet),
			helpers.NewEchoRoute(http.MethodPut, "/htmx/metal-sheets/edit",
				h.PutUpdateMetalSheet),
			helpers.NewEchoRoute(http.MethodDelete, "/htmx/metal-sheets/delete",
				h.DeleteMetalSheet),
		},
	)
}

func (h *MetalSheets) GetEditDialog(c echo.Context) error {
	renderProps := &dialogs.EditMetalSheetProps{}

	var (
		toolID int64
		err    error
	)

	// Open edit dialog for adding or editing a metal sheet entry
	// First get the metal sheet id from query param
	if metalSheetID, _ := h.ParseInt64Query(c, "id"); metalSheetID > 0 {
		// Render dialog content for editing an existing metal sheet
		// Store metal sheet to render props
		if renderProps.MetalSheet, err = h.DB.MetalSheets.Get(metalSheetID); err != nil {
			return h.HandleError(c, err, "failed to fetch metal sheet from database")
		}
		toolID = renderProps.MetalSheet.ToolID
	} else {
		// No ID, render dialog content for adding a new metal sheet
		if toolID, err = h.ParseInt64Query(c, "tool_id"); err != nil {
			return h.HandleError(c, err, "failed to get the tool id from query")
		}
	}

	// Store tool to render props
	if renderProps.Tool, err = h.DB.Tools.Get(toolID); err != nil {
		return h.HandleError(c, err, "failed to get tool from database")
	}

	d := dialogs.EditMetalSheet(renderProps)
	if err := d.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c, "failed to render edit metal sheet dialog: "+err.Error())
	}

	return nil
}

func (h *MetalSheets) PostCreateMetalSheet(c echo.Context) error {
	// Get user from context
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	// Get tool ID from query parameter
	toolID, err := h.ParseInt64Query(c, "tool_id")
	if err != nil {
		return h.HandleError(c, err, "failed to get tool_id from query")
	}

	// Get tool for feed content
	tool, err := h.DB.Tools.Get(toolID)
	if err != nil {
		return h.HandleError(c, err, "failed to get tool from database")
	}

	// Parse form data
	metalSheet, err := h.parseMetalSheetForm(c)
	if err != nil {
		return h.HandleError(c, err, "failed to parse metal sheet form data")
	}

	// Set the tool ID
	metalSheet.ToolID = toolID

	// Create the metal sheet in database
	if _, err := h.DB.MetalSheets.Add(metalSheet); err != nil {
		return h.HandleError(c, err, "failed to create metal sheet in database")
	}

	// Create feed entry
	title := "Blech erstellt"
	content := fmt.Sprintf("Werkzeug: %s\nStärke: %.1f mm\nBlech: %.1f mm",
		tool.String(), metalSheet.TileHeight, metalSheet.Value)

	// Add position-specific details for bottom tools
	if tool.Position == models.PositionBottom {
		content += fmt.Sprintf("\nMarke: %d mm\nStf.: %.1f\nStf. Max: %.1f",
			metalSheet.MarkeHeight, metalSheet.STF, metalSheet.STFMax)
	}

	feed := models.NewFeed(title, content, user.TelegramID)
	if err := h.DB.Feeds.Add(feed); err != nil {
		h.LogError("Failed to create feed for metal sheet creation: %v", err)
	}

	// Return success response that closes dialog and refreshes the page
	c.Response().Header().Set("HX-Refresh", "true")
	return c.NoContent(http.StatusOK)
}

func (h *MetalSheets) PutUpdateMetalSheet(c echo.Context) error {
	// Get user from context
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	// Get metal sheet ID from query parameter
	metalSheetID, err := h.ParseInt64Query(c, "id")
	if err != nil {
		return h.HandleError(c, err, "failed to get id from query")
	}

	// Get existing metal sheet
	existingSheet, err := h.DB.MetalSheets.Get(metalSheetID)
	if err != nil {
		return h.HandleError(c, err, "failed to get existing metal sheet from database")
	}

	// Get tool for feed content
	tool, err := h.DB.Tools.Get(existingSheet.ToolID)
	if err != nil {
		return h.HandleError(c, err, "failed to get tool from database")
	}

	// Parse form data
	metalSheet, err := h.parseMetalSheetForm(c)
	if err != nil {
		return h.HandleError(c, err, "failed to parse metal sheet form data")
	}

	// Keep the original ID and tool ID
	metalSheet.ID = existingSheet.ID
	metalSheet.ToolID = existingSheet.ToolID

	// Update the metal sheet in database
	if err := h.DB.MetalSheets.Update(metalSheet); err != nil {
		return h.HandleError(c, err, "failed to update metal sheet in database")
	}

	// Create feed entry
	title := "Blech aktualisiert"
	content := fmt.Sprintf("Werkzeug: %s\nStärke: %.1f mm\nBlech: %.1f mm",
		tool.String(), metalSheet.TileHeight, metalSheet.Value)

	// Add position-specific details for bottom tools
	if tool.Position == models.PositionBottom {
		content += fmt.Sprintf("\nMarke: %d mm\nStf.: %.1f\nStf. Max: %.1f",
			metalSheet.MarkeHeight, metalSheet.STF, metalSheet.STFMax)
	}

	feed := models.NewFeed(title, content, user.TelegramID)
	if err := h.DB.Feeds.Add(feed); err != nil {
		h.LogError("Failed to create feed for metal sheet update: %v", err)
	}

	// Return success response that closes dialog and refreshes the page
	c.Response().Header().Set("HX-Refresh", "true")
	return c.NoContent(http.StatusOK)
}

func (h *MetalSheets) DeleteMetalSheet(c echo.Context) error {
	// Get user from context
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	// Get metal sheet ID from query parameter
	metalSheetID, err := h.ParseInt64Query(c, "id")
	if err != nil {
		return h.HandleError(c, err, "failed to get id from query")
	}

	// Get existing metal sheet for feed content before deletion
	existingSheet, err := h.DB.MetalSheets.Get(metalSheetID)
	if err != nil {
		return h.HandleError(c, err, "failed to get existing metal sheet from database")
	}

	// Get tool for feed content
	tool, err := h.DB.Tools.Get(existingSheet.ToolID)
	if err != nil {
		return h.HandleError(c, err, "failed to get tool from database")
	}

	// Delete the metal sheet from database
	if err := h.DB.MetalSheets.Delete(metalSheetID); err != nil {
		return h.HandleError(c, err, "failed to delete metal sheet from database")
	}

	// Create feed entry
	title := "Blech gelöscht"
	content := fmt.Sprintf("Werkzeug: %s\nStärke: %.1f mm\nBlech: %.1f mm",
		tool.String(), existingSheet.TileHeight, existingSheet.Value)

	// Add position-specific details for bottom tools
	if tool.Position == models.PositionBottom {
		content += fmt.Sprintf("\nMarke: %d mm\nStf.: %.1f\nStf. Max: %.1f",
			existingSheet.MarkeHeight, existingSheet.STF, existingSheet.STFMax)
	}

	feed := models.NewFeed(title, content, user.TelegramID)
	if err := h.DB.Feeds.Add(feed); err != nil {
		h.LogError("Failed to create feed for metal sheet deletion: %v", err)
	}

	// Return empty response (the row will be removed by HTMX)
	return c.NoContent(http.StatusOK)
}

func (h *MetalSheets) parseMetalSheetForm(c echo.Context) (*models.MetalSheet, error) {
	metalSheet := &models.MetalSheet{}

	// Parse tile height (required)
	if tileHeight, err := strconv.ParseFloat(c.FormValue("tile_height"), 64); err != nil {
		return nil, err
	} else {
		metalSheet.TileHeight = tileHeight
	}

	// Parse value (required)
	if value, err := strconv.ParseFloat(c.FormValue("value"), 64); err != nil {
		return nil, err
	} else {
		metalSheet.Value = value
	}

	// Parse marke height (optional, for bottom position only)
	if markeHeightStr := c.FormValue("marke_height"); markeHeightStr != "" {
		if markeHeight, err := strconv.Atoi(markeHeightStr); err == nil {
			metalSheet.MarkeHeight = markeHeight
		}
	}

	// Parse STF (optional, for bottom position only)
	if stfStr := c.FormValue("stf"); stfStr != "" {
		if stf, err := strconv.ParseFloat(stfStr, 64); err == nil {
			metalSheet.STF = stf
		}
	}

	// Parse STF Max (optional, for bottom position only)
	if stfMaxStr := c.FormValue("stf_max"); stfMaxStr != "" {
		if stfMax, err := strconv.ParseFloat(stfMaxStr, 64); err == nil {
			metalSheet.STFMax = stfMax
		}
	}

	return metalSheet, nil
}
