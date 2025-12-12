package urlb

import (
	"net/http"

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
