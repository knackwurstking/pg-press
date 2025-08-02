package router

import (
	"context"
	"embed"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/handler"
	"github.com/knackwurstking/pgpress/internal/htmxhandler"
	"github.com/knackwurstking/pgpress/internal/wshandler"
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

	// Initialize base handler configuration
	base := &handler.Base{
		DB:               o.DB,
		ServerPathPrefix: o.ServerPathPrefix,
		Templates:        templates,
	}

	// Initialize and configure feed notification system
	wsFeedHandler := wshandler.NewFeedHandler(o.DB, templates)
	{
		// Start the feed notification manager in a goroutine
		ctx := context.Background()
		go wsFeedHandler.Start(ctx)

		// Set the notifier on the feeds for real-time updates
		o.DB.Feeds.SetBroadcaster(wsFeedHandler)
	}

	// Register all application routes
	(&handler.Auth{Base: base}).RegisterRoutes(e)
	(&handler.Home{Base: base}).RegisterRoutes(e)
	(&handler.Feed{Base: base}).RegisterRoutes(e)
	(&handler.Profile{Base: base}).RegisterRoutes(e)
	(&handler.TroubleReports{Base: base}).RegisterRoutes(e)

	(&htmxhandler.Nav{
		Base: &htmxhandler.Base{
			DB:               o.DB,
			ServerPathPrefix: o.ServerPathPrefix + "/nav",
			Templates:        templates,
		},
		FeedNotifier: wsFeedHandler,
	}).RegisterRoutes(e)
}
