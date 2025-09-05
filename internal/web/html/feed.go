package html

import (
	"net/http"

	"github.com/labstack/echo/v4"

	database "github.com/knackwurstking/pgpress/internal/database/core"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/web/htmx"
	pagesfeed "github.com/knackwurstking/pgpress/internal/web/templates/pages/feed"
	"github.com/knackwurstking/pgpress/internal/web/webhelpers"
)

type Feed struct {
	DB *database.DB
}

func (h *Feed) RegisterRoutes(e *echo.Echo) {
	webhelpers.RegisterEchoRoutes(
		e,
		[]*webhelpers.EchoRoute{
			webhelpers.NewEchoRoute(http.MethodGet, "/feed", h.handleFeed),
		},
	)

	htmxFeed := htmx.Feed{DB: h.DB}
	htmxFeed.RegisterRoutes(e)
}

func (h *Feed) handleFeed(c echo.Context) error {
	logger.HandlerFeed().Debug("Rendering feed page")

	page := pagesfeed.Page()
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		logger.HandlerFeed().Error("Failed to render feed page: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render feed page: "+err.Error())
	}
	return nil
}
