package dialogs

import (
	"strconv"

	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/dialogs/templates"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/utils"

	"github.com/labstack/echo/v4"
)

func GetEditPress(c echo.Context) *echo.HTTPError {
	pn, merr := utils.GetQueryInt8(c, "id")
	if merr != nil && merr.IsNotFoundError() {
		pn = -1
	} else if merr != nil {
		return merr.Echo()
	}

	if pn > -1 {
		press, merr := db.GetPress(shared.PressNumber(pn))
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
	data, merr := GetEditPressFormData(c)
	if merr != nil {
		return merr.Echo()
	}

	merr = db.AddPress(&shared.Press{
		ID:           data.PressNumber,
		Type:         data.MachineType,
		CyclesOffset: data.CyclesOffset,
	})
	if merr != nil {
		return merr.Echo()
	}

	utils.SetHXTrigger(c, "tools-tab")

	return nil
}

func PutPress(c echo.Context) *echo.HTTPError {
	id, merr := utils.GetQueryInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}

	data, merr := GetEditPressFormData(c)
	if merr != nil {
		return merr.Echo()
	}

	merr = db.UpdatePress(&shared.Press{
		ID:           shared.PressNumber(id),
		Type:         data.MachineType,
		CyclesOffset: data.CyclesOffset,
	})
	if merr != nil {
		return merr.Echo()
	}

	utils.SetHXTrigger(c, "tools-tab")

	return nil
}

type EditPressFormData struct {
	PressNumber  shared.PressNumber
	MachineType  shared.MachineType
	CyclesOffset int64
}

func GetEditPressFormData(c echo.Context) (*EditPressFormData, *errors.HTTPError) {
	vPressNumber := c.FormValue("press_number")
	if vPressNumber == "" {
		return nil, errors.NewValidationError("press number is required").HTTPError()
	}
	pressNumber, err := strconv.Atoi(vPressNumber)
	if err != nil {
		return nil, errors.NewValidationError("invalid press number %s: %v", vPressNumber, err).HTTPError()
	}

	cyclesOffset, err := strconv.ParseInt(c.FormValue("cycles_offset"), 10, 64)
	if err != nil {
		return nil, errors.NewValidationError(
			"invalid cycles offset %s: %v", c.FormValue("cycles_offset"), err,
		).HTTPError()
	}

	return &EditPressFormData{
		PressNumber:  shared.PressNumber(pressNumber),
		MachineType:  shared.MachineType(c.FormValue("machine_type")),
		CyclesOffset: cyclesOffset,
	}, nil
}
