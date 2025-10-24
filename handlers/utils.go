package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/knackwurstking/pgpress/env"
	"github.com/knackwurstking/pgpress/errors"
	"github.com/knackwurstking/pgpress/models"
	"github.com/labstack/echo/v4"
)

func HandleNotFound(err error, message string) error {
	if err == nil {
		return echo.NewHTTPError(http.StatusNotFound, message)
	}

	return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("%s: %v", message, err))
}

func HandleBadRequest(err error, message string) error {
	if err == nil {
		return echo.NewHTTPError(http.StatusBadRequest, message)
	}

	return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("%s: %v", message, err))
}

// HandleError creates an HTTP error with the appropriate status code
func HandleError(err error, context string) error {
	if err == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, context)
	}

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

func ParseQueryInt64(c echo.Context, paramName string) (int64, error) {
	idStr := c.QueryParam(paramName)
	if idStr == "" {
		return 0, fmt.Errorf("missing %s query parameter", paramName)
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s query parameter: must be a number", paramName)
	}

	return id, nil
}

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

func ParseParamInt64(c echo.Context, paramName string) (int64, error) {
	idStr := c.Param(paramName)
	if idStr == "" {
		return 0, fmt.Errorf("missing %s parameter", paramName)
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s parameter: must be a number", paramName)
	}
	return id, nil
}

func ParseParamInt8(c echo.Context, paramName string) (int8, error) {
	idStr := c.Param(paramName)
	if idStr == "" {
		return 0, fmt.Errorf("missing %s parameter", paramName)
	}

	id, err := strconv.ParseInt(idStr, 10, 8)
	if err != nil {
		return 0, fmt.Errorf("invalid %s parameter: must be a number", paramName)
	}
	return int8(id), nil
}

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

func SetHXTrigger(c echo.Context, events ...string) {
	c.Response().Header().Set("HX-Trigger", strings.Join(events, ", "))
}

func SetHXRedirect(c echo.Context, path string) {
	c.Response().Header().Set("HX-Redirect", env.ServerPathPrefix+path)
}

// SetHXAfterSettle will set data passed to it as (json) data, which can be used to trigger client-side events after the response is settled.
func SetHXAfterSettle(c echo.Context, data map[string]any) {
	triggerDataJSON, _ := json.Marshal(data)
	c.Response().Header().Set("HX-Trigger-After-Settle", string(triggerDataJSON))
}
