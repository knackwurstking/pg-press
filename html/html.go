package html

import (
	"embed"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/pgvis"
)

const (
	CookieName = "pgvis-api-key"
)

//go:embed templates
var templates embed.FS

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

func Serve(e *echo.Echo, options Options) {
	e.StaticFS(options.ServerPathPrefix+"/", echo.MustSubFS(static, "static"))

	ServeHome(e, options)
	ServeFeed(e, options)
	ServeLogin(e, options)
	ServeLogout(e, options)
	ServeProfile(e, options)
	ServeTroubleReports(e, options)
}
