// NOTE: Cleaned up by AI
// Package routes provides HTTP route handlers and web interface for the pgvis application.
package routes

import (
	"embed"
	"html/template"
	"net/http"
	"time"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/knackwurstking/pg-vis/routes/feed"
	"github.com/knackwurstking/pg-vis/routes/nav"
	"github.com/knackwurstking/pg-vis/routes/profile"
	"github.com/knackwurstking/pg-vis/routes/shared"
	"github.com/knackwurstking/pg-vis/routes/troublereports"
)

const (
	CookieName               = "pgvis-api-key"
	CookieExpirationDuration = time.Hour * 24 * 31 * 6
)

var (
	//go:embed templates
	templates embed.FS

	//go:embed static
	static embed.FS
)

type Options struct {
	ServerPathPrefix string
	DB               *pgvis.DB
}

func Serve(e *echo.Echo, o Options) {
	e.StaticFS(o.ServerPathPrefix+"/", echo.MustSubFS(static, "static"))

	serveHome(e, o)
	serveLogin(e, o)
	serveLogout(e, o)

	feed.Serve(templates, o.ServerPathPrefix, e, o.DB)
	profile.Serve(templates, o.ServerPathPrefix, e, o.DB)
	troublereports.Serve(templates, o.ServerPathPrefix, e, o.DB)
	nav.Serve(templates, o.ServerPathPrefix, e, o.DB)
}

func serveHome(e *echo.Echo, options Options) {
	e.GET(options.ServerPathPrefix+"/", func(c echo.Context) error {
		t, err := template.ParseFS(templates,
			shared.LayoutTemplatePath,
			shared.HomeTemplatePath,
			shared.NavFeedTemplatePath,
		)
		if err != nil {
			return echo.NewHTTPError(
				http.StatusInternalServerError,
				"failed to parse templates",
			)
		}

		if err := t.Execute(c.Response(), nil); err != nil {
			return echo.NewHTTPError(
				http.StatusInternalServerError,
				"failed to execute template",
			)
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
		formParams, _ := c.FormParams()
		apiKey := formParams.Get(shared.APIKeyFormField)

		if apiKey != "" {
			if handleApiKeyLogin(apiKey, options.DB, c) {
				if err := c.Redirect(http.StatusSeeOther, "./profile"); err != nil {
					return echo.NewHTTPError(
						http.StatusInternalServerError,
						"redirect failed",
					)
				}
				return nil
			}
		}

		pageData := LoginPageData{
			ApiKey:        apiKey,
			InvalidApiKey: apiKey != "",
		}

		t, err := template.ParseFS(templates,
			shared.LayoutTemplatePath,
			shared.LoginTemplatePath,
		)
		if err != nil {
			return echo.NewHTTPError(
				http.StatusInternalServerError,
				"failed to load page templates",
			)
		}

		if err := t.Execute(c.Response(), pageData); err != nil {
			return echo.NewHTTPError(
				http.StatusInternalServerError,
				"failed to render page",
			)
		}

		return nil
	})
}

func serveLogout(e *echo.Echo, options Options) {
	e.GET(options.ServerPathPrefix+"/logout", func(c echo.Context) error {
		if cookie, err := c.Cookie(CookieName); err == nil {
			if err := options.DB.Cookies.Remove(cookie.Value); err != nil {
				log.Errorf("Failed to remove cookie from database: %s", err)
			}
		}

		if err := c.Redirect(http.StatusSeeOther, "./login"); err != nil {
			log.Errorf("Failed to redirect after logout: %s", err)

			return echo.NewHTTPError(
				http.StatusInternalServerError,
				"redirect failed",
			)
		}

		return nil
	})
}

func handleApiKeyLogin(apiKey string, db *pgvis.DB, ctx echo.Context) (ok bool) {
	if apiKey == "" {
		return false
	}

	user, err := db.Users.GetUserFromApiKey(apiKey)
	if err != nil {
		if pgvis.IsNotFound(err) {
			return false
		}

		log.Errorf("Failed to get user from API key: %s", err)

		return false
	}

	if user.ApiKey != apiKey {
		return false
	}

	if existingCookie, err := ctx.Cookie(CookieName); err == nil {
		log.Debug("Removing existing authentication cookie")

		if err := db.Cookies.Remove(existingCookie.Value); err != nil {
			log.Warnf("Failed to remove existing cookie: %s", err)
		}
	}

	log.Debugf("Creating new session for user %s (Telegram ID: %d)",
		user.UserName, user.TelegramID)

	cookie := &http.Cookie{
		Name:    CookieName,
		Value:   uuid.New().String(),
		Expires: time.Now().Add(CookieExpirationDuration),
	}

	ctx.SetCookie(cookie)

	sessionCookie := pgvis.NewCookie(
		ctx.Request().UserAgent(),
		cookie.Value,
		apiKey,
	)

	if err := db.Cookies.Add(sessionCookie); err != nil {
		log.Errorf("Failed to create session: %s", err)
		return false
	}

	return true
}
