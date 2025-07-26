package handler

import (
	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/internal/constants"
	"github.com/knackwurstking/pg-vis/internal/utils"
)

type Home struct {
	*Base
}

// NewHome creates a new home handler.
func NewHome(base *Base) *Home {
	return &Home{base}
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
