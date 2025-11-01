package handlers

import (
	"net/http"

	"github.com/knackwurstking/pg-press/components"
	"github.com/knackwurstking/pg-press/env"
	"github.com/knackwurstking/pg-press/models"
	"github.com/knackwurstking/pg-press/services"
	"github.com/knackwurstking/pg-press/utils"
	"github.com/labstack/echo/v4"
)

type Feed struct {
	registry *services.Registry
}

func NewFeed(r *services.Registry) *Feed {
	return &Feed{
		registry: r,
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
	feeds, err := h.registry.Feeds.ListRange(0, env.MaxFeedsPerPage)
	if err != nil {
		return HandleError(err, "failed to get feeds")
	}

	user, err := GetUserFromContext(c)
	if err != nil {
		return HandleError(err, "failed to get user from context")
	}

	userMap := make(map[models.TelegramID]*models.User)
	for _, feed := range feeds {
		feedUser, err := h.registry.Users.Get(feed.UserID)
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

		if err := h.registry.Users.Update(user); err != nil {
			return HandleError(err, "failed to update user's last feed")
		}
	}

	return nil
}
