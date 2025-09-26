package web

import (
	"context"
	"embed"
	"os"

	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/web/features/auth"
	"github.com/knackwurstking/pgpress/internal/web/features/feed"
	"github.com/knackwurstking/pgpress/internal/web/features/home"
	"github.com/knackwurstking/pgpress/internal/web/features/profile"
	"github.com/knackwurstking/pgpress/internal/web/features/tools"
	"github.com/knackwurstking/pgpress/internal/web/features/troublereports"
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

	// HTMX Handler (Old)
	htmx.NewNav(db, wsh).RegisterRoutes(e)
	htmx.NewMetalSheets(db).RegisterRoutes(e)
	htmx.NewCycles(db).RegisterRoutes(e)

	// TODO: New Handlers here
	auth.NewRoutes(db).RegisterRoutes(e)
	feed.NewRoutes(db).RegisterRoutes(e)
	home.NewRoutes(db).RegisterRoutes(e)
	profile.NewRoutes(db).RegisterRoutes(e)
	tools.NewRoutes(db).RegisterRoutes(e)
	troublereports.NewRoutes(db).RegisterRoutes(e)
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
