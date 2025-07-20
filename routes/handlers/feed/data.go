package feed

import (
	"io/fs"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/knackwurstking/pg-vis/routes/constants"
	"github.com/knackwurstking/pg-vis/routes/internal/utils"
)

type Data struct {
	Feeds []*pgvis.Feed
}

func GETData(templates fs.FS, c echo.Context, db *pgvis.DB) *echo.HTTPError {
	data := &Data{
		Feeds: make([]*pgvis.Feed, 0),
	}

	// Get feeds
	feeds, err := db.Feeds.ListRange(0, 100)
	if err != nil {
		return utils.HandlePgvisError(c, err)
	}
	data.Feeds = feeds

	// Update user's last feed
	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	if len(data.Feeds) > 0 {
		user.LastFeed = data.Feeds[0].ID
		err := db.Users.Update(user.TelegramID, user)
		if err != nil {
			return utils.HandlePgvisError(c, err)
		}
	}

	return utils.HandleTemplate(c, data,
		templates,
		[]string{
			constants.LegacyFeedDataTemplatePath,
		},
	)
}
