package home

import (
	"net/http"

	"github.com/knackwurstking/pg-press/handlers/home/components"
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
		utils.NewEchoRoute(http.MethodGet, "", h.GetHomePage),
	})
}

func (h *Handler) GetHomePage(c echo.Context) error {
	page := components.PageHome()
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return utils.HandleError(err, "failed to render home page")
	}
	return nil
}
