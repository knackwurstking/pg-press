package tool

import (
	"net/http"

	"github.com/knackwurstking/pg-press/internal/env"
	"github.com/knackwurstking/pg-press/internal/logger"

	"github.com/knackwurstking/ui"
	"github.com/labstack/echo/v4"
)

var (
	log = logger.New("handler: tool")
)

func Register(e *echo.Echo, path string) {
	ui.RegisterEchoRoutes(e, env.ServerPathPrefix, []*ui.EchoRoute{
		// Main Page
		ui.NewEchoRoute(http.MethodGet, path+"/:id", GetToolPage), // "is_cassette" defines the tool type

		// Regenerations Table
		ui.NewEchoRoute(http.MethodDelete, path+"/:id/delete-regeneration", DeleteRegeneration), // "id" is regeneration ID

		// Tool status and regenerations management
		ui.NewEchoRoute(http.MethodGet, path+"/:id/regeneration-edit", RegenerationEditable),
		ui.NewEchoRoute(http.MethodGet, path+"/:id/regeneration-display", RegenerationNonEditable),
		ui.NewEchoRoute(http.MethodPut, path+"/:id/regeneration", Regeneration),

		// Section loading
		ui.NewEchoRoute(http.MethodGet, path+"/:id/notes", GetToolNotes),
		ui.NewEchoRoute(http.MethodGet, path+"/:id/metal-sheets", GetToolMetalSheets),

		// Cycles table rows
		ui.NewEchoRoute(http.MethodGet, path+"/:id/cycles", GetCyclesSectionContent),
		ui.NewEchoRoute(http.MethodGet, path+"/:id/total-cycles", GetToolTotalCycles),

		// Update tools binding data
		ui.NewEchoRoute(http.MethodPatch, path+"/:id/bind", ToolBinding),
		ui.NewEchoRoute(http.MethodPatch, path+"/:id/unbind", ToolUnBinding),

		// Delete a cycle table entry
		ui.NewEchoRoute(http.MethodDelete, path+"/cycle/delete", DeleteToolCycle),
	})
}
