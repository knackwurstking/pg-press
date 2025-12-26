package dialogs

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/labstack/echo/v4"
)

func GetEditMetalSheet(c echo.Context) *echo.HTTPError {
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

		t := EditMetalSheetDialog(metalSheet, tool)
		err := t.Render(c.Request().Context(), c.Response())
		if err != nil {
			return errors.NewRenderError(err, "EditMetalSheetDialog")
		}
		return nil
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

	t := NewMetalSheetDialog(tool)
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "NewMetalSheetDialog")
	}
	return nil
}

func PostMetalSheet(c echo.Context) *echo.HTTPError {
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
	metalSheet, merr := GetMetalSheetDialogForm(c)
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

func PutMetalSheet(c echo.Context) *echo.HTTPError {
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
	metalSheet, merr := GetMetalSheetDialogForm(c)
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

	utils.SetHXTrigger(c, env.HXGlobalTrigger)

	return nil
}

func parseUpperMetalSheetForm(c echo.Context, ms *shared.UpperMetalSheet) (*shared.UpperMetalSheet, *errors.MasterError) {
	// ...
}

func parseLowerMetalSheetForm(c echo.Context, ms *shared.LowerMetalSheet) (*shared.LowerMetalSheet, *errors.MasterError) {
	// ...
}

//func getMetalSheetDialogForm(c echo.Context) (*shared.MetalSheet, *errors.MasterError) {
//	metalSheet := &models.MetalSheet{}
//
//	// Parse required tile height field
//	tileHeight, err := strconv.ParseFloat(c.FormValue("tile_height"), 64)
//	if err != nil {
//		return nil, errors.NewMasterError(err, http.StatusBadRequest)
//	}
//	metalSheet.TileHeight = tileHeight
//
//	// Parse required value field
//	value, err := strconv.ParseFloat(c.FormValue("value"), 64)
//	if err != nil {
//		return nil, errors.NewMasterError(err, http.StatusBadRequest)
//	}
//	metalSheet.Value = value
//
//	// Parse optional marke height field
//	if markeHeightStr := c.FormValue("marke_height"); markeHeightStr != "" {
//		if markeHeight, err := strconv.Atoi(markeHeightStr); err == nil {
//			metalSheet.MarkeHeight = markeHeight
//		}
//	}
//
//	// Parse optional STF field
//	if stfStr := c.FormValue("stf"); stfStr != "" {
//		if stf, err := strconv.ParseFloat(stfStr, 64); err == nil {
//			metalSheet.STF = stf
//		}
//	}
//
//	// Parse optional STF Max field
//	if stfMaxStr := c.FormValue("stf_max"); stfMaxStr != "" {
//		if stfMax, err := strconv.ParseFloat(stfMaxStr, 64); err == nil {
//			metalSheet.STFMax = stfMax
//		}
//	}
//
//	// Parse identifier field with validation
//	identifierStr := c.FormValue("identifier")
//	if machineType, err := models.ParseMachineType(identifierStr); err == nil {
//		metalSheet.Identifier = machineType
//	} else {
//		// Log the invalid value but don't fail - default to SACMI
//		metalSheet.Identifier = models.MachineTypeSACMI // Default to SACMI
//	}
//
//	return metalSheet, nil
//}
