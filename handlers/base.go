package handlers

import (
	"github.com/knackwurstking/pg-press/logger"
	"github.com/knackwurstking/pg-press/services"
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
//func (b *BaseHandler) SetHXTrigger(c echo.Context) {
//	c.Response().Header().Set("HX-Trigger", "pageLoaded")
//}
//
//func (b *BaseHandler) SetHXRedirect(c echo.Context, path string) {
//	c.Response().Header().Set("HX-Redirect", env.ServerPathPrefix+path)
//}
//
