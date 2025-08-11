package router

import (
	"context"
	"embed"

	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/handler"
	"github.com/knackwurstking/pgpress/internal/htmxhandler"
	"github.com/knackwurstking/pgpress/internal/wshandler"
	"github.com/labstack/echo/v4"
)

var (
	//go:embed assets
	assets embed.FS
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
	}

	htmxBase := &htmxhandler.Base{
		DB:               base.DB,
		ServerPathPrefix: base.ServerPathPrefix + "/nav",
	}

	(&htmxhandler.Nav{Base: htmxBase, WSHandler: wsh}).RegisterRoutes(e)

	(&handler.Auth{Base: base}).RegisterRoutes(e)
	(&handler.Home{Base: base}).RegisterRoutes(e)
	(&handler.Feed{Base: base}).RegisterRoutes(e)
	(&handler.Profile{Base: base}).RegisterRoutes(e)
	(&handler.TroubleReports{Base: base}).RegisterRoutes(e)
}

func startWebSocketHandlers(db *database.DB) *wshandler.WSHandlers {
	wsh := &wshandler.WSHandlers{
		Feed: wshandler.NewFeedHandler(db),
	}

	// Start the feed notification manager in a goroutine
	ctx := context.Background()
	go wsh.Feed.Start(ctx)

	// Set the notifier on the feeds for real-time updates
	db.Feeds.SetBroadcaster(wsh.Feed)

	return wsh
}
