package handler

import (
	"net/http"

	"github.com/knackwurstking/pgpress/internal/constants"
	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/htmxhandler"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/templates/pages"
	"github.com/knackwurstking/pgpress/internal/utils"
	"github.com/labstack/echo/v4"
)

type Tools struct {
	DB *database.DB
}

func (h *Tools) RegisterRoutes(e *echo.Echo) {
	path := "/tools"

	e.GET(serverPathPrefix+path, h.handleToolsPage)
	e.GET(serverPathPrefix+path+"/", h.handleToolsPage)

	e.GET(serverPathPrefix+path+"/active/:press", h.handleToolsActivePage)
	e.GET(serverPathPrefix+path+"/active/:press/", h.handleToolsActivePage)
	e.GET(serverPathPrefix+path+"/all/:id", h.handleToolPage)
	e.GET(serverPathPrefix+path+"/all/:id/", h.handleToolPage)

	htmxTroubleReports := htmxhandler.Tools{DB: h.DB}
	htmxTroubleReports.RegisterRoutes(e)
}

func (h *Tools) handleToolsPage(c echo.Context) error {
	logger.HandlerTools().Debug("Rendering tools page")

	tools, err := h.DB.ToolsHelper.ListWithNotes()
	if err != nil {
		logger.HandlerTools().Error("Failed to fetch tools: %v", err)
		return echo.NewHTTPError(database.GetHTTPStatusCode(err),
			"failed to get tools: "+err.Error())
	}

	logger.HandlerTools().Debug("Retrieved %d tools", len(tools))

	page := pages.ToolsPage(tools)
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		logger.HandlerTools().Error("Failed to render tools page: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render tools page: "+err.Error())
	}
	return nil
}

func (h *Tools) handleToolsActivePage(c echo.Context) error {
	press, err := utils.ParseInt64Param(c, constants.QueryParamPress)
	if err != nil {
		logger.HandlerTools().Error("Failed to parse press parameter: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest,
			"failed to parse id: "+err.Error())
	}

	logger.HandlerTools().Debug("Rendering tools active page for press %d", press)

	page := pages.ToolsActivePage(press)
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		logger.HandlerTools().Error("Failed to render tools active page: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render tools active page: "+err.Error())
	}
	return nil
}

func (h *Tools) handleToolPage(c echo.Context) error {
	id, err := utils.ParseInt64Param(c, constants.QueryParamID)
	if err != nil {
		logger.HandlerTools().Error("Failed to parse tool id parameter: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest,
			"failed to parse id: "+err.Error())
	}

	logger.HandlerTools().Debug("Fetching tool %d with notes", id)

	tool, err := h.DB.ToolsHelper.GetWithNotes(id)
	if err != nil {
		logger.HandlerTools().Error("Failed to fetch tool %d: %v", id, err)
		return echo.NewHTTPError(database.GetHTTPStatusCode(err),
			"failed to get tool: "+err.Error())
	}

	logger.HandlerTools().Debug("Successfully fetched tool %d: Type=%s, Code=%s", id, tool.Type, tool.Code)

	// Fetch metal sheets assigned to this tool
	metalSheets, err := h.DB.MetalSheets.GetByToolID(id)
	if err != nil {
		// Log error but don't fail - metal sheets are supplementary data
		logger.HandlerTools().Error("Failed to fetch metal sheets: " + err.Error())
		metalSheets = []*database.MetalSheet{}
	}

	logger.HandlerTools().Debug("Rendering tool page for tool %d with %d metal sheets", id, len(metalSheets))

	page := pages.ToolPage(tool, metalSheets)
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		logger.HandlerTools().Error("Failed to render tool page: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render tools all page: "+err.Error())
	}
	return nil
}
