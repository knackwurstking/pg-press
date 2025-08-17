package router

import (
	"context"
	"embed"
	"os"

	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/handler"
	"github.com/knackwurstking/pgpress/internal/htmxhandler"
	"github.com/knackwurstking/pgpress/internal/wshandler"
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

	(&htmxhandler.Nav{DB: db, WSHandler: wsh}).RegisterRoutes(e)

	(&handler.Auth{DB: db}).RegisterRoutes(e)
	(&handler.Home{}).RegisterRoutes(e)
	(&handler.Feed{DB: db}).RegisterRoutes(e)
	(&handler.Profile{DB: db}).RegisterRoutes(e)
	(&handler.TroubleReports{DB: db}).RegisterRoutes(e)
	(&handler.Tools{DB: db}).RegisterRoutes(e)
}

func startWebSocketHandlers(db *database.DB) *wshandler.WSHandlers {
	wsh := &wshandler.WSHandlers{
		Feed: wshandler.NewFeedHandler(db),
	}

	// Start the feed notification manager in a goroutine
	ctx := context.Background()
	go wsh.Feed.Start(ctx)

	// Set the notifier on the feeds for real-time updates
	db.Feeds.SetBroadcaster(wsh.Feed)

	return wsh
}
