package tools

import (
	"net/http"

	"github.com/knackwurstking/pg-press/internal/env"
	"github.com/knackwurstking/pg-press/internal/logger"

	"github.com/knackwurstking/ui/templ/ui"
	"github.com/labstack/echo/v4"
)

var (
	log = logger.New("handler: tools")
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
