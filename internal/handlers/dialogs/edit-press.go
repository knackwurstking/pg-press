package dialogs

import (
	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/dialogs/templates"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/utils"

	"github.com/labstack/echo/v4"
)

func GetEditPress(c echo.Context) *echo.HTTPError {
	id, merr := utils.GetQueryInt8(c, "id")
	if merr != nil && !merr.IsNotFoundError() {
		return merr.Echo()
	}

	if id > 0 {
		press, merr := db.GetPress(shared.EntityID(id))
		if merr != nil {
			return merr.Echo()
		}

		t := templates.EditPressDialog(press)
		err := t.Render(c.Request().Context(), c.Response())
		if err != nil {
			return errors.NewRenderError(err, "EditPressDialog")
		}
		return nil
	}

	t := templates.NewPressDialog()
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "NewPressDialog")
	}
	return nil
}

func PostPress(c echo.Context) *echo.HTTPError {
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

func PutPress(c echo.Context) *echo.HTTPError {
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
		Number:       press.Number,
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
	// Press ID
	vPressID, err := utils.SanitizeInt64(c.FormValue("press_id"))
	if err != nil {
		return nil, errors.NewValidationError("invalid press ID: %v", err)
	}

	// Press Number
	vPressNumber, err := utils.SanitizeInt8(c.FormValue("press_id"))
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
