package dialogs

import (
	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/env"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/urlb"
	"github.com/labstack/echo/v4"
)

func GetEditMetalSheet(c echo.Context) *echo.HTTPError {
	// Check if we're editing an existing metal sheet (has ID) or creating new one
	idQuery, merr := shared.ParseQueryInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}

	if metalSheetID := shared.EntityID(idQuery); metalSheetID > 0 {
		positionQuery, merr := shared.ParseQueryInt(c, "position")
		if merr != nil {
			return merr.Echo()
		}
		switch shared.Slot(positionQuery) {
		case shared.SlotUpper:
			metalSheet, merr := db.GetUpperMetalSheet(metalSheetID)
			if merr != nil {
				return merr.Echo()
			}
			tool, merr := db.GetTool(metalSheet.ToolID)
			if merr != nil {
				return merr.Echo()
			}

			t := EditUpperMetalSheetDialog(metalSheet, tool)
			err := t.Render(c.Request().Context(), c.Response())
			if err != nil {
				return errors.NewRenderError(err, "EditUpperMetalSheetDialog")
			}
			return nil

		case shared.SlotLower:
			metalSheet, merr := db.GetLowerMetalSheet(metalSheetID)
			if merr != nil {
				return merr.Echo()
			}
			tool, merr := db.GetTool(metalSheet.ToolID)
			if merr != nil {
				return merr.Echo()
			}

			t := EditLowerMetalSheetDialog(metalSheet, tool)
			err := t.Render(c.Request().Context(), c.Response())
			if err != nil {
				return errors.NewRenderError(err, "EditLowerMetalSheetDialog")
			}
			return nil

		default:
			return errors.NewValidationError("invalid slot position").MasterError().Echo()
		}
	}

	// Creating new metal sheet, get tool_id from query
	toolIDQuery, merr := shared.ParseQueryInt64(c, "tool_id")
	if merr != nil {
		return merr.Echo()
	}
	// Fetch the associated tool for the dialog
	tool, merr := db.GetTool(shared.EntityID(toolIDQuery))
	if merr != nil {
		return merr.Echo()
	}

	switch tool.Position {
	case shared.SlotUpper:
		t := NewUpperMetalSheetDialog(tool)
		err := t.Render(c.Request().Context(), c.Response())
		if err != nil {
			return errors.NewRenderError(err, "NewUpperMetalSheetDialog")
		}
		return nil

	case shared.SlotLower:
		t := NewLowerMetalSheetDialog(tool)
		err := t.Render(c.Request().Context(), c.Response())
		if err != nil {
			return errors.NewRenderError(err, "NewLowerMetalSheetDialog")
		}
		return nil

	default:
		return errors.NewValidationError("invalid slot position").MasterError().Echo()
	}
}

func PostMetalSheet(c echo.Context) *echo.HTTPError {
	// Extract tool ID from query parameters
	id, merr := shared.ParseQueryInt64(c, "tool_id")
	if merr != nil {
		return merr.Echo()
	}
	// Fetch the associated tool
	tool, merr := db.GetTool(shared.EntityID(id))
	if merr != nil {
		return merr.Echo()
	}

	switch tool.Position {
	case shared.SlotUpper:
		ums, merr := parseUpperMetalSheetForm(c, nil)
		if merr != nil {
			return merr.Wrap("could not parse upper metal sheet form").Echo()
		}
		ums.ToolID = tool.ID
		merr = db.AddUpperMetalSheet(ums)
		if merr != nil {
			return merr.Wrap("could not add upper metal sheet to database").Echo()
		}

	case shared.SlotLower:
		lms, merr := parseLowerMetalSheetForm(c, nil)
		if merr != nil {
			return merr.Wrap("could not parse lower metal sheet form").Echo()
		}
		lms.ToolID = tool.ID
		merr = db.AddLowerMetalSheet(lms)
		if merr != nil {
			return merr.Wrap("could not add lower metal sheet to database").Echo()
		}

	default:
		return errors.NewValidationError("invalid tool position").MasterError().Echo()
	}

	urlb.SetHXTrigger(c, "reload-metal-sheets")

	return nil
}

func PutMetalSheet(c echo.Context) *echo.HTTPError {
	// Extract metal sheet ID from query parameters
	id, merr := shared.ParseQueryInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	position, merr := shared.ParseQueryInt(c, "position")
	if merr != nil {
		return merr.Echo()
	}
	switch shared.Slot(position) {
	case shared.SlotUpper:
		ums, merr := db.GetUpperMetalSheet(shared.EntityID(id))
		if merr != nil {
			return merr.Wrap("could not fetch existing upper metal sheet").Echo()
		}
		ums, merr = parseUpperMetalSheetForm(c, ums)
		if merr != nil {
			return merr.Wrap("could not parse upper metal sheet form").Echo()
		}
		merr = db.UpdateUpperMetalSheet(ums)
		if merr != nil {
			return merr.Wrap("could not update upper metal sheet in database").Echo()
		}

	case shared.SlotLower:
		lms, merr := db.GetLowerMetalSheet(shared.EntityID(id))
		if merr != nil {
			return merr.Wrap("could not fetch existing upper metal sheet").Echo()
		}
		lms, merr = parseLowerMetalSheetForm(c, lms)
		if merr != nil {
			return merr.Wrap("could not parse lower metal sheet form").Echo()
		}
		merr = db.UpdateLowerMetalSheet(lms)
		if merr != nil {
			return merr.Wrap("could not update upper metal sheet in database").Echo()
		}

	default:
		return errors.NewValidationError("invalid slot position").MasterError().Echo()
	}

	urlb.SetHXTrigger(c, "reload-metal-sheets")

	return nil
}

func parseUpperMetalSheetForm(c echo.Context, ms *shared.UpperMetalSheet) (*shared.UpperMetalSheet, *errors.MasterError) {
	var (
		tileHeightStr = c.FormValue("tile_height")
		valueStr      = c.FormValue("value")
	)

	// ...
}

func parseLowerMetalSheetForm(c echo.Context, ms *shared.LowerMetalSheet) (*shared.LowerMetalSheet, *errors.MasterError) {
	// TODO: ...
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
