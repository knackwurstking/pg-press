package profile

import (
	"net/http"

	"github.com/knackwurstking/pg-press/internal/env"

	"github.com/knackwurstking/ui"
	"github.com/labstack/echo/v4"
)

func Register(e *echo.Echo, path string) {
	ui.RegisterEchoRoutes(
		e,
		env.ServerPathPrefix,
		[]*ui.EchoRoute{
			ui.NewEchoRoute(http.MethodGet, path, GetProfilePage),
			ui.NewEchoRoute(http.MethodPost, path, PostProfilePage),
			ui.NewEchoRoute(http.MethodGet, path+"/cookies", HTMXGetCookies),
			ui.NewEchoRoute(http.MethodDelete, path+"/cookies", HTMXDeleteCookies),
		},
	)
}
