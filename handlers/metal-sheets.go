package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/knackwurstking/pg-press/components"
	"github.com/knackwurstking/pg-press/env"
	"github.com/knackwurstking/pg-press/models"
	"github.com/knackwurstking/pg-press/services"
	"github.com/knackwurstking/pg-press/utils"
	"github.com/labstack/echo/v4"
)

type MetalSheets struct {
	registry *services.Registry
}

func NewMetalSheets(r *services.Registry) *MetalSheets {
	return &MetalSheets{
		registry: r,
	}
}

func (h *MetalSheets) RegisterRoutes(e *echo.Echo) {
	utils.RegisterEchoRoutes(
		e,
		[]*utils.EchoRoute{
			// HTMX
			// GET route for displaying the edit dialog
			utils.NewEchoRoute(http.MethodGet, "/htmx/metal-sheets/edit",
				h.HTMXGetEditMetalSheetDialog),

			// POST route for creating a new metal sheet
			utils.NewEchoRoute(http.MethodPost, "/htmx/metal-sheets/edit",
				h.HTMXPostEditMetalSheetDialog),

			// PUT route for updating an existing metal sheet
			utils.NewEchoRoute(http.MethodPut, "/htmx/metal-sheets/edit",
				h.HTMXPutEditMetalSheetDialog),

			// DELETE route for removing a metal sheet
			utils.NewEchoRoute(http.MethodDelete, "/htmx/metal-sheets/delete",
				h.HTMXDeleteMetalSheet),
		},
	)
}

// GetEditDialog renders the edit/create dialog for metal sheets
func (h *MetalSheets) HTMXGetEditMetalSheetDialog(c echo.Context) error {
	renderProps := &components.DialogEditMetalSheetProps{}
	var toolID models.ToolID
	var err error

	// Check if we're editing an existing metal sheet (has ID) or creating new one
	if metalSheetIDQuery, _ := utils.ParseQueryInt64(c, "id"); metalSheetIDQuery > 0 {
		metalSheetID := models.MetalSheetID(metalSheetIDQuery)

		// Fetch existing metal sheet for editing
		if renderProps.MetalSheet, err = h.registry.MetalSheets.Get(metalSheetID); err != nil {
			return utils.HandleError(err, "failed to fetch metal sheet from database")
		}
		toolID = renderProps.MetalSheet.ToolID
	} else {
		// Creating new metal sheet, get tool_id from query
		toolIDQuery, err := utils.ParseQueryInt64(c, "tool_id")
		if err != nil {
			return utils.HandleError(err, "failed to get the tool id from query")
		}
		toolID = models.ToolID(toolIDQuery)
	}

	// Fetch the associated tool for the dialog
	if renderProps.Tool, err = h.registry.Tools.Get(toolID); err != nil {
		return utils.HandleError(err, "failed to get tool from database")
	}

	// Render the edit dialog template
	d := components.DialogEditMetalSheet(renderProps)
	if err := d.Render(c.Request().Context(), c.Response()); err != nil {
		return utils.HandleError(err, "failed to render edit metal sheet dialog")
	}

	return nil
}

// PostCreateMetalSheet handles the creation of a new metal sheet
func (h *MetalSheets) HTMXPostEditMetalSheetDialog(c echo.Context) error {
	// Get current user for feed creation
	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return utils.HandleBadRequest(err, "failed to get user from context")
	}

	// Extract tool ID from query parameters
	toolIDQuery, err := utils.ParseQueryInt64(c, "tool_id")
	if err != nil {
		return utils.HandleError(err, "failed to get tool_id from query")
	}
	toolID := models.ToolID(toolIDQuery)

	// Fetch the associated tool
	tool, err := h.registry.Tools.Get(toolID)
	if err != nil {
		return utils.HandleError(err, "failed to get tool from database")
	}

	// Parse form data into metal sheet model
	metalSheet, err := h.parseMetalSheetForm(c)
	if err != nil {
		return utils.HandleError(err, "failed to parse metal sheet form data")
	}

	// Associate metal sheet with the tool
	metalSheet.ToolID = toolID

	// Save new metal sheet to database
	if _, err := h.registry.MetalSheets.Add(metalSheet); err != nil {
		return utils.HandleError(err, "failed to create metal sheet in database")
	}

	// Create feed entry for the new metal sheet
	h.createFeed(user, tool, metalSheet, "Blech erstellt")

	utils.SetHXTrigger(c, env.HXGlobalTrigger)

	return nil
}

// PutUpdateMetalSheet handles updates to existing metal sheets
func (h *MetalSheets) HTMXPutEditMetalSheetDialog(c echo.Context) error {
	// Get current user for feed creation
	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return utils.HandleBadRequest(err, "failed to get user from context")
	}

	// Extract metal sheet ID from query parameters
	metalSheetIDQuery, err := utils.ParseQueryInt64(c, "id")
	if err != nil {
		return utils.HandleBadRequest(err, "failed to get id from query")
	}
	metalSheetID := models.MetalSheetID(metalSheetIDQuery)

	// Fetch the existing metal sheet to preserve ID and tool association
	existingSheet, err := h.registry.MetalSheets.Get(metalSheetID)
	if err != nil {
		return utils.HandleError(err, "failed to get existing metal sheet from database")
	}

	// Fetch the associated tool for feed creation
	tool, err := h.registry.Tools.Get(existingSheet.ToolID)
	if err != nil {
		return utils.HandleError(err, "failed to get tool from database")
	}

	// Parse updated form data
	metalSheet, err := h.parseMetalSheetForm(c)
	if err != nil {
		return utils.HandleError(err, "failed to parse metal sheet form data")
	}

	// Preserve the original ID and tool association
	metalSheet.ID = existingSheet.ID
	metalSheet.ToolID = existingSheet.ToolID

	// Update the metal sheet in database
	if err := h.registry.MetalSheets.Update(metalSheet); err != nil {
		return utils.HandleError(err, "failed to update metal sheet in database")
	}

	// Create feed entry for the updated metal sheet showing changes
	h.createUpdateFeed(user, tool, existingSheet, metalSheet)

	utils.SetHXTrigger(c, env.HXGlobalTrigger)

	return nil
}

// DeleteMetalSheet handles the deletion of metal sheets
func (h *MetalSheets) HTMXDeleteMetalSheet(c echo.Context) error {
	// Get current user for feed creation
	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return utils.HandleBadRequest(err, "failed to get user from context")
	}

	// Extract metal sheet ID from query parameters
	metalSheetIDQuery, err := utils.ParseQueryInt64(c, "id")
	if err != nil {
		return utils.HandleBadRequest(err, "failed to get id from query")
	}
	metalSheetID := models.MetalSheetID(metalSheetIDQuery)

	// Fetch the existing metal sheet before deletion for feed creation
	existingSheet, err := h.registry.MetalSheets.Get(metalSheetID)
	if err != nil {
		return utils.HandleError(err, "failed to get existing metal sheet from database")
	}

	// Fetch the associated tool for feed creation
	tool, err := h.registry.Tools.Get(existingSheet.ToolID)
	if err != nil {
		return utils.HandleError(err, "failed to get tool from database")
	}

	// Delete the metal sheet from database
	if err := h.registry.MetalSheets.Delete(metalSheetID); err != nil {
		return utils.HandleError(err, "failed to delete metal sheet from database")
	}

	// Create feed entry for the deleted metal sheet
	h.createFeed(user, tool, existingSheet, "Blech gelöscht")

	utils.SetHXTrigger(c, env.HXGlobalTrigger)

	return nil
}

// createFeed creates a feed entry for metal sheet operations
func (h *MetalSheets) createFeed(user *models.User, tool *models.Tool, metalSheet *models.MetalSheet, title string) {
	// Build base feed content with tool and metal sheet info
	content := fmt.Sprintf("Werkzeug: %s\nStärke: %.1f mm\nBlech: %.1f mm\nTyp: %s",
		tool.String(), metalSheet.TileHeight, metalSheet.Value, metalSheet.Identifier.String())

	// Add additional fields for bottom position tools
	if tool.Position == models.PositionBottom {
		content += fmt.Sprintf("\nMarke: %d mm\nStf.: %.1f\nStf. Max: %.1f",
			metalSheet.MarkeHeight, metalSheet.STF, metalSheet.STFMax)
	}

	// Create and save the feed entry
	feed := models.NewFeed(title, content, user.TelegramID)
	if err := h.registry.Feeds.Add(feed); err != nil {
		slog.Error("Failed to create feed", "error", err)
	}
}

// createUpdateFeed creates a feed entry for metal sheet updates showing old vs new values
func (h *MetalSheets) createUpdateFeed(user *models.User, tool *models.Tool, oldSheet, newSheet *models.MetalSheet) {
	content := fmt.Sprintf("Werkzeug: %s", tool.String())

	// Check for changes in TileHeight
	if oldSheet.TileHeight != newSheet.TileHeight {
		content += fmt.Sprintf("\nStärke: %.1f mm → %.1f mm", oldSheet.TileHeight, newSheet.TileHeight)
	} else {
		content += fmt.Sprintf("\nStärke: %.1f mm", newSheet.TileHeight)
	}

	// Check for changes in Value
	if oldSheet.Value != newSheet.Value {
		content += fmt.Sprintf("\nBlech: %.1f mm → %.1f mm", oldSheet.Value, newSheet.Value)
	} else {
		content += fmt.Sprintf("\nBlech: %.1f mm", newSheet.Value)
	}

	// Check for changes in machine type
	if oldSheet.Identifier != newSheet.Identifier {
		content += fmt.Sprintf("\nTyp: %s → %s", oldSheet.Identifier.String(), newSheet.Identifier.String())
	} else {
		content += fmt.Sprintf("\nTyp: %s", newSheet.Identifier.String())
	}

	// Add additional fields for bottom position tools
	if tool.Position == models.PositionBottom {
		// Check for changes in MarkeHeight
		if oldSheet.MarkeHeight != newSheet.MarkeHeight {
			content += fmt.Sprintf("\nMarke: %d mm → %d mm", oldSheet.MarkeHeight, newSheet.MarkeHeight)
		} else {
			content += fmt.Sprintf("\nMarke: %d mm", newSheet.MarkeHeight)
		}

		// Check for changes in STF
		if oldSheet.STF != newSheet.STF {
			content += fmt.Sprintf("\nStf.: %.1f → %.1f", oldSheet.STF, newSheet.STF)
		} else {
			content += fmt.Sprintf("\nStf.: %.1f", newSheet.STF)
		}

		// Check for changes in STFMax
		if oldSheet.STFMax != newSheet.STFMax {
			content += fmt.Sprintf("\nStf. Max: %.1f → %.1f", oldSheet.STFMax, newSheet.STFMax)
		} else {
			content += fmt.Sprintf("\nStf. Max: %.1f", newSheet.STFMax)
		}
	}

	// Create and save the feed entry
	feed := models.NewFeed("Blech aktualisiert", content, user.TelegramID)
	if err := h.registry.Feeds.Add(feed); err != nil {
		slog.Error("Failed to create update feed", "error", err)
	}
}

// parseMetalSheetForm extracts metal sheet data from form submission
func (h *MetalSheets) parseMetalSheetForm(c echo.Context) (*models.MetalSheet, error) {
	metalSheet := &models.MetalSheet{}

	// Parse required tile height field
	tileHeight, err := strconv.ParseFloat(c.FormValue("tile_height"), 64)
	if err != nil {
		return nil, err
	}
	metalSheet.TileHeight = tileHeight

	// Parse required value field
	value, err := strconv.ParseFloat(c.FormValue("value"), 64)
	if err != nil {
		return nil, err
	}
	metalSheet.Value = value

	// Parse optional marke height field
	if markeHeightStr := c.FormValue("marke_height"); markeHeightStr != "" {
		if markeHeight, err := strconv.Atoi(markeHeightStr); err == nil {
			metalSheet.MarkeHeight = markeHeight
		}
	}

	// Parse optional STF field
	if stfStr := c.FormValue("stf"); stfStr != "" {
		if stf, err := strconv.ParseFloat(stfStr, 64); err == nil {
			metalSheet.STF = stf
		}
	}

	// Parse optional STF Max field
	if stfMaxStr := c.FormValue("stf_max"); stfMaxStr != "" {
		if stfMax, err := strconv.ParseFloat(stfMaxStr, 64); err == nil {
			metalSheet.STFMax = stfMax
		}
	}

	// Parse identifier field with validation
	identifierStr := c.FormValue("identifier")
	if machineType, err := models.ParseMachineType(identifierStr); err == nil {
		metalSheet.Identifier = machineType
	} else {
		// Log the invalid value but don't fail - default to SACMI
		metalSheet.Identifier = models.MachineTypeSACMI // Default to SACMI
	}

	return metalSheet, nil
}
