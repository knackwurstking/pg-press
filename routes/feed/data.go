package feed

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/knackwurstking/pg-vis/routes/utils"
)

type Data struct {
	Feeds []*pgvis.Feed
}

func (d *Data) Render(f *pgvis.Feed) template.HTML {
	return template.HTML(
		fmt.Sprintf(
			`<article id="feed-%d">
         		<main>%s</main>
         		<footer><small>%s</small></footer>
          	</article>`,
			f.ID,
			f.Main,
			time.UnixMilli(f.Time).Local().String(),
		),
	)
}

func GETData(templates fs.FS, c echo.Context, db *pgvis.DB) *echo.HTTPError {
	data := &Data{
		Feeds: make([]*pgvis.Feed, 0),
	}

	{ // Create Feeds
		feeds, err := db.Feeds.ListRange(0, 100)
		if err != nil {
			return utils.HandlePgvisError(c, err)
		}

		data.Feeds = feeds
	}

	{ // Update Users Last Feed
		user, herr := utils.GetUserFromContext(c)
		if herr != nil {
			return herr
		}

		for _, feed := range data.Feeds {
			user.LastFeed = feed.ID
			break
		}

		err := db.Users.Update(user.TelegramID, user)
		if err != nil {
			return utils.HandlePgvisError(c, err)
		}
	}

	t, err := template.ParseFS(templates, "templates/feed/data.html")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err = t.Execute(c.Response(), data); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return nil
}
