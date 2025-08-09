package router

import (
	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/handler"
	"github.com/labstack/echo/v4"
)

type Options struct {
	ServerPathPrefix string
	DB               *database.DB
}

func Serve(e *echo.Echo, o Options) {
	e.StaticFS(o.ServerPathPrefix+"/", echo.MustSubFS(assets, "assets"))

	base := &handler.Base{
		DB:               o.DB,
		ServerPathPrefix: o.ServerPathPrefix,
		Templates:        templates,
	}

	(&handler.Auth{Base: base}).RegisterRoutes(e)
}
