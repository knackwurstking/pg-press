package htmx

import (
	"net/http"

	"github.com/knackwurstking/pgpress/internal/constants"
	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/web/helpers"
	"github.com/knackwurstking/pgpress/internal/web/templates/feedpage"
	"github.com/knackwurstking/pgpress/pkg/utils"

	"github.com/labstack/echo/v4"
)

type Feed struct {
	DB *database.DB
}

func (h *Feed) RegisterRoutes(e *echo.Echo) {
	helpers.RegisterEchoRoutes(
		e,
		[]*helpers.EchoRoute{
			helpers.NewEchoRoute(http.MethodGet, "/htmx/feed/list", h.handleListGET),
		},
	)
}

func (h *Feed) handleListGET(c echo.Context) error {
	// Get feeds
	feeds, err := h.DB.Feeds.ListRange(0, constants.MaxFeedsPerPage)
	if err != nil {
		return echo.NewHTTPError(utils.GetHTTPStatusCode(err),
			"error getting feeds: "+err.Error())
	}

	// Update user's last feed
	user, err := helpers.GetUserFromContext(c)
	if err != nil {
		return err
	}

	feedData := feedpage.List(feeds, user.LastFeed)
	err = feedData.Render(c.Request().Context(), c.Response())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"error rendering feed data: "+err.Error())
	}

	if len(feeds) > 0 {
		oldLastFeed := user.LastFeed
		user.LastFeed = feeds[0].ID
		logger.HTMXHandlerFeed().Debug("Updating user %s last feed from %d to %d",
			user.Name, oldLastFeed, user.LastFeed)

		if err := h.DB.Users.Update(user, user); err != nil {
			return echo.NewHTTPError(utils.GetHTTPStatusCode(err),
				"error updating user's last feed: "+err.Error())
		}

	}

	return nil
}
