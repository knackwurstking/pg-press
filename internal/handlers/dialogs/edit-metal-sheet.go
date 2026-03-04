package dialogs

import (
	"fmt"
	"net/http"

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

			t := EditUpperMetalSheetDialog(metalSheet.ID, UpperMetalSheetDialogProps{
				UpperMetalSheetFormData: UpperMetalSheetFormData{
					TileHeight: metalSheet.TileHeight,
					Value:      metalSheet.Value,
				},
				ToolID:       metalSheet.ToolID,
				ToolPosition: tool.Position,
				Open:         true,
				OOB:          true,
			})
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

			t := EditLowerMetalSheetDialog(metalSheet.ID, LowerMetalSheetDialogProps{
				LowerMetalSheetFormData: LowerMetalSheetFormData{
					Identifier:  metalSheet.Identifier,
					TileHeight:  metalSheet.TileHeight,
					Value:       metalSheet.Value,
					MarkeHeight: metalSheet.MarkeHeight,
					STF:         metalSheet.STF,
					STFMax:      metalSheet.STFMax,
				},
				ToolID:       metalSheet.ToolID,
				ToolPosition: tool.Position,
				Open:         true,
				OOB:          true,
			})
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
		t := NewUpperMetalSheetDialog(UpperMetalSheetDialogProps{
			ToolID:       tool.ID,
			ToolPosition: tool.Position,
			Open:         true,
			OOB:          true,
		})
		err := t.Render(c.Request().Context(), c.Response())
		if err != nil {
			return errors.NewRenderError(err, "NewUpperMetalSheetDialog")
		}
		return nil

	case shared.SlotLower:
		t := NewLowerMetalSheetDialog(LowerMetalSheetDialogProps{
			ToolID:       tool.ID,
			ToolPosition: tool.Position,
			Open:         true,
			OOB:          true,
		})
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
	// Extract metal sheet ID from query parameters
	id, _ := utils.GetQueryInt64(c, "id")
	if id > 0 {
		return updateMetalSheet(c, shared.EntityID(id))
	}

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
		data, ierrs := parseUpperMetalSheetForm(c)
		if len(ierrs) > 0 {
			return reRenderNewUpperMetalSheetDialog(tool.ID, data, renderProps{c, true, ierrs})
		}
		ums := &shared.UpperMetalSheet{
			BaseMetalSheet: shared.BaseMetalSheet{
				ToolID:     tool.ID,
				TileHeight: data.TileHeight,
				Value:      data.Value,
			},
		}
		if merr = db.AddUpperMetalSheet(ums); merr != nil {
			ierr := errors.NewInputError("", fmt.Sprintf("could not add upper metal sheet to database: %v", merr))
			return reRenderNewUpperMetalSheetDialog(tool.ID, data, renderProps{c, true, []*errors.InputError{ierr}})
		}

	case shared.SlotLower:
		data, ierrs := parseLowerMetalSheetForm(c)
		if len(ierrs) > 0 {
			return reRenderNewLowerMetalSheetDialog(tool.ID, data, renderProps{c, true, ierrs})
		}
		lms := &shared.LowerMetalSheet{
			BaseMetalSheet: shared.BaseMetalSheet{
				ToolID:     tool.ID,
				TileHeight: data.TileHeight,
				Value:      data.Value,
			},
			MarkeHeight: data.MarkeHeight,
			STF:         data.STF,
			STFMax:      data.STFMax,
			Identifier:  data.Identifier,
		}
		if merr = db.AddLowerMetalSheet(lms); merr != nil {
			ierr := errors.NewInputError("", fmt.Sprintf("could not add lower metal sheet to database: %v", merr))
			return reRenderNewLowerMetalSheetDialog(tool.ID, data, renderProps{c, true, []*errors.InputError{ierr}})
		}

	default:
		return errors.NewValidationError("invalid tool position").HTTPError().Echo()
	}

	utils.SetHXTrigger(c, "reload-metal-sheets")

	// Close dialog
	switch tool.Position {
	case shared.SlotUpper:
		return reRenderNewUpperMetalSheetDialog(tool.ID, UpperMetalSheetFormData{}, renderProps{c, false, nil})
	case shared.SlotLower:
		return reRenderNewLowerMetalSheetDialog(tool.ID, LowerMetalSheetFormData{}, renderProps{c, false, nil})
	}

	return nil
}

func updateMetalSheet(c echo.Context, id shared.EntityID) *echo.HTTPError {
	position, merr := utils.GetQueryInt(c, "position")
	if merr != nil {
		return merr.Echo()
	}

	switch shared.Slot(position) {
	case shared.SlotUpper:
		ums, merr := db.GetUpperMetalSheet(shared.EntityID(id))
		if merr != nil {
			return merr.WrapEcho("could not fetch existing upper metal sheet with ID %d", id)
		}

		data, ierrs := parseUpperMetalSheetForm(c)
		if len(ierrs) > 0 {
			return reRenderEditUpperMetalSheetDialog(ums.ToolID, id, data, renderProps{c, true, ierrs})
		}

		if merr = db.UpdateUpperMetalSheet(ums); merr != nil {
			ierr := errors.NewInputError("", fmt.Sprintf("could not update upper metal sheet in database: %v", merr))
			return reRenderEditUpperMetalSheetDialog(ums.ToolID, id, data, renderProps{c, true, []*errors.InputError{ierr}})
		}

	case shared.SlotLower:
		lms, merr := db.GetLowerMetalSheet(shared.EntityID(id))
		if merr != nil {
			return merr.WrapEcho("could not fetch existing lower metal sheet with ID %d", id)
		}

		data, ierrs := parseLowerMetalSheetForm(c)
		if len(ierrs) > 0 {
			return reRenderEditLowerMetalSheetDialog(lms.ToolID, id, data, renderProps{c, true, ierrs})
		}

		merr = db.UpdateLowerMetalSheet(lms)
		if merr != nil {
			ierr := errors.NewInputError("", fmt.Sprintf("could not update lower metal sheet in database: %v", merr))
			return reRenderEditLowerMetalSheetDialog(lms.ToolID, id, data, renderProps{c, true, []*errors.InputError{ierr}})
		}

	default:
		return errors.NewValidationError("invalid slot position").HTTPError().Echo()
	}

	utils.SetHXTrigger(c, "reload-metal-sheets")

	// Close dialog
	switch shared.Slot(position) {
	case shared.SlotUpper:
		return reRenderEditUpperMetalSheetDialog(0, id, UpperMetalSheetFormData{}, renderProps{c, false, nil})
	case shared.SlotLower:
		return reRenderEditLowerMetalSheetDialog(0, id, LowerMetalSheetFormData{}, renderProps{c, false, nil})
	}

	return nil
}

func parseUpperMetalSheetForm(c echo.Context) (data UpperMetalSheetFormData, ierrs []*errors.InputError) {
	var err error
	data.TileHeight, err = utils.SanitizeFloat(c.FormValue("tile_height"))
	if err != nil {
		ierr := errors.NewInputError("tile_height", fmt.Sprintf("invalid tile height: %v", err))
		ierrs = append(ierrs, ierr)
	}

	data.Value, err = utils.SanitizeFloat(c.FormValue("value"))
	if err != nil {
		ierr := errors.NewInputError("value", fmt.Sprintf("invalid value: %v", err))
		ierrs = append(ierrs, ierr)
	}

	return
}

func parseLowerMetalSheetForm(c echo.Context) (data LowerMetalSheetFormData, ierrs []*errors.InputError) {
	var err error
	data.TileHeight, err = utils.SanitizeFloat(c.FormValue("tile_height"))
	if err != nil {
		ierr := errors.NewInputError("tile_height", fmt.Sprintf("invalid tile height: %v", err))
		ierrs = append(ierrs, ierr)
	}

	data.Value, err = utils.SanitizeFloat(c.FormValue("value"))
	if err != nil {
		ierr := errors.NewInputError("value", fmt.Sprintf("invalid value: %v", err))
		ierrs = append(ierrs, ierr)
	}

	data.MarkeHeight, err = utils.SanitizeInt(c.FormValue("marke_height"))
	if err != nil {
		ierr := errors.NewInputError("marke_height", fmt.Sprintf("invalid marke height: %v", err))
		ierrs = append(ierrs, ierr)
	}

	data.STF, err = utils.SanitizeFloat(c.FormValue("stf"))
	if err != nil {
		ierr := errors.NewInputError("stf", fmt.Sprintf("invalid STF value: %v", err))
		ierrs = append(ierrs, ierr)
	}

	data.STFMax, err = utils.SanitizeFloat(c.FormValue("stf_max"))
	if err != nil {
		ierr := errors.NewInputError("stf_max", fmt.Sprintf("invalid STF max value: %v", err))
		ierrs = append(ierrs, ierr)
	}

	switch v := shared.MachineType(utils.SanitizeText(c.FormValue("machine_type"))); v {
	case shared.MachineTypeSACMI, shared.MachineTypeSITI:
		data.Identifier = v
	default:
		ierr := errors.NewInputError("machine_type", fmt.Sprintf("invalid machine type machine_type: %v", v))
		ierrs = append(ierrs, ierr)
	}

	return
}

type renderProps struct {
	c     echo.Context
	Open  bool
	Error []*errors.InputError
}

func reRenderNewUpperMetalSheetDialog(toolID shared.EntityID, data UpperMetalSheetFormData, prop renderProps) *echo.HTTPError {
	t := NewUpperMetalSheetDialog(UpperMetalSheetDialogProps{
		UpperMetalSheetFormData: data,
		ToolID:                  toolID,
		ToolPosition:            shared.SlotUpper,
		Open:                    prop.Open,
		OOB:                     true,
		Error:                   prop.Error,
	})
	if err := t.Render(prop.c.Request().Context(), prop.c.Response()); err != nil {
		return errors.NewRenderError(err, "NewUpperMetalSheetDialog")
	}
	if len(prop.Error) > 0 {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid input"))
	}
	return nil
}

func reRenderNewLowerMetalSheetDialog(toolID shared.EntityID, data LowerMetalSheetFormData, prop renderProps) *echo.HTTPError {
	t := NewLowerMetalSheetDialog(LowerMetalSheetDialogProps{
		LowerMetalSheetFormData: data,
		ToolID:                  toolID,
		ToolPosition:            shared.SlotLower,
		Open:                    prop.Open,
		OOB:                     true,
		Error:                   prop.Error,
	})
	if err := t.Render(prop.c.Request().Context(), prop.c.Response()); err != nil {
		return errors.NewRenderError(err, "NewLowerMetalSheetDialog")
	}
	if len(prop.Error) > 0 {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid input"))
	}
	return nil
}

func reRenderEditUpperMetalSheetDialog(toolID shared.EntityID, msID shared.EntityID, data UpperMetalSheetFormData, prop renderProps) *echo.HTTPError {
	t := EditUpperMetalSheetDialog(msID, UpperMetalSheetDialogProps{
		UpperMetalSheetFormData: data,
		ToolID:                  toolID,
		ToolPosition:            shared.SlotUpper,
		Open:                    prop.Open,
		OOB:                     true,
		Error:                   prop.Error,
	})
	if err := t.Render(prop.c.Request().Context(), prop.c.Response()); err != nil {
		return errors.NewRenderError(err, "EditUpperMetalSheetDialog")
	}
	if len(prop.Error) > 0 {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid input"))
	}
	return nil
}

func reRenderEditLowerMetalSheetDialog(toolID shared.EntityID, msID shared.EntityID, data LowerMetalSheetFormData, prop renderProps) *echo.HTTPError {
	t := EditLowerMetalSheetDialog(msID, LowerMetalSheetDialogProps{
		LowerMetalSheetFormData: data,
		ToolID:                  toolID,
		ToolPosition:            shared.SlotLower,
		Open:                    prop.Open,
		OOB:                     true,
		Error:                   prop.Error,
	})
	if err := t.Render(prop.c.Request().Context(), prop.c.Response()); err != nil {
		return errors.NewRenderError(err, "EditLowerMetalSheetDialog")
	}
	if len(prop.Error) > 0 {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid input"))
	}
	return nil
}
