package handlers

import (
	"net/http"

	"github.com/knackwurstking/pgpress/components"
	"github.com/knackwurstking/pgpress/logger"
	"github.com/knackwurstking/pgpress/services"
	"github.com/knackwurstking/pgpress/utils"
	"github.com/labstack/echo/v4"
)

type Help struct {
	*Base
}

func NewHelp(db *services.Registry) *Help {
	return &Help{
		Base: NewBase(db, logger.NewComponentLogger("Help")),
	}
}

func (h *Help) RegisterRoutes(e *echo.Echo) {
	utils.RegisterEchoRoutes(e, []*utils.EchoRoute{
		utils.NewEchoRoute(http.MethodGet, "/help/markdown", h.GetMarkdown),
	})
}

func (h *Help) GetMarkdown(c echo.Context) error {
	page := components.PageHelpMarkdown()
	if err := page.Render(c.Request().Context(), c.Response().Writer); err != nil {
		return HandleError(err, "render help page failed")
	}
	return nil
}
