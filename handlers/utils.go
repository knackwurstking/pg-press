package handlers

import (
	"fmt"
	"net/http"

	"github.com/knackwurstking/pgpress/errors"
	"github.com/knackwurstking/pgpress/models"
	"github.com/labstack/echo/v4"
)

func HandleBadRequest(err error, message string) error {
	if err == nil {
		return echo.NewHTTPError(http.StatusBadRequest, message)
	}

	return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("%s: %v", message, err))
}

// HandleError creates an HTTP error with the appropriate status code
func HandleError(err error, context string) error {
	statusCode := errors.GetHTTPStatusCode(err)
	if statusCode == 0 {
		statusCode = http.StatusInternalServerError
	}
	return echo.NewHTTPError(statusCode, fmt.Sprintf("%s: %v", context, err))
}

// GetUserFromContext retrieves the authenticated user from the request context
func GetUserFromContext(c echo.Context) (*models.User, error) {
	userInterface := c.Get("user")
	if userInterface == nil {
		return nil, fmt.Errorf("user not found in context")
	}

	user, ok := userInterface.(*models.User)
	if !ok {
		return nil, fmt.Errorf("invalid user type in context")
	}

	return user, nil
}

// RedirectTo performs an HTTP redirect to the specified path
func RedirectTo(c echo.Context, path string) error {
	return c.Redirect(http.StatusSeeOther, path)
}

func ParseQueryString(c echo.Context, paramName string) (string, error) {
	s := c.QueryParam(paramName)
	if s == "" {
		return "", fmt.Errorf("missing %s query parameter", paramName)
	}

	return s, nil
}
