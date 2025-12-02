package help

import (
	"net/http"

	"github.com/knackwurstking/pg-press/env"
	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/handlers/help/templates"
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
		ui.NewEchoRoute(http.MethodGet, path+"/markdown", h.GetMarkdown),
	})
}

func (h *Handler) GetMarkdown(c echo.Context) error {
	page := templates.MarkdownPage()
	if err := page.Render(c.Request().Context(), c.Response().Writer); err != nil {
		return errors.Handler(err, "render help page")
	}
	return nil
}
