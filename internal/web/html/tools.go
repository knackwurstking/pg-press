package html

import (
	"errors"
	"net/http"

	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/web/helpers"
	"github.com/knackwurstking/pgpress/internal/web/templates/toolspage"
	"github.com/knackwurstking/pgpress/internal/web/templates/toolspage/presspage"
	"github.com/knackwurstking/pgpress/internal/web/templates/toolspage/toolpage"
	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"

	"github.com/labstack/echo/v4"
)

type Tools struct {
	DB *database.DB
}

func (h *Tools) RegisterRoutes(e *echo.Echo) {
	helpers.RegisterEchoRoutes(
		e,
		[]*helpers.EchoRoute{
			helpers.NewEchoRoute(http.MethodGet, "/tools", h.handleTools),
			helpers.NewEchoRoute(http.MethodGet, "/tools/press/:press", h.handlePressPage),
			helpers.NewEchoRoute(http.MethodGet, "/tools/press/:press/umbau", h.handlePressUmbauPage),
			helpers.NewEchoRoute(http.MethodGet, "/tools/tool/:id", h.handleToolPage),
		},
	)
}

func (h *Tools) handleTools(c echo.Context) error {
	logger.HandlerTools().Info("Rendering tools page")

	tools, err := h.DB.Tools.ListWithNotes()
	if err != nil {
		return echo.NewHTTPError(utils.GetHTTPStatusCode(err),
			"failed to get tools: "+err.Error())
	}

	logger.HandlerTools().Debug("Retrieved %d tools", len(tools))

	pressUtilization, err := h.DB.Tools.GetPressUtilization()
	if err != nil {
		return echo.NewHTTPError(utils.GetHTTPStatusCode(err),
			"failed to get press utilization: "+err.Error())
	}

	page := toolspage.Page(&toolspage.PageProps{
		Tools:            tools,
		PressUtilization: pressUtilization,
	})

	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render tools page: "+err.Error())
	}

	return nil
}

func (h *Tools) handlePressPage(c echo.Context) error {
	// Get user from context
	user, err := helpers.GetUserFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to get user from context: "+err.Error())
	}

	// Get press number from param
	var pn models.PressNumber
	pns, err := helpers.ParseInt64Param(c, "press")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest,
			"failed to parse id: "+err.Error())
	}
	pn = models.PressNumber(pns)
	if !models.IsValidPressNumber(&pn) {
		return echo.NewHTTPError(http.StatusBadRequest,
			"invalid press number")
	}

	// Get cycles for this press
	cycles, err := h.DB.PressCycles.GetPressCycles(pn, nil, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to get press cycles: "+err.Error())
	}

	// Get tools
	tools, err := h.DB.Tools.List()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to get tools map: "+err.Error())
	}
	// Convert tools to map[int64]*Tool
	toolsMap := make(map[int64]*models.Tool)
	for _, tool := range tools {
		toolsMap[tool.ID] = tool
	}

	// Render page
	logger.HandlerTools().Debug("Rendering page for press %d", pn)
	page := presspage.Page(presspage.PageProps{
		Press:    pn,
		Cycles:   cycles,
		User:     user,
		ToolsMap: toolsMap,
	})

	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render press page: "+err.Error())
	}

	return nil
}

func (h *Tools) handlePressUmbauPage(c echo.Context) error {
	// Get user from context
	user, err := helpers.GetUserFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to get user from context: "+err.Error())
	}

	// Get press number from param
	var pn models.PressNumber
	pns, err := helpers.ParseInt64Param(c, "press")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest,
			"failed to parse id: "+err.Error())
	}
	pn = models.PressNumber(pns)
	if !models.IsValidPressNumber(&pn) {
		return echo.NewHTTPError(http.StatusBadRequest,
			"invalid press number")
	}

	// TODO: Implement press umbau page logic

	return errors.New("under construction")
}

func (h *Tools) handleToolPage(c echo.Context) error {
	user, err := helpers.GetUserFromContext(c)
	if err != nil {
		return err
	}

	id, err := helpers.ParseInt64Param(c, "id")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest,
			"failed to parse id from query parameter: "+err.Error())
	}

	logger.HandlerTools().Debug("Fetching tool %d with notes", id)
	tool, err := h.DB.Tools.GetWithNotes(id)
	if err != nil {
		return echo.NewHTTPError(utils.GetHTTPStatusCode(err),
			"failed to get tool: "+err.Error())
	}

	logger.HandlerTools().Debug("Successfully fetched tool %d: Type=%s, Code=%s", id, tool.Type, tool.Code)

	// Fetch metal sheets assigned to this tool
	metalSheets, err := h.DB.MetalSheets.GetByToolID(id)
	if err != nil {
		// Log error but don't fail - metal sheets are supplementary data
		logger.HandlerTools().Error("Failed to fetch metal sheets: %v", err)
		metalSheets = []*models.MetalSheet{}
	}

	logger.HandlerTools().Debug("Rendering tool page for tool %d with %d metal sheets", id, len(metalSheets))

	page := toolpage.Page(user, tool, metalSheets)
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render tool page: "+err.Error())
	}

	return nil
}
