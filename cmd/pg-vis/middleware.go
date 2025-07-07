package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/knackwurstking/pg-vis/html"
	"github.com/knackwurstking/pg-vis/pkg/pgvis"
)

var (
	// FIXME: Find a better way to to this !!!
	keyAuthSkipperRegExp = regexp.MustCompile(
		`(.*/login.*|.*\.css|manifest.json|.*\.png|.*\.ico)`,
	)
)

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
		Skipper:    keyAuthSkipper,
		KeyLookup:  "header:" + echo.HeaderAuthorization + ",query:access_token,cookie:" + html.CookieName,
		AuthScheme: "Bearer",

		Validator: func(auth string, ctx echo.Context) (bool, error) {
			return keyAuthValidator(auth, ctx, db)
		},

		ErrorHandler: func(err error, c echo.Context) error {
			log.Errorf("KeyAuth -> ErrorHandler -> %s", err.Error())
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
		if cookie, err := ctx.Cookie(html.CookieName); err == nil {
			if c, err := db.Cookies.Get(cookie.Value); err != nil {
				return false, nil
			} else {
				log.Debugf("KeyAuth -> Validator -> Cookie found, try to search for the user again")

				if user, err = db.Users.GetUserFromApiKey(c.ApiKey); err != nil {
					return false, nil
				}
			}
		} else {
			return false, nil
		}
	} else {
		// Api Key found in users table (database), remove old existing cookie
		if cookie, err := ctx.Cookie(html.CookieName); err == nil {
			db.Cookies.Remove(cookie.Value)
		}
	}

	log.Debugf(
		"KeyAuth -> Validator -> Authorized: telegram_id=%#v; user_name=%#v",
		user.TelegramID, user.UserName,
	)

	ctx.Set("user", user)

	return true, nil
}
