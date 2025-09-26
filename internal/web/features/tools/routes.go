package tools

import (
	"net/http"

	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/web/shared/helpers"

	"github.com/labstack/echo/v4"
)

type Routes struct {
	handler *Handler
}

func NewRoutes(db *database.DB) *Routes {
	return &Routes{
		handler: NewHandler(db),
	}
}

func (r *Routes) RegisterRoutes(e *echo.Echo) {
	helpers.RegisterEchoRoutes(
		e,
		[]*helpers.EchoRoute{
			// Pages
			helpers.NewEchoRoute(http.MethodGet, "/tools",
				r.handler.GetToolsPage),

			helpers.NewEchoRoute(http.MethodGet, "/tools/press/:press",
				r.handler.GetPressPage),

			helpers.NewEchoRoute(http.MethodGet, "/tools/press/:press/umbau",
				r.handler.GetUmbauPage),
			helpers.NewEchoRoute(http.MethodPost, "/tools/press/:press/umbau",
				r.handler.PostUmbauPage),

			helpers.NewEchoRoute(http.MethodGet, "/tools/tool/:id",
				r.handler.GetToolPage),

			// HTMX
			// List tools
			helpers.NewEchoRoute(http.MethodGet, "/htmx/tools/list", r.handler.HTMXGetToolsList),

			// Get, Post or Edit a tool
			helpers.NewEchoRoute(http.MethodGet, "/htmx/tools/edit",
				r.handler.HTMXGetEditToolDialog),
			helpers.NewEchoRoute(http.MethodPost, "/htmx/tools/edit",
				r.handler.HTMXPostEditToolDialog),
			helpers.NewEchoRoute(http.MethodPut, "/htmx/tools/edit",
				r.handler.HTMXPutEditToolDialog),

			// Delete a tool
			helpers.NewEchoRoute(http.MethodDelete, "/htmx/tools/delete",
				r.handler.HTMXDeleteTool),

			// Tool status management
			helpers.NewEchoRoute(http.MethodGet, "/htmx/tools/status-edit",
				r.handler.HTMXGetStatusEdit),
			helpers.NewEchoRoute(http.MethodGet, "/htmx/tools/status-display",
				r.handler.HTMXGetStatusDisplay),
			helpers.NewEchoRoute(http.MethodPut, "/htmx/tools/status",
				r.handler.HTMXUpdateToolStatus),

			// Press page sections
			helpers.NewEchoRoute(http.MethodGet, "/htmx/tools/press/active-tools",
				r.handler.HTMXGetPressActiveTools),
			helpers.NewEchoRoute(http.MethodGet, "/htmx/tools/press/metal-sheets",
				r.handler.HTMXGetPressMetalSheets),
			helpers.NewEchoRoute(http.MethodGet, "/htmx/tools/press/cycles",
				r.handler.HTMXGetPressCycles),

			// Cycles table rows
			helpers.NewEchoRoute(http.MethodGet, "/htmx/tools/cycles",
				r.handler.HTMXGetToolCycles),

			helpers.NewEchoRoute(http.MethodGet, "/htmx/tools/total-cycles",
				r.handler.HTMXGetToolTotalCycles),

			// Get, add or edit a cycles table entry
			helpers.NewEchoRoute(http.MethodGet, "/htmx/tools/cycle/edit",
				r.handler.HTMXGetToolCycleEditDialog),

			helpers.NewEchoRoute(http.MethodPost, "/htmx/tools/cycle/edit",
				r.handler.HTMXPostToolCycleEditDialog),

			helpers.NewEchoRoute(http.MethodPut, "/htmx/tools/cycle/edit",
				r.handler.HTMXPutToolCycleEditDialog),

			// Delete a cycle table entry
			helpers.NewEchoRoute(http.MethodDelete, "/htmx/tools/cycle/delete",
				r.handler.HTMXDeleteToolCycle),
		},
	)
}
