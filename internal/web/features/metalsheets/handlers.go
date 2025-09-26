package metalsheets

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/web/shared/dialogs"
	"github.com/knackwurstking/pgpress/internal/web/shared/handlers"
	"github.com/knackwurstking/pgpress/pkg/logger"
	"github.com/knackwurstking/pgpress/pkg/models"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	*handlers.BaseHandler
}

func NewHandler(db *database.DB) *Handler {
	return &Handler{
		BaseHandler: handlers.NewBaseHandler(db, logger.NewComponentLogger("Metal Sheets")),
	}
}

// GetEditDialog renders the edit/create dialog for metal sheets
func (h *Handler) HTMXGetEditMetalSheetDialog(c echo.Context) error {
	renderProps := &dialogs.EditMetalSheetProps{}
	var toolID int64
	var err error

	// Check if we're editing an existing metal sheet (has ID) or creating new one
	if metalSheetID, _ := h.ParseInt64Query(c, "id"); metalSheetID > 0 {
		// Fetch existing metal sheet for editing
		if renderProps.MetalSheet, err = h.DB.MetalSheets.Get(metalSheetID); err != nil {
			return h.HandleError(c, err, "failed to fetch metal sheet from database")
		}
		toolID = renderProps.MetalSheet.ToolID
	} else {
		// Creating new metal sheet, get tool_id from query
		if toolID, err = h.ParseInt64Query(c, "tool_id"); err != nil {
			return h.HandleError(c, err, "failed to get the tool id from query")
		}
	}

	// Fetch the associated tool for the dialog
	if renderProps.Tool, err = h.DB.Tools.Get(toolID); err != nil {
		return h.HandleError(c, err, "failed to get tool from database")
	}

	// Render the edit dialog template
	if err := dialogs.EditMetalSheet(renderProps).Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c, "failed to render edit metal sheet dialog: "+err.Error())
	}

	return nil
}

// PostCreateMetalSheet handles the creation of a new metal sheet
func (h *Handler) HTMXPostEditMetalSheetDialog(c echo.Context) error {
	// Get current user for feed creation
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	// Extract tool ID from query parameters
	toolID, err := h.ParseInt64Query(c, "tool_id")
	if err != nil {
		return h.HandleError(c, err, "failed to get tool_id from query")
	}

	// Fetch the associated tool
	tool, err := h.DB.Tools.Get(toolID)
	if err != nil {
		return h.HandleError(c, err, "failed to get tool from database")
	}

	// Parse form data into metal sheet model
	metalSheet, err := h.parseMetalSheetForm(c)
	if err != nil {
		return h.HandleError(c, err, "failed to parse metal sheet form data")
	}

	// Associate metal sheet with the tool
	metalSheet.ToolID = toolID

	// Save new metal sheet to database
	if _, err := h.DB.MetalSheets.Add(metalSheet); err != nil {
		return h.HandleError(c, err, "failed to create metal sheet in database")
	}

	// Create feed entry for the new metal sheet
	h.createFeed(user, tool, metalSheet, "Blech erstellt")
	// Refresh the page via HTMX
	c.Response().Header().Set("HX-Refresh", "true")
	return c.NoContent(http.StatusOK)
}

// PutUpdateMetalSheet handles updates to existing metal sheets
func (h *Handler) HTMXPutEditMetalSheetDialog(c echo.Context) error {
	// Get current user for feed creation
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	// Extract metal sheet ID from query parameters
	metalSheetID, err := h.ParseInt64Query(c, "id")
	if err != nil {
		return h.HandleError(c, err, "failed to get id from query")
	}

	// Fetch the existing metal sheet to preserve ID and tool association
	existingSheet, err := h.DB.MetalSheets.Get(metalSheetID)
	if err != nil {
		return h.HandleError(c, err, "failed to get existing metal sheet from database")
	}

	// Fetch the associated tool for feed creation
	tool, err := h.DB.Tools.Get(existingSheet.ToolID)
	if err != nil {
		return h.HandleError(c, err, "failed to get tool from database")
	}

	// Parse updated form data
	metalSheet, err := h.parseMetalSheetForm(c)
	if err != nil {
		return h.HandleError(c, err, "failed to parse metal sheet form data")
	}

	// Preserve the original ID and tool association
	metalSheet.ID = existingSheet.ID
	metalSheet.ToolID = existingSheet.ToolID

	// Update the metal sheet in database
	if err := h.DB.MetalSheets.Update(metalSheet); err != nil {
		return h.HandleError(c, err, "failed to update metal sheet in database")
	}

	// Create feed entry for the updated metal sheet showing changes
	h.createUpdateFeed(user, tool, existingSheet, metalSheet)
	// Refresh the page via HTMX
	c.Response().Header().Set("HX-Refresh", "true")
	return c.NoContent(http.StatusOK)
}

// DeleteMetalSheet handles the deletion of metal sheets
func (h *Handler) HTMXDeleteMetalSheet(c echo.Context) error {
	// Get current user for feed creation
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	// Extract metal sheet ID from query parameters
	metalSheetID, err := h.ParseInt64Query(c, "id")
	if err != nil {
		return h.HandleError(c, err, "failed to get id from query")
	}

	// Fetch the existing metal sheet before deletion for feed creation
	existingSheet, err := h.DB.MetalSheets.Get(metalSheetID)
	if err != nil {
		return h.HandleError(c, err, "failed to get existing metal sheet from database")
	}

	// Fetch the associated tool for feed creation
	tool, err := h.DB.Tools.Get(existingSheet.ToolID)
	if err != nil {
		return h.HandleError(c, err, "failed to get tool from database")
	}

	// Delete the metal sheet from database
	if err := h.DB.MetalSheets.Delete(metalSheetID); err != nil {
		return h.HandleError(c, err, "failed to delete metal sheet from database")
	}

	// Create feed entry for the deleted metal sheet
	h.createFeed(user, tool, existingSheet, "Blech gelöscht")
	return c.NoContent(http.StatusOK)
}

// createFeed creates a feed entry for metal sheet operations
func (h *Handler) createFeed(user *models.User, tool *models.Tool, metalSheet *models.MetalSheet, title string) {
	// Build base feed content with tool and metal sheet info
	content := fmt.Sprintf("Werkzeug: %s\nStärke: %.1f mm\nBlech: %.1f mm",
		tool.String(), metalSheet.TileHeight, metalSheet.Value)

	// Add additional fields for bottom position tools
	if tool.Position == models.PositionBottom {
		content += fmt.Sprintf("\nMarke: %d mm\nStf.: %.1f\nStf. Max: %.1f",
			metalSheet.MarkeHeight, metalSheet.STF, metalSheet.STFMax)
	}

	// Create and save the feed entry
	feed := models.NewFeed(title, content, user.TelegramID)
	if err := h.DB.Feeds.Add(feed); err != nil {
		h.LogError("Failed to create feed: %v", err)
	}
}

// createUpdateFeed creates a feed entry for metal sheet updates showing old vs new values
func (h *Handler) createUpdateFeed(user *models.User, tool *models.Tool, oldSheet, newSheet *models.MetalSheet) {
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
	if err := h.DB.Feeds.Add(feed); err != nil {
		h.LogError("Failed to create update feed: %v", err)
	}
}

// parseMetalSheetForm extracts metal sheet data from form submission
func (h *Handler) parseMetalSheetForm(c echo.Context) (*models.MetalSheet, error) {
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

	return metalSheet, nil
}
