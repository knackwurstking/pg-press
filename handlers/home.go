package handlers

import (
	"net/http"

	"github.com/knackwurstking/pgpress/components"
	"github.com/knackwurstking/pgpress/internal/services"
	"github.com/knackwurstking/pgpress/logger"
	"github.com/knackwurstking/pgpress/utils"

	"github.com/labstack/echo/v4"
)

type HomeHandler struct {
	*baseHandler
}

func NewHomeHandler(db *services.Registry) *HomeHandler {
	return &HomeHandler{
		baseHandler: newBaseHandler(db, logger.NewComponentLogger("Home")),
	}
}

func (h *HomeHandler) RegisterRoutes(e *echo.Echo) {
	utils.RegisterEchoRoutes(
		e,
		[]*utils.EchoRoute{
			// Pages
			utils.NewEchoRoute(http.MethodGet, "", h.GetHomePage),
		},
	)
}

func (h *HomeHandler) GetHomePage(c echo.Context) error {
	page := components.PageHome()
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return h.renderInternalError(c,
			"failed to render home page: "+err.Error())
	}
	return nil
}
