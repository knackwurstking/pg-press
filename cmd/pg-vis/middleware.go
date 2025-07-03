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
		Format: "${custom} ---> ${status} ${method} ${uri} (${remote_ip}) ${latency_human}\n",
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
	// FIXME: Find a better way to to this
	skipperRegExp := regexp.MustCompile(
		`(.*/login.*|.*pico.lime.min.css|manifest.json|.*\.png|.*\.ico)`,
	)

	return middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
		Skipper: func(c echo.Context) bool {
			url := c.Request().URL.String()
			log.Debugf("Auth: Skipper: %s", url)

			return skipperRegExp.MatchString(url)
		},

		KeyLookup: "header:" + echo.HeaderAuthorization + ",query:access_token,cookie:" + html.CookieName,

		AuthScheme: "Bearer",

		Validator: func(auth string, c echo.Context) (bool, error) {
			log.Debugf("Auth: Validator: %s", c.Request().UserAgent())

			if cookie, err := c.Cookie(html.CookieName); err == nil {
				c, err := db.Cookies.Get(cookie.Value)
				if err == nil {
					log.Debugf("Auth: Validator: cookie found")
					auth = c.ApiKey
				}
			}

			user, err := db.Users.GetUserFromApiKey(auth)
			if err != nil {
				return false, fmt.Errorf("get user from db: %s (%#v)", err.Error(), auth)
			}

			return user.ApiKey == auth, nil
		},

		ErrorHandler: func(err error, c echo.Context) error {
			log.Debugf("Auth ErrorHandler: %s", err.Error())

			if err != nil {
				return c.Redirect(http.StatusSeeOther, serverPathPrefix+"/login")
			}

			return nil
		},
	})
}
