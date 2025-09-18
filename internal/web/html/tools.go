package html

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/web/handlers"
	"github.com/knackwurstking/pgpress/internal/web/helpers"
	"github.com/knackwurstking/pgpress/internal/web/templates/toolspage"
	"github.com/knackwurstking/pgpress/internal/web/templates/toolspage/presspage"
	"github.com/knackwurstking/pgpress/internal/web/templates/toolspage/toolpage"

	"github.com/knackwurstking/pgpress/pkg/logger"
	"github.com/knackwurstking/pgpress/pkg/models"

	"github.com/labstack/echo/v4"
)

type Tools struct {
	*handlers.BaseHandler
}

func NewTools(db *database.DB, logger *logger.Logger) *Tools {
	return &Tools{
		BaseHandler: handlers.NewBaseHandler(db, logger),
	}
}

func (h *Tools) RegisterRoutes(e *echo.Echo) {
	helpers.RegisterEchoRoutes(
		e,
		[]*helpers.EchoRoute{
			helpers.NewEchoRoute(http.MethodGet, "/tools",
				h.HandleTools),

			helpers.NewEchoRoute(http.MethodGet, "/tools/press/:press",
				h.HandlePressPage),
			helpers.NewEchoRoute(http.MethodGet, "/tools/press/:press/umbau",
				h.HandlePressUmbauPage),

			helpers.NewEchoRoute(http.MethodGet, "/tools/tool/:id",
				h.HandleToolPage),
		},
	)
}

func (h *Tools) HandleTools(c echo.Context) error {
	h.LogInfo("Rendering tools page")

	tools, err := h.DB.Tools.ListWithNotes()
	if err != nil {
		return h.HandleError(c, err, "failed to get tools")
	}

	h.LogDebug("Retrieved %d tools", len(tools))

	pressUtilization, err := h.DB.Tools.GetPressUtilization()
	if err != nil {
		return h.HandleError(c, err, "failed to get press utilization")
	}

	page := toolspage.Page(&toolspage.PageProps{
		Tools:            tools,
		PressUtilization: pressUtilization,
	})

	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c, "failed to render tools page: "+err.Error())
	}

	return nil
}

func (h *Tools) HandlePressPage(c echo.Context) error {
	// Get user from context
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	// Get press number from param
	var pn models.PressNumber
	pns, err := h.ParseInt64Param(c, "press")
	if err != nil {
		return h.RenderBadRequest(c, "failed to parse id: "+err.Error())
	}
	pn = models.PressNumber(pns)
	if !models.IsValidPressNumber(&pn) {
		return h.RenderBadRequest(c, fmt.Sprintf("invalid press number: %d", pn))
	}

	// Get cycles for this press
	cycles, err := h.DB.PressCycles.GetPressCycles(pn, nil, nil)
	if err != nil {
		return h.HandleError(c, err, "failed to get press cycles")
	}

	// Get tools
	tools, err := h.DB.Tools.List()
	if err != nil {
		return h.HandleError(c, err, "failed to get tools map")
	}
	// Convert tools to map[int64]*Tool
	toolsMap := make(map[int64]*models.Tool)
	for _, tool := range tools {
		toolsMap[tool.ID] = tool
	}

	// Render page
	h.LogDebug("Rendering page for press %d", pn)
	page := presspage.Page(presspage.PageProps{
		Press:    pn,
		Cycles:   cycles,
		User:     user,
		ToolsMap: toolsMap,
	})

	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c, "failed to render press page: "+err.Error())
	}

	return nil
}

func (h *Tools) HandlePressUmbauPage(c echo.Context) error {
	// Get user from context
	_, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	// Get press number from param
	var pn models.PressNumber
	pns, err := h.ParseInt64Param(c, "press")
	if err != nil {
		return h.RenderBadRequest(c, "failed to parse id: "+err.Error())
	}
	pn = models.PressNumber(pns)
	if !models.IsValidPressNumber(&pn) {
		return h.RenderBadRequest(c, "invalid press number")
	}

	// TODO: Implement press umbau page logic

	return errors.New("under construction")
}

func (h *Tools) HandleToolPage(c echo.Context) error {
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	id, err := h.ParseInt64Param(c, "id")
	if err != nil {
		return h.RenderBadRequest(c,
			"failed to parse id from query parameter:"+err.Error())
	}

	h.LogDebug("Fetching tool %d with notes", id)

	tool, err := h.DB.Tools.GetWithNotes(id)
	if err != nil {
		return h.HandleError(c, err, "failed to get tool")
	}

	h.LogDebug("Successfully fetched tool %d: Type=%s, Code=%s",
		id, tool.Type, tool.Code)

	// Fetch metal sheets assigned to this tool
	metalSheets, err := h.DB.MetalSheets.GetByToolID(id)
	if err != nil {
		// Log error but don't fail - metal sheets are supplementary data
		h.LogError("Failed to fetch metal sheets: %v", err)
		metalSheets = []*models.MetalSheet{}
	}

	h.LogDebug("Rendering tool page for tool %d with %d metal sheets", id, len(metalSheets))

	page := toolpage.Page(user, tool, metalSheets)
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c, "failed to render tool page: "+err.Error())
	}

	return nil
}
