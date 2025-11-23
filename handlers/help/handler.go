package help

import (
	"net/http"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/handlers/help/components"
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
		utils.NewEchoRoute(http.MethodGet, "/help/markdown", h.GetMarkdown),
	})
}

func (h *Handler) GetMarkdown(c echo.Context) error {
	page := components.PageHelpMarkdown()
	if err := page.Render(c.Request().Context(), c.Response().Writer); err != nil {
		return errors.Handler(err, "render help page")
	}
	return nil
}
