package html

import (
	"embed"
	"html/template"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
	"github.com/knackwurstking/pg-vis/internal/cookies"
	"github.com/knackwurstking/pg-vis/pkg/pgvis"
	"github.com/labstack/echo/v4"
)

const (
	CookieName = "pgvis-api-key"
)

var (
	// TODO: Create a special cookie handler, holding the value and the device information,
	//       also store this this in the database table "cookies"
	Cookies = cookies.New()
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

	e.GET(options.ServerPathPrefix+"/login", func(c echo.Context) error {
		return handleLogin(c, options.DB)
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

func handleLogin(c echo.Context, db *pgvis.DB) error {
	v, err := c.FormParams()
	apiKey := v.Get("api-key")

	invalidApiKey := false

	if apiKey != "" {
		log.Debugf("Form: Api Key: %#v", apiKey)

		if u, err := db.Users.GetUserFromApiKey(apiKey); err == nil {
			if u.ApiKey == apiKey {
				log.Debugf("Form: set cookie and redirect")

				cookie := new(http.Cookie)

				cookie.Name = CookieName
				cookie.Value = uuid.New().String()
				c.SetCookie(cookie)

				Cookies.Add(c.Request().UserAgent(), cookie.Value)

				return c.Redirect(http.StatusSeeOther, "/")
			}

			if u.ApiKey != "" {
				invalidApiKey = true
			}
		} else {
			log.Warnf("Get user for api key failed: %s", err.Error())
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

	err = t.Execute(c.Response(), LoginData{
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

type LoginData struct {
	ApiKey        string
	InvalidApiKey bool
}
