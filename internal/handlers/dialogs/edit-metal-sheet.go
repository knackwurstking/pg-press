package dialogs

import (
	"strconv"
	"strings"

	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/urlb"
	"github.com/labstack/echo/v4"
)

func GetEditMetalSheet(c echo.Context) *echo.HTTPError {
	// Check if we're editing an existing metal sheet (has ID) or creating new one
	idQuery, _ := shared.ParseQueryInt64(c, "id")
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

func parseUpperMetalSheetForm(c echo.Context, ums *shared.UpperMetalSheet) (*shared.UpperMetalSheet, *errors.MasterError) {
	if ums == nil {
		ums = &shared.UpperMetalSheet{}
	}

	var (
		tileHeightStr = c.FormValue("tile_height")
		valueStr      = c.FormValue("value")
		err           error
	)

	ums.TileHeight, err = strconv.ParseFloat(tileHeightStr, 64)
	if err != nil {
		return nil, errors.NewValidationError("invalid tile height: %v", err).MasterError()
	}

	ums.Value, err = strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return nil, errors.NewValidationError("invalid value: %v", err).MasterError()
	}

	return ums, nil
}

func parseLowerMetalSheetForm(c echo.Context, lms *shared.LowerMetalSheet) (*shared.LowerMetalSheet, *errors.MasterError) {
	if lms == nil {
		lms = &shared.LowerMetalSheet{}
	}

	var (
		tileHeightStr  = c.FormValue("tile_height")
		valueStr       = c.FormValue("value")
		markeHeightStr = c.FormValue("marke_height")
		stfStr         = c.FormValue("stf")
		stfMaxStr      = c.FormValue("stf_max")
		identifierStr  = strings.ToUpper(strings.TrimSpace(c.FormValue("identifier")))
		err            error
	)

	lms.TileHeight, err = strconv.ParseFloat(tileHeightStr, 64)
	if err != nil {
		return nil, errors.NewValidationError("invalid tile height: %v", err).MasterError()
	}

	lms.Value, err = strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return nil, errors.NewValidationError("invalid value: %v", err).MasterError()
	}

	lms.MarkeHeight, err = strconv.Atoi(markeHeightStr)
	if err != nil {
		return nil, errors.NewValidationError("invalid marke height: %v", err).MasterError()
	}

	lms.STF, err = strconv.ParseFloat(stfStr, 64)
	if err != nil {
		return nil, errors.NewValidationError("invalid STF value: %v", err).MasterError()
	}

	lms.STFMax, err = strconv.ParseFloat(stfMaxStr, 64)
	if err != nil {
		return nil, errors.NewValidationError("invalid STF max value: %v", err).MasterError()
	}

	switch v := shared.MachineType(identifierStr); v {
	case shared.MachineTypeSACMI, shared.MachineTypeSITI:
		lms.Identifier = v
	default:
		return nil, errors.NewValidationError("identifier must be either 'SACMI' or 'SITI'").MasterError()
	}

	return lms, nil
}
