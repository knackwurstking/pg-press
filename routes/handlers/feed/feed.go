// Package feed provides HTTP route handlers for feed management.
package feed

import (
	"io/fs"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/knackwurstking/pg-vis/routes/constants"
	"github.com/knackwurstking/pg-vis/routes/internal/utils"
)

type Handler struct {
	db               *pgvis.DB
	serverPathPrefix string
	templates        fs.FS
}

// NewHandler creates a new feed handler.
func NewHandler(db *pgvis.DB, serverPathPrefix string, templates fs.FS) *Handler {
	return &Handler{
		db:               db,
		serverPathPrefix: serverPathPrefix,
		templates:        templates,
	}
}

func (h *Handler) RegisterRoutes(e *echo.Echo) {
	e.GET(h.serverPathPrefix+"/feed", h.handleFeed)
	e.GET(h.serverPathPrefix+"/feed/data", h.handleGetData)
}

func (h *Handler) handleFeed(c echo.Context) error {
	return utils.HandleTemplate(c, nil,
		h.templates,
		constants.FeedPageTemplates,
	)
}

type DataTemplateData struct {
	Feeds      []*pgvis.Feed
	LastFeedID int64
}

func (h *Handler) handleGetData(c echo.Context) error {
	data := &DataTemplateData{
		Feeds: make([]*pgvis.Feed, 0),
	}

	// Get feeds
	feeds, err := h.db.Feeds.ListRange(0, 100)
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
		err := h.db.Users.Update(user.TelegramID, user)
		if err != nil {
			return utils.HandlePgvisError(c, err)
		}
	}

	return utils.HandleTemplate(c, data,
		h.templates,
		[]string{
			constants.FeedDataComponentTemplatePath,
		},
	)
}
