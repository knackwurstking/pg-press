package router

import (
	"context"
	"embed"
	"os"

	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/web/html"
	"github.com/knackwurstking/pgpress/internal/web/htmx"
	"github.com/knackwurstking/pgpress/internal/web/wshandlers"

	"github.com/labstack/echo/v4"
)

var (
	//go:embed assets
	assets embed.FS

	serverPathPrefix = os.Getenv("SERVER_PATH_PREFIX")
)

func Serve(e *echo.Echo, db *database.DB) {
	// Static File Server
	e.StaticFS(serverPathPrefix+"/", echo.MustSubFS(assets, "assets"))

	// WebSocket Handlers
	wsh := startWebSocketHandlers(db)

	// HTML Handler (Old)
	(&html.Feed{DB: db}).RegisterRoutes(e)
	(&html.Profile{DB: db}).RegisterRoutes(e)
	(&html.TroubleReports{DB: db}).RegisterRoutes(e)

	// HTML Handler (Migrated)
	html.NewAuth(db, logger.HandlerAuth()).RegisterRoutes(e)
	html.NewHome(db, logger.HandlerHome()).RegisterRoutes(e)
	html.NewTools(db, logger.HandlerTools()).RegisterRoutes(e)

	// HTMX Handler
	(&htmx.Nav{DB: db, WSFeedHandler: wsh}).RegisterRoutes(e)
	(&htmx.Feed{DB: db}).RegisterRoutes(e)
	(&htmx.Profile{DB: db}).RegisterRoutes(e)
	(&htmx.TroubleReports{DB: db}).RegisterRoutes(e)
	(&htmx.Tools{DB: db}).RegisterRoutes(e)
	(&htmx.Cycles{DB: db}).RegisterRoutes(e)
	(&htmx.MetalSheets{DB: db}).RegisterRoutes(e)
}

// NOTE: If i have more then just this on handler i need to change the return type
func startWebSocketHandlers(db *database.DB) *wshandlers.FeedHandler {
	wsfh := wshandlers.NewFeedHandler(db)

	// Start the feed notification manager in a goroutine
	ctx := context.Background()
	go wsfh.Start(ctx)

	// Set the notifier on the feeds for real-time updates
	db.Feeds.SetBroadcaster(wsfh)

	return wsfh
}
