package main

import (
	"github.com/knackwurstking/pg-press/internal/assets"
	"github.com/knackwurstking/pg-press/internal/common"
	"github.com/knackwurstking/pg-press/internal/env"
	"github.com/knackwurstking/pg-press/internal/handlers"

	"github.com/labstack/echo/v4"
)

func Serve(e *echo.Echo, r *common.DB) {
	// Static File Server
	e.StaticFS(env.ServerPathPrefix+"/", assets.GetAssets())

	handlers.RegisterAll(r, e)
}
