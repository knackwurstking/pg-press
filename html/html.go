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
		return handleLogin(c, options.ServerPathPrefix, options.DB)
	})

	// TODO: /logout route, just remove the cookie and redirect to the /login page

	e.GET(options.ServerPathPrefix+"/profile", func(c echo.Context) error {
		return handleProfile(c, options.DB)
	})

	e.GET(options.ServerPathPrefix+"/trouble-reports", func(c echo.Context) error {
		return handleTroubleReports(c)
	})
}

func handleHomePage(c echo.Context) error {
	pageData := PageData{}

	t, err := template.ParseFS(routes,
		pageData.TemplatePatterns(
			"routes/layout.html",
			"routes/home.html",
		)...,
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

func handleLogin(ctx echo.Context, serverPathPrefix string, db *pgvis.DB) error {
	v, err := ctx.FormParams()
	apiKey := v.Get("api-key")

	if ok, err := handleLoginApiKey(apiKey, db, ctx); ok {
		return ctx.Redirect(http.StatusSeeOther, serverPathPrefix+"/profile")
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
		pageData.TemplatePatterns(
			"routes/layout.html",
			"routes/login.html",
		)...,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	err = t.Execute(ctx.Response(), LoginPageData{
		ApiKey:        apiKey,
		InvalidApiKey: apiKey != "",
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return nil
}

func handleFeed(c echo.Context) error {
	pageData := PageData{}

	t, err := template.ParseFS(routes,
		pageData.TemplatePatterns(
			"routes/layout.html",
			"routes/feed.html",
		)...,
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

func handleProfile(ctx echo.Context, db *pgvis.DB) error {
	var pageData ProfilePageData

	user, err := getUserFromContext(ctx)
	if err != nil {
		return err
	}
	pageData.User = user

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

	t, err := template.ParseFS(routes,
		pageData.TemplatePatterns(
			"routes/layout.html",
			"routes/profile.html",
		)...,
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

func handleTroubleReports(ctx echo.Context) error {
	pageData := PageData{}

	t, err := template.ParseFS(routes,
		pageData.TemplatePatterns(
			"routes/layout.html",
			"routes/trouble-reports.html",
		)...,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err = t.Execute(ctx.Response(), pageData); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return nil
}
