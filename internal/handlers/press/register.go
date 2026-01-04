package press

import (
	"net/http"

	"github.com/knackwurstking/pg-press/internal/env"
	ui "github.com/knackwurstking/ui/ui-templ"
	"github.com/labstack/echo/v4"
)

func Register(e *echo.Echo, path string) {
	ui.RegisterEchoRoutes(
		e,
		env.ServerPathPrefix,
		[]*ui.EchoRoute{
			// Press page
			ui.NewEchoRoute(http.MethodGet, path+"/:press", GetPage),

			// HTMX endpoints for press content
			ui.NewEchoRoute(http.MethodGet, path+"/:press/active-tools", GetActiveTools),
			ui.NewEchoRoute(http.MethodGet, path+"/:press/metal-sheets", GetPressMetalSheets),
			ui.NewEchoRoute(http.MethodGet, path+"/:press/cycles", GetCycles),
			ui.NewEchoRoute(http.MethodGet, path+"/:press/notes", GetPressNotes),
			ui.NewEchoRoute(http.MethodGet, path+"/:press/press-regenerations", GetPressRegenerations),

			ui.NewEchoRoute(http.MethodDelete, path+"/:press", DeletePress),

			// PDF Handlers
			//ui.NewEchoRoute(http.MethodGet, path+"/:press/cycle-summary-pdf", GetCycleSummaryPDF),
		},
	)
}
