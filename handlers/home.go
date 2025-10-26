package handlers

import (
	"net/http"

	"github.com/knackwurstking/pg-press/components"
	"github.com/knackwurstking/pg-press/logger"
	"github.com/knackwurstking/pg-press/services"
	"github.com/knackwurstking/pg-press/utils"

	"github.com/labstack/echo/v4"
)

type Home struct {
	*Base
}

func NewHome(r *services.Registry) *Home {
	return &Home{
		Base: NewBase(r, logger.NewComponentLogger("Home")),
	}
}

func (h *Home) RegisterRoutes(e *echo.Echo) {
	utils.RegisterEchoRoutes(e, []*utils.EchoRoute{
		utils.NewEchoRoute(http.MethodGet, "", h.GetHomePage),
	})
}

func (h *Home) GetHomePage(c echo.Context) error {
	page := components.PageHome()
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return HandleError(err, "failed to render home page")
	}
	return nil
}
