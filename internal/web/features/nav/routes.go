package nav

import (
	"net/http"

	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/web/shared/helpers"
	"github.com/knackwurstking/pgpress/internal/web/wshandlers"

	"github.com/labstack/echo/v4"
)

type Routes struct {
	handler *Handler
}

func NewRoutes(db *database.DB, ws *wshandlers.FeedHandler) *Routes {
	return &Routes{
		handler: NewHandler(db, ws),
	}
}

func (r *Routes) RegisterRoutes(e *echo.Echo) {
	helpers.RegisterEchoRoutes(
		e,
		[]*helpers.EchoRoute{
			// HTMX
			helpers.NewEchoRoute(http.MethodGet, "/htmx/nav/feed-counter",
				r.handler.GetFeedCounter),
		},
	)
}
