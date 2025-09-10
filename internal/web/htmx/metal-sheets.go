package htmx

import (
	"errors"
	"net/http"

	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/web/helpers"

	"github.com/labstack/echo/v4"
)

type MetalSheets struct {
	DB *database.DB
}

func (h *MetalSheets) RegisterRoutes(e *echo.Echo) {
	helpers.RegisterEchoRoutes(
		e,
		[]*helpers.EchoRoute{
			helpers.NewEchoRoute(http.MethodGet, "/htmx/metal-sheets/edit", h.handleEditGET),
		},
	)
}

func (h *MetalSheets) handleEditGET(c echo.Context) error {
	// TODO: Open edit dialog for adding or editing a metal sheet entry

	return errors.New("under construction")
}
