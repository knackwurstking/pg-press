package router

import (
	"context"
	"embed"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/internal/database"
	"github.com/knackwurstking/pg-vis/internal/handler"
	"github.com/knackwurstking/pg-vis/internal/handler/home"
	"github.com/knackwurstking/pg-vis/internal/handler/nav"
	"github.com/knackwurstking/pg-vis/internal/handler/profile"
	"github.com/knackwurstking/pg-vis/internal/handler/troublereports"
	"github.com/knackwurstking/pg-vis/internal/notifications"
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

	// Initialize feed notification system
	feedNotifier := notifications.NewFeedNotifier(o.DB, templates)

	// Start the feed notification manager in a goroutine
	ctx := context.Background()
	go feedNotifier.Start(ctx)

	// Set the notifier on the feeds for real-time updates
	o.DB.Feeds.SetNotifier(feedNotifier)

	// Initialize handlers
	base := &handler.Base{
		DB:               o.DB,
		ServerPathPrefix: o.ServerPathPrefix,
		Templates:        templates,
	}

	homeHandler := home.NewHandler(o.DB, o.ServerPathPrefix, templates)                     // TODO: ...
	profileHandler := profile.NewHandler(o.DB, o.ServerPathPrefix, templates)               // TODO: ...
	troublereportsHandler := troublereports.NewHandler(o.DB, o.ServerPathPrefix, templates) // TODO: ...
	navHandler := nav.NewHandler(o.DB, o.ServerPathPrefix, templates, feedNotifier)         // TODO: ...

	// Register routes
	handler.NewAuth(base).RegisterRoutes(e)
	homeHandler.RegisterRoutes(e)
	handler.NewFeed(base).RegisterRoutes(e)
	profileHandler.RegisterRoutes(e)
	troublereportsHandler.RegisterRoutes(e)
	navHandler.RegisterRoutes(e)
}
