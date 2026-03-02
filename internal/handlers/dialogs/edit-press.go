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

func GetEditPress(c echo.Context) *echo.HTTPError {
	id, merr := utils.GetQueryInt64(c, "id")
	if merr != nil && !merr.IsNotFoundError() {
		return merr.Echo()
	}

	if id > 0 {
		press, merr := db.GetPress(shared.EntityID(id))
		if merr != nil {
			return merr.Echo()
		}

		t := EditPressDialog(EditPressDialogProps{
			PressFormData: PressFormData{
				Number:       press.Number,
				Type:         press.Type,
				Code:         press.Code,
				CyclesOffset: press.CyclesOffset,
			},
			PressID: press.ID,
			OOB:     true,
			Open:    true,
		})
		err := t.Render(c.Request().Context(), c.Response())
		if err != nil {
			return errors.NewRenderError(err, "EditPressDialog")
		}
		return nil
	}

	t := NewPressDialog(NewPressDialogProps{
		OOB:  true,
		Open: true,
	})
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "NewPressDialog")
	}
	return nil
}

func PostPress(c echo.Context) *echo.HTTPError {
	// Get press number from query first
	id, _ := utils.GetQueryInt64(c, "id")
	if id > 0 {
		return updatePress(c, shared.EntityID(id))
	}

	data, ierr := parseEditPressForm(c)
	if ierr != nil {
		return ReRenderNewPressDialog(c, true, data, ierr)
	}

	merr := db.AddPress(&shared.Press{
		Number:       data.Number,
		Type:         data.Type,
		Code:         data.Code,
		CyclesOffset: data.CyclesOffset,
	})
	if merr != nil {
		ierr = errors.NewInputError("form", fmt.Sprintf("failed to add press: %v", merr))
		return ReRenderNewPressDialog(c, true, data, ierr)
	}

	utils.SetHXTrigger(c, "tools-tab")

	return ReRenderNewPressDialog(c, false, data, nil)
}

func updatePress(c echo.Context, id shared.EntityID) *echo.HTTPError {
	data, ierr := parseEditPressForm(c)
	if ierr != nil {
		return ReRenderEditPressDialog(c, id, true, data, ierr)
	}

	press, herr := db.GetPress(id)
	if herr != nil {
		ierr = errors.NewInputError("form", fmt.Sprintf("failed to get press: %v", herr))
		return ReRenderEditPressDialog(c, id, true, data, ierr)
	}

	merr := db.UpdatePress(&shared.Press{
		ID:           press.ID,
		Number:       data.Number,
		Type:         data.Type,
		Code:         data.Code,
		CyclesOffset: data.CyclesOffset,
		SlotUp:       press.SlotUp,
		SlotDown:     press.SlotDown,
	})
	if merr != nil {
		ierr = errors.NewInputError("form", fmt.Sprintf("failed to update press: %v", merr))
		return ReRenderEditPressDialog(c, id, true, data, nil)
	}

	utils.SetHXTrigger(c, "tools-tab")

	return ReRenderEditPressDialog(c, id, false, data, nil)
}

func parseEditPressForm(c echo.Context) (data PressFormData, ierr *errors.InputError) {
	// Press Number
	vPressNumber, err := utils.SanitizeInt8(c.FormValue("press_number"))
	if err != nil {
		ierr = errors.NewInputError("press_number", fmt.Sprintf("invalid press number: %v", err))
		return
	}
	data.Number = shared.PressNumber(vPressNumber)

	// Code
	data.Code = utils.SanitizeText(c.FormValue("code"))
	if data.Code == "" {
		ierr = errors.NewInputError("code", "code cannot be empty")
		return
	}

	// Cycles Offset
	data.CyclesOffset, err = utils.SanitizeInt64(c.FormValue("cycles_offset"))
	if err != nil {
		ierr = errors.NewInputError("cycles_offset", fmt.Sprintf("invalid cycles offset: %v", err))
		return
	}

	// Machine Type
	data.Type = shared.MachineType(utils.SanitizeText(c.FormValue("machine_type")))

	return
}

func ReRenderNewPressDialog(c echo.Context, open bool, data PressFormData, ierr *errors.InputError) *echo.HTTPError {
	t := NewPressDialog(NewPressDialogProps{
		PressFormData: data,
		Open:          open,
		OOB:           true,
		Error:         ierr,
	})
	if err := t.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.NewRenderError(err, "NewPressDialog")
	}
	if ierr != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid input: %v", ierr))
	}
	return nil
}

func ReRenderEditPressDialog(c echo.Context, pressID shared.EntityID, open bool, data PressFormData, ierr *errors.InputError) *echo.HTTPError {
	t := EditPressDialog(EditPressDialogProps{
		PressFormData: data,
		PressID:       pressID,
		Open:          open,
		OOB:           true,
		Error:         ierr,
	})
	if err := t.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.NewRenderError(err, "EditPressDialog")
	}
	if ierr != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid input: %v", ierr))
	}
	return nil
}
