package nav

import (
	"fmt"
	"io/fs"
	"net/http"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/knackwurstking/pg-vis/routes/constants"
	"github.com/knackwurstking/pg-vis/routes/internal/utils"
	"github.com/labstack/echo/v4"
)

type FeedCounter struct {
	Count int
}

func GETFeedCounter(templates fs.FS, c echo.Context, db *pgvis.DB) *echo.HTTPError {
	data := &FeedCounter{}

	feeds, err := db.Feeds.ListRange(0, 100)
	if err != nil {
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			fmt.Errorf("list feeds: %w", err),
		)
	}

	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	for _, feed := range feeds {
		if feed.ID > user.LastFeed {
			data.Count++
		} else {
			break
		}
	}

	return utils.HandleTemplate(c, data,
		templates,
		[]string{
			constants.LegacyFeedCounterTemplatePath,
		},
	)
}
