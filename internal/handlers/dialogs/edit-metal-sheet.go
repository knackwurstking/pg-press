package dialogs

import (
	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/utils"

	"github.com/labstack/echo/v4"
)

func GetEditMetalSheet(c echo.Context) *echo.HTTPError {
	// Check if we're editing an existing metal sheet (has ID) or creating new one
	idQuery, _ := utils.GetQueryInt64(c, "id")
	if metalSheetID := shared.EntityID(idQuery); metalSheetID > 0 {
		positionQuery, merr := utils.GetQueryInt(c, "position")
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
			return errors.NewValidationError("invalid slot position").HTTPError().Echo()
		}
	}

	// Creating new metal sheet, get tool_id from query
	toolIDQuery, merr := utils.GetQueryInt64(c, "tool_id")
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
		return errors.NewValidationError("invalid slot position").HTTPError().Echo()
	}
}

func PostMetalSheet(c echo.Context) *echo.HTTPError {
	// Extract tool ID from query parameters
	id, merr := utils.GetQueryInt64(c, "tool_id")
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
		ums, verr := parseUpperMetalSheetForm(c, nil)
		if verr != nil {
			return verr.HTTPError().Echo()
		}
		ums.ToolID = tool.ID
		merr = db.AddUpperMetalSheet(ums)
		if merr != nil {
			return merr.Wrap("could not add upper metal sheet to database").Echo()
		}

	case shared.SlotLower:
		lms, verr := parseLowerMetalSheetForm(c, nil)
		if verr != nil {
			return verr.HTTPError().Echo()
		}
		lms.ToolID = tool.ID
		merr = db.AddLowerMetalSheet(lms)
		if merr != nil {
			return merr.Wrap("could not add lower metal sheet to database").Echo()
		}

	default:
		return errors.NewValidationError("invalid tool position").HTTPError().Echo()
	}

	utils.SetHXTrigger(c, "reload-metal-sheets")

	return nil
}

func PutMetalSheet(c echo.Context) *echo.HTTPError {
	// Extract metal sheet ID from query parameters
	id, merr := utils.GetQueryInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	position, merr := utils.GetQueryInt(c, "position")
	if merr != nil {
		return merr.Echo()
	}
	switch shared.Slot(position) {
	case shared.SlotUpper:
		ums, merr := db.GetUpperMetalSheet(shared.EntityID(id))
		if merr != nil {
			return merr.Wrap("could not fetch existing upper metal sheet").Echo()
		}
		ums, verr := parseUpperMetalSheetForm(c, ums)
		if verr != nil {
			return verr.HTTPError().Echo()
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
		lms, verr := parseLowerMetalSheetForm(c, lms)
		if verr != nil {
			return verr.HTTPError().Echo()
		}
		merr = db.UpdateLowerMetalSheet(lms)
		if merr != nil {
			return merr.Wrap("could not update upper metal sheet in database").Echo()
		}

	default:
		return errors.NewValidationError("invalid slot position").HTTPError().Echo()
	}

	utils.SetHXTrigger(c, "reload-metal-sheets")

	return nil
}

func parseUpperMetalSheetForm(c echo.Context, ums *shared.UpperMetalSheet) (*shared.UpperMetalSheet, *errors.ValidationError) {
	if ums == nil {
		ums = &shared.UpperMetalSheet{}
	}

	var err error
	ums.TileHeight, err = utils.SanitizeFloat(c.FormValue("tile_height"))
	if err != nil {
		return nil, errors.NewValidationError("invalid tile height: %v", err)
	}

	ums.Value, err = utils.SanitizeFloat(c.FormValue("value"))
	if err != nil {
		return nil, errors.NewValidationError("invalid value: %v", err)
	}

	return ums, nil
}

func parseLowerMetalSheetForm(c echo.Context, lms *shared.LowerMetalSheet) (*shared.LowerMetalSheet, *errors.ValidationError) {
	if lms == nil {
		lms = &shared.LowerMetalSheet{}
	}

	var err error
	lms.TileHeight, err = utils.SanitizeFloat(c.FormValue("tile_height"))
	if err != nil {
		return nil, errors.NewValidationError("invalid tile height: %v", err)
	}

	lms.Value, err = utils.SanitizeFloat(c.FormValue("value"))
	if err != nil {
		return nil, errors.NewValidationError("invalid value: %v", err)
	}

	lms.MarkeHeight, err = utils.SanitizeInt(c.FormValue("marke_height"))
	if err != nil {
		return nil, errors.NewValidationError("invalid marke height: %v", err)
	}

	lms.STF, err = utils.SanitizeFloat(c.FormValue("stf"))
	if err != nil {
		return nil, errors.NewValidationError("invalid STF value: %v", err)
	}

	lms.STFMax, err = utils.SanitizeFloat(c.FormValue("stf_max"))
	if err != nil {
		return nil, errors.NewValidationError("invalid STF max value: %v", err)
	}

	switch v := shared.MachineType(utils.SanitizeText(c.FormValue("identifier"))); v {
	case shared.MachineTypeSACMI, shared.MachineTypeSITI:
		lms.Identifier = v
	default:
		return nil, errors.NewValidationError("identifier must be either 'SACMI' or 'SITI'")
	}

	return lms, nil
}
