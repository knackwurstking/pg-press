package feed

import (
	"html/template"
	"io/fs"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/pgvis"
)

// TODO: Websocket handler for the nav icon button count per user
// TODO: Websocket handler for live feeds

func Serve(templates fs.FS, serverPathPrefix string, e *echo.Echo, db *pgvis.DB) {
	// TODO: Server side rendering feeds for user
	e.GET(serverPathPrefix+"/feed", func(c echo.Context) error {
		t, err := template.ParseFS(templates,
			"templates/layout.html",
			"templates/feed.html",
		)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		err = t.Execute(c.Response(), nil)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		return nil
	})
}
