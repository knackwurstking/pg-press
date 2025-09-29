package tool

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
			// HTML
			helpers.NewEchoRoute(http.MethodGet, "/tools/tool/:id",
				r.handler.GetToolPage),

			// HTMX
			// Tool status management
			helpers.NewEchoRoute(http.MethodGet, "/htmx/tools/status-edit",
				r.handler.HTMXGetStatusEdit),
			helpers.NewEchoRoute(http.MethodGet, "/htmx/tools/status-display",
				r.handler.HTMXGetStatusDisplay),
			helpers.NewEchoRoute(http.MethodPut, "/htmx/tools/status",
				r.handler.HTMXUpdateToolStatus),

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
