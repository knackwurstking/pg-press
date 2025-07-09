package html

import (
	"embed"
	"html/template"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/pkg/pgvis"
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
		return handleLogin(c, options.DB)
	})

	e.GET(options.ServerPathPrefix+"/logout", func(c echo.Context) error {
		return handleLogout(c, options.DB)
	})

	e.GET(options.ServerPathPrefix+"/profile", func(c echo.Context) error {
		return handleProfile(c, options.DB)
	})

	e.GET(options.ServerPathPrefix+"/trouble-reports", func(c echo.Context) error {
		return handleTroubleReports(c)
	})
}

func handleHomePage(c echo.Context) *echo.HTTPError {
	pageData := PageData{}

	t, err := template.ParseFS(routes,
		"routes/layout.html",
		"routes/home.html",
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	err = t.Execute(c.Response(), pageData)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return nil
}

func handleLogin(ctx echo.Context, db *pgvis.DB) *echo.HTTPError {
	v, err := ctx.FormParams()
	apiKey := v.Get("api-key")

	if ok, err := handleLoginApiKey(apiKey, db, ctx); ok {
		if err = ctx.Redirect(http.StatusSeeOther, "./profile"); err != nil {
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

	t, err := template.ParseFS(routes,
		"routes/layout.html",
		"routes/login.html",
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	err = t.Execute(ctx.Response(), pageData)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return nil
}

func handleLogout(ctx echo.Context, db *pgvis.DB) *echo.HTTPError {
	if cookie, err := ctx.Cookie(CookieName); err == nil {
		db.Cookies.Remove(cookie.Value)

		err = ctx.Redirect(http.StatusSeeOther, "./login")
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}

	return nil
}

func handleFeed(c echo.Context) *echo.HTTPError {
	pageData := PageData{}

	t, err := template.ParseFS(routes,
		"routes/layout.html",
		"routes/feed.html",
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	err = t.Execute(c.Response(), pageData)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return nil
}

func handleProfile(ctx echo.Context, db *pgvis.DB) *echo.HTTPError {
	var pageData ProfilePageData

	if user, err := getUserFromContext(ctx); err != nil {
		return err
	} else {
		pageData.User = user
	}

	// Get "user-name" from form data (optional), and update database user
	v, err := ctx.FormParams()
	userName := v.Get("user-name")

	// Database update
	if userName != "" && userName != pageData.User.UserName {
		log.Debugf(
			"/profile -> Change user name in database: %s => %s",
			pageData.User.UserName, userName,
		)

		pageData.User.UserName = userName
		if err = db.Users.Update(pageData.User.TelegramID, pageData.User); err != nil {
			pageData.ErrorMessages = []string{err.Error()}
		}
	}

	// TODO: Add cookies to `pageData`, check for errors and append to `pageData.ErrorMessages`

	t, err := template.ParseFS(routes,
		"routes/layout.html",
		"routes/profile.html",
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	err = t.Execute(ctx.Response(), pageData)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return nil
}

func handleTroubleReports(ctx echo.Context) *echo.HTTPError {
	pageData := PageData{}

	t, err := template.ParseFS(routes,
		"routes/layout.html",
		"routes/trouble-reports.html",
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err = t.Execute(ctx.Response(), pageData); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return nil
}
