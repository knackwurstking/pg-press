package router

import (
	"context"
	"embed"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/internal/database"
	"github.com/knackwurstking/pg-vis/internal/handler"
	"github.com/knackwurstking/pg-vis/internal/wshandler"
)

var (
	//go:embed templates
	templates embed.FS

	//go:embed assets
	assets embed.FS
)

// Options contains configuration for the routes package.
type Options struct {
	ServerPathPrefix string
	DB               *database.DB
}

// Serve configures and registers all HTTP routes for the application.
func Serve(e *echo.Echo, o Options) {
	// Serve static assets
	e.StaticFS(o.ServerPathPrefix+"/", echo.MustSubFS(assets, "assets"))

	// Initialize handlers
	base := &handler.Base{
		DB:               o.DB,
		ServerPathPrefix: o.ServerPathPrefix,
		Templates:        templates,
	}

	// Initialize feed notification system
	wsFeedHandler := wshandler.NewFeedHandler(o.DB, templates)

	{ // Initialize feed notification system
		// Start the feed notification manager in a goroutine
		ctx := context.Background()
		go wsFeedHandler.Start(ctx)

		// Set the notifier on the feeds for real-time updates
		o.DB.Feeds.SetBroadcaster(wsFeedHandler)
	}

	// Register routes
	handler.NewAuth(base).RegisterRoutes(e)
	handler.NewHome(base).RegisterRoutes(e)
	handler.NewFeed(base).RegisterRoutes(e)
	handler.NewProfile(base).RegisterRoutes(e)
	handler.NewTroubleReports(base).RegisterRoutes(e)
	handler.NewNav(base, wsFeedHandler).RegisterRoutes(e)
}
