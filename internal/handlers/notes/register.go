package notes

import (
	"net/http"

	"github.com/knackwurstking/pg-press/internal/common"
	"github.com/knackwurstking/pg-press/internal/env"
	"github.com/knackwurstking/pg-press/internal/logger"
	ui "github.com/knackwurstking/ui/ui-templ"
	"github.com/labstack/echo/v4"
)

var (
	db  *common.DB
	log = logger.New("handler: notes")
)

func Register(cdb *common.DB, e *echo.Echo, path string) {
	db = cdb

	ui.RegisterEchoRoutes(e, env.ServerPathPrefix, []*ui.EchoRoute{
		// Notes page
		ui.NewEchoRoute(http.MethodGet, path, GetPage),

		// HTMX routes for notes deletion
		ui.NewEchoRoute(http.MethodDelete, path+"/delete", DeleteNote),

		// Render Notes Grid
		ui.NewEchoRoute(http.MethodGet, path+"/grid", GetNotesGrid),
	})
}
