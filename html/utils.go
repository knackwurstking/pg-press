package html

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
	"github.com/knackwurstking/pg-vis/pkg/pgvis"
	"github.com/labstack/echo/v4"
)

func getUserFromContext(ctx echo.Context) (*pgvis.User, *echo.HTTPError) {
	user, ok := ctx.Get("user").(*pgvis.User)
	if !ok {
		return nil, echo.NewHTTPError(http.StatusInternalServerError,
			errors.New("user is missing in context"))
	}

	return user, nil
}

func handleLoginApiKey(apiKey string, db *pgvis.DB, ctx echo.Context) (ok bool, err error) {
	if apiKey == "" {
		return false, nil
	}

	u, err := db.Users.GetUserFromApiKey(apiKey)
	if err != nil {
		return false, fmt.Errorf("database error: %s", err.Error())
	}

	// Set cookie and redirect to "/"
	if u.ApiKey == apiKey {
		if cookie, err := ctx.Cookie(CookieName); err == nil {
			log.Debug("HandleLoginApiKey -> Removing the old cookie...")
			if err = db.Cookies.Remove(cookie.Value); err != nil {
				log.Warnf("HandleLoginApiKey -> Removing the old cookie failed: %s", err.Error())
			}
		}

		log.Debugf(
			"HandleLoginApiKey -> Set cookie and redirect to /profile (id: %#v; user: %#v)",
			u.TelegramID, u.UserName,
		)

		cookie := new(http.Cookie)

		cookie.Name = CookieName
		cookie.Value = uuid.New().String()
		ctx.SetCookie(cookie)

		db.Cookies.Add(&pgvis.Cookie{
			UserAgent: ctx.Request().UserAgent(),
			Value:     cookie.Value,
			ApiKey:    apiKey,
			LastLogin: time.Now().UnixMilli(),
		})

		return true, nil
	}

	return false, nil
}
