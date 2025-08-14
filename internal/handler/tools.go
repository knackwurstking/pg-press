package handler

import (
	"net/http"

	"github.com/knackwurstking/pgpress/internal/htmxhandler"
	"github.com/knackwurstking/pgpress/internal/templates/pages"
	"github.com/labstack/echo/v4"
)

type Tools struct {
	*Base
}

func (h *Tools) RegisterRoutes(e *echo.Echo) {
	path := "/tools"

	e.GET(h.ServerPathPrefix+path, h.handleToolsPage)

	htmxTroubleReports := htmxhandler.Tools{Base: h.NewHTMXHandlerBase(path)}
	htmxTroubleReports.RegisterRoutes(e)
}

func (h *Tools) handleToolsPage(c echo.Context) error {
	page := pages.ToolsPage()
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render tools page: "+err.Error())
	}
	return nil
}
