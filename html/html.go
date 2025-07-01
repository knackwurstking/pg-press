package html

import (
	"embed"
	"html/template"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/knackwurstking/pg-vis/pkg/pgvis"
	"github.com/labstack/echo/v4"
)

const (
	CookieName = "pgvis-api-key"
)

//go:embed routes
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

	e.GET(options.ServerPathPrefix+"/signup", func(c echo.Context) error {
		return handleSignUp(c, options.DB)
	})

	e.GET(options.ServerPathPrefix+"/feed", func(c echo.Context) error {
		return handleFeed(c)
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

func handleSignUp(c echo.Context, db *pgvis.DB) error {
	v, err := c.FormParams()
	apiKey := v.Get("api-key")

	invalidApiKey := false

	if apiKey != "" {
		log.Debugf("Form: Api Key: %#v", apiKey)

		u, err := db.Users.GetUserFromApiKey(apiKey)
		if err == nil {
			if u.ApiKey == apiKey {
				cookie := new(http.Cookie)

				cookie.Name = CookieName
				cookie.Value = u.ApiKey
				c.SetCookie(cookie)

				c.Redirect(http.StatusSeeOther, "/")
			}

			invalidApiKey = true
		}
	}

	t, err := template.ParseFS(routes,
		"routes/layout.html",
		"routes/signup/page.html",
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	err = t.Execute(c.Response(), SignUpData{
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

type SignUpData struct {
	ApiKey        string
	InvalidApiKey bool
}
