package dialogs

import (
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

	data, verr := parseEditPressForm(c)
	if verr != nil {
		return verr.HTTPError().Echo()
	}

	merr := db.AddPress(&shared.Press{
		Number:       data.PressNumber,
		Type:         data.MachineType,
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
	data, verr := parseEditPressForm(c)
	if verr != nil {
		return verr.HTTPError().Echo()
	}

	press, herr := db.GetPress(data.PressID)
	if herr != nil {
		return herr.Echo()
	}

	merr := db.UpdatePress(&shared.Press{
		ID:           press.ID,
		Number:       data.PressNumber,
		Type:         data.MachineType,
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

type editPressForm struct {
	PressID      shared.EntityID
	PressNumber  shared.PressNumber
	MachineType  shared.MachineType
	Code         string
	CyclesOffset int64
}

func parseEditPressForm(c echo.Context) (*editPressForm, *errors.ValidationError) {
	// Press Number
	vPressNumber, err := utils.SanitizeInt8(c.FormValue("press_number"))
	if err != nil {
		return nil, errors.NewValidationError("invalid press number: %v", err)
	}

	code := utils.SanitizeText(c.FormValue("code"))
	if code == "" {
		return nil, errors.NewValidationError("code cannot be empty")
	}

	// Cycles Offset
	cyclesOffset, err := utils.SanitizeInt64(c.FormValue("cycles_offset"))
	if err != nil {
		return nil, errors.NewValidationError("invalid cycles offset: %v", err)
	}

	return &editPressForm{
		PressID:      shared.EntityID(vPressID),
		PressNumber:  shared.PressNumber(vPressNumber),
		MachineType:  shared.MachineType(utils.SanitizeText(c.FormValue("machine_type"))),
		Code:         code,
		CyclesOffset: cyclesOffset,
	}, nil
}
