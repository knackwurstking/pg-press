package html

import (
	"net/http"

	database "github.com/knackwurstking/pgpress/internal/database/core"
	"github.com/knackwurstking/pgpress/internal/database/dberror"
	metalsheetmodels "github.com/knackwurstking/pgpress/internal/database/models/metalsheet"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/web/constants"
	webhelpers "github.com/knackwurstking/pgpress/internal/web/helpers"
	toolspage "github.com/knackwurstking/pgpress/internal/web/templates/pages/tools"
	presspage "github.com/knackwurstking/pgpress/internal/web/templates/pages/tools/press"
	toolpage "github.com/knackwurstking/pgpress/internal/web/templates/pages/tools/tool"

	"github.com/labstack/echo/v4"
)

type Tools struct {
	DB *database.DB
}

func (h *Tools) RegisterRoutes(e *echo.Echo) {
	webhelpers.RegisterEchoRoutes(
		e,
		[]*webhelpers.EchoRoute{
			webhelpers.NewEchoRoute(http.MethodGet, "/tools", h.handleTools),
			webhelpers.NewEchoRoute(http.MethodGet, "/tools/press/:press", h.handlePressPage),
			webhelpers.NewEchoRoute(http.MethodGet, "/tools/tool/:id", h.handleToolPage),
		},
	)
}

func (h *Tools) handleTools(c echo.Context) error {
	logger.HandlerTools().Debug("Rendering tools page")

	tools, err := h.DB.ToolsHelper.ListWithNotes()
	if err != nil {
		logger.HandlerTools().Error("Failed to fetch tools: %v", err)
		return echo.NewHTTPError(dberror.GetHTTPStatusCode(err),
			"failed to get tools: "+err.Error())
	}

	logger.HandlerTools().Debug("Retrieved %d tools", len(tools))

	page := toolspage.Page(tools)
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		logger.HandlerTools().Error("Failed to render tools page: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render tools page: "+err.Error())
	}
	return nil
}

func (h *Tools) handlePressPage(c echo.Context) error {
	press, err := webhelpers.ParseInt64Param(c, constants.QueryParamPress)
	if err != nil {
		logger.HandlerTools().Error("Failed to parse press parameter: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest,
			"failed to parse id: "+err.Error())
	}

	logger.HandlerTools().Debug("Rendering tools active page for press %d", press)

	page := presspage.Page(press)
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		logger.HandlerTools().Error("Failed to render tools active page: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render tools active page: "+err.Error())
	}
	return nil
}

func (h *Tools) handleToolPage(c echo.Context) error {
	user, err := webhelpers.GetUserFromContext(c)
	if err != nil {
		return err
	}

	id, err := webhelpers.ParseInt64Param(c, constants.QueryParamID)
	if err != nil {
		logger.HandlerTools().Error("Failed to parse tool id parameter: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest,
			"failed to parse id: "+err.Error())
	}

	logger.HandlerTools().Debug("Fetching tool %d with notes", id)

	tool, err := h.DB.ToolsHelper.GetWithNotes(id)
	if err != nil {
		logger.HandlerTools().Error("Failed to fetch tool %d: %v", id, err)
		return echo.NewHTTPError(dberror.GetHTTPStatusCode(err),
			"failed to get tool: "+err.Error())
	}

	logger.HandlerTools().Debug("Successfully fetched tool %d: Type=%s, Code=%s", id, tool.Type, tool.Code)

	// Fetch metal sheets assigned to this tool
	metalSheets, err := h.DB.MetalSheets.GetByToolID(id)
	if err != nil {
		// Log error but don't fail - metal sheets are supplementary data
		logger.HandlerTools().Error("Failed to fetch metal sheets: %v", err)
		metalSheets = []*metalsheetmodels.MetalSheet{}
	}

	logger.HandlerTools().Debug("Rendering tool page for tool %d with %d metal sheets", id, len(metalSheets))

	page := toolpage.Page(user, tool, metalSheets)
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		logger.HandlerTools().Error("Failed to render tool page: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render tools all page: "+err.Error())
	}
	return nil
}
