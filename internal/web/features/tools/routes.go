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
			// TODO: Move to tool/
			helpers.NewEchoRoute(http.MethodGet, "/htmx/tools/status-edit",
				r.handler.HTMXGetStatusEdit),
			// TODO: Move to tool/
			helpers.NewEchoRoute(http.MethodGet, "/htmx/tools/status-display",
				r.handler.HTMXGetStatusDisplay),
			// TODO: Move to tool/
			helpers.NewEchoRoute(http.MethodPut, "/htmx/tools/status",
				r.handler.HTMXUpdateToolStatus),

			// Cycles table rows
			// TODO: Move to tool/
			helpers.NewEchoRoute(http.MethodGet, "/htmx/tools/cycles",
				r.handler.HTMXGetToolCycles),

			// TODO: Move to tool/
			helpers.NewEchoRoute(http.MethodGet, "/htmx/tools/total-cycles",
				r.handler.HTMXGetToolTotalCycles),

			// Get, add or edit a cycles table entry
			// TODO: Move to tool/
			helpers.NewEchoRoute(http.MethodGet, "/htmx/tools/cycle/edit",
				r.handler.HTMXGetToolCycleEditDialog),

			// TODO: Move to tool/
			helpers.NewEchoRoute(http.MethodPost, "/htmx/tools/cycle/edit",
				r.handler.HTMXPostToolCycleEditDialog),

			// TODO: Move to tool/
			helpers.NewEchoRoute(http.MethodPut, "/htmx/tools/cycle/edit",
				r.handler.HTMXPutToolCycleEditDialog),

			// Delete a cycle table entry
			// TODO: Move to tool/
			helpers.NewEchoRoute(http.MethodDelete, "/htmx/tools/cycle/delete",
				r.handler.HTMXDeleteToolCycle),
		},
	)
}
