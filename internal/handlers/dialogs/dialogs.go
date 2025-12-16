package dialogs

import (
	"net/http"

	"github.com/knackwurstking/pg-press/internal/common"
	"github.com/knackwurstking/pg-press/internal/env"
	"github.com/knackwurstking/pg-press/internal/logger"

	ui "github.com/knackwurstking/ui/ui-templ"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	DB  *common.DB
	Log *ui.Logger
}

func NewHandler(db *common.DB) *Handler {
	return &Handler{
		DB:  db,
		Log: logger.New("handler: dialogs"),
	}
}

func (h *Handler) RegisterRoutes(e *echo.Echo, path string) {
	ui.RegisterEchoRoutes(e, env.ServerPathPrefix, []*ui.EchoRoute{
		// Edit cycle dialog
		//ui.NewEchoRoute(http.MethodGet, path+"/edit-cycle", h.GetEditCycle),
		//ui.NewEchoRoute(http.MethodPost, path+"/edit-cycle", h.PostEditCycle),
		//ui.NewEchoRoute(http.MethodPut, path+"/edit-cycle", h.PutEditCycle),

		// Edit tool dialog
		ui.NewEchoRoute(http.MethodGet, path+"/edit-tool", h.GetToolDialog),
		ui.NewEchoRoute(http.MethodPost, path+"/edit-tool", h.PostTool),
		ui.NewEchoRoute(http.MethodPut, path+"/edit-tool", h.PutTool),

		// TODO: Add a cassette dialog handler for edit/new, just like tools but for cassettes
		ui.NewEchoRoute(http.MethodGet, path+"/edit-cassette", h.GetCassetteDialog),
		ui.NewEchoRoute(http.MethodPost, path+"/edit-cassette", h.PostCassette),
		ui.NewEchoRoute(http.MethodPut, path+"/edit-cassette", h.PutCassette),

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
