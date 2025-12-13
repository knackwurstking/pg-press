package urlb

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/labstack/echo/v4"
)

// RedirectTo performs an HTTP redirect to the specified path
func RedirectTo(c echo.Context, path templ.SafeURL) *errors.MasterError {
	err := c.Redirect(http.StatusSeeOther, string(path))
	if err != nil {
		return errors.NewMasterError(err, 0)
	}
	return nil
}

// SetHXTrigger sets HX-Trigger header
func SetHXTrigger(c echo.Context, events ...string) {
	c.Response().Header().Set("HX-Trigger", strings.Join(events, ", "))
}

// SetHXRedirect sets HX-Redirect header
func SetHXRedirect(c echo.Context, path templ.SafeURL) {
	c.Response().Header().Set("HX-Redirect", string(path))
}

// SetHXAfterSettle sets HX-Trigger-After-Settle header with JSON data
func SetHXAfterSettle(c echo.Context, data map[string]any) {
	triggerDataJSON, _ := json.Marshal(data)
	c.Response().Header().Set("HX-Trigger-After-Settle", string(triggerDataJSON))
}
