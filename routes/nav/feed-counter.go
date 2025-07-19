package nav

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/http"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/knackwurstking/pg-vis/routes/shared"
	"github.com/knackwurstking/pg-vis/routes/utils"
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

	t, err := template.ParseFS(templates, shared.FeedCounterTemplatePath)
	if err != nil {
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			fmt.Errorf("template parsing: %w", err),
		)
	}

	if err := t.Execute(c.Response(), data); err != nil {
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			fmt.Errorf("template executing: %w", err),
		)
	}

	return nil
}
