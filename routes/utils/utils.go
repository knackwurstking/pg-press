// Package utils provides utility functions for HTTP route handlers.
//
// This package contains helper functions commonly used across different
// route handlers, including context management, request validation,
// response formatting, and authentication utilities.
package utils

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/pgvis"
)

// Context keys for storing values in Echo context
const (
	// UserContextKey is the key used to store user data in context
	UserContextKey = "user"
	// APIKeyContextKey is the key used to store API key in context
	APIKeyContextKey = "api_key"
)

// Common validation constants
const (
	// MaxPageSize defines the maximum number of items per page for pagination
	MaxPageSize = 100
	// DefaultPageSize is the default number of items per page
	DefaultPageSize = 20
	// MaxSearchLength is the maximum length for search queries
	MaxSearchLength = 500
)

// User Context Management

// GetUserFromContext retrieves the authenticated user from the Echo context.
// This function should be used in handlers that require user authentication.
//
// Parameters:
//   - ctx: Echo context containing the authenticated user
//
// Returns:
//   - *pgvis.User: The authenticated user
//   - *echo.HTTPError: HTTP error if user is not found in context
func GetUserFromContext(ctx echo.Context) (*pgvis.User, *echo.HTTPError) {
	user, ok := ctx.Get(UserContextKey).(*pgvis.User)
	if !ok {
		return nil, echo.NewHTTPError(http.StatusUnauthorized,
			"authentication required")
	}

	if user == nil {
		return nil, echo.NewHTTPError(http.StatusUnauthorized,
			"invalid user session")
	}

	return user, nil
}

// SetUserInContext stores a user in the Echo context for use by handlers.
//
// Parameters:
//   - ctx: Echo context to store the user in
//   - user: User to store in context
func SetUserInContext(ctx echo.Context, user *pgvis.User) {
	ctx.Set(UserContextKey, user)
}

// GetAPIKeyFromContext retrieves the API key from the Echo context.
//
// Parameters:
//   - ctx: Echo context containing the API key
//
// Returns:
//   - string: The API key
//   - *echo.HTTPError: HTTP error if API key is not found
func GetAPIKeyFromContext(ctx echo.Context) (string, *echo.HTTPError) {
	apiKey, ok := ctx.Get(APIKeyContextKey).(string)
	if !ok || apiKey == "" {
		return "", echo.NewHTTPError(http.StatusUnauthorized,
			"API key required")
	}

	return apiKey, nil
}

// SetAPIKeyInContext stores an API key in the Echo context.
//
// Parameters:
//   - ctx: Echo context to store the API key in
//   - apiKey: API key to store in context
func SetAPIKeyInContext(ctx echo.Context, apiKey string) {
	ctx.Set(APIKeyContextKey, apiKey)
}

// Request Parameter Utilities

// ParseIDParam extracts and validates an ID parameter from the URL.
//
// Parameters:
//   - ctx: Echo context containing the request
//   - paramName: Name of the parameter to extract
//
// Returns:
//   - int64: The parsed ID
//   - *echo.HTTPError: HTTP error if ID is invalid or missing
func ParseIDParam(ctx echo.Context, paramName string) (int64, *echo.HTTPError) {
	idStr := ctx.Param(paramName)
	if idStr == "" {
		return 0, echo.NewHTTPError(http.StatusBadRequest,
			"missing "+paramName+" parameter")
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, echo.NewHTTPError(http.StatusBadRequest,
			"invalid "+paramName+": must be a number")
	}

	if id <= 0 {
		return 0, echo.NewHTTPError(http.StatusBadRequest,
			"invalid "+paramName+": must be positive")
	}

	return id, nil
}

// ParseIntQuery extracts and validates an integer query parameter.
//
// Parameters:
//   - ctx: Echo context containing the request
//   - paramName: Name of the query parameter
//   - defaultValue: Default value if parameter is missing
//   - min: Minimum allowed value
//   - max: Maximum allowed value
//
// Returns:
//   - int: The parsed integer value
//   - *echo.HTTPError: HTTP error if parameter is invalid
func ParseIntQuery(ctx echo.Context, paramName string, defaultValue, min, max int) (int, *echo.HTTPError) {
	valueStr := ctx.QueryParam(paramName)
	if valueStr == "" {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, echo.NewHTTPError(http.StatusBadRequest,
			"invalid "+paramName+": must be a number")
	}

	if value < min {
		return 0, echo.NewHTTPError(http.StatusBadRequest,
			"invalid "+paramName+": minimum value is "+strconv.Itoa(min))
	}

	if value > max {
		return 0, echo.NewHTTPError(http.StatusBadRequest,
			"invalid "+paramName+": maximum value is "+strconv.Itoa(max))
	}

	return value, nil
}

// ParsePaginationParams extracts and validates pagination parameters from query string.
//
// Parameters:
//   - ctx: Echo context containing the request
//
// Returns:
//   - offset: Starting offset for pagination
//   - limit: Number of items per page
//   - *echo.HTTPError: HTTP error if parameters are invalid
func ParsePaginationParams(ctx echo.Context) (offset, limit int, httpErr *echo.HTTPError) {
	// Parse page number (1-based)
	page, httpErr := ParseIntQuery(ctx, "page", 1, 1, 10000)
	if httpErr != nil {
		return 0, 0, httpErr
	}

	// Parse page size
	limit, httpErr = ParseIntQuery(ctx, "limit", DefaultPageSize, 1, MaxPageSize)
	if httpErr != nil {
		return 0, 0, httpErr
	}

	// Calculate offset (0-based)
	offset = (page - 1) * limit

	return offset, limit, nil
}

// ParseRequiredIDQuery extracts and validates a required ID from query parameters.
//
// Parameters:
//   - ctx: Echo context containing the request
//   - paramName: Name of the query parameter
//
// Returns:
//   - int64: The parsed ID
//   - *echo.HTTPError: HTTP error if ID is invalid or missing
func ParseRequiredIDQuery(ctx echo.Context, paramName string) (int64, *echo.HTTPError) {
	idStr := ctx.QueryParam(paramName)
	if idStr == "" {
		return 0, echo.NewHTTPError(http.StatusBadRequest,
			"missing "+paramName+" query parameter")
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, echo.NewHTTPError(http.StatusBadRequest,
			"invalid "+paramName+": must be a number")
	}

	if id <= 0 {
		return 0, echo.NewHTTPError(http.StatusBadRequest,
			"invalid "+paramName+": must be positive")
	}

	return id, nil
}

// ParseSearchQuery extracts and validates a search query parameter.
//
// Parameters:
//   - ctx: Echo context containing the request
//
// Returns:
//   - string: The sanitized search query
//   - *echo.HTTPError: HTTP error if query is invalid
func ParseSearchQuery(ctx echo.Context) (string, *echo.HTTPError) {
	query := strings.TrimSpace(ctx.QueryParam("q"))

	if len(query) > MaxSearchLength {
		return "", echo.NewHTTPError(http.StatusBadRequest,
			"search query too long: maximum "+strconv.Itoa(MaxSearchLength)+" characters")
	}

	return query, nil
}

// Error Handling Utilities

// HandlePgvisError converts pgvis errors to appropriate HTTP responses.
//
// Parameters:
//   - ctx: Echo context for the response
//   - err: The pgvis error to handle
//
// Returns:
//   - *echo.HTTPError: Echo HTTP error with appropriate status code
func HandlePgvisError(ctx echo.Context, err error) *echo.HTTPError {
	if err == nil {
		return nil
	}

	code := pgvis.GetHTTPStatusCode(err)
	message := err.Error()

	// Special handling for different error types
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

	if pgvis.IsAuthError(err) {
		return echo.NewHTTPError(code, "Authentication failed")
	}

	// Generic error response
	return echo.NewHTTPError(code, message)
}

// Security Utilities

// SanitizeInput performs basic input sanitization for user-provided strings.
//
// Parameters:
//   - input: String to sanitize
//
// Returns:
//   - string: Sanitized string
func SanitizeInput(input string) string {
	// Trim whitespace
	sanitized := strings.TrimSpace(input)

	// Remove null bytes
	sanitized = strings.ReplaceAll(sanitized, "\x00", "")

	return sanitized
}

// ValidateStringLength validates that a string is within specified length bounds.
//
// Parameters:
//   - value: String to validate
//   - fieldName: Name of the field for error messages
//   - min: Minimum allowed length
//   - max: Maximum allowed length
//
// Returns:
//   - *echo.HTTPError: HTTP error if validation fails
func ValidateStringLength(value, fieldName string, min, max int) *echo.HTTPError {
	length := len(value)

	if length < min {
		return echo.NewHTTPError(http.StatusBadRequest,
			fieldName+" must be at least "+strconv.Itoa(min)+" characters")
	}

	if length > max {
		return echo.NewHTTPError(http.StatusBadRequest,
			fieldName+" must not exceed "+strconv.Itoa(max)+" characters")
	}

	return nil
}
