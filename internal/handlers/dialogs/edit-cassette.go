package dialogs

import (
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"

	"github.com/labstack/echo/v4"
)

func (h *Handler) GetCassetteDialog(c echo.Context) *echo.HTTPError {
	return nil
}

func (h *Handler) PostCassette(c echo.Context) *echo.HTTPError {
	return nil
}

func (h *Handler) PutCassette(c echo.Context) *echo.HTTPError {
	return nil
}

func (h *Handler) getCassetteDialogForm(c echo.Context) (*shared.Cassette, *errors.ValidationError) {
	return nil, nil
}
