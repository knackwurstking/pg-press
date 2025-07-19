// Package utils provides utility functions for HTTP route handlers.
package utils

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/labstack/echo/v4"
)

const (
	UserContextKey   = "user"
	APIKeyContextKey = "api_key"

	MaxPageSize     = 100
	DefaultPageSize = 20
	MaxSearchLength = 500
)

func GetUserFromContext(ctx echo.Context) (*pgvis.User, *echo.HTTPError) {
	user, ok := ctx.Get(UserContextKey).(*pgvis.User)
	if !ok {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}
	if user == nil {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "invalid user session")
	}
	return user, nil
}

func SetUserInContext(ctx echo.Context, user *pgvis.User) {
	ctx.Set(UserContextKey, user)
}

func GetAPIKeyFromContext(ctx echo.Context) (string, *echo.HTTPError) {
	apiKey, ok := ctx.Get(APIKeyContextKey).(string)
	if !ok || apiKey == "" {
		return "", echo.NewHTTPError(http.StatusUnauthorized, "API key required")
	}
	return apiKey, nil
}

func SetAPIKeyInContext(ctx echo.Context, apiKey string) {
	ctx.Set(APIKeyContextKey, apiKey)
}

func ParseIDParam(ctx echo.Context, paramName string) (int64, *echo.HTTPError) {
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

func ParseIntQuery(ctx echo.Context, paramName string, defaultValue, min, max int) (int, *echo.HTTPError) {
	valueStr := ctx.QueryParam(paramName)
	if valueStr == "" {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, echo.NewHTTPError(http.StatusBadRequest, "invalid "+paramName+": must be a number")
	}
	if value < min {
		return 0, echo.NewHTTPError(http.StatusBadRequest, "invalid "+paramName+": minimum value is "+strconv.Itoa(min))
	}
	if value > max {
		return 0, echo.NewHTTPError(http.StatusBadRequest, "invalid "+paramName+": maximum value is "+strconv.Itoa(max))
	}
	return value, nil
}

func ParsePaginationParams(ctx echo.Context) (offset, limit int, httpErr *echo.HTTPError) {
	page, httpErr := ParseIntQuery(ctx, "page", 1, 1, 10000)
	if httpErr != nil {
		return 0, 0, httpErr
	}

	limit, httpErr = ParseIntQuery(ctx, "limit", DefaultPageSize, 1, MaxPageSize)
	if httpErr != nil {
		return 0, 0, httpErr
	}

	offset = (page - 1) * limit
	return offset, limit, nil
}

func ParseRequiredIDQuery(ctx echo.Context, paramName string) (int64, *echo.HTTPError) {
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

func ParseSearchQuery(ctx echo.Context) (string, *echo.HTTPError) {
	query := strings.TrimSpace(ctx.QueryParam("q"))
	if len(query) > MaxSearchLength {
		return "", echo.NewHTTPError(http.StatusBadRequest, "search query too long: maximum "+strconv.Itoa(MaxSearchLength)+" characters")
	}
	return query, nil
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

	if pgvis.IsNotFound(err) {
		return echo.NewHTTPError(code, "Resource not found")
	}

	return echo.NewHTTPError(code, message)
}

func SanitizeInput(input string) string {
	sanitized := strings.TrimSpace(input)
	sanitized = strings.ReplaceAll(sanitized, "\x00", "")
	return sanitized
}

func ValidateStringLength(value, fieldName string, min, max int) *echo.HTTPError {
	length := len(value)
	if length < min {
		return echo.NewHTTPError(http.StatusBadRequest, fieldName+" must be at least "+strconv.Itoa(min)+" characters")
	}
	if length > max {
		return echo.NewHTTPError(http.StatusBadRequest, fieldName+" must not exceed "+strconv.Itoa(max)+" characters")
	}
	return nil
}
