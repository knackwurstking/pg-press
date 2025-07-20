// Package routes provides HTTP route handlers and web interface for the pgvis application.
package routes

import (
	"embed"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/knackwurstking/pg-vis/routes/components/nav"
	"github.com/knackwurstking/pg-vis/routes/handlers/auth"
	"github.com/knackwurstking/pg-vis/routes/handlers/feed"
	"github.com/knackwurstking/pg-vis/routes/handlers/home"
	"github.com/knackwurstking/pg-vis/routes/handlers/profile"
	"github.com/knackwurstking/pg-vis/routes/handlers/troublereports"
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

	// Initialize handlers
	authHandler := auth.NewHandler(o.DB, o.ServerPathPrefix, templates)
	homeHandler := home.NewHandler(o.DB, o.ServerPathPrefix, templates)
	feedHandler := feed.NewHandler(o.DB, o.ServerPathPrefix, templates)

	// Register routes
	authHandler.RegisterRoutes(e)
	homeHandler.RegisterRoutes(e)
	feedHandler.RegisterRoutes(e)

	// Legacy handlers (to be migrated)
	profile.Serve(templates, o.ServerPathPrefix, e, o.DB)
	troublereports.Serve(templates, o.ServerPathPrefix, e, o.DB)
	nav.Serve(templates, o.ServerPathPrefix, e, o.DB)
}
