package handlers

import (
	"context"

	"github.com/knackwurstking/pg-press/internal/common"
	"github.com/knackwurstking/pg-press/internal/handlers/auth"
	"github.com/knackwurstking/pg-press/internal/handlers/dialogs"
	"github.com/knackwurstking/pg-press/internal/handlers/editor"
	"github.com/knackwurstking/pg-press/internal/handlers/feed"
	"github.com/knackwurstking/pg-press/internal/handlers/help"
	"github.com/knackwurstking/pg-press/internal/handlers/home"
	"github.com/knackwurstking/pg-press/internal/handlers/metalsheets"
	"github.com/knackwurstking/pg-press/internal/handlers/nav"
	"github.com/knackwurstking/pg-press/internal/handlers/notes"
	"github.com/knackwurstking/pg-press/internal/handlers/press"
	"github.com/knackwurstking/pg-press/internal/handlers/pressregenerations"
	"github.com/knackwurstking/pg-press/internal/handlers/profile"
	"github.com/knackwurstking/pg-press/internal/handlers/tool"
	"github.com/knackwurstking/pg-press/internal/handlers/tools"
	"github.com/knackwurstking/pg-press/internal/handlers/troublereports"
	"github.com/knackwurstking/pg-press/internal/handlers/umbau"
	"github.com/knackwurstking/pg-press/internal/handlers/wsfeed"

	"github.com/labstack/echo/v4"
)

func RegisterAll(r *common.DB, e *echo.Echo) {
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

func StartWsFeedHandler(r *common.DB) *wsfeed.Handler {
	wsfh := wsfeed.NewHandler(r)

	// Start the feed notification manager in a goroutine
	ctx := context.Background()
	go wsfh.Start(ctx)

	// Set the notifier on the feeds for real-time updates
	r.Feeds.SetBroadcaster(wsfh)

	return wsfh
}
