// Package utils provides utility functions for HTTP route handlers.
package utils

import (
	"bytes"
	"errors"
	"html/template"
	"io/fs"
	"net/http"
	"strconv"
	"strings"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/labstack/echo/v4"
)

const (
	userContextKey   = "user"
	apiKeyContextKey = "api_key"

	authenticationRequiredMessage = "authentication required"
	invalidUserSessionMessage     = "invalid user session"
	templateParseErrorMessage     = "failed to parse templates"
	templateExecuteErrorMessage   = "failed to render page"
)

func GetUserFromContext(ctx echo.Context) (*pgvis.User, *echo.HTTPError) {
	user, ok := ctx.Get(userContextKey).(*pgvis.User)
	if !ok {
		return nil, echo.NewHTTPError(
			http.StatusUnauthorized,
			authenticationRequiredMessage,
		)
	}
	if user == nil {
		return nil, echo.NewHTTPError(
			http.StatusUnauthorized,
			invalidUserSessionMessage,
		)
	}
	return user, nil
}

func SetUserInContext(ctx echo.Context, user *pgvis.User) {
	ctx.Set(userContextKey, user)
}

func GetAPIKeyFromContext(ctx echo.Context) (string, *echo.HTTPError) {
	apiKey, ok := ctx.Get(apiKeyContextKey).(string)
	if !ok || apiKey == "" {
		return "", echo.NewHTTPError(http.StatusUnauthorized, "API key required")
	}
	return apiKey, nil
}

func SetAPIKeyInContext(ctx echo.Context, apiKey string) {
	ctx.Set(apiKeyContextKey, apiKey)
}

func ParseInt64Param(ctx echo.Context, paramName string) (int64, *echo.HTTPError) {
	idStr := ctx.Param(paramName)
	if idStr == "" {
		return 0, echo.NewHTTPError(http.StatusBadRequest, "missing "+paramName+" parameter")
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, echo.NewHTTPError(http.StatusBadRequest, "invalid "+paramName+": must be a number")
	}
	if id <= 0 {
		return 0, echo.NewHTTPError(http.StatusBadRequest, "invalid "+paramName+": must be positive")
	}
	return id, nil
}

func ParseInt64Query(ctx echo.Context, paramName string) (int64, *echo.HTTPError) {
	idStr := ctx.QueryParam(paramName)
	if idStr == "" {
		return 0, echo.NewHTTPError(http.StatusBadRequest, "missing "+paramName+" query parameter")
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, echo.NewHTTPError(http.StatusBadRequest, "invalid "+paramName+": must be a number")
	}
	if id <= 0 {
		return 0, echo.NewHTTPError(http.StatusBadRequest, "invalid "+paramName+": must be positive")
	}
	return id, nil
}

func HandlePgvisError(ctx echo.Context, err error) *echo.HTTPError {
	if err == nil {
		return nil
	}

	code := pgvis.GetHTTPStatusCode(err)
	message := err.Error()

	if pgvis.IsValidationError(err) {
		return echo.NewHTTPError(code, map[string]any{
			"error":   "Validation failed",
			"code":    code,
			"status":  http.StatusText(code),
			"details": err,
		})
	}

	if errors.Is(err, pgvis.ErrNotFound) {
		return echo.NewHTTPError(code, "Resource not found")
	}

	return echo.NewHTTPError(code, message)
}

func SanitizeInput(input string) string {
	sanitized := strings.TrimSpace(input)
	sanitized = strings.ReplaceAll(sanitized, "\x00", "")
	return sanitized
}

func HandleTemplate(c echo.Context, pageData any, templates fs.FS, patterns []string) *echo.HTTPError {
	t, err := template.ParseFS(templates, patterns...)
	if err != nil {
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			pgvis.WrapError(err, templateParseErrorMessage),
		)
	}

	if err := t.Execute(c.Response(), pageData); err != nil {
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			pgvis.WrapError(err, templateExecuteErrorMessage),
		)
	}

	return nil
}

func RenderTemplateToString(templates fs.FS, patterns []string, pageData any) (string, error) {
	t, err := template.ParseFS(templates, patterns...)
	if err != nil {
		return "", pgvis.WrapError(err, templateParseErrorMessage)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, pageData); err != nil {
		return "", pgvis.WrapError(err, templateExecuteErrorMessage)
	}

	return buf.String(), nil
}
