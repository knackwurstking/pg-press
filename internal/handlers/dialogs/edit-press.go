package dialogs

import (
	"strconv"

	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/labstack/echo/v4"
)

func GetEditPress(c echo.Context) *echo.HTTPError {
	id, merr := shared.ParseQueryInt64(c, "id")
	if merr != nil && !merr.IsNotFoundError() {
		return merr.Echo()
	}

	if id > 0 {
		press, merr := db.GetPress(shared.EntityID(id))
		if merr != nil {
			return merr.Echo()
		}

		t := EditPressDialog(press)
		err := t.Render(c.Request().Context(), c.Response())
		if err != nil {
			return errors.NewRenderError(err, "EditPressDialog")
		}
		return nil
	}

	t := NewPressDialog()
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "NewPressDialog")
	}
	return nil
}

func PostPress(c echo.Context) *echo.HTTPError {
	// TODO: ...
}

func PutPress(c echo.Context) *echo.HTTPError {
	// TODO: ...
}

type EditPressFormData struct {
	PressNumber shared.PressNumber
	MachineType shared.MachineType
}

func GetEditPressFormData(c echo.Context) (*EditPressFormData, *errors.MasterError) {
	vPressNumber := c.FormValue("press_number")
	if vPressNumber == "" {
		return nil, errors.NewValidationError("press number is required").MasterError()
	}
	pressNumber, err := strconv.Atoi(vPressNumber)
	if err != nil {
		return nil, errors.NewValidationError("invalid press number %s: %v", vPressNumber, err).MasterError()
	}

	return &EditPressFormData{
		PressNumber: shared.PressNumber(pressNumber),
		MachineType: shared.MachineType(c.FormValue("machine_type")),
	}, nil
}
