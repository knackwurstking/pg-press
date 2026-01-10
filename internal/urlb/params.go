package urlb

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
		return nil, errors.NewAuthorizationError("no user in context").HTTPError()
	}

	user, ok := u.(*shared.User)
	if !ok || user.Validate() != nil {
		return nil, errors.NewAuthorizationError("invalid user in context").HTTPError()
	}

	return user, nil
}

// -----------------------------------------------------------------------------
// Query Parameter Parsing
// -----------------------------------------------------------------------------

// ParseQueryString parses a string query parameter from the request
func ParseQueryString(c echo.Context, paramName string) (string, *errors.HTTPError) {
	s := c.QueryParam(paramName)
	if s == "" {
		return s, errors.NewNotFoundError("missing %s", paramName).HTTPError()
	}

	return s, nil
}

// ParseQueryBool parses a boolean query parameter from the request
func ParseQueryBool(c echo.Context, paramName string) bool {
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

// ParseQueryInt64 parses an int64 query parameter from the request
func ParseQueryInt64(c echo.Context, paramName string) (int64, *errors.HTTPError) {
	idStr := c.QueryParam(paramName)
	if idStr == "" {
		return 0, errors.NewNotFoundError("missing %s", paramName).HTTPError()
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, errors.NewValidationError("invalid %s query parameter: must be a number", paramName).HTTPError()
	}

	return id, nil
}

func ParseQueryInt(c echo.Context, paramName string) (int, *errors.HTTPError) {
	idStr := c.QueryParam(paramName)
	if idStr == "" {
		return 0, errors.NewNotFoundError("missing %s", paramName).HTTPError()
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, errors.NewValidationError("invalid %s query parameter: must be a number", paramName).HTTPError()
	}

	return id, nil
}

// -----------------------------------------------------------------------------
// Path Parameter Parsing
// -----------------------------------------------------------------------------

// ParseParamInt64 parses an int64 parameter from the request
func ParseParamInt64(c echo.Context, paramName string) (int64, *errors.HTTPError) {
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

// ParseParamInt8 parses an int8 parameter from the request
func ParseParamInt8(c echo.Context, paramName string) (int8, *errors.HTTPError) {
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
