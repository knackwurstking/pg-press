package html

import (
	"embed"
	"errors"
	"html/template"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
	"github.com/knackwurstking/pg-vis/pkg/pgvis"
	"github.com/labstack/echo/v4"
)

const (
	CookieName = "pgvis-api-key"
)

//go:embed routes
//go:embed svg
var routes embed.FS

//go:embed static
var static embed.FS

type Options struct {
	ServerPathPrefix string
	DB               *pgvis.DB
}

func Serve(e *echo.Echo, options Options) {
	e.StaticFS(options.ServerPathPrefix+"/", echo.MustSubFS(static, "static"))

	e.GET(options.ServerPathPrefix+"/", func(c echo.Context) error {
		return handleHomePage(c)
	})

	e.GET(options.ServerPathPrefix+"/feed", func(c echo.Context) error {
		return handleFeed(c)
	})

	e.GET(options.ServerPathPrefix+"/login", func(c echo.Context) error {
		return handleLogin(c, options.ServerPathPrefix, options.DB)
	})

	e.GET(options.ServerPathPrefix+"/profile", func(c echo.Context) error {
		return handleProfile(c, options.DB)
	})
}

func handleHomePage(c echo.Context) error {
	t, err := template.ParseFS(routes,
		"routes/layout.html",
		"routes/page.html",
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	err = t.Execute(c.Response(), nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return nil
}

func handleLogin(ctx echo.Context, serverPathPrefix string, db *pgvis.DB) error {
	v, err := ctx.FormParams()
	apiKey := v.Get("api-key")

	invalidApiKey := false

	if apiKey != "" {
		if u, err := db.Users.GetUserFromApiKey(apiKey); err == nil {
			// Set cookie and redirect to "/"
			if u.ApiKey == apiKey {
				log.Debugf(
					"/login -> Set cookie and redirect to /profile (id: %#v; user: %#v)",
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
				})

				return ctx.Redirect(http.StatusSeeOther, serverPathPrefix+"/profile")
			}

			if u.ApiKey != "" {
				invalidApiKey = true
			}
		} else {
			log.Warnf("/login -> User not found, invalid api key: %s", err.Error())
			invalidApiKey = true
		}
	}

	t, err := template.ParseFS(routes,
		"routes/layout.html",
		"routes/login/page.html",
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	err = t.Execute(ctx.Response(), LoginPageData{
		ApiKey:        apiKey,
		InvalidApiKey: invalidApiKey,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return nil
}

func handleFeed(c echo.Context) error {
	t, err := template.ParseFS(routes,
		"routes/layout.html",
		"routes/feed/page.html",
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	err = t.Execute(c.Response(), nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return nil
}

func handleProfile(ctx echo.Context, db *pgvis.DB) error {
	var data ProfilePageData

	{ // Get user from the echos context
		user, ok := ctx.Get("user").(*pgvis.User)
		if !ok {
			return echo.NewHTTPError(http.StatusInternalServerError,
				errors.New("this should never happen, user is missing in context"))
		}
		data.User = user
	}

	{ // Get "user-name" from form data (optional), and update database user
		v, err := ctx.FormParams()
		if userName := v.Get("user-name"); userName != "" && userName != data.User.UserName {
			log.Debugf(
				"/profile -> Change user name in database: %s => %s",
				data.User.UserName, userName,
			)

			data.User.UserName = userName
			if err = db.Users.Update(data.User.TelegramID, data.User); err != nil {
				data.ErrorMessages = []string{err.Error()}
			} else {
				data.User.UserName = userName
			}
		}
	}

	t, err := template.ParseFS(routes,
		"routes/layout.html",
		"routes/profile/page.html",
		"svg/pencil.html",
		"svg/triangle-alert.html",
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	err = t.Execute(ctx.Response(), data)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return nil
}

type LoginPageData struct {
	ApiKey        string
	InvalidApiKey bool
}

type ProfilePageData struct {
	User          *pgvis.User
	ErrorMessages []string
}
