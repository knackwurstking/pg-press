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

// TODO: Re-Render dialog with error or close at the end...
func PostPress(c echo.Context) *echo.HTTPError {
	// Get press number from query first
	id, _ := utils.GetQueryInt64(c, "id")
	if id > 0 {
		return updatePress(c, shared.EntityID(id))
	}

	data, ierr := parseEditPressForm(c)
	if ierr != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid input: %v", ierr))
	}

	merr := db.AddPress(&shared.Press{
		Number:       data.Number,
		Type:         data.Type,
		Code:         data.Code,
		CyclesOffset: data.CyclesOffset,
	})
	if merr != nil {
		return merr.Echo()
	}

	utils.SetHXTrigger(c, "tools-tab")

	return nil
}

// TODO: Re-Render dialog with error or close at the end
func updatePress(c echo.Context, id shared.EntityID) *echo.HTTPError {
	data, ierr := parseEditPressForm(c)
	if ierr != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid input: %v", ierr))
	}

	press, herr := db.GetPress(id)
	if herr != nil {
		return herr.Echo()
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
		return merr.Echo()
	}

	utils.SetHXTrigger(c, "tools-tab")

	return nil
}

func parseEditPressForm(c echo.Context) (data *PressFormData, ierr *errors.InputError) {
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
