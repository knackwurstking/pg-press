package dialogs

import (
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/labstack/echo/v4"
)

func GetEditPress(c echo.Context) *echo.HTTPError {
}

func PostPress(c echo.Context) *echo.HTTPError {
}

func PutPress(c echo.Context) *echo.HTTPError {
}

type EditPressFormData struct {
}

func GetEditPressFormData() (*EditPressFormData, *errors.MasterError) {
}
