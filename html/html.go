package html

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/html/handler/profile"
	"github.com/knackwurstking/pg-vis/html/handler/troublereports"
	"github.com/knackwurstking/pg-vis/pgvis"
)

const (
	CookieName               = "pgvis-api-key"
	CookieExpirationDuration = time.Hour * 24 * 31 * 6
)

//go:embed templates
var templates embed.FS

//go:embed static
var static embed.FS

type Options struct {
	ServerPathPrefix string
	DB               *pgvis.DB
}

func Serve(e *echo.Echo, o Options) {
	e.StaticFS(o.ServerPathPrefix+"/", echo.MustSubFS(static, "static"))

	serveHome(e, o)
	serveFeed(e, o)
	serveLogin(e, o)
	serveLogout(e, o)

	profile.Serve(templates, o.ServerPathPrefix, e, o.DB)
	troublereports.Serve(templates, o.ServerPathPrefix, e, o.DB)
}

func serveHome(e *echo.Echo, options Options) {
	e.GET(options.ServerPathPrefix+"/", func(c echo.Context) error {
		t, err := template.ParseFS(templates,
			"templates/layout.html",
			"templates/home.html",
		)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		err = t.Execute(c.Response(), nil)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		return nil
	})
}

func serveFeed(e *echo.Echo, options Options) {
	e.GET(options.ServerPathPrefix+"/feed", func(c echo.Context) error {
		t, err := template.ParseFS(templates,
			"templates/layout.html",
			"templates/feed.html",
		)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		err = t.Execute(c.Response(), nil)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		return nil
	})
}

type LoginPageData struct {
	ApiKey        string
	InvalidApiKey bool
}

func serveLogin(e *echo.Echo, options Options) {
	e.GET(options.ServerPathPrefix+"/login", func(c echo.Context) error {
		v, err := c.FormParams()
		apiKey := v.Get("api-key")

		if ok, err := handleApiKeyLogin(apiKey, options.DB, c); ok {
			if err = c.Redirect(http.StatusSeeOther, "./profile"); err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
			} else {
				return nil
			}
		} else {
			if err != nil {
				log.Errorf("/login -> Invalid Api Key: %s", err.Error())
			}
		}

		pageData := LoginPageData{
			ApiKey:        apiKey,
			InvalidApiKey: apiKey != "",
		}

		t, err := template.ParseFS(templates,
			"templates/layout.html",
			"templates/login.html",
		)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		err = t.Execute(c.Response(), pageData)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		return nil
	})
}

func serveLogout(e *echo.Echo, options Options) {
	e.GET(options.ServerPathPrefix+"/logout", func(c echo.Context) error {
		if cookie, err := c.Cookie(CookieName); err == nil {
			options.DB.Cookies.Remove(cookie.Value)

			err = c.Redirect(http.StatusSeeOther, "./login")
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
			}
		}

		return nil
	})
}

func handleApiKeyLogin(apiKey string, db *pgvis.DB, ctx echo.Context) (ok bool, err error) {
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
			log.Debug("Removing the old cookie...")
			if err = db.Cookies.Remove(cookie.Value); err != nil {
				log.Warnf("Removing the old cookie failed: %s", err.Error())
			}
		}

		log.Debugf(
			"Set cookie and redirect to /profile (id: %#v; user: %#v)",
			u.TelegramID, u.UserName,
		)

		cookie := new(http.Cookie)

		cookie.Name = CookieName
		cookie.Value = uuid.New().String()
		cookie.Expires = time.Now().Add(CookieExpirationDuration)

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
