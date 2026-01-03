package dialogs

import (
	"net/http"

	"github.com/knackwurstking/pg-press/internal/env"
	"github.com/knackwurstking/pg-press/internal/logger"
	ui "github.com/knackwurstking/ui/ui-templ"
	"github.com/labstack/echo/v4"
)

var (
	log = logger.New("handler: dialogs")
)

func Register(e *echo.Echo, path string) {
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
		ui.NewEchoRoute(http.MethodPost, path+"/edit-note", PostNote),
		ui.NewEchoRoute(http.MethodPut, path+"/edit-note", PutNote),

		// Edit cycle dialog
		ui.NewEchoRoute(http.MethodGet, path+"/edit-cycle", GetEditCycle),
		ui.NewEchoRoute(http.MethodPost, path+"/edit-cycle", PostCycle),
		ui.NewEchoRoute(http.MethodPut, path+"/edit-cycle", PutCycle),

		// Edit metal sheet dialog
		ui.NewEchoRoute(http.MethodGet, path+"/edit-metal-sheet", GetEditMetalSheet),
		ui.NewEchoRoute(http.MethodPost, path+"/edit-metal-sheet", PostMetalSheet),
		ui.NewEchoRoute(http.MethodPut, path+"/edit-metal-sheet", PutMetalSheet),

		// New/Edit a Press
		ui.NewEchoRoute(http.MethodGet, path+"/edit-press", GetEditPress),
		ui.NewEchoRoute(http.MethodPost, path+"/edit-press", PostPress),
		ui.NewEchoRoute(http.MethodPut, path+"/edit-press", PutPress),

		// Edit tool regeneration dialog
		ui.NewEchoRoute(http.MethodGet, path+"/edit-tool-regeneration", GetEditToolRegeneration),
		ui.NewEchoRoute(http.MethodPut, path+"/edit-tool-regeneration", PostToolRegeneration),
		ui.NewEchoRoute(http.MethodPut, path+"/edit-tool-regeneration", PutToolRegeneration),

		// Edit press regeneration dialog
		//ui.NewEchoRoute(http.MethodGet, path+"/edit-press-regeneration", h.GetEditPressRegeneration),
		//ui.NewEchoRoute(http.MethodPut, path+"/edit-press-regeneration", h.PutEditPressRegeneration),
	})
}
