package profile

import (
	"net/http"

	"github.com/knackwurstking/pg-press/internal/common"
	"github.com/knackwurstking/pg-press/internal/env"
	"github.com/knackwurstking/pg-press/internal/logger"
	ui "github.com/knackwurstking/ui/ui-templ"
	"github.com/labstack/echo/v4"
)

var (
	Log = logger.New("handler: profile")
	DB  *common.DB
)

func Register(db *common.DB, e *echo.Echo, path string) {
	DB = db

	ui.RegisterEchoRoutes(
		e,
		env.ServerPathPrefix,
		[]*ui.EchoRoute{
			ui.NewEchoRoute(http.MethodGet, path, GetProfilePage),
			ui.NewEchoRoute(http.MethodGet, path+"/cookies", HTMXGetCookies),
			ui.NewEchoRoute(http.MethodDelete, path+"/cookies", HTMXDeleteCookies),
		},
	)
}
