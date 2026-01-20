package handlers

import (
	"github.com/knackwurstking/pg-press/internal/handlers/auth"
	"github.com/knackwurstking/pg-press/internal/handlers/dialogs"
	"github.com/knackwurstking/pg-press/internal/handlers/editor"
	"github.com/knackwurstking/pg-press/internal/handlers/help"
	"github.com/knackwurstking/pg-press/internal/handlers/home"
	"github.com/knackwurstking/pg-press/internal/handlers/metalsheets"
	"github.com/knackwurstking/pg-press/internal/handlers/notes"
	"github.com/knackwurstking/pg-press/internal/handlers/press"
	"github.com/knackwurstking/pg-press/internal/handlers/pressregenerations"
	"github.com/knackwurstking/pg-press/internal/handlers/profile"
	"github.com/knackwurstking/pg-press/internal/handlers/tool"
	"github.com/knackwurstking/pg-press/internal/handlers/tools"
	"github.com/knackwurstking/pg-press/internal/handlers/troublereports"
	"github.com/knackwurstking/pg-press/internal/handlers/umbau"

	"github.com/labstack/echo/v4"
)

func RegisterAll(e *echo.Echo) {
	registers := []struct {
		handler func(e *echo.Echo, path string)
		subPath string
	}{
		{handler: home.Register, subPath: ""},
		{handler: auth.Register, subPath: ""},
		{handler: profile.Register, subPath: "/profile"},
		{handler: tools.Register, subPath: "/tools"},
		{handler: dialogs.Register, subPath: "/dialog"},
		{handler: tool.Register, subPath: "/tool"},
		{handler: notes.Register, subPath: "/notes"},
		{handler: press.Register, subPath: "/press"},
		{handler: umbau.Register, subPath: "/umbau"},
		{handler: metalsheets.Register, subPath: "/metal-sheets"},
		{handler: troublereports.Register, subPath: "/trouble-reports"},
		{handler: editor.Register, subPath: "/editor"},
		{handler: help.Register, subPath: "/help"},
		{handler: pressregenerations.Register, subPath: "/press-regeneration"},
	}
	for _, reg := range registers {
		reg.handler(e, reg.subPath)
	}
}
