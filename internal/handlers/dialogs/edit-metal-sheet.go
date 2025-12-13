package dialogs

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/knackwurstking/pg-press/internal/env"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/dialogs/templates"
	"github.com/knackwurstking/pg-press/models"
	"github.com/knackwurstking/pg-press/utils"

	"github.com/labstack/echo/v4"
)

func (h *Handler) GetEditMetalSheet(c echo.Context) *echo.HTTPError {
	// Check if we're editing an existing metal sheet (has ID) or creating new one
	metalSheetIDQuery, _ := utils.ParseQueryInt64(c, "id")
	if metalSheetIDQuery > 0 {
		metalSheetID := models.MetalSheetID(metalSheetIDQuery)

		// Fetch existing metal sheet for editing
		metalSheet, merr := h.registry.MetalSheets.Get(metalSheetID)
		if merr != nil {
			return merr.Echo()
		}

		// Fetch the associated tool for the dialog
		tool, merr := h.registry.Tools.Get(metalSheet.ToolID)
		if merr != nil {
			return merr.Echo()
		}

		t := templates.EditMetalSheetDialog(metalSheet, tool)
		err := t.Render(c.Request().Context(), c.Response())
		if err != nil {
			return errors.NewRenderError(err, "EditMetalSheetDialog")
		}
	}

	// Creating new metal sheet, get tool_id from query
	toolIDQuery, merr := utils.ParseQueryInt64(c, "tool_id")
	if merr != nil {
		return merr.Echo()
	}
	toolID := models.ToolID(toolIDQuery)

	// Fetch the associated tool for the dialog
	tool, merr := h.registry.Tools.Get(toolID)
	if merr != nil {
		return merr.Echo()
	}

	t := templates.NewMetalSheetDialog(tool)
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "NewMetalSheetDialog")
	}

	return nil
}

func (h *Handler) PostEditMetalSheet(c echo.Context) *echo.HTTPError {
	slog.Info("Metal sheet creation request received")

	// Get current user for feed creation
	user, merr := utils.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	// Extract tool ID from query parameters
	toolIDQuery, merr := utils.ParseQueryInt64(c, "tool_id")
	if merr != nil {
		return merr.Echo()
	}
	toolID := models.ToolID(toolIDQuery)

	// Fetch the associated tool
	tool, merr := h.registry.Tools.Get(toolID)
	if merr != nil {
		return merr.Echo()
	}

	// Parse form data into metal sheet model
	metalSheet, merr := GetMetalSheetFormData(c)
	if merr != nil {
		return merr.Echo()
	}

	// Associate metal sheet with the tool
	metalSheet.ToolID = toolID

	// Save new metal sheet to database
	_, merr = h.registry.MetalSheets.Add(metalSheet)
	if merr != nil {
		return merr.Echo()
	}

	h.createNewMetalSheetFeed(user, tool, metalSheet)

	utils.SetHXTrigger(c, env.HXGlobalTrigger)

	return nil
}

func (h *Handler) PutEditMetalSheet(c echo.Context) *echo.HTTPError {
	slog.Info("Updating metal sheet")

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

	// Fetch the existing metal sheet to preserve ID and tool association
	existingSheet, merr := h.registry.MetalSheets.Get(metalSheetID)
	if merr != nil {
		return merr.Echo()
	}

	// Fetch the associated tool for feed creation
	tool, merr := h.registry.Tools.Get(existingSheet.ToolID)
	if merr != nil {
		return merr.Echo()
	}

	// Parse updated form data
	metalSheet, merr := GetMetalSheetFormData(c)
	if merr != nil {
		return merr.Echo()
	}

	// Preserve the original ID and tool association
	metalSheet.ID = existingSheet.ID
	metalSheet.ToolID = existingSheet.ToolID

	// Update the metal sheet in database
	merr = h.registry.MetalSheets.Update(metalSheet)
	if merr != nil {
		return merr.Echo()
	}

	h.createUpdateMetalSheetFeed(user, tool, existingSheet, metalSheet)

	utils.SetHXTrigger(c, env.HXGlobalTrigger)

	return nil
}

func (h *Handler) createNewMetalSheetFeed(user *models.User, tool *models.Tool, metalSheet *models.MetalSheet) {
	// Build base feed content with tool and metal sheet info
	content := fmt.Sprintf("Werkzeug: %s\nStärke: %.1f mm\nBlech: %.1f mm\nTyp: %s",
		tool.String(), metalSheet.TileHeight, metalSheet.Value, metalSheet.Identifier.String())

	// Add additional fields for bottom position tools
	if tool.Position == models.PositionBottom {
		content += fmt.Sprintf("\nMarke: %d mm\nStf.: %.1f\nStf. Max: %.1f",
			metalSheet.MarkeHeight, metalSheet.STF, metalSheet.STFMax)
	}

	// Create and save the feed entry
	dberr := h.registry.Feeds.Add("Blech erstellt", content, user.TelegramID)
	if dberr != nil {
		slog.Warn("Failed to create feed", "error", dberr)
	}
}

func (h *Handler) createUpdateMetalSheetFeed(user *models.User, tool *models.Tool, oldSheet, newSheet *models.MetalSheet) {
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
	dberr := h.registry.Feeds.Add("Blech aktualisiert", content, user.TelegramID)
	if dberr != nil {
		slog.Warn("Failed to create update feed", "error", dberr)
	}
}

func GetMetalSheetFormData(c echo.Context) (*models.MetalSheet, *errors.MasterError) {
	metalSheet := &models.MetalSheet{}

	// Parse required tile height field
	tileHeight, err := strconv.ParseFloat(c.FormValue("tile_height"), 64)
	if err != nil {
		return nil, errors.NewMasterError(err, http.StatusBadRequest)
	}
	metalSheet.TileHeight = tileHeight

	// Parse required value field
	value, err := strconv.ParseFloat(c.FormValue("value"), 64)
	if err != nil {
		return nil, errors.NewMasterError(err, http.StatusBadRequest)
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
