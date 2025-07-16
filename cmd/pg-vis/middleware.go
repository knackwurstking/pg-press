package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"slices"
	"time"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/knackwurstking/pg-vis/routes"
)

var (
	// FIXME: Find a better way to to this !!!
	keyAuthSkipperRegExp = regexp.MustCompile(
		`(.*/login.*|.*\.css|manifest.json|.*\.png|.*\.ico|.*service-worker\.js|.*\.woff|.*\.woff2|.*htmx.min.js)`,
	)

	pages []string
)

func init() {
	pages = []string{
		serverPathPrefix + "/",
		serverPathPrefix + "/feed",
		serverPathPrefix + "/profile",
		serverPathPrefix + "/trouble-reports",
	}
}

func middlewareLogger() echo.MiddlewareFunc {
	return middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${custom} ${status} ${method} ${uri} (${remote_ip}) ${latency_human}\n",
		Output: os.Stderr,

		CustomTagFunc: func(c echo.Context, buf *bytes.Buffer) (int, error) {
			t := time.Now()
			buf.Write(fmt.Appendf(nil,
				"%d/%02d/%02d %02d:%02d:%02d",
				t.Year(), int(t.Month()), t.Day(),
				t.Hour(), t.Minute(), t.Second(),
			))

			return 0, nil
		},
	})
}

func middlewareKeyAuth(db *pgvis.DB) echo.MiddlewareFunc {
	return middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
		Skipper: keyAuthSkipper,

		KeyLookup: "header:" + echo.HeaderAuthorization +
			",query:access_token,cookie:" + routes.CookieName,

		AuthScheme: "Bearer",

		Validator: func(auth string, ctx echo.Context) (bool, error) {
			return keyAuthValidator(auth, ctx, db)
		},

		ErrorHandler: func(err error, c echo.Context) error {
			log.Errorf("KeyAuth -> ErrorHandler -> %#v", err.Error())
			log.Debugf("KeyAuth -> ErrorHandler -> User-Agent=%#v", c.Request().UserAgent())
			return c.Redirect(http.StatusSeeOther, serverPathPrefix+"/login")
		},
	})
}

func keyAuthSkipper(ctx echo.Context) bool {
	url := ctx.Request().URL.String()

	if ok := keyAuthSkipperRegExp.MatchString(url); ok {
		log.Debugf("KeyAuth -> Skipper -> Skip: %s", url)
		return true
	}

	return false
}

func keyAuthValidator(auth string, ctx echo.Context, db *pgvis.DB) (bool, error) {
	log.Debugf("KeyAuth -> Validator -> User-Agent=%#v", ctx.Request().UserAgent())

	user, err := db.Users.GetUserFromApiKey(auth)
	if err != nil {
		if cookie, err := ctx.Cookie(routes.CookieName); err == nil {
			if c, err := db.Cookies.Get(cookie.Value); err != nil {
				return false, nil
			} else {
				log.Debugf("KeyAuth -> Validator -> Cookie found, try to search for the user...")

				if user, err = db.Users.GetUserFromApiKey(c.ApiKey); err != nil {
					return false, nil
				}

				log.Debugf("KeyAuth -> Validator -> ctx.Request().URL.Path=%#v", ctx.Request().URL.Path)
				if slices.Contains(pages, ctx.Request().URL.Path) {
					log.Debugf("KeyAuth -> Validator -> Update cookies last login timestamp")

					c.LastLogin = time.Now().UnixMilli()

					// Update expiration for the browser cookie
					cookie.Expires = time.Now().Add(routes.CookieExpirationDuration)

					if err := db.Cookies.Update(c.Value, c); err != nil {
						log.Errorf("KeyAuth -> Validator -> Update cookies database error: %#v", err)
					}
				}
			}
		} else {
			return false, nil
		}
	}

	log.Debugf(
		"KeyAuth -> Validator -> Authorized: telegram_id=%#v; user_name=%#v",
		user.TelegramID, user.UserName,
	)

	ctx.Set("user", user)

	return true, nil
}
