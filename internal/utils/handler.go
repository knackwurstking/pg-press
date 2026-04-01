package utils

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/labstack/echo/v4"
)

// RedirectTo performs an HTTP redirect to the specified path
func RedirectTo(c echo.Context, path templ.SafeURL) *errors.HTTPError {
	err := c.Redirect(http.StatusSeeOther, string(path))
	if err != nil {
		return errors.NewHTTPError(err)
	}
	return nil
}

// SetHXTrigger sets HX-Trigger header
func SetHXTrigger(c echo.Context, events ...string) {
	slog.Debug("Setting HX-Trigger events", "events", strings.Join(events, ", "))
	c.Response().Header().Set("HX-Trigger", strings.Join(events, ", "))
}

// SetHXRedirect sets HX-Redirect header
func SetHXRedirect(c echo.Context, path templ.SafeURL) {
	slog.Debug("Setting HX-Redirect", "path", path)
	c.Response().Header().Set("HX-Redirect", string(path))
}

// SetHXAfterSettle sets HX-Trigger-After-Settle header with JSON data
func SetHXAfterSettle(c echo.Context, data map[string]any) {
	triggerDataJSON, _ := json.Marshal(data)
	slog.Debug("Setting HX-Trigger-After-Settle", "trigger_data", string(triggerDataJSON))
	c.Response().Header().Set("HX-Trigger-After-Settle", string(triggerDataJSON))
}
