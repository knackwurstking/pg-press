package router

import (
	"context"
	"embed"
	"os"

	database "github.com/knackwurstking/pgpress/internal/database/core"
	"github.com/knackwurstking/pgpress/internal/web/html"
	"github.com/knackwurstking/pgpress/internal/web/htmx"
	"github.com/knackwurstking/pgpress/internal/web/ws"
	"github.com/labstack/echo/v4"
)

var (
	//go:embed assets
	assets embed.FS

	serverPathPrefix = os.Getenv("SERVER_PATH_PREFIX")
)

func Serve(e *echo.Echo, db *database.DB) {
	e.StaticFS(serverPathPrefix+"/", echo.MustSubFS(assets, "assets"))

	wsh := startWebSocketHandlers(db)

	(&htmx.Nav{DB: db, WSHandler: wsh}).RegisterRoutes(e)

	(&html.Auth{DB: db}).RegisterRoutes(e)
	(&html.Home{}).RegisterRoutes(e)
	(&html.Feed{DB: db}).RegisterRoutes(e)
	(&html.Profile{DB: db}).RegisterRoutes(e)
	(&html.TroubleReports{DB: db}).RegisterRoutes(e)
	(&html.Tools{DB: db}).RegisterRoutes(e)
}

func startWebSocketHandlers(db *database.DB) *ws.WSHandlers {
	wsh := &ws.WSHandlers{
		Feed: ws.NewFeedHandler(db),
	}

	// Start the feed notification manager in a goroutine
	ctx := context.Background()
	go wsh.Feed.Start(ctx)

	// Set the notifier on the feeds for real-time updates
	db.Feeds.SetBroadcaster(wsh.Feed)

	return wsh
}
