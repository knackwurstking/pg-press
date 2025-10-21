package profile

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
			helpers.NewEchoRoute(http.MethodGet, "/profile", r.handler.GetProfilePage),

			// HTMX
			helpers.NewEchoRoute(http.MethodGet, "/htmx/profile/cookies", r.handler.HTMXGetCookies),

			helpers.NewEchoRoute(http.MethodDelete, "/htmx/profile/cookies",
				r.handler.HTMXDeleteCookies),
		},
	)
}
