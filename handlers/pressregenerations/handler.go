package pressregenerations

import (
	"net/http"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/handlers/pressregenerations/templates"
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
			h.GetRegenerationPage,
		),

		utils.NewEchoRoute(
			http.MethodPost,
			path+"/:press",
			h.HxAddRegeneration,
		),

		utils.NewEchoRoute(
			http.MethodPost,
			path+"/:press/delete",
			h.HxDeleteRegeneration,
		),
	})
}

func (h *Handler) GetRegenerationPage(c echo.Context) error {
	press, eerr := h.parseParamPress(c)
	if eerr != nil {
		return eerr
	}

	if err := templates.Page(templates.PageProps{
		Press: press,
	}).Render(c.Request().Context(), c.Response()); err != nil {
		return errors.Handler(err, "render press page template")
	}

	return nil
}

func (h *Handler) HxAddRegeneration(c echo.Context) (err error) {
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

	utils.SetHXRedirect(c, utils.UrlPress(press).Page)

	return nil
}

func (h *Handler) HxDeleteRegeneration(c echo.Context) (err error) {
	id, err := utils.ParseQueryInt64(c, "id") // PressRegenerationID
	if err != nil {
		return errors.BadRequest(err, "missing id query")
	}

	if err := h.registry.PressRegenerations.Delete(models.PressRegenerationID(id)); err != nil {
		return errors.Handler(err, "delete press regeneration")
	}

	return nil
}
