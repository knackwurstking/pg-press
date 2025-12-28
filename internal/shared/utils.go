package shared

import (
	"strconv"
	"strings"

	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/labstack/echo/v4"
)

// -----------------------------------------------------------------------------
// String Masking
// -----------------------------------------------------------------------------

// MaskString masks sensitive strings by showing only the first and last 4 characters.
// For strings with 8 or fewer characters, all characters are masked.
func MaskString(s string) string {
	if len(s) <= 8 {
		return strings.Repeat("*", len(s))
	}
	return s[:4] + strings.Repeat("*", len(s)-8) + s[len(s)-4:]
}

// -----------------------------------------------------------------------------
// User Retrieval from Context
// -----------------------------------------------------------------------------

func GetUserFromContext(c echo.Context) (*User, *errors.MasterError) {
	u := c.Get("user")
	if u == nil {
		return nil, errors.NewAuthorizationError("no user in context").MasterError()
	}

	user, ok := u.(*User)
	if !ok || user.Validate() != nil {
		return nil, errors.NewAuthorizationError("invalid user in context").MasterError()
	}

	return user, nil
}

// -----------------------------------------------------------------------------
// Query Parameter Parsing
// -----------------------------------------------------------------------------

// ParseQueryString parses a string query parameter from the request
func ParseQueryString(c echo.Context, paramName string) (string, *errors.MasterError) {
	s := c.QueryParam(paramName)
	if s == "" {
		return s, errors.NewNotFoundError("missing %s", paramName).MasterError()
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
func ParseQueryInt64(c echo.Context, paramName string) (int64, *errors.MasterError) {
	idStr := c.QueryParam(paramName)
	if idStr == "" {
		return 0, errors.NewNotFoundError("missing %s", paramName).MasterError()
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, errors.NewValidationError("invalid %s query parameter: must be a number", paramName).MasterError()
	}

	return id, nil
}

func ParseQueryInt(c echo.Context, paramName string) (int, *errors.MasterError) {
	idStr := c.QueryParam(paramName)
	if idStr == "" {
		return 0, errors.NewNotFoundError("missing %s", paramName).MasterError()
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, errors.NewValidationError("invalid %s query parameter: must be a number", paramName).MasterError()
	}

	return id, nil
}

// -----------------------------------------------------------------------------
// Path Parameter Parsing
// -----------------------------------------------------------------------------

// ParseParamInt64 parses an int64 parameter from the request
func ParseParamInt64(c echo.Context, paramName string) (int64, *errors.MasterError) {
	idStr := c.Param(paramName)
	if idStr == "" {
		return 0, errors.NewNotFoundError("missing %s", paramName).MasterError()
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, errors.NewValidationError("invalid %s: must be a number", paramName).MasterError()
	}
	return id, nil
}

// ParseParamInt8 parses an int8 parameter from the request
func ParseParamInt8(c echo.Context, paramName string) (int8, *errors.MasterError) {
	idStr := c.Param(paramName)
	if idStr == "" {
		return 0, errors.NewNotFoundError("missing %s", paramName).MasterError()
	}

	id, err := strconv.ParseInt(idStr, 10, 8)
	if err != nil {
		return 0, errors.NewValidationError("invalid %s: must be a number", paramName).MasterError()
	}
	return int8(id), nil
}
