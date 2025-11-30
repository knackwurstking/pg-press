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

func (h *Handler) RegisterRoutes(e *echo.Echo, path string) {
	utils.RegisterEchoRoutes(e, []*utils.EchoRoute{
		// Press regeneration page
		utils.NewEchoRoute(
			http.MethodGet,
			path+"/:press",
			h.GetPressRegenerationsPage,
		),

		utils.NewEchoRoute(
			http.MethodPost,
			path+"/:press",
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
