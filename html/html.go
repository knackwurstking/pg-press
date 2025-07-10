package html

import (
	"embed"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/pgvis"
)

const (
	CookieName = "pgvis-api-key"
)

//go:embed routes
var routes embed.FS

//go:embed static
var static embed.FS

type PageData struct {
	ErrorMessages []string
}

func NewPageData() PageData {
	return PageData{
		ErrorMessages: make([]string, 0),
	}
}

type Options struct {
	ServerPathPrefix string
	DB               *pgvis.DB
}

// TODO: Clean up this mess
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

	e.GET(options.ServerPathPrefix+"/profile/cookies", func(c echo.Context) error {
		return handleProfileCookiesGET(c, options.DB)
	})

	e.DELETE(options.ServerPathPrefix+"/profile/cookies", func(c echo.Context) error {
		return handleProfileCookiesDELETE(c, options.DB)
	})

	e.GET(options.ServerPathPrefix+"/trouble-reports", func(c echo.Context) error {
		return handleTroubleReports(c)
	})
}
