package auth

import (
	"net/http"

	"github.com/knackwurstking/pg-press/internal/common"
	"github.com/knackwurstking/pg-press/internal/env"
	"github.com/knackwurstking/pg-press/internal/logger"

	ui "github.com/knackwurstking/ui/ui-templ"

	"github.com/labstack/echo/v4"
)

const (
	CookieName = "pgpress-api-key"
)

var (
	Log = logger.New("handler: auth")
	DB  *common.DB
)

func Register(db *common.DB, e *echo.Echo, path string) {
	DB = db

	ui.RegisterEchoRoutes(e, env.ServerPathPrefix, []*ui.EchoRoute{
		ui.NewEchoRoute(http.MethodGet, path+"/login", GetLoginPage),
		ui.NewEchoRoute(http.MethodPost, path+"/login", PostLoginPage),
		ui.NewEchoRoute(http.MethodGet, path+"/logout", GetLogout),
	})
}
