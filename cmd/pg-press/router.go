package main

import (
	"context"

	"github.com/knackwurstking/pgpress"
	"github.com/knackwurstking/pgpress/handlers"
	"github.com/knackwurstking/pgpress/services"
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

	//editor.NewRoutes(db).RegisterRoutes(e)
	//help.NewRoutes(db).RegisterRoutes(e) // TODO: Continue here
	//profile.NewRoutes(db).RegisterRoutes(e)
	//tools.NewRoutes(db).RegisterRoutes(e)
	//tool.NewRoutes(db).RegisterRoutes(e)
	//press.NewRoutes(db).RegisterRoutes(e)
	//umbau.NewRoutes(db).RegisterRoutes(e)
	//troublereports.NewRoutes(db).RegisterRoutes(e)
	//nav.NewRoutes(db, wsFeedHandler).RegisterRoutes(e)
	//notes.NewRoutes(db).RegisterRoutes(e)
	//metalsheets.NewRoutes(db).RegisterRoutes(e)
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
