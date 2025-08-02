package handler

import (
	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pgpress/internal/constants"
	"github.com/knackwurstking/pgpress/internal/htmxhandler"
	"github.com/knackwurstking/pgpress/internal/utils"
)

type TroubleReports struct {
	*Base
}

func (h *TroubleReports) RegisterRoutes(e *echo.Echo) {
	path := "/trouble-reports"

	e.GET(h.ServerPathPrefix+path, h.handleMainPage)

	htmxTroubleReports := htmxhandler.TroubleReports{Base: h.NewHTMX(path)}
	htmxTroubleReports.RegisterRoutes(e)
}

func (h *TroubleReports) handleMainPage(c echo.Context) error {
	return utils.HandleTemplate(c, nil,
		h.Templates,
		constants.TroubleReportsPageTemplates,
	)
}
