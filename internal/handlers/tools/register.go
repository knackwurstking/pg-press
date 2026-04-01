package tools

import (
	"net/http"

	"github.com/knackwurstking/pg-press/internal/env"

	"github.com/knackwurstking/ui"
	"github.com/labstack/echo/v4"
)

func Register(e *echo.Echo, path string) {
	ui.RegisterEchoRoutes(e, env.ServerPathPrefix, []*ui.EchoRoute{
		ui.NewEchoRoute(http.MethodGet, path, GetToolsPage),
		ui.NewEchoRoute(http.MethodDelete, path+"/delete", Delete),
		ui.NewEchoRoute(http.MethodPatch, path+"/mark-dead", MarkAsDead),
		ui.NewEchoRoute(http.MethodGet, path+"/section/press", PressSection),
		ui.NewEchoRoute(http.MethodGet, path+"/section/tools", ToolsSection),
	})
}
