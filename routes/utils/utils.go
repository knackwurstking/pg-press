package utils

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/pgvis"
)

func GetUserFromContext(ctx echo.Context) (*pgvis.User, *echo.HTTPError) {
	user, ok := ctx.Get("user").(*pgvis.User)
	if !ok {
		return nil, echo.NewHTTPError(http.StatusInternalServerError,
			errors.New("user is missing in context"))
	}

	return user, nil
}
