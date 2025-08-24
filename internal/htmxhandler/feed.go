package htmxhandler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/templates/components"
	"github.com/knackwurstking/pgpress/internal/utils"
)

type Feed struct {
	DB *database.DB
}

func (h *Feed) RegisterRoutes(e *echo.Echo) {
	e.GET(serverPathPrefix+"/htmx/feed/data", h.handleGetData)
	e.GET(serverPathPrefix+"/htmx/feed/data/", h.handleGetData)
}

func (h *Feed) handleGetData(c echo.Context) error {
	logger.HTMXHandlerFeed().Debug("Fetching feed data")

	// Get feeds
	feeds, err := h.DB.Feeds.ListRange(0, 100)
	if err != nil {
		logger.HTMXHandlerFeed().Error("Failed to fetch feeds: %v", err)
		return echo.NewHTTPError(database.GetHTTPStatusCode(err),
			"error getting feeds: "+err.Error())
	}

	logger.HTMXHandlerFeed().Debug("Retrieved %d feed items", len(feeds))

	// Update user's last feed
	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return err
	}

	logger.HTMXHandlerFeed().Debug("Rendering feed data for user %s", user.UserName)

	feedData := components.FeedData(feeds, user.LastFeed)
	err = feedData.Render(c.Request().Context(), c.Response())
	if err != nil {
		logger.HTMXHandlerFeed().Error("Failed to render feed data: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			"error rendering feed data: "+err.Error())
	}

	if len(feeds) > 0 {
		oldLastFeed := user.LastFeed
		user.LastFeed = feeds[0].ID
		logger.HTMXHandlerFeed().Info("Updating user %s last feed from %d to %d",
			user.UserName, oldLastFeed, user.LastFeed)

		if err := h.DB.Users.Update(user.TelegramID, user); err != nil {
			logger.HTMXHandlerFeed().Error("Failed to update user's last feed: %v", err)
			return echo.NewHTTPError(database.GetHTTPStatusCode(err),
				"error updating user's last feed: "+err.Error())
		}

		logger.HTMXHandlerFeed().Debug("Successfully updated last feed for user %s", user.UserName)
	}

	return nil
}
