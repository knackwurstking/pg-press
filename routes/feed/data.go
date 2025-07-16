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

func (d *Data) HTML(f *pgvis.Feed) (html string) {
	switch v := f.Data.(type) {
	case *pgvis.User:
		return fmt.Sprintf(
			`
				<article id="feed%d">
					<main><p>New user: %s</p></main>
					<footer>%s</footer>
				</article>
			`,
			f.ID,
			v.UserName,
			time.UnixMilli(f.Time).Local().String(),
		)
	default:
		return ""
	}
}

func GETData(templates fs.FS, c echo.Context, db *pgvis.DB) *echo.HTTPError {
	data := &Data{
		Feeds: make([]*pgvis.Feed, 0),
	}

	// TODO: List the last 100 feeds from the database

	t, err := template.ParseFS(templates, "templates/feed/cookies.html")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("template parsing failed: %s", err))
	}

	if err := t.Execute(c.Response(), data); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("template executing failed: %s", err))
	}

	return nil
}
