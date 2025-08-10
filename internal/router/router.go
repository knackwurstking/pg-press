package router

import (
	"context"

	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/handler"
	"github.com/knackwurstking/pgpress/internal/htmxhandler"
	"github.com/knackwurstking/pgpress/internal/wshandler"
	"github.com/labstack/echo/v4"
)

type Options struct {
	ServerPathPrefix string
	DB               *database.DB
}

func Serve(e *echo.Echo, o Options) {
	e.StaticFS(o.ServerPathPrefix+"/", echo.MustSubFS(assets, "assets"))

	wsh := startWebSocketHandlers(o.DB)

	base := &handler.Base{
		DB:               o.DB,
		ServerPathPrefix: o.ServerPathPrefix,
		Templates:        templates,
	}

	htmxBase := &htmxhandler.Base{
		DB:               base.DB,
		ServerPathPrefix: base.ServerPathPrefix + "/nav",
		Templates:        base.Templates,
	}

	(&htmxhandler.Nav{Base: htmxBase, WSHandler: wsh}).RegisterRoutes(e)

	(&handler.Auth{Base: base}).RegisterRoutes(e)
	(&handler.Home{Base: base}).RegisterRoutes(e)
	(&handler.Feed{Base: base}).RegisterRoutes(e)
	// TODO: Continue with the profile page handler
}

func startWebSocketHandlers(db *database.DB) *wshandler.WSHandlers {
	wsh := &wshandler.WSHandlers{
		Feed: wshandler.NewFeedHandler(db, templates),
	}

	// Start the feed notification manager in a goroutine
	ctx := context.Background()
	go wsh.Feed.Start(ctx)

	// Set the notifier on the feeds for real-time updates
	db.Feeds.SetBroadcaster(wsh.Feed)

	return wsh
}
