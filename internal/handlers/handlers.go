package handlers

import (
	"github.com/knackwurstking/pg-press/internal/common"
	"github.com/knackwurstking/pg-press/internal/handlers/auth"
	"github.com/knackwurstking/pg-press/internal/handlers/home"
	"github.com/knackwurstking/pg-press/internal/handlers/profile"

	"github.com/labstack/echo/v4"
)

func RegisterAll(r *common.DB, e *echo.Echo) {
	registers := []struct {
		handler func(e *echo.Echo, path string)
		subPath string
	}{
		{handler: home.NewHandler(r).RegisterRoutes, subPath: ""},
		{handler: auth.NewHandler(r).RegisterRoutes, subPath: ""},
		{handler: profile.NewHandler(r).RegisterRoutes, subPath: "/profile"},
		{handler: tools.NewHandler(r).RegisterRoutes, subPath: "/tools"},
	}
	for _, reg := range registers {
		reg.handler(e, reg.subPath)
	}

	//nav.NewHandler(r, wsFeedHandler).RegisterRoutes(e, "/nav")
	//feed.NewHandler(r).RegisterRoutes(e, "/feed")
	//help.NewHandler(r).RegisterRoutes(e, "/help")
	//editor.NewHandler(r).RegisterRoutes(e, "/editor")
	//notes.NewHandler(r).RegisterRoutes(e, "/notes")
	//metalsheets.NewHandler(r).RegisterRoutes(e, "/metal-sheets")
	//umbau.NewHandler(r).RegisterRoutes(e, "/umbau")
	//troublereports.NewHandler(r).RegisterRoutes(e, "/trouble-reports")
	//tool.NewHandler(r).RegisterRoutes(e, "/tool")
	//press.NewHandler(r).RegisterRoutes(e, "/press")
	//pressregenerations.NewHandler(r).RegisterRoutes(e, "/press-regeneration")
	//dialogs.NewHandler(r).RegisterRoutes(e, "/dialog")
}
