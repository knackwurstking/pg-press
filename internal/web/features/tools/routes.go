package tools

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
			// Pages
			helpers.NewEchoRoute(http.MethodGet, "/tools",
				r.handler.GetToolsPage),

			// HTMX
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

			// Mark a tool as dead
			helpers.NewEchoRoute(http.MethodPatch, "/htmx/tools/mark-dead",
				r.handler.HTMXMarkToolAsDead),

			// Section handlers
			helpers.NewEchoRoute(http.MethodGet, "/htmx/tools/section/press",
				r.handler.HTMXGetSectionPress),
			helpers.NewEchoRoute(http.MethodGet, "/htmx/tools/section/tools",
				r.handler.HTMXGetSectionTools),

			// Admin section handlers
			helpers.NewEchoRoute(http.MethodGet, "/htmx/tools/admin/overlapping-tools",
				r.handler.HTMXGetAdminOverlappingTools),
		},
	)
}
