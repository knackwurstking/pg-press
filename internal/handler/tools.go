package handler

import (
	"net/http"

	"github.com/knackwurstking/pgpress/internal/constants"
	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/htmxhandler"
	"github.com/knackwurstking/pgpress/internal/templates/pages"
	"github.com/knackwurstking/pgpress/internal/utils"
	"github.com/labstack/echo/v4"
)

type Tools struct {
	*Base
}

func (h *Tools) RegisterRoutes(e *echo.Echo) {
	path := "/tools"

	e.GET(h.ServerPathPrefix+path, h.handleToolsPage)

	e.GET(h.ServerPathPrefix+path+"/all/:id", h.handleToolsAllPage)
	e.GET(h.ServerPathPrefix+path+"/active/:press", h.handleToolsActivePage)

	htmxTroubleReports := htmxhandler.Tools{Base: h.NewHTMXHandlerBase(path)}
	htmxTroubleReports.RegisterRoutes(e)
}

func (h *Tools) handleToolsPage(c echo.Context) error {
	tools, err := h.DB.Tools.List()
	if err != nil {
		return echo.NewHTTPError(database.GetHTTPStatusCode(err),
			"failed to get tools: "+err.Error())
	}

	page := pages.ToolsPage(tools)
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render tools page: "+err.Error())
	}
	return nil
}

func (h *Tools) handleToolsAllPage(c echo.Context) error {
	id, err := utils.ParseInt64Param(c, constants.QueryParamID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest,
			"failed to parse id: "+err.Error())
	}

	tool, err := h.DB.Tools.Get(id)
	if err != nil {
		return echo.NewHTTPError(database.GetHTTPStatusCode(err),
			"failed to get tool: "+err.Error())
	}

	page := pages.ToolsAllPage(tool)
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render tools all page: "+err.Error())
	}
	return nil
}

func (h *Tools) handleToolsActivePage(c echo.Context) error {
	press, err := utils.ParseInt64Param(c, constants.QueryParamPress)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest,
			"failed to parse id: "+err.Error())
	}

	page := pages.ToolsActivePage(press)
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render tools active page: "+err.Error())
	}
	return nil
}
