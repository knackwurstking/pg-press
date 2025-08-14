package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pgpress/internal/htmxhandler"
	"github.com/knackwurstking/pgpress/internal/templates/pages"
)

type TroubleReports struct {
	*Base
}

func (h *TroubleReports) RegisterRoutes(e *echo.Echo) {
	path := "/trouble-reports"

	e.GET(h.ServerPathPrefix+path, h.handleMainPage)

	htmxTroubleReports := htmxhandler.TroubleReports{Base: h.NewHTMXHandlerBase(path)}
	htmxTroubleReports.RegisterRoutes(e)
}

func (h *TroubleReports) handleMainPage(c echo.Context) error {
	page := pages.TroubleReportsPage()
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render trouble reports page: "+err.Error())
	}
	return nil
}
