package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/knackwurstking/pgpress/internal/services"
	"github.com/knackwurstking/pgpress/pkg/logger"
	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"

	"github.com/labstack/echo/v4"
)

// BaseHandler provides common functionality for all handlers
type BaseHandler struct {
	DB     *services.Registry
	Logger *logger.Logger
}

// NewBaseHandler creates a new base handler with database and logger
func NewBaseHandler(db *services.Registry, logger *logger.Logger) *BaseHandler {
	return &BaseHandler{
		DB:     db,
		Logger: logger,
	}
}

// GetUserFromContext retrieves the authenticated user from the request context
func (b *BaseHandler) GetUserFromContext(c echo.Context) (*models.User, error) {
	user, ok := c.Get("user").(*models.User)
	if !ok {
		return nil, fmt.Errorf("missing user context")
	}
	if user == nil {
		return nil, fmt.Errorf("invalid user session")
	}
	return user, nil
}

// RenderError creates a standardized HTTP error response
func (b *BaseHandler) RenderError(c echo.Context, statusCode int, message string) error {
	return echo.NewHTTPError(statusCode, message)
}

// RenderInternalError creates a 500 Internal Server Error response
func (b *BaseHandler) RenderInternalError(c echo.Context, message string) error {
	return echo.NewHTTPError(http.StatusInternalServerError, message)
}

// RenderBadRequest creates a 400 Bad Request response
func (b *BaseHandler) RenderBadRequest(c echo.Context, message string) error {
	return echo.NewHTTPError(http.StatusBadRequest, message)
}

// RenderUnauthorized creates a 401 Unauthorized response
// For example when an action requires admin privileges
func (b *BaseHandler) RenderUnauthorized(c echo.Context, message string) error {
	return echo.NewHTTPError(http.StatusUnauthorized, message)
}

// RenderNotFound creates a 404 Not Found response
func (b *BaseHandler) RenderNotFound(c echo.Context, message string) error {
	return echo.NewHTTPError(http.StatusNotFound, message)
}

// HandleError processes errors and returns appropriate HTTP responses
func (b *BaseHandler) HandleError(c echo.Context, err error, context string) error {
	message := fmt.Sprintf("%s: %v", context, err)
	if b.Logger != nil {
		b.Logger.Error(message)
	}

	statusCode := utils.GetHTTPStatusCode(err)
	if statusCode == 0 {
		statusCode = http.StatusInternalServerError
	}

	return b.RenderError(c, statusCode, message)
}

// RedirectTo redirects the user to a specific path
func (b *BaseHandler) RedirectTo(c echo.Context, path string) error {
	return c.Redirect(http.StatusSeeOther, path)
}

// GetSanitizedFormValue retrieves and sanitizes form input
func (b *BaseHandler) GetSanitizedFormValue(c echo.Context, fieldName string) string {
	formParams, _ := c.FormParams()
	sanitized := strings.TrimSpace(formParams.Get(fieldName))
	sanitized = strings.ReplaceAll(sanitized, "\x00", "")
	return sanitized
}

func (b *BaseHandler) SanitizeFilename(filename string) string {
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

func (b *BaseHandler) ParseInt8Param(c echo.Context, paramName string) (int8, error) {
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

// ParseInt64Param parses an int64 parameter from the URL path
func (b *BaseHandler) ParseInt64Param(c echo.Context, paramName string) (int64, error) {
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

// ParseInt64Query parses an int64 parameter from the query string
func (b *BaseHandler) ParseInt64Query(c echo.Context, paramName string) (int64, error) {
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

// ParseStringQuery parses a string parameter from the query string
func (b *BaseHandler) ParseStringQuery(c echo.Context, paramName string) (string, error) {
	s := c.QueryParam(paramName)
	if s == "" {
		return "", fmt.Errorf("missing %s query parameter", paramName)
	}

	return s, nil
}

// ParseBoolQuery parses a boolean parameter from the query string
func (b *BaseHandler) ParseBoolQuery(c echo.Context, paramName string) bool {
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

// LogDebug logs a debug message if logger is available
func (b *BaseHandler) LogDebug(format string, args ...any) {
	if b.Logger != nil {
		b.Logger.Debug(format, args...)
	}
}

// LogInfo logs an informational message if logger is available
func (b *BaseHandler) LogInfo(format string, args ...any) {
	if b.Logger != nil {
		b.Logger.Info(format, args...)
	}
}

// LogError logs an error message if logger is available
func (b *BaseHandler) LogWarn(format string, args ...any) {
	if b.Logger != nil {
		b.Logger.Warn(format, args...)
	}
}

// LogError logs an error message if logger is available
func (b *BaseHandler) LogError(format string, args ...any) {
	if b.Logger != nil {
		b.Logger.Error(format, args...)
	}
}
