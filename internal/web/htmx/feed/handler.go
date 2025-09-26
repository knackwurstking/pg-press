package feed

import (
	"net/http"

	"github.com/knackwurstking/pgpress/internal/constants"
	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/web/shared/components"
	"github.com/knackwurstking/pgpress/internal/web/shared/handlers"
	"github.com/knackwurstking/pgpress/internal/web/shared/helpers"
	"github.com/knackwurstking/pgpress/pkg/models"

	"github.com/labstack/echo/v4"
)

type Feed struct {
	*handlers.BaseHandler
}

func NewFeed(db *database.DB) *Feed {
	return &Feed{
		BaseHandler: handlers.NewBaseHandler(db, logger.HTMXHandlerFeed()),
	}
}

func (h *Feed) RegisterRoutes(e *echo.Echo) {
	helpers.RegisterEchoRoutes(
		e,
		[]*helpers.EchoRoute{
			helpers.NewEchoRoute(http.MethodGet, "/htmx/feed/list", h.HandleListGET),
		},
	)
}

func (h *Feed) HandleListGET(c echo.Context) error {
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
			h.LogError("failed to getting user: %v", err)
		}
		userMap[feed.UserID] = user
	}

	feedData := components.FeedsList(feeds, user.LastFeed, userMap)
	err = feedData.Render(c.Request().Context(), c.Response())
	if err != nil {
		return h.RenderInternalError(c,
			"error rendering feed data: "+err.Error())
	}

	// Update user's last feed
	if len(feeds) > 0 {
		oldLastFeed := user.LastFeed
		user.LastFeed = feeds[0].ID
		h.LogDebug("Updating user %s last feed from %d to %d",
			user.Name, oldLastFeed, user.LastFeed)

		if err := h.DB.Users.Update(user); err != nil {
			return h.HandleError(c, err, "failed to update user's last feed")
		}
	}

	return nil
}
