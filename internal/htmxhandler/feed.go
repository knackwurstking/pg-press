package htmxhandler

import (
	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/internal/constants"
	"github.com/knackwurstking/pg-vis/internal/database"
	"github.com/knackwurstking/pg-vis/internal/utils"
)

type FeedTemplateData struct {
	Feeds      []*database.Feed
	LastFeedID int64
}

type Feed struct {
	*Base
}

func (h *Feed) RegisterRoutes(e *echo.Echo) {
	e.GET(h.ServerPathPrefix+"/data", h.handleGetData)
}

func (h *Feed) handleGetData(c echo.Context) error {
	// Get feeds
	feeds, err := h.DB.Feeds.ListRange(0, 100)
	if err != nil {
		return utils.HandlePgvisError(c, err)
	}

	// Update user's last feed
	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	data := &FeedTemplateData{
		Feeds:      feeds,
		LastFeedID: user.LastFeed,
	}

	if len(feeds) > 0 {
		user.LastFeed = feeds[0].ID
		if err := h.DB.Users.Update(user.TelegramID, user); err != nil {
			return utils.HandlePgvisError(c, err)
		}
	}

	return utils.HandleTemplate(c, data, h.Templates,
		[]string{constants.HTMXFeedDataTemplatePath})
}
