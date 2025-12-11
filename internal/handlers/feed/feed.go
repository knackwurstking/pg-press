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
	ui "github.com/knackwurstking/ui/ui-templ"
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
	ui.RegisterEchoRoutes(e, env.ServerPathPrefix, []*ui.EchoRoute{
		ui.NewEchoRoute(http.MethodGet, path, h.GetFeedPage),
		ui.NewEchoRoute(http.MethodGet, path+"/list", h.HTMXGetFeedsList),
	})
}

func (h *Handler) GetFeedPage(c echo.Context) error {
	t := templates.Page()
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "Feed Page")
	}

	return nil
}

func (h *Handler) HTMXGetFeedsList(c echo.Context) error {
	slog.Debug("Retrieving feed list", "offset", 0, "limit", env.MaxFeedsPerPage)

	feeds, merr := h.registry.Feeds.ListRange(0, env.MaxFeedsPerPage)
	if merr != nil {
		return merr.Echo()
	}

	user, merr := utils.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	userMap := make(map[models.TelegramID]*models.User)
	for _, feed := range feeds {
		feedUser, merr := h.registry.Users.Get(feed.UserID)
		if merr != nil {
			slog.Error("failed to get user", "error", merr)
			continue
		}
		userMap[feed.UserID] = feedUser
	}

	t := templates.FeedsList(feeds, user.LastFeed, userMap)
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "FeedsList")
	}

	if len(feeds) > 0 && feeds[0].ID != user.LastFeed {
		oldLastFeed := user.LastFeed
		user.LastFeed = feeds[0].ID
		slog.Debug(
			"update users last viewed feed",
			"user_name", user.Name,
			"last_feed_from", oldLastFeed,
			"last_feed_to", user.LastFeed,
		)

		merr = h.registry.Users.Update(user)
		if merr != nil {
			return merr.WrapEcho("update user's last feed")
		}
	}

	return nil
}
