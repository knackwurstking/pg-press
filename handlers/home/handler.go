package home

import (
	"net/http"

	"github.com/knackwurstking/pg-press/env"
	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/handlers/home/templates"
	"github.com/knackwurstking/pg-press/services"
	"github.com/knackwurstking/ui"

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
	ui.RegisterEchoRoutes(e, env.ServerPathPrefix, []*ui.EchoRoute{
		ui.NewEchoRoute(http.MethodGet, path, h.GetHomePage),
	})
}

func (h *Handler) GetHomePage(c echo.Context) error {
	page := templates.Page()
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.Handler(err, "render home page")
	}
	return nil
}
