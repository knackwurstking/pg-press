package html

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pgpress/internal/database/core"
	"github.com/knackwurstking/pgpress/internal/web/htmx"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/web/templates/pages"
	"github.com/knackwurstking/pgpress/internal/utils"
)

type Feed struct {
	DB *database.DB
}

func (h *Feed) RegisterRoutes(e *echo.Echo) {
	utils.RegisterEchoRoutes(
		e,
		[]*utils.EchoRoute{
			utils.NewEchoRoute(http.MethodGet, "/feed", h.handleFeed),
		},
	)

	htmxFeed := htmx.Feed{DB: h.DB}
	htmxFeed.RegisterRoutes(e)
}

func (h *Feed) handleFeed(c echo.Context) error {
	logger.HandlerFeed().Debug("Rendering feed page")

	page := pages.FeedPage()
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		logger.HandlerFeed().Error("Failed to render feed page: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render feed page: "+err.Error())
	}
	return nil
}
