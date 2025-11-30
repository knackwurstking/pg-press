package feed

import (
	"log/slog"
	"net/http"

	"github.com/knackwurstking/pg-press/env"
	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/handlers/feed/templates"
	"github.com/knackwurstking/pg-press/models"
	"github.com/knackwurstking/pg-press/services"
	"github.com/knackwurstking/pg-press/utils"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	registry *services.Registry
}

func NewHandler(r *services.Registry) *Handler {
	return &Handler{
		registry: r,
	}
}

func (h *Handler) RegisterRoutes(e *echo.Echo, path string) {
	utils.RegisterEchoRoutes(e, []*utils.EchoRoute{
		utils.NewEchoRoute(http.MethodGet, path, h.GetFeedPage),
		utils.NewEchoRoute(http.MethodGet, path+"/list", h.HTMXGetFeedsList),
	})
}

func (h *Handler) GetFeedPage(c echo.Context) error {
	page := templates.Page()
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.Handler(err, "render feed page")
	}
	return nil
}

func (h *Handler) HTMXGetFeedsList(c echo.Context) error {
	feeds, err := h.registry.Feeds.ListRange(0, env.MaxFeedsPerPage)
	if err != nil {
		return errors.Handler(err, "get feeds")
	}

	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	userMap := make(map[models.TelegramID]*models.User)
	for _, feed := range feeds {
		feedUser, err := h.registry.Users.Get(feed.UserID)
		if err != nil {
			slog.Error("failed to get user", "error", err)
			continue
		}
		userMap[feed.UserID] = feedUser
	}

	feedData := templates.FeedsList(feeds, user.LastFeed, userMap)
	if err := feedData.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.Handler(err, "render feed data")
	}

	if len(feeds) > 0 && feeds[0].ID != user.LastFeed {
		oldLastFeed := user.LastFeed
		user.LastFeed = feeds[0].ID
		slog.Info("Updating users last feed",
			"user_name", user.Name, "last_feed_from", oldLastFeed, "last_feed_to", user.LastFeed)

		if err := h.registry.Users.Update(user); err != nil {
			return errors.Handler(err, "update user's last feed")
		}
	}

	return nil
}
