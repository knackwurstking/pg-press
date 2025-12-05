package umbau

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/knackwurstking/pg-press/env"
	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/handlers/umbau/templates"
	"github.com/knackwurstking/pg-press/models"
	"github.com/knackwurstking/pg-press/services"
	"github.com/knackwurstking/pg-press/utils"
	ui "github.com/knackwurstking/ui/ui-templ"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	registry *services.Registry
}

func NewHandler(r *services.Registry) *Handler {
	return &Handler{
		registry: r,
	}
}

func (h *Handler) RegisterRoutes(e *echo.Echo, path string) {
	ui.RegisterEchoRoutes(
		e,
		env.ServerPathPrefix,
		[]*ui.EchoRoute{
			ui.NewEchoRoute(http.MethodGet, path+"/:press", h.GetUmbauPage),
			ui.NewEchoRoute(http.MethodPost, path+"/:press", h.PostUmbauPage),
		},
	)
}

func (h *Handler) GetUmbauPage(c echo.Context) error {
	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	pressNumberParam, err := utils.ParseParamInt8(c, "press")
	if err != nil {
		return errors.NewBadRequestError(err, "parse press number")
	}

	pressNumber := models.PressNumber(pressNumberParam)
	if !models.IsValidPressNumber(&pressNumber) {
		return errors.NewBadRequestError(nil, "invalid press number: %d", pressNumber)
	}

	slog.Info("Rendering the umbau page", "user_name", user.Name)

	tools, dberr := h.registry.Tools.List()
	if dberr != nil {
		return errors.HandlerError(dberr, "list tools")
	}

	umbauPage := templates.Page(&templates.PageProps{
		PressNumber: pressNumber,
		User:        user,
		Tools:       tools,
	})

	err = umbauPage.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "UmbauPage")
	}

	return nil
}

func (h *Handler) PostUmbauPage(c echo.Context) error {
	slog.Info("Change the active tools")

	// Get the user from the request context
	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	// Get the press number from the request context
	pressNumberParam, err := utils.ParseParamInt8(c, "press")
	if err != nil {
		return errors.NewBadRequestError(err, "parse press number")
	}

	// Validate the press number
	pressNumber := models.PressNumber(pressNumberParam)
	if !models.IsValidPressNumber(&pressNumber) {
		return errors.NewBadRequestError(nil, "invalid press number: %d", pressNumber)
	}

	// Get form value for the press cycles
	totalCyclesStr := c.FormValue("press-total-cycles")
	if totalCyclesStr == "" {
		return errors.NewBadRequestError(nil, "missing total cycles")
	}

	// Parse this press cycles to int64
	totalCycles, err := strconv.ParseInt(totalCyclesStr, 10, 64)
	if err != nil {
		return errors.NewBadRequestError(err, "invalid total cycles")
	}

	// Get form value for the top tool
	var topToolID models.ToolID
	if id, err := strconv.ParseInt(c.FormValue("top"), 10, 64); err != nil {
		return errors.NewBadRequestError(nil, "missing top tool")
	} else {
		topToolID = models.ToolID(id)
	}

	// Get form value for the bottom tool
	var bottomToolID models.ToolID
	if id, err := strconv.ParseInt(c.FormValue("bottom"), 10, 64); err != nil {
		return errors.NewBadRequestError(nil, "missing bottom tool")
	} else {
		bottomToolID = models.ToolID(id)
	}

	// Get a list with all tools
	tools, dberr := h.registry.Tools.List()
	if dberr != nil {
		return errors.HandlerError(dberr, "get tools")
	}

	// Get the top tool
	topTool, err := h.findToolByID(tools, topToolID)
	if err != nil {
		return errors.NewBadRequestError(err, "invalid top tool")
	}

	// Get the bottom tool
	bottomTool, err := h.findToolByID(tools, bottomToolID)
	if err != nil {
		return errors.NewBadRequestError(err, "invalid bottom tool")
	}

	// Check if the tools are compatible with each other
	if topTool.Format.String() != bottomTool.Format.String() {
		return errors.NewBadRequestError(nil, "tools are not compatible")
	}

	// Get current tools for press
	currentTools, dberr := h.registry.Tools.ListByPress(&pressNumber)
	if dberr != nil {
		return errors.HandlerError(dberr, "get current tools for press")
	}

	// Create final cycle entries for current tools with total cycles
	for _, tool := range currentTools {
		cycle := models.NewCycle(pressNumber, tool.ID, tool.Position, totalCycles, user.TelegramID)

		_, dberr = h.registry.PressCycles.Add(cycle, user)
		if dberr != nil {
			return errors.HandlerError(dberr, "create final cycle for tool %d", tool.ID)
		}
	}

	// Unassign current tools from press
	for _, tool := range currentTools {
		dberr := h.registry.Tools.UpdatePress(tool.ID, nil, user)
		if dberr != nil {
			return errors.HandlerError(dberr, "unassign tool %d", tool.ID)
		}
	}

	// Assign new tools to press
	newTools := []*models.Tool{topTool, bottomTool}
	for _, tool := range newTools {
		dberr := h.registry.Tools.UpdatePress(tool.ID, &pressNumber, user)
		if dberr != nil {
			return errors.HandlerError(dberr, "assign tool %d to press", tool.ID)
		}
	}

	// Create feed entry
	title := fmt.Sprintf("Werkzeugwechsel Presse %d", pressNumber)
	content := fmt.Sprintf(
		"Umbau abgeschlossen für Presse %d.\nEingebautes Oberteil: %s\nEingebautes Unterteil: %s\nGesamtzyklen: %d",
		pressNumber, topTool.String(), bottomTool.String(), totalCycles,
	)

	// Create feed entry
	if _, err := h.registry.Feeds.AddSimple(title, content, user.TelegramID); err != nil {
		slog.Warn("Failed to create feed", "press", pressNumber, "error", err)
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) findToolByID(tools []*models.Tool, toolID models.ToolID) (*models.Tool, error) {
	for _, tool := range tools {
		if tool.ID == toolID {
			return tool, nil
		}
	}
	return nil, fmt.Errorf("tool not found: %d", toolID)
}
