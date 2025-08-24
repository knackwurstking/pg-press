package handler

import (
	"net/http"

	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/templates/pages"
	"github.com/labstack/echo/v4"
)

type Home struct{}

func (h *Home) RegisterRoutes(e *echo.Echo) {
	e.GET(serverPathPrefix, h.handleHome)
	e.GET(serverPathPrefix+"/", h.handleHome)
}

// handleHome handles the home page request.
func (h *Home) handleHome(c echo.Context) error {
	logger.HandlerHome().Debug("Rendering home page")

	page := pages.HomePage()
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		logger.HandlerHome().Error("Failed to render home page: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render home page: "+err.Error())
	}
	return nil
}
