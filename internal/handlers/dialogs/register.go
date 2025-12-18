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
	Log = logger.New("handler: dialogs")
	DB  *common.DB
)

func Register(db *common.DB, e *echo.Echo, path string) {
	DB = db

	ui.RegisterEchoRoutes(e, env.ServerPathPrefix, []*ui.EchoRoute{
		// Edit cycle dialog
		//ui.NewEchoRoute(http.MethodGet, path+"/edit-cycle", h.GetEditCycle),
		//ui.NewEchoRoute(http.MethodPost, path+"/edit-cycle", h.PostEditCycle),
		//ui.NewEchoRoute(http.MethodPut, path+"/edit-cycle", h.PutEditCycle),

		// Edit tool dialog
		ui.NewEchoRoute(http.MethodGet, path+"/edit-tool", GetToolDialog),
		ui.NewEchoRoute(http.MethodPost, path+"/edit-tool", PostTool),
		ui.NewEchoRoute(http.MethodPut, path+"/edit-tool", PutTool),

		// Edit cassette dialog
		ui.NewEchoRoute(http.MethodGet, path+"/edit-cassette", GetCassetteDialog),
		ui.NewEchoRoute(http.MethodPost, path+"/edit-cassette", PostCassette),
		ui.NewEchoRoute(http.MethodPut, path+"/edit-cassette", PutCassette),

		// Edit metal sheet dialog
		//ui.NewEchoRoute(http.MethodGet, path+"/edit-metal-sheet", h.GetEditMetalSheet),
		//ui.NewEchoRoute(http.MethodPost, path+"/edit-metal-sheet", h.PostEditMetalSheet),
		//ui.NewEchoRoute(http.MethodPut, path+"/edit-metal-sheet", h.PutEditMetalSheet),

		// Edit note dialog
		//ui.NewEchoRoute(http.MethodGet, path+"/edit-note", h.GetEditNote),
		//ui.NewEchoRoute(http.MethodPost, path+"/edit-note", h.PostEditNote),
		//ui.NewEchoRoute(http.MethodPut, path+"/edit-note", h.PutEditNote),

		// Edit tool regeneration dialog
		//ui.NewEchoRoute(http.MethodGet, path+"/edit-tool-regeneration", h.GetEditToolRegeneration),
		//ui.NewEchoRoute(http.MethodPut, path+"/edit-tool-regeneration", h.PutEditToolRegeneration),

		// Edit press regeneration dialog
		//ui.NewEchoRoute(http.MethodGet, path+"/edit-press-regeneration", h.GetEditPressRegeneration),
		//ui.NewEchoRoute(http.MethodPut, path+"/edit-press-regeneration", h.PutEditPressRegeneration),
	})
}
