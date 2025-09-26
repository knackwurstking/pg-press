package feed

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/web/shared/handlers"
	"github.com/knackwurstking/pgpress/internal/web/shared/helpers"
)

type Feed struct {
	*handlers.BaseHandler
}

func NewFeed(db *database.DB) *Feed {
	return &Feed{
		BaseHandler: handlers.NewBaseHandler(db, logger.HandlerFeed()),
	}
}

func (h *Feed) RegisterRoutes(e *echo.Echo) {
	helpers.RegisterEchoRoutes(
		e,
		[]*helpers.EchoRoute{
			helpers.NewEchoRoute(http.MethodGet, "/feed", h.HandleFeed),
		},
	)
}

func (h *Feed) HandleFeed(c echo.Context) error {
	page := FeedPage()
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c,
			"failed to render feed page: "+err.Error())
	}
	return nil
}
