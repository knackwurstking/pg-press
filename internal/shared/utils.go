package shared

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/labstack/echo/v4"
)

/*******************************************************************************
 * String Utils
 ******************************************************************************/

// MaskString masks sensitive strings by showing only the first and last 4 characters.
// For strings with 8 or fewer characters, all characters are masked.
func MaskString(s string) string {
	if len(s) <= 8 {
		return strings.Repeat("*", len(s))
	}
	return s[:4] + strings.Repeat("*", len(s)-8) + s[len(s)-4:]
}

/*******************************************************************************
 * Handler Utils
 ******************************************************************************/

func GetUserFromContext(c echo.Context) (*User, *errors.MasterError) {
	u := c.Get("user")
	if u == nil {
		return nil, errors.NewMasterError(
			fmt.Errorf("no user"), http.StatusUnauthorized,
		)
	}

	user, ok := u.(*User)
	if !ok || user.Validate() != nil {
		return nil, errors.NewMasterError(
			fmt.Errorf("invalid user"), http.StatusUnauthorized,
		)
	}

	return user, nil
}

/*******************************************************************************
 * Echo Parse Utils
 ******************************************************************************/

// TODO: Add echo parse utilities

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
		return 0, errors.NewMasterError(
			fmt.Errorf("missing %s query parameter", paramName),
			http.StatusNotFound,
		)
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, errors.NewMasterError(
			fmt.Errorf("invalid %s query parameter: must be a number", paramName),
			http.StatusBadRequest,
		)
	}

	return id, nil
}

// ParseParamInt64 parses an int64 parameter from the request
func ParseParamInt64(c echo.Context, paramName string) (int64, *errors.MasterError) {
	idStr := c.Param(paramName)
	if idStr == "" {
		return 0, errors.NewMasterError(fmt.Errorf("missing %s", paramName), http.StatusBadRequest)
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, errors.NewMasterError(fmt.Errorf("invalid %s", paramName), http.StatusBadRequest)
	}
	return id, nil
}

// ParseParamInt8 parses an int8 parameter from the request
func ParseParamInt8(c echo.Context, paramName string) (int8, *errors.MasterError) {
	idStr := c.Param(paramName)
	if idStr == "" {
		return 0, errors.NewMasterError(
			fmt.Errorf("missing %s", paramName), http.StatusNotFound,
		)
	}

	id, err := strconv.ParseInt(idStr, 10, 8)
	if err != nil {
		return 0, errors.NewMasterError(
			fmt.Errorf("invalid %s", paramName), http.StatusBadRequest,
		)

	}
	return int8(id), nil
}

// ParseQueryString parses a string query parameter from the request
func ParseQueryString(c echo.Context, paramName string) (string, *errors.MasterError) {
	s := c.QueryParam(paramName)
	if s == "" {
		return s, errors.NewMasterError(fmt.Errorf("missing %s", paramName), http.StatusNotFound)
	}

	return s, nil
}

// ParseFormValueTime parses a date from a form value
func ParseFormValueTime(c echo.Context, paramName string) (time.Time, *errors.MasterError) {
	v := c.FormValue(paramName)
	if v == "" {
		return time.Time{}, errors.NewMasterError(
			fmt.Errorf("missing %s", paramName), http.StatusNotFound,
		)
	}

	t, err := time.Parse("2006-01-02", v) // YYYY-MM-DD
	if err != nil {
		return t, errors.NewMasterError(errors.Wrap(err, "parsing date input"), http.StatusBadRequest)
	}

	return t, nil
}
