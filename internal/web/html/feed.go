package html

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/web/helpers"
	"github.com/knackwurstking/pgpress/internal/web/templates/feedpage"
)

type Feed struct {
	DB *database.DB
}

func (h *Feed) RegisterRoutes(e *echo.Echo) {
	helpers.RegisterEchoRoutes(
		e,
		[]*helpers.EchoRoute{
			helpers.NewEchoRoute(http.MethodGet, "/feed", h.handleFeed),
		},
	)
}

func (h *Feed) handleFeed(c echo.Context) error {
	logger.HandlerFeed().Debug("Rendering feed page")

	page := feedpage.Page()
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		logger.HandlerFeed().Error("Failed to render feed page: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render feed page: "+err.Error())
	}
	return nil
}
