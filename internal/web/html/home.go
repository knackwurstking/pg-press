package html

import (
	"net/http"

	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/web/helpers"
	"github.com/knackwurstking/pgpress/internal/web/templates/homepage"
	"github.com/labstack/echo/v4"
)

type Home struct{}

func (h *Home) RegisterRoutes(e *echo.Echo) {
	helpers.RegisterEchoRoutes(
		e,
		[]*helpers.EchoRoute{
			helpers.NewEchoRoute(http.MethodGet, "", h.handleHome),
		},
	)
}

// handleHome handles the home page request.
func (h *Home) handleHome(c echo.Context) error {
	page := homepage.Page()
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		logger.HandlerHome().Error("Failed to render home page: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render home page: "+err.Error())
	}
	return nil
}
