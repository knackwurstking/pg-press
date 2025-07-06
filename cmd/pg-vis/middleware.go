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
	// FIXME: Find a better way to to this !!!
	skipperRegExp := regexp.MustCompile(
		`(.*/login.*|.*\.css|manifest.json|.*\.png|.*\.ico)`,
	)

	return middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
		Skipper: func(c echo.Context) bool {
			url := c.Request().URL.String()

			if ok := skipperRegExp.MatchString(url); ok {
				log.Debugf("KeyAuth -> Skipper -> Skip: %s", url)
				return true
			}

			return false
		},

		KeyLookup: "header:" + echo.HeaderAuthorization + ",query:access_token,cookie:" + html.CookieName,

		AuthScheme: "Bearer",

		Validator: func(auth string, c echo.Context) (bool, error) {
			log.Debugf("KeyAuth -> Validator -> User-Agent=%#v", c.Request().UserAgent())

			if cookie, err := c.Cookie(html.CookieName); err == nil {
				c, err := db.Cookies.Get(cookie.Value)
				if err == nil {
					log.Debugf("KeyAuth -> Validator -> cookie found")
					auth = c.ApiKey
				}
			}

			user, err := db.Users.GetUserFromApiKey(auth)
			if err != nil {
				return false, fmt.Errorf("get user from db with auth %#v failed: %s", auth, err.Error())
			}

			log.Debugf(
				"KeyAuth -> Validator -> telegram_id=%#v; user_name=%#v",
				user.TelegramID, user.UserName,
			)

			if user.ApiKey == auth {
				c.Set("user", user)
				return true, nil
			}

			return false, nil
		},

		ErrorHandler: func(err error, c echo.Context) error {
			log.Errorf("KeyAuth -> ErrorHandler -> %s", err.Error())
			log.Debugf("KeyAuth -> ErrorHandler -> User-Agent=%#v", c.Request().UserAgent())
			return c.Redirect(http.StatusSeeOther, serverPathPrefix+"/login")
		},
	})
}
