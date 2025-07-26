package handler

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

// NewHandler creates a new feed handler.
func NewFeed(base *Base) *Feed {
	return &Feed{base}
}

func (h *Feed) RegisterRoutes(e *echo.Echo) {
	e.GET(h.ServerPathPrefix+"/feed", h.handleFeed)
	e.GET(h.ServerPathPrefix+"/feed/data", h.handleGetData)
}

func (h *Feed) handleFeed(c echo.Context) error {
	return utils.HandleTemplate(c, nil,
		h.Templates,
		constants.FeedPageTemplates,
	)
}

func (h *Feed) handleGetData(c echo.Context) error {
	data := &FeedTemplateData{
		Feeds: make([]*database.Feed, 0),
	}

	// Get feeds
	feeds, err := h.DB.Feeds.ListRange(0, 100)
	if err != nil {
		return utils.HandlePgvisError(c, err)
	}
	data.Feeds = feeds

	// Update user's last feed
	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	data.LastFeedID = user.LastFeed

	if len(data.Feeds) > 0 {
		user.LastFeed = data.Feeds[0].ID
		err := h.DB.Users.Update(user.TelegramID, user)
		if err != nil {
			return utils.HandlePgvisError(c, err)
		}
	}

	return utils.HandleTemplate(c, data,
		h.Templates,
		[]string{
			constants.FeedDataComponentTemplatePath,
		},
	)
}
