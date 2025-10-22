package handlers

import (
	"errors"
	"net/http"

	"github.com/knackwurstking/pgpress/models"
	"github.com/labstack/echo/v4"
)

func GetInternalServerError(err error, message string) *echo.HTTPError {
	return echo.NewHTTPError(http.StatusInternalServerError, message+": "+err.Error())
}

func GetUserFromContext(c echo.Context) (*models.User, error) {
	userInterface := c.Get("user")
	if userInterface == nil {
		return nil, errors.New("user not found in context")
	}

	user, ok := userInterface.(*models.User)
	if !ok {
		return nil, errors.New("invalid user type in context")
	}

	return user, nil
}
