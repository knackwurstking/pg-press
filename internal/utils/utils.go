package utils

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pgpress/internal/dberror"
	"github.com/knackwurstking/pgpress/internal/models"
)

func GetUserFromContext(ctx echo.Context) (*models.User, error) {
	user, ok := ctx.Get("user").(*models.User)
	if !ok {
		return nil, echo.NewHTTPError(
			http.StatusUnauthorized,
			"authentication required",
		)
	}
	if user == nil {
		return nil, echo.NewHTTPError(
			http.StatusUnauthorized,
			"invalid user session",
		)
	}
	return user, nil
}

func ParseStringQuery(ctx echo.Context, paramName string) (string, error) {
	s := ctx.QueryParam(paramName)
	if s == "" {
		return "", fmt.Errorf("missing %s query parameter", paramName)
	}

	return s, nil
}

func ParseInt64Param(ctx echo.Context, paramName string) (int64, error) {
	idStr := ctx.Param(paramName)
	if idStr == "" {
		return 0, echo.NewHTTPError(http.StatusBadRequest, "missing "+paramName+" parameter")
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, echo.NewHTTPError(http.StatusBadRequest, "invalid "+paramName+": must be a number")
	}
	return id, nil
}

func ParseInt64Query(ctx echo.Context, paramName string) (int64, error) {
	idStr := ctx.QueryParam(paramName)
	if idStr == "" {
		return 0, echo.NewHTTPError(http.StatusBadRequest, "missing "+paramName+" query parameter")
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, echo.NewHTTPError(http.StatusBadRequest, "invalid "+paramName+": must be a number")
	}
	return id, nil
}

func ParseBoolQuery(ctx echo.Context, paramName string) bool {
	value := ctx.QueryParam(paramName)
	if value == "" {
		return false
	}

	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return false
	}
	return boolValue
}

func SanitizeInput(input string) string {
	sanitized := strings.TrimSpace(input)
	sanitized = strings.ReplaceAll(sanitized, "\x00", "")
	return sanitized
}

func HandleTemplate(
	c echo.Context,
	pageData any,
	templates fs.FS,
	patterns []string,
) error {
	t, err := template.ParseFS(templates, patterns...)
	if err != nil {
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			dberror.WrapError(err, "failed to parse templates"),
		)
	}

	if err := t.Execute(c.Response(), pageData); err != nil {
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			dberror.WrapError(err, "failed to render page"),
		)
	}

	return nil
}
