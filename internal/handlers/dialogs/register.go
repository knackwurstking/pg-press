package dialogs

import (
	"net/http"

	"github.com/knackwurstking/pg-press/internal/common"
	"github.com/knackwurstking/pg-press/internal/env"
	"github.com/knackwurstking/pg-press/internal/logger"
	ui "github.com/knackwurstking/ui/ui-templ"
	"github.com/labstack/echo/v4"
)

var (
	log = logger.New("handler: dialogs")
	db  *common.DB
)

func Register(cdb *common.DB, e *echo.Echo, path string) {
	db = cdb

	ui.RegisterEchoRoutes(e, env.ServerPathPrefix, []*ui.EchoRoute{
		// Edit tool dialog
		ui.NewEchoRoute(http.MethodGet, path+"/edit-tool", GetToolDialog),
		ui.NewEchoRoute(http.MethodPost, path+"/edit-tool", PostTool),
		ui.NewEchoRoute(http.MethodPut, path+"/edit-tool", PutTool),

		// Edit cassette dialog
		ui.NewEchoRoute(http.MethodGet, path+"/edit-cassette", GetCassetteDialog),
		ui.NewEchoRoute(http.MethodPost, path+"/edit-cassette", PostCassette),
		ui.NewEchoRoute(http.MethodPut, path+"/edit-cassette", PutCassette),

		// Edit note dialog
		ui.NewEchoRoute(http.MethodGet, path+"/edit-note", GetEditNote),
		ui.NewEchoRoute(http.MethodPost, path+"/edit-note", PostEditNote),
		ui.NewEchoRoute(http.MethodPut, path+"/edit-note", PutEditNote),

		// Edit cycle dialog
		//ui.NewEchoRoute(http.MethodGet, path+"/edit-cycle", h.GetEditCycle),
		//ui.NewEchoRoute(http.MethodPost, path+"/edit-cycle", h.PostEditCycle),
		//ui.NewEchoRoute(http.MethodPut, path+"/edit-cycle", h.PutEditCycle),

		// Edit metal sheet dialog
		//ui.NewEchoRoute(http.MethodGet, path+"/edit-metal-sheet", h.GetEditMetalSheet),
		//ui.NewEchoRoute(http.MethodPost, path+"/edit-metal-sheet", h.PostEditMetalSheet),
		//ui.NewEchoRoute(http.MethodPut, path+"/edit-metal-sheet", h.PutEditMetalSheet),		// Edit tool regeneration dialog
		//ui.NewEchoRoute(http.MethodGet, path+"/edit-tool-regeneration", h.GetEditToolRegeneration),
		//ui.NewEchoRoute(http.MethodPut, path+"/edit-tool-regeneration", h.PutEditToolRegeneration),

		// Edit press regeneration dialog
		//ui.NewEchoRoute(http.MethodGet, path+"/edit-press-regeneration", h.GetEditPressRegeneration),
		//ui.NewEchoRoute(http.MethodPut, path+"/edit-press-regeneration", h.PutEditPressRegeneration),
	})
}
