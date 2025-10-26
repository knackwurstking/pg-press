package handlers

import (
	"net/http"

	"github.com/knackwurstking/pgpress/components"
	"github.com/knackwurstking/pgpress/env"
	"github.com/knackwurstking/pgpress/logger"
	"github.com/knackwurstking/pgpress/models"
	"github.com/knackwurstking/pgpress/services"
	"github.com/knackwurstking/pgpress/utils"
	"github.com/labstack/echo/v4"
)

type Feed struct {
	*Base
}

func NewFeed(db *services.Registry) *Feed {
	return &Feed{
		Base: NewBase(db, logger.NewComponentLogger("Feed")),
	}
}

func (h *Feed) RegisterRoutes(e *echo.Echo) {
	utils.RegisterEchoRoutes(e, []*utils.EchoRoute{
		utils.NewEchoRoute(http.MethodGet, "/feed", h.GetFeedPage),
		utils.NewEchoRoute(http.MethodGet, "/htmx/feed/list", h.HTMXGetFeedsList),
	})
}

func (h *Feed) GetFeedPage(c echo.Context) error {
	page := components.PageFeed()
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return HandleError(err, "failed to render feed page")
	}
	return nil
}

func (h *Feed) HTMXGetFeedsList(c echo.Context) error {
	feeds, err := h.Registry.Feeds.ListRange(0, env.MaxFeedsPerPage)
	if err != nil {
		return HandleError(err, "failed to get feeds")
	}

	user, err := GetUserFromContext(c)
	if err != nil {
		return HandleError(err, "failed to get user from context")
	}

	userMap := make(map[models.TelegramID]*models.User)
	for _, feed := range feeds {
		feedUser, err := h.Registry.Users.Get(feed.UserID)
		if err != nil {
			h.Log.Error("failed to get user: %v", err)
			continue
		}
		userMap[feed.UserID] = feedUser
	}

	feedData := components.FeedsList(feeds, user.LastFeed, userMap)
	if err := feedData.Render(c.Request().Context(), c.Response()); err != nil {
		return HandleError(err, "failed to render feed data")
	}

	if len(feeds) > 0 && feeds[0].ID != user.LastFeed {
		oldLastFeed := user.LastFeed
		user.LastFeed = feeds[0].ID
		h.Log.Info("Updating user %s last feed from %d to %d",
			user.Name, oldLastFeed, user.LastFeed)

		if err := h.Registry.Users.Update(user); err != nil {
			return HandleError(err, "failed to update user's last feed")
		}
	}

	return nil
}
