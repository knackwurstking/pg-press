package router

import (
	"context"
	"embed"
	"os"

	"github.com/knackwurstking/pgpress/internal/database"
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

	// HTML Handler
	html.NewAuth(db).RegisterRoutes(e)
	html.NewFeed(db).RegisterRoutes(e)
	html.NewHome(db).RegisterRoutes(e)
	html.NewProfile(db).RegisterRoutes(e)
	html.NewTroubleReports(db).RegisterRoutes(e)
	html.NewTools(db).RegisterRoutes(e)

	// HTMX Handler (Old)
	(&htmx.TroubleReports{DB: db}).RegisterRoutes(e) // TODO: Migrate
	(&htmx.Cycles{DB: db}).RegisterRoutes(e)         // TODO: Migrate

	// HTMX Handler (Migrated)
	htmx.NewNav(db, wsh).RegisterRoutes(e)
	htmx.NewFeed(db).RegisterRoutes(e)
	htmx.NewProfile(db).RegisterRoutes(e)
	htmx.NewTools(db).RegisterRoutes(e) // TODO: Migrate
	htmx.NewMetalSheets(db).RegisterRoutes(e)
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
