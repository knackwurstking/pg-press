package nav

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/http"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/knackwurstking/pg-vis/routes/utils"
	"github.com/labstack/echo/v4"
)

type FeedCounter struct {
	Count int
}

func GETFeedCounter(templates fs.FS, c echo.Context, db *pgvis.DB) *echo.HTTPError {
	data := &FeedCounter{}

	_, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	// TODO: Get number of feeds the user has not viewed yet
	data.Count = 1 // Just for testing

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
