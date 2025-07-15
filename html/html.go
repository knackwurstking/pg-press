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

	"github.com/knackwurstking/pg-vis/html/handler"
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

func Serve(e *echo.Echo, options Options) {
	e.StaticFS(options.ServerPathPrefix+"/", echo.MustSubFS(static, "static"))

	serveHome(e, options)
	serveFeed(e, options)
	serveLogin(e, options)
	serveLogout(e, options)
	serveProfile(e, options)
	serveTroubleReports(e, options)
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

type ProfilePageData struct {
	User    *pgvis.User
	Cookies []*pgvis.Cookie
}

func NewProfilePageData() ProfilePageData {
	return ProfilePageData{
		Cookies: make([]*pgvis.Cookie, 0),
	}
}

func (p ProfilePageData) CookiesSorted() []*pgvis.Cookie {
	return pgvis.SortCookies(p.Cookies)
}

func serveProfile(e *echo.Echo, options Options) {
	e.GET(options.ServerPathPrefix+"/profile", func(c echo.Context) error {
		pageData := NewProfilePageData()

		if user, err := handler.GetUserFromContext(c); err != nil {
			return err
		} else {
			pageData.User = user
		}

		if err := handleUserNameChange(c, &pageData, options.DB); err != nil {
			return echo.NewHTTPError(
				http.StatusInternalServerError,
				fmt.Errorf("change username: %s", err.Error()),
			)
		}

		if cookies, err := options.DB.Cookies.ListApiKey(pageData.User.ApiKey); err != nil {
			log.Error("/profile -> List cookies for Api Key failed: %s", err.Error())
		} else {
			pageData.Cookies = cookies
		}

		t, err := template.ParseFS(templates,
			"templates/layout.html",
			"templates/profile.html",
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

	e.GET(options.ServerPathPrefix+"/profile/cookies", func(c echo.Context) error {
		return profile.GETCookies(templates, c, options.DB)
	})

	e.DELETE(options.ServerPathPrefix+"/profile/cookies", func(c echo.Context) error {
		return profile.DELETECookies(templates, c, options.DB)
	})
}

func serveTroubleReports(e *echo.Echo, options Options) {
	e.GET(options.ServerPathPrefix+"/trouble-reports", func(c echo.Context) error {
		t, err := template.ParseFS(templates,
			"templates/layout.html",
			"templates/trouble-reports.html",
		)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		if err = t.Execute(c.Response(), nil); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		return nil
	})

	// HTMX: Dialog Edit

	e.GET(options.ServerPathPrefix+"/trouble-reports/dialog-edit", func(c echo.Context) error {
		return troublereports.GETDialogEdit(templates, c, options.DB, nil)
	})

	// FormValues:
	//   - title: string
	//   - content: multiline-string
	e.POST(options.ServerPathPrefix+"/trouble-reports/dialog-edit", func(c echo.Context) error {
		return troublereports.POSTDialogEdit(templates, c, options.DB)
	})

	// QueryParam:
	//   - id: int
	//
	// FormValue:
	//   - title: string
	//   - content: multiline-string
	e.PUT(options.ServerPathPrefix+"/trouble-reports/dialog-edit", func(c echo.Context) error {
		return troublereports.PUTDialogEdit(templates, c, options.DB)
	})

	// HTMX: Data

	e.GET(options.ServerPathPrefix+"/trouble-reports/data", func(c echo.Context) error {
		return troublereports.GETData(templates, c, options.DB)
	})

	// QueryParam:
	//   - id: int
	e.DELETE(options.ServerPathPrefix+"/trouble-reports/data", func(c echo.Context) error {
		return troublereports.DELETEData(templates, c, options.DB)
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

func handleUserNameChange(ctx echo.Context, pageData *ProfilePageData, db *pgvis.DB) error {
	v, err := ctx.FormParams()
	userName := v.Get("user-name")

	// Database update
	if userName != "" && userName != pageData.User.UserName {
		log.Debugf(
			"Change user name in database: %s => %s",
			pageData.User.UserName, userName,
		)

		user := pgvis.NewUser(pageData.User.TelegramID, userName, pageData.User.ApiKey)
		if err = db.Users.Update(pageData.User.TelegramID, user); err != nil {
			return err
		} else {
			pageData.User.UserName = userName
		}
	}

	return nil
}
