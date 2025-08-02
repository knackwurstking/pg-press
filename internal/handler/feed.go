package handler

import (
	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pgpress/internal/constants"
	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/htmxhandler"
	"github.com/knackwurstking/pgpress/internal/utils"
)

type FeedTemplateData struct {
	Feeds      []*database.Feed
	LastFeedID int64
}

type Feed struct {
	*Base
}

func (h *Feed) RegisterRoutes(e *echo.Echo) {
	prefix := "/feed"

	e.GET(h.ServerPathPrefix+prefix, h.handleFeed)

	htmxFeed := htmxhandler.Feed{Base: h.NewHTMX(prefix)}
	htmxFeed.RegisterRoutes(e)
}

func (h *Feed) handleFeed(c echo.Context) error {
	return utils.HandleTemplate(c, nil, h.Templates,
		constants.FeedPageTemplates)
}
