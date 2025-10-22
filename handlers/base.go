package handlers

import (
	"github.com/knackwurstking/pgpress/logger"
	"github.com/knackwurstking/pgpress/services"
)

type Base struct {
	Registry *services.Registry
	Log      *logger.Logger
}

func NewBase(r *services.Registry, l *logger.Logger) *Base {
	return &Base{
		Registry: r,
		Log:      l,
	}
}

// TODO: Just create functions in ./utils.go instead
//func (b *BaseHandler) RenderError(c echo.Context, statusCode int, message string) error {
//	return echo.NewHTTPError(statusCode, message)
//}
//
//func (b *BaseHandler) RenderBadRequest(c echo.Context, message string) error {
//	return echo.NewHTTPError(http.StatusBadRequest, message)
//}
//
//// RenderUnauthorized creates a 401 Unauthorized response
//// For example when an action requires admin privileges
//func (b *BaseHandler) RenderUnauthorized(c echo.Context, message string) error {
//	return echo.NewHTTPError(http.StatusUnauthorized, message)
//}
//
//// RenderNotFound creates a 404 Not Found response
//func (b *BaseHandler) RenderNotFound(c echo.Context, message string) error {
//	return echo.NewHTTPError(http.StatusNotFound, message)
//}
//
//// getSanitizedFormValue retrieves and sanitizes form input
//func (b *BaseHandler) GetSanitizedFormValue(c echo.Context, fieldName string) string {
//	formParams, _ := c.FormParams()
//	sanitized := strings.TrimSpace(formParams.Get(fieldName))
//	sanitized = strings.ReplaceAll(sanitized, "\x00", "")
//	return sanitized
//}
//
//func (b *BaseHandler) SanitizeFilename(filename string) string {
//	if idx := strings.LastIndex(filename, "."); idx > 0 {
//		filename = filename[:idx]
//	}
//
//	filename = strings.ReplaceAll(filename, " ", "_")
//	filename = strings.ReplaceAll(filename, "-", "_")
//	filename = strings.ReplaceAll(filename, "(", "_")
//	filename = strings.ReplaceAll(filename, ")", "_")
//	filename = strings.ReplaceAll(filename, "[", "_")
//	filename = strings.ReplaceAll(filename, "]", "_")
//
//	for strings.Contains(filename, "__") {
//		filename = strings.ReplaceAll(filename, "__", "_")
//	}
//
//	filename = strings.Trim(filename, "_")
//
//	if filename == "" {
//		filename = "attachment"
//	}
//
//	return filename
//}
//
//func (b *BaseHandler) ParseInt8Param(c echo.Context, paramName string) (int8, error) {
//	idStr := c.Param(paramName)
//	if idStr == "" {
//		return 0, fmt.Errorf("missing %s parameter", paramName)
//	}
//
//	id, err := strconv.ParseInt(idStr, 10, 8)
//	if err != nil {
//		return 0, fmt.Errorf("invalid %s parameter: must be a number", paramName)
//	}
//	return int8(id), nil
//}
//
//// ParseInt64Param parses an int64 parameter from the URL path
//func (b *BaseHandler) ParseInt64Param(c echo.Context, paramName string) (int64, error) {
//	idStr := c.Param(paramName)
//	if idStr == "" {
//		return 0, fmt.Errorf("missing %s parameter", paramName)
//	}
//
//	id, err := strconv.ParseInt(idStr, 10, 64)
//	if err != nil {
//		return 0, fmt.Errorf("invalid %s parameter: must be a number", paramName)
//	}
//	return id, nil
//}
//
//// ParseInt64Query parses an int64 parameter from the query string
//func (b *BaseHandler) ParseInt64Query(c echo.Context, paramName string) (int64, error) {
//	idStr := c.QueryParam(paramName)
//	if idStr == "" {
//		return 0, fmt.Errorf("missing %s query parameter", paramName)
//	}
//
//	id, err := strconv.ParseInt(idStr, 10, 64)
//	if err != nil {
//		return 0, fmt.Errorf("invalid %s query parameter: must be a number", paramName)
//	}
//
//	return id, nil
//}
//
//// ParseBoolQuery parses a boolean parameter from the query string
//func (b *BaseHandler) ParseBoolQuery(c echo.Context, paramName string) bool {
//	value := c.QueryParam(paramName)
//	if value == "" {
//		return false
//	}
//
//	boolValue, err := strconv.ParseBool(value)
//	if err != nil {
//		return false
//	}
//	return boolValue
//}
//
//func (b *BaseHandler) SetHXTrigger(c echo.Context) {
//	c.Response().Header().Set("HX-Trigger", "pageLoaded")
//}
//
//func (b *BaseHandler) SetHXRedirect(c echo.Context, path string) {
//	c.Response().Header().Set("HX-Redirect", env.ServerPathPrefix+path)
//}
//
