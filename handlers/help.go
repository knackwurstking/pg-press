package handlers

import (
	"net/http"

	"github.com/knackwurstking/pg-press/components"
	"github.com/knackwurstking/pg-press/logger"
	"github.com/knackwurstking/pg-press/services"
	"github.com/knackwurstking/pg-press/utils"
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
