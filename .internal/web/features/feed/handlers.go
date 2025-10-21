package feed

import (
	"github.com/knackwurstking/pgpress/internal/constants"
	"github.com/knackwurstking/pgpress/internal/services"
	"github.com/knackwurstking/pgpress/internal/web/features/feed/templates"
	"github.com/knackwurstking/pgpress/internal/web/shared/base"
	"github.com/knackwurstking/pgpress/pkg/logger"
	"github.com/knackwurstking/pgpress/pkg/models"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	*base.Handler
}

func NewHandler(db *services.Registry) *Handler {
	return &Handler{
		Handler: base.NewHandler(db, logger.NewComponentLogger("Feed")),
	}
}

func (h *Handler) GetFeedPage(c echo.Context) error {
	page := templates.FeedPage()
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c,
			"failed to render feed page: "+err.Error())
	}
	return nil
}

func (h *Handler) HTMXGetFeedsList(c echo.Context) error {
	// Get feeds
	feeds, err := h.DB.Feeds.ListRange(0, constants.MaxFeedsPerPage)
	if err != nil {
		return h.HandleError(c, err, "failed to getting feeds")
	}

	// Update user's last feed
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed getting user from context")
	}

	// Create users map map[int64]*models.User
	userMap := make(map[int64]*models.User)

	// Populate users map
	for _, feed := range feeds {
		user, err := h.DB.Users.Get(feed.UserID)
		if err != nil {
			h.Log.Error("failed to getting user: %v", err)
		}
		userMap[feed.UserID] = user
	}

	feedData := templates.FeedsList(feeds, user.LastFeed, userMap)
	err = feedData.Render(c.Request().Context(), c.Response())
	if err != nil {
		return h.RenderInternalError(c,
			"error rendering feed data: "+err.Error())
	}

	// Update user's last feed
	if len(feeds) > 0 {
		oldLastFeed := user.LastFeed
		user.LastFeed = feeds[0].ID
		h.Log.Info("Updating user %s last feed from %d to %d",
			user.Name, oldLastFeed, user.LastFeed)

		if err := h.DB.Users.Update(user); err != nil {
			return h.HandleError(c, err, "failed to update user's last feed")
		}
	}

	return nil
}
