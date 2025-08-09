package router

import (
	"context"

	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/handler"
	"github.com/knackwurstking/pgpress/internal/wshandler"
	"github.com/labstack/echo/v4"
)

type Options struct {
	ServerPathPrefix string
	DB               *database.DB
}

func Serve(e *echo.Echo, o Options) {
	e.StaticFS(o.ServerPathPrefix+"/", echo.MustSubFS(assets, "assets"))

	startWebSocketHandlers(o.DB)

	base := &handler.Base{
		DB:               o.DB,
		ServerPathPrefix: o.ServerPathPrefix,
		Templates:        templates,
	}

	(&handler.Auth{Base: base}).RegisterRoutes(e)
	// TODO: Continue with the home page
}

type wsHandlers struct {
	feed *wshandler.FeedHandler
}

func startWebSocketHandlers(db *database.DB) {
	wsh := wsHandlers{
		// TOOD: The templates fs needs to be replaced with templ components
		feed: wshandler.NewFeedHandler(db, templates),
	}

	// Start the feed notification manager in a goroutine
	ctx := context.Background()
	go wsh.feed.Start(ctx)

	// Set the notifier on the feeds for real-time updates
	db.Feeds.SetBroadcaster(wsh.feed)
}
