package nav

import (
	"io/fs"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/labstack/echo/v4"
)

func Serve(templates fs.FS, serverPathPrefix string, e *echo.Echo, db *pgvis.DB) {
	e.GET(serverPathPrefix+"/nav/feed-counter", func(c echo.Context) error {
		return GETFeedCounter(templates, c, db)
	})
}
