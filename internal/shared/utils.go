package shared

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/labstack/echo/v4"
)

/*******************************************************************************
 * String Utils
 ******************************************************************************/

// MaskString masks sensitive strings by showing only the first and last 4 characters.
// For strings with 8 or fewer characters, all characters are masked.
func MaskString(s string) string {
	if len(s) <= 8 {
		return strings.Repeat("*", len(s))
	}
	return s[:4] + strings.Repeat("*", len(s)-8) + s[len(s)-4:]
}

/*******************************************************************************
 * Handler Utils
 ******************************************************************************/

func GetUserFromContext(c echo.Context) (*User, *errors.MasterError) {
	u := c.Get("user")
	if u == nil {
		return nil, errors.NewMasterError(
			fmt.Errorf("no user"), http.StatusUnauthorized,
		)
	}

	user, ok := u.(*User)
	if !ok || user.Validate() != nil {
		return nil, errors.NewMasterError(
			fmt.Errorf("invalid user"), http.StatusUnauthorized,
		)
	}

	return user, nil
}
