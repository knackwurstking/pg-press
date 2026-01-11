package utils

import (
	"strconv"

	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"

	"github.com/labstack/echo/v4"
)

// -----------------------------------------------------------------------------
// User Retrieval from Context
// -----------------------------------------------------------------------------

func GetUserFromContext(c echo.Context) (*shared.User, *errors.HTTPError) {
	u := c.Get("user")
	if u == nil {
		return nil, errors.NewAuthorizationError("you must be logged in to access this resource").HTTPError()
	}

	user, ok := u.(*shared.User)
	if !ok || user.Validate() != nil {
		return nil, errors.NewAuthorizationError("invalid user session. Please log in again").HTTPError()
	}

	return user, nil
}

// -----------------------------------------------------------------------------
// Query Parameter
// -----------------------------------------------------------------------------

func GetQueryString(c echo.Context, paramName string) (string, *errors.HTTPError) {
	s := c.QueryParam(paramName)
	if s == "" {
		return s, errors.NewValidationError("query parameter '%s' is required", paramName).HTTPError()
	}

	return s, nil
}

func GetQueryBool(c echo.Context, paramName string) bool {
	value := c.QueryParam(paramName)
	if value == "" {
		return false
	}

	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return false
	}

	return boolValue
}

func GetQueryInt64(c echo.Context, paramName string) (int64, *errors.HTTPError) {
	idStr := c.QueryParam(paramName)
	if idStr == "" {
		return 0, errors.NewValidationError("query parameter '%s' is required", paramName).HTTPError()
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, errors.NewValidationError("invalid '%s' value '%s': must be a valid integer", paramName, idStr).HTTPError()
	}

	return id, nil
}

func GetQueryInt(c echo.Context, paramName string) (int, *errors.HTTPError) {
	idStr := c.QueryParam(paramName)
	if idStr == "" {
		return 0, errors.NewValidationError("query parameter '%s' is required", paramName).HTTPError()
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, errors.NewValidationError("invalid '%s' value '%s': must be a valid integer", paramName, idStr).HTTPError()
	}

	return id, nil
}

// -----------------------------------------------------------------------------
// Path Parameter
// -----------------------------------------------------------------------------

func GetParamInt64(c echo.Context, paramName string) (int64, *errors.HTTPError) {
	idStr := c.Param(paramName)
	if idStr == "" {
		return 0, errors.NewNotFoundError("missing %s", paramName).HTTPError()
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, errors.NewValidationError("invalid %s: must be a number", paramName).HTTPError()
	}
	return id, nil
}

func GetParamInt8(c echo.Context, paramName string) (int8, *errors.HTTPError) {
	idStr := c.Param(paramName)
	if idStr == "" {
		return 0, errors.NewNotFoundError("missing %s", paramName).HTTPError()
	}

	id, err := strconv.ParseInt(idStr, 10, 8)
	if err != nil {
		return 0, errors.NewValidationError("invalid %s: must be a number", paramName).HTTPError()
	}
	return int8(id), nil
}
