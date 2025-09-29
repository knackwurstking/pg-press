package umbau

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
			helpers.NewEchoRoute(http.MethodGet, "/tools/press/:press/umbau",
				r.handler.GetUmbauPage),
			helpers.NewEchoRoute(http.MethodPost, "/tools/press/:press/umbau",
				r.handler.PostUmbauPage),

			// HTMX
		},
	)
}
