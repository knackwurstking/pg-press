package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/services/shared"
	"github.com/labstack/echo/v4"
)

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

// SanitizeFilename cleans up a filename by replacing characters and removing duplicates
func SanitizeFilename(filename string) string {
	if idx := strings.LastIndex(filename, "."); idx > 0 {
		filename = filename[:idx]
	}

	filename = strings.ReplaceAll(filename, " ", "_")
	filename = strings.ReplaceAll(filename, "-", "_")
	filename = strings.ReplaceAll(filename, "(", "_")
	filename = strings.ReplaceAll(filename, ")", "_")
	filename = strings.ReplaceAll(filename, "[", "_")
	filename = strings.ReplaceAll(filename, "]", "_")

	for strings.Contains(filename, "__") {
		filename = strings.ReplaceAll(filename, "__", "_")
	}

	filename = strings.Trim(filename, "_")

	if filename == "" {
		filename = "attachment"
	}

	return filename
}

// SetHXTrigger sets HX-Trigger header
func SetHXTrigger(c echo.Context, events ...string) {
	c.Response().Header().Set("HX-Trigger", strings.Join(events, ", "))
}

// SetHXRedirect sets HX-Redirect header
func SetHXRedirect(c echo.Context, path templ.SafeURL) {
	c.Response().Header().Set("HX-Redirect", string(path))
}

// SetHXAfterSettle sets HX-Trigger-After-Settle header with JSON data
func SetHXAfterSettle(c echo.Context, data map[string]any) {
	triggerDataJSON, _ := json.Marshal(data)
	c.Response().Header().Set("HX-Trigger-After-Settle", string(triggerDataJSON))
}
