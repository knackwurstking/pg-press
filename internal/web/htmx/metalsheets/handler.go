package metalsheets

import (
	"errors"
	"net/http"

	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/web/handlers"
	"github.com/knackwurstking/pgpress/internal/web/helpers"

	"github.com/labstack/echo/v4"
)

type MetalSheets struct {
	*handlers.BaseHandler
}

func NewMetalSheets(db *database.DB) *MetalSheets {
	return &MetalSheets{
		BaseHandler: handlers.NewBaseHandler(db, logger.HTMXHandlerMetalSheets()),
	}
}

func (h *MetalSheets) RegisterRoutes(e *echo.Echo) {
	helpers.RegisterEchoRoutes(
		e,
		[]*helpers.EchoRoute{
			helpers.NewEchoRoute(http.MethodGet, "/htmx/metal-sheets/edit",
				h.HandleEditGET),
		},
	)
}

func (h *MetalSheets) HandleEditGET(c echo.Context) error {
	// TODO: Open edit dialog for adding or editing a metal sheet entry

	return errors.New("under construction")
}
