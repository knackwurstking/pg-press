package pressregenerations

import (
	"fmt"
	"net/http"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/handlers/pressregenerations/components"
	"github.com/knackwurstking/pg-press/models"
	"github.com/knackwurstking/pg-press/services"
	"github.com/knackwurstking/pg-press/utils"

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

func (h *Handler) RegisterRoutes(e *echo.Echo) {
	utils.RegisterEchoRoutes(e, []*utils.EchoRoute{
		// Press regeneration page
		utils.NewEchoRoute(
			http.MethodGet,
			"/tools/press/:press/regenerations",
			h.GetPressRegenerationsPage,
		),

		utils.NewEchoRoute(
			http.MethodPost,
			"/hx/tools/press/:press/regenerations",
			h.HxPostPressRegenerationsPage,
		),
	})
}

func (h *Handler) GetPressRegenerationsPage(c echo.Context) error {
	press, eerr := h.parseParamPress(c)
	if eerr != nil {
		return eerr
	}

	if err := components.PagePressRegenerations(components.PagePressRegenerationsProps{
		Press: press,
	}).Render(c.Request().Context(), c.Response()); err != nil {
		return errors.Handler(err, "render press page template")
	}

	return nil
}

func (h *Handler) HxPostPressRegenerationsPage(c echo.Context) (err error) {
	var (
		press models.PressNumber
		eerr  *echo.HTTPError
	)

	press, eerr = h.parseParamPress(c)
	if eerr != nil {
		return eerr
	}

	var (
		data *models.PressRegeneration
	)

	data, eerr = ParseFormRegenerationsPage(c, press)
	if eerr != nil {
		return eerr
	}

	if _, err = h.registry.PressRegenerations.Add(data); err != nil {
		return errors.Handler(err, "add press regeneration")
	}

	utils.SetHXRedirect(c, fmt.Sprintf("/tools/press/%d", press))

	return nil
}

func (h *Handler) parseParamPress(c echo.Context) (models.PressNumber, *echo.HTTPError) {
	pressNum, err := utils.ParseParamInt8(c, "press")
	if err != nil {
		return -1, errors.BadRequest(err, "invalid or missing press parameter")
	}

	press := models.PressNumber(pressNum)
	if !models.IsValidPressNumber(&press) {
		return -1, errors.BadRequest(err, "invalid press number")
	}

	return press, nil
}
