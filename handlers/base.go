package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/knackwurstking/pgpress/env"
	"github.com/knackwurstking/pgpress/errors"
	"github.com/knackwurstking/pgpress/internal/services" // TODO: Continue here
	"github.com/knackwurstking/pgpress/logger"
	"github.com/knackwurstking/pgpress/models"

	"github.com/labstack/echo/v4"
)

type baseHandler struct {
	db  *services.Registry
	log *logger.Logger
}

func newBaseHandler(db *services.Registry, logger *logger.Logger) *baseHandler {
	return &baseHandler{
		db:  db,
		log: logger,
	}
}

func (b *baseHandler) getUserFromContext(c echo.Context) (*models.User, error) {
	user, ok := c.Get("user").(*models.User)
	if !ok {
		return nil, fmt.Errorf("missing user context")
	}
	if user == nil {
		return nil, fmt.Errorf("invalid user session")
	}
	return user, nil
}

func (b *baseHandler) renderError(c echo.Context, statusCode int, message string) error {
	return echo.NewHTTPError(statusCode, message)
}

func (b *baseHandler) renderInternalError(c echo.Context, message string) error {
	return echo.NewHTTPError(http.StatusInternalServerError, message)
}

func (b *baseHandler) renderBadRequest(c echo.Context, message string) error {
	return echo.NewHTTPError(http.StatusBadRequest, message)
}

// renderUnauthorized creates a 401 Unauthorized response
// For example when an action requires admin privileges
func (b *baseHandler) renderUnauthorized(c echo.Context, message string) error {
	return echo.NewHTTPError(http.StatusUnauthorized, message)
}

// renderNotFound creates a 404 Not Found response
func (b *baseHandler) renderNotFound(c echo.Context, message string) error {
	return echo.NewHTTPError(http.StatusNotFound, message)
}

// handleError processes errors and returns appropriate HTTP responses
func (b *baseHandler) handleError(c echo.Context, err error, context string) error {
	message := fmt.Sprintf("%s: %v", context, err)
	b.log.Error("%s", message)

	statusCode := errors.GetHTTPStatusCode(err)
	if statusCode == 0 {
		statusCode = http.StatusInternalServerError
	}

	return b.renderError(c, statusCode, message)
}

// redirectTo redirects the user to a specific path
func (b *baseHandler) redirectTo(c echo.Context, path string) error {
	return c.Redirect(http.StatusSeeOther, path)
}

// getSanitizedFormValue retrieves and sanitizes form input
func (b *baseHandler) getSanitizedFormValue(c echo.Context, fieldName string) string {
	formParams, _ := c.FormParams()
	sanitized := strings.TrimSpace(formParams.Get(fieldName))
	sanitized = strings.ReplaceAll(sanitized, "\x00", "")
	return sanitized
}

func (b *baseHandler) sanitizeFilename(filename string) string {
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

func (b *baseHandler) parseInt8Param(c echo.Context, paramName string) (int8, error) {
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

// parseInt64Param parses an int64 parameter from the URL path
func (b *baseHandler) parseInt64Param(c echo.Context, paramName string) (int64, error) {
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

// parseInt64Query parses an int64 parameter from the query string
func (b *baseHandler) parseInt64Query(c echo.Context, paramName string) (int64, error) {
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

// parseStringQuery parses a string parameter from the query string
func (b *baseHandler) parseStringQuery(c echo.Context, paramName string) (string, error) {
	s := c.QueryParam(paramName)
	if s == "" {
		return "", fmt.Errorf("missing %s query parameter", paramName)
	}

	return s, nil
}

// parseBoolQuery parses a boolean parameter from the query string
func (b *baseHandler) parseBoolQuery(c echo.Context, paramName string) bool {
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

func (b *baseHandler) setHXTrigger(c echo.Context) {
	c.Response().Header().Set("HX-Trigger", "pageLoaded")
}

func (b *baseHandler) setHXRedirect(c echo.Context, path string) {
	c.Response().Header().Set("HX-Redirect", env.ServerPathPrefix+path)
}
