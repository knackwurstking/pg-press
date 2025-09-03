package htmx

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pgpress/internal/constants"
	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/dberror"
	"github.com/knackwurstking/pgpress/internal/logger"
	feedscomp "github.com/knackwurstking/pgpress/internal/web/templates/components/feeds"
	"github.com/knackwurstking/pgpress/internal/utils"
)

type Feed struct {
	DB *database.DB
}

func (h *Feed) RegisterRoutes(e *echo.Echo) {
	utils.RegisterEchoRoutes(
		e,
		[]*utils.EchoRoute{
			utils.NewEchoRoute(http.MethodGet, "/htmx/feed/list", h.handleListGET),
		},
	)
}

func (h *Feed) handleListGET(c echo.Context) error {
	logger.HTMXHandlerFeed().Debug("Fetching feed data")

	// Get feeds
	feeds, err := h.DB.Feeds.ListRange(0, constants.MaxFeedsPerPage)
	if err != nil {
		logger.HTMXHandlerFeed().Error("Failed to fetch feeds: %v", err)
		return echo.NewHTTPError(dberror.GetHTTPStatusCode(err),
			"error getting feeds: "+err.Error())
	}

	logger.HTMXHandlerFeed().Debug("Retrieved %d feed items", len(feeds))

	// Update user's last feed
	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return err
	}

	logger.HTMXHandlerFeed().Debug("Rendering feed data for user %s", user.UserName)

	feedData := feedscomp.List(feeds, user.LastFeed)
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

		if err := h.DB.Users.Update(user, user); err != nil {
			logger.HTMXHandlerFeed().Error("Failed to update user's last feed: %v", err)
			return echo.NewHTTPError(dberror.GetHTTPStatusCode(err),
				"error updating user's last feed: "+err.Error())
		}

		logger.HTMXHandlerFeed().Debug("Successfully updated last feed for user %s", user.UserName)
	}

	return nil
}
