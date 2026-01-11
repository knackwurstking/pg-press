package utils

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/logger"
	"github.com/labstack/echo/v4"
)

var log = logger.New("urlb")

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
	log.Debug("Setting HX-Trigger to %s", strings.Join(events, ", "))
	c.Response().Header().Set("HX-Trigger", strings.Join(events, ", "))
}

// SetHXRedirect sets HX-Redirect header
func SetHXRedirect(c echo.Context, path templ.SafeURL) {
	log.Debug("Setting HX-Redirect to %s", path)
	c.Response().Header().Set("HX-Redirect", string(path))
}

// SetHXAfterSettle sets HX-Trigger-After-Settle header with JSON data
func SetHXAfterSettle(c echo.Context, data map[string]any) {
	triggerDataJSON, _ := json.Marshal(data)
	log.Debug("Setting HX-Trigger-After-Settle to %s", string(triggerDataJSON))
	c.Response().Header().Set("HX-Trigger-After-Settle", string(triggerDataJSON))
}
