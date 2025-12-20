package auth

import (
	"net/http"

	"github.com/knackwurstking/pg-press/internal/env"
	"github.com/knackwurstking/pg-press/internal/logger"

	ui "github.com/knackwurstking/ui/ui-templ"

	"github.com/labstack/echo/v4"
)

const (
	CookieName = "pgpress-api-key"
)

var (
	log = logger.New("handler: auth")
)

func Register(e *echo.Echo, path string) {
	ui.RegisterEchoRoutes(e, env.ServerPathPrefix, []*ui.EchoRoute{
		ui.NewEchoRoute(http.MethodGet, path+"/login", GetLoginPage),
		ui.NewEchoRoute(http.MethodPost, path+"/login", PostLoginPage),
		ui.NewEchoRoute(http.MethodGet, path+"/logout", GetLogout),
	})
}
