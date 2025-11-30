package handlers

import (
	"context"

	"github.com/knackwurstking/pg-press/handlers/auth"
	"github.com/knackwurstking/pg-press/handlers/dialogs"
	"github.com/knackwurstking/pg-press/handlers/editor"
	"github.com/knackwurstking/pg-press/handlers/feed"
	"github.com/knackwurstking/pg-press/handlers/help"
	"github.com/knackwurstking/pg-press/handlers/home"
	"github.com/knackwurstking/pg-press/handlers/metalsheets"
	"github.com/knackwurstking/pg-press/handlers/nav"
	"github.com/knackwurstking/pg-press/handlers/notes"
	"github.com/knackwurstking/pg-press/handlers/press"
	"github.com/knackwurstking/pg-press/handlers/pressregenerations"

	"github.com/knackwurstking/pg-press/handlers/profile"
	"github.com/knackwurstking/pg-press/handlers/tool"
	"github.com/knackwurstking/pg-press/handlers/tools"
	"github.com/knackwurstking/pg-press/handlers/troublereports"
	"github.com/knackwurstking/pg-press/handlers/umbau"
	"github.com/knackwurstking/pg-press/handlers/wsfeed"
	"github.com/knackwurstking/pg-press/services"
	"github.com/labstack/echo/v4"
)

func RegisterAll(r *services.Registry, e *echo.Echo) {
	// WebSocket Handlers
	wsFeedHandler := StartWsFeedHandler(r)

	nav.NewHandler(r, wsFeedHandler).RegisterRoutes(e, "/nav")
	home.NewHandler(r).RegisterRoutes(e, "")
	auth.NewHandler(r).RegisterRoutes(e, "")
	feed.NewHandler(r).RegisterRoutes(e, "/feed")
	help.NewHandler(r).RegisterRoutes(e, "/help")
	editor.NewHandler(r).RegisterRoutes(e, "/editor")
	profile.NewHandler(r).RegisterRoutes(e, "/profile")
	notes.NewHandler(r).RegisterRoutes(e, "/notes")
	metalsheets.NewHandler(r).RegisterRoutes(e, "/metal-sheets")
	umbau.NewHandler(r).RegisterRoutes(e, "/umbau")
	troublereports.NewHandler(r).RegisterRoutes(e, "/trouble-reports")
	tools.NewHandler(r).RegisterRoutes(e, "/tools")
	tool.NewHandler(r).RegisterRoutes(e, "/tool")
	press.NewHandler(r).RegisterRoutes(e, "/press")
	pressregenerations.NewHandler(r).RegisterRoutes(e, "/press-regeneration")

	dialogs.NewHandler(r).RegisterRoutes(e, "/dialog")
}

func StartWsFeedHandler(r *services.Registry) *wsfeed.Handler {
	wsfh := wsfeed.NewHandler(r)

	// Start the feed notification manager in a goroutine
	ctx := context.Background()
	go wsfh.Start(ctx)

	// Set the notifier on the feeds for real-time updates
	r.Feeds.SetBroadcaster(wsfh)

	return wsfh
}
