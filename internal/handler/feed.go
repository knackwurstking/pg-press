package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/htmxhandler"
	"github.com/knackwurstking/pgpress/internal/templates/pages"
)

type Feed struct {
	DB *database.DB
}

func (h *Feed) RegisterRoutes(e *echo.Echo) {
	prefix := "/feed"

	e.GET(serverPathPrefix+prefix, h.handleFeed)

	htmxFeed := htmxhandler.Feed{DB: h.DB}
	htmxFeed.RegisterRoutes(e)
}

func (h *Feed) handleFeed(c echo.Context) error {
	page := pages.FeedPage()
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render feed page: "+err.Error())
	}
	return nil
}
