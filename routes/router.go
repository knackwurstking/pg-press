// Package routes provides HTTP route handlers and web interface for the pgvis application.
package routes

import (
	"context"
	"embed"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/knackwurstking/pg-vis/routes/handlers/auth"
	"github.com/knackwurstking/pg-vis/routes/handlers/feed"
	"github.com/knackwurstking/pg-vis/routes/handlers/home"
	"github.com/knackwurstking/pg-vis/routes/handlers/nav"
	"github.com/knackwurstking/pg-vis/routes/handlers/profile"
	"github.com/knackwurstking/pg-vis/routes/handlers/troublereports"
	"github.com/knackwurstking/pg-vis/routes/internal/notifications"
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
	DB               *pgvis.DB
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
	authHandler := auth.NewHandler(
		o.DB, o.ServerPathPrefix, templates)

	homeHandler := home.NewHandler(
		o.DB, o.ServerPathPrefix, templates)

	feedHandler := feed.NewHandler(
		o.DB, o.ServerPathPrefix, templates)

	profileHandler := profile.NewHandler(
		o.DB, o.ServerPathPrefix, templates)

	troublereportsHandler := troublereports.NewHandler(
		o.DB, o.ServerPathPrefix, templates)

	navHandler := nav.NewHandler(
		o.DB, o.ServerPathPrefix, templates, feedNotifier)

	// Register routes
	authHandler.RegisterRoutes(e)
	homeHandler.RegisterRoutes(e)
	feedHandler.RegisterRoutes(e)
	profileHandler.RegisterRoutes(e)
	troublereportsHandler.RegisterRoutes(e)
	navHandler.RegisterRoutes(e)
}
