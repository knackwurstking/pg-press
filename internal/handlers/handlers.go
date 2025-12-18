package handlers

import (
	"github.com/knackwurstking/pg-press/internal/common"
	"github.com/knackwurstking/pg-press/internal/handlers/auth"
	"github.com/knackwurstking/pg-press/internal/handlers/dialogs"
	"github.com/knackwurstking/pg-press/internal/handlers/home"
	"github.com/knackwurstking/pg-press/internal/handlers/profile"
	"github.com/knackwurstking/pg-press/internal/handlers/tool"
	"github.com/knackwurstking/pg-press/internal/handlers/tools"

	"github.com/labstack/echo/v4"
)

func RegisterAll(db *common.DB, e *echo.Echo) {
	registers := []struct {
		handler func(db *common.DB, e *echo.Echo, path string)
		subPath string
	}{
		{handler: home.Register, subPath: ""},
		{handler: auth.Register, subPath: ""},
		{handler: profile.Register, subPath: "/profile"},
		{handler: tools.Register, subPath: "/tools"},
		{handler: dialogs.Register, subPath: "/dialog"},
		{handler: tool.Register, subPath: "/tool"},
	}
	for _, reg := range registers {
		reg.handler(db, e, reg.subPath)
	}

	//help.NewHandler(r).RegisterRoutes(e, "/help")
	//editor.NewHandler(r).RegisterRoutes(e, "/editor")
	//notes.NewHandler(r).RegisterRoutes(e, "/notes")
	//metalsheets.NewHandler(r).RegisterRoutes(e, "/metal-sheets")
	//umbau.NewHandler(r).RegisterRoutes(e, "/umbau")
	//troublereports.NewHandler(r).RegisterRoutes(e, "/trouble-reports")
	//press.NewHandler(r).RegisterRoutes(e, "/press")
	//pressregenerations.NewHandler(r).RegisterRoutes(e, "/press-regeneration")
}
