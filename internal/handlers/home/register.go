package home

import (
	"net/http"

	"github.com/knackwurstking/pg-press/internal/env"

	"github.com/knackwurstking/ui/pkg/ui"
	"github.com/labstack/echo/v4"
)

func Register(e *echo.Echo, path string) {
	ui.RegisterEchoRoutes(e, env.ServerPathPrefix, []*ui.EchoRoute{
		ui.NewEchoRoute(http.MethodGet, path, GetHomePage),
	})
}
