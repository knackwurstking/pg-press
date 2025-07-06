package html

import (
	"errors"
	"net/http"

	"github.com/knackwurstking/pg-vis/pkg/pgvis"
	"github.com/labstack/echo/v4"
)

func getUserFromContext(ctx echo.Context) (*pgvis.User, error) {
	user, ok := ctx.Get("user").(*pgvis.User)
	if !ok {
		return nil, echo.NewHTTPError(http.StatusInternalServerError,
			errors.New("user is missing in context"))
	}

	return user, nil
}
