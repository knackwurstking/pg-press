package tool

import (
	"net/http"

	"github.com/knackwurstking/pgpress/internal/services"
	"github.com/knackwurstking/pgpress/internal/web/shared/helpers"

	"github.com/labstack/echo/v4"
)

type Routes struct {
	handler *Handler
}

func NewRoutes(db *services.Registry) *Routes {
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

			// Regnerations Table
			helpers.NewEchoRoute(
				http.MethodGet,
				"/htmx/tools/tool/:id/edit-regeneration",
				r.handler.HTMXGetEditRegeneration,
			),
			helpers.NewEchoRoute(
				http.MethodPut,
				"/htmx/tools/tool/:id/edit-regeneration",
				r.handler.HTMXPutEditRegeneration,
			),

			// TODO: Add route for "/htmx/tools/tool/:id/delete-regeneration"
			helpers.NewEchoRoute(
				http.MethodDelete,
				"/htmx/tools/tool/:id/delete-regeneration",
				r.handler.HTMXDeleteRegeneration,
			),

			// Tool status and regenerations management
			helpers.NewEchoRoute(
				http.MethodGet,
				"/htmx/tools/status-edit",
				r.handler.HTMXGetStatusEdit,
			),

			helpers.NewEchoRoute(
				http.MethodGet,
				"/htmx/tools/status-display",
				r.handler.HTMXGetStatusDisplay,
			),

			helpers.NewEchoRoute(
				http.MethodPut,
				"/htmx/tools/status",
				r.handler.HTMXUpdateToolStatus,
			),

			// Section loading
			helpers.NewEchoRoute(
				http.MethodGet,
				"/htmx/tools/notes",
				r.handler.HTMXGetToolNotes,
			),
			helpers.NewEchoRoute(
				http.MethodGet,
				"/htmx/tools/metal-sheets",
				r.handler.HTMXGetToolMetalSheets,
			),

			// Cycles table rows
			helpers.NewEchoRoute(
				http.MethodGet,
				"/htmx/tools/cycles",
				r.handler.HTMXGetCycles,
			),

			helpers.NewEchoRoute(
				http.MethodGet,
				"/htmx/tools/total-cycles",
				r.handler.HTMXGetToolTotalCycles,
			),

			// Get, add or edit a cycles table entry
			helpers.NewEchoRoute(
				http.MethodGet,
				"/htmx/tools/cycle/edit",
				r.handler.HTMXGetToolCycleEditDialog,
			),

			helpers.NewEchoRoute(
				http.MethodPost,
				"/htmx/tools/cycle/edit",
				r.handler.HTMXPostToolCycleEditDialog,
			),

			helpers.NewEchoRoute(
				http.MethodPut,
				"/htmx/tools/cycle/edit",
				r.handler.HTMXPutToolCycleEditDialog,
			),

			// Delete a cycle table entry
			helpers.NewEchoRoute(
				http.MethodDelete,
				"/htmx/tools/cycle/delete",
				r.handler.HTMXDeleteToolCycle,
			),

			// Update tools binding data
			helpers.NewEchoRoute(
				http.MethodPatch,
				"/htmx/tools/tool/:id/bind",
				r.handler.HTMXPatchToolBinding,
			),

			helpers.NewEchoRoute(
				http.MethodPatch,
				"/htmx/tools/tool/:id/unbind",
				r.handler.HTMXPatchToolUnBinding,
			),
		},
	)
}
