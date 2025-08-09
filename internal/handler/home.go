package handler

import (
	"net/http"

	"github.com/knackwurstking/pgpress/internal/templates/pages"
	"github.com/labstack/echo/v4"
)

type Home struct {
	*Base
}

func (h *Home) RegisterRoutes(e *echo.Echo) {
	e.GET(h.ServerPathPrefix+"/", h.handleHome)
}

// handleHome handles the home page request.
func (h *Home) handleHome(c echo.Context) error {
	page := pages.Home()
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render home page: "+err.Error())
	}
	return nil
}
