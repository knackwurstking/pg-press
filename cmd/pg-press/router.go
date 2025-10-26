package main

import (
	"context"

	pgpress "github.com/knackwurstking/pg-press"
	"github.com/knackwurstking/pg-press/handlers"
	"github.com/knackwurstking/pg-press/services"
	"github.com/labstack/echo/v4"
)

func Serve(e *echo.Echo, r *services.Registry) {
	// Static File Server
	e.StaticFS(serverPathPrefix+"/", pgpress.GetAssets())

	// WebSocket Handlers
	wsFeedHandler := startWsFeedHandler(r)

	handlers.NewNav(r, wsFeedHandler).RegisterRoutes(e)
	handlers.NewHome(r).RegisterRoutes(e)
	handlers.NewAuth(r).RegisterRoutes(e)
	handlers.NewFeed(r).RegisterRoutes(e)
	handlers.NewHelp(r).RegisterRoutes(e)
	handlers.NewEditor(r).RegisterRoutes(e)
	handlers.NewProfile(r).RegisterRoutes(e)
	handlers.NewNotes(r).RegisterRoutes(e)
	handlers.NewMetalSheets(r).RegisterRoutes(e)
	handlers.NewUmbau(r).RegisterRoutes(e)
	handlers.NewTroubleReports(r).RegisterRoutes(e)
	handlers.NewTools(r).RegisterRoutes(e)
	handlers.NewTool(r).RegisterRoutes(e)
	handlers.NewPress(r).RegisterRoutes(e)
}

func startWsFeedHandler(r *services.Registry) *handlers.FeedHandler {
	wsfh := handlers.NewFeedHandler(r)

	// Start the feed notification manager in a goroutine
	ctx := context.Background()
	go wsfh.Start(ctx)

	// Set the notifier on the feeds for real-time updates
	r.Feeds.SetBroadcaster(wsfh)

	return wsfh
}
