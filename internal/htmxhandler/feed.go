package htmxhandler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/templates/components"
	"github.com/knackwurstking/pgpress/internal/utils"
)

type Feed struct {
	*Base
}

func (h *Feed) RegisterRoutes(e *echo.Echo) {
	e.GET(h.ServerPathPrefix+"/data", h.handleGetData)
}

func (h *Feed) handleGetData(c echo.Context) error {
	// Get feeds
	feeds, err := h.DB.Feeds.ListRange(0, 100)
	if err != nil {
		return echo.NewHTTPError(database.GetHTTPStatusCode(err),
			"error getting feeds: "+err.Error())
	}

	// Update user's last feed
	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	if len(feeds) > 0 {
		user.LastFeed = feeds[0].ID
		if err := h.DB.Users.Update(user.TelegramID, user); err != nil {
			return echo.NewHTTPError(database.GetHTTPStatusCode(err),
				"error updating user's last feed: "+err.Error())
		}
	}

	feedData := components.FeedData(feeds, user.LastFeed)
	err = feedData.Render(c.Request().Context(), c.Response())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"error rendering feed data: "+err.Error())
	}

	return nil
}
