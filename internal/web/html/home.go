package html

import (
	"net/http"

	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/web/handlers"
	"github.com/knackwurstking/pgpress/internal/web/helpers"
	"github.com/knackwurstking/pgpress/internal/web/templates/homepage"
	"github.com/knackwurstking/pgpress/pkg/logger"
	"github.com/labstack/echo/v4"
)

type Home struct {
	*handlers.BaseHandler
}

func NewHome(db *database.DB, logger *logger.Logger) *Tools {
	return &Tools{
		BaseHandler: handlers.NewBaseHandler(db, logger),
	}
}

// TODO: Refactor

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
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render home page: "+err.Error())
	}

	return nil
}
