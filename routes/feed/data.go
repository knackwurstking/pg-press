package feed

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/pgvis"
)

type Data struct {
	Feeds []*pgvis.Feed
}

func (d *Data) Render(f *pgvis.Feed) (html string) {
	return fmt.Sprintf(
		`<article id="feed%d">
			<main>%s</main>
			<footer>%s</footer>
		</article>`,
		f.ID,
		f.Main,
		time.UnixMilli(f.Time).Local().String(),
	)
}

func GETData(templates fs.FS, c echo.Context, db *pgvis.DB) *echo.HTTPError {
	data := &Data{
		Feeds: make([]*pgvis.Feed, 0),
	}

	feeds, err := db.Feeds.ListRange(0, 100)
	if err != nil {
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			fmt.Errorf("list feeds from range 0..100: %s", err.Error()),
		)
	}
	data.Feeds = feeds

	t, err := template.ParseFS(templates, "templates/feed/data.html")
	if err != nil {
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			fmt.Errorf("template parsing failed: %s", err.Error()),
		)
	}

	if err := t.Execute(c.Response(), data); err != nil {
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			fmt.Errorf("template executing failed: %s", err.Error()),
		)
	}

	return nil
}
