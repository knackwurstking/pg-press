package home

import (
	"net/http"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/handlers/home/templates"
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
		utils.NewEchoRoute(http.MethodGet, path, h.GetHomePage),
	})
}

func (h *Handler) GetHomePage(c echo.Context) error {
	page := templates.PageHome()
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.Handler(err, "render home page")
	}
	return nil
}
