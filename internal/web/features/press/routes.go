package press

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
			// Press page
			helpers.NewEchoRoute(http.MethodGet, "/tools/press/:press",
				r.handler.GetPressPage),

			// HTMX endpoints for press content
			helpers.NewEchoRoute(http.MethodGet, "/htmx/tools/press/:press/active-tools",
				r.handler.HTMXGetPressActiveTools),

			helpers.NewEchoRoute(http.MethodGet, "/htmx/tools/press/:press/metal-sheets",
				r.handler.HTMXGetPressMetalSheets),

			helpers.NewEchoRoute(http.MethodGet, "/htmx/tools/press/:press/cycles",
				r.handler.HTMXGetPressCycles),

			helpers.NewEchoRoute(http.MethodGet, "/htmx/tools/press/:press/notes",
				r.handler.HTMXGetPressNotes),
		},
	)
}
