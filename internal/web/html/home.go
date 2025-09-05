package html

import (
	"net/http"

	"github.com/knackwurstking/pgpress/internal/logger"
	pageshome "github.com/knackwurstking/pgpress/internal/web/templates/pages/home"
	"github.com/knackwurstking/pgpress/internal/web/webhelpers"
	"github.com/labstack/echo/v4"
)

type Home struct{}

func (h *Home) RegisterRoutes(e *echo.Echo) {
	webhelpers.RegisterEchoRoutes(
		e,
		[]*webhelpers.EchoRoute{
			webhelpers.NewEchoRoute(http.MethodGet, "", h.handleHome),
		},
	)
}

// handleHome handles the home page request.
func (h *Home) handleHome(c echo.Context) error {
	logger.HandlerHome().Debug("Rendering home page")

	page := pageshome.Page()
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		logger.HandlerHome().Error("Failed to render home page: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render home page: "+err.Error())
	}
	return nil
}
