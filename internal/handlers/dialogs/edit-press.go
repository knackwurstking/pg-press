package dialogs

import (
	"strconv"

	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/labstack/echo/v4"
)

func GetEditPress(c echo.Context) *echo.HTTPError {
}

func PostPress(c echo.Context) *echo.HTTPError {
}

func PutPress(c echo.Context) *echo.HTTPError {
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
