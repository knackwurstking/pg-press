package main

import (
	pgpress "github.com/knackwurstking/pg-press"
	"github.com/knackwurstking/pg-press/handlers"
	"github.com/knackwurstking/pg-press/services"
	"github.com/labstack/echo/v4"
)

func Serve(e *echo.Echo, r *services.Registry) {
	// Static File Server
	e.StaticFS(serverPathPrefix+"/", pgpress.GetAssets())

	handlers.RegisterAll(r, e)
}
