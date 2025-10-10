package metalsheets

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
			// HTMX
			// GET route for displaying the edit dialog
			helpers.NewEchoRoute(http.MethodGet, "/htmx/metal-sheets/edit",
				r.handler.HTMXGetEditMetalSheetDialog),

			// POST route for creating a new metal sheet
			helpers.NewEchoRoute(http.MethodPost, "/htmx/metal-sheets/edit",
				r.handler.HTMXPostEditMetalSheetDialog),

			// PUT route for updating an existing metal sheet
			helpers.NewEchoRoute(http.MethodPut, "/htmx/metal-sheets/edit",
				r.handler.HTMXPutEditMetalSheetDialog),

			// DELETE route for removing a metal sheet
			helpers.NewEchoRoute(http.MethodDelete, "/htmx/metal-sheets/delete",
				r.handler.HTMXDeleteMetalSheet),
		},
	)
}
