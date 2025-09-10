package htmx

import (
	"net/http"

	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/web/helpers"

	"github.com/labstack/echo/v4"
)

type TroubleReports struct {
	DB *database.DB
}

func (h *TroubleReports) RegisterRoutes(e *echo.Echo) {
	helpers.RegisterEchoRoutes(
		e,
		[]*helpers.EchoRoute{
			// Dialog edit routes
			helpers.NewEchoRoute(http.MethodGet, "/htmx/trouble-reports/dialog-edit", func(c echo.Context) error {
				return h.handleGetDialogEdit(c, nil)
			}),
			helpers.NewEchoRoute(http.MethodPost, "/htmx/trouble-reports/dialog-edit", h.handlePostDialogEdit),
			helpers.NewEchoRoute(http.MethodPut, "/htmx/trouble-reports/dialog-edit", h.handlePutDialogEdit),

			// Data routes
			helpers.NewEchoRoute(http.MethodGet, "/htmx/trouble-reports/data", h.handleGetData),
			helpers.NewEchoRoute(http.MethodDelete, "/htmx/trouble-reports/data", h.handleDeleteData),

			// Attachments preview routes
			helpers.NewEchoRoute(http.MethodGet, "/htmx/trouble-reports/attachments-preview", h.handleGetAttachmentsPreview),

			// Modifications routes
			helpers.NewEchoRoute(http.MethodGet, "/htmx/trouble-reports/modifications/:id", func(c echo.Context) error {
				return h.handleGetModifications(c, nil)
			}),
			helpers.NewEchoRoute(http.MethodPost, "/htmx/trouble-reports/modifications/:id", h.handlePostModifications),
		},
	)
}
