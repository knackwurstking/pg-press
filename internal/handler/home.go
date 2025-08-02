package handler

import (
	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pgpress/internal/constants"
	"github.com/knackwurstking/pgpress/internal/utils"
)

type Home struct {
	*Base
}

func (h *Home) RegisterRoutes(e *echo.Echo) {
	e.GET(h.ServerPathPrefix+"/", h.handleHome)
}

// handleHome handles the home page request.
func (h *Home) handleHome(c echo.Context) error {
	return utils.HandleTemplate(
		c,
		nil,
		h.Templates,
		constants.HomePageTemplates,
	)
}
