package htmx

import (
	"net/http"

	database "github.com/knackwurstking/pgpress/internal/database/core"
	"github.com/knackwurstking/pgpress/internal/web/webhelpers"

	"github.com/labstack/echo/v4"
)

type TroubleReports struct {
	DB *database.DB
}

func (h *TroubleReports) RegisterRoutes(e *echo.Echo) {
	webhelpers.RegisterEchoRoutes(
		e,
		[]*webhelpers.EchoRoute{
			// Dialog edit routes
			webhelpers.NewEchoRoute(http.MethodGet, "/htmx/trouble-reports/dialog-edit", func(c echo.Context) error {
				return h.handleGetDialogEdit(c, nil)
			}),
			webhelpers.NewEchoRoute(http.MethodPost, "/htmx/trouble-reports/dialog-edit", h.handlePostDialogEdit),
			webhelpers.NewEchoRoute(http.MethodPut, "/htmx/trouble-reports/dialog-edit", h.handlePutDialogEdit),

			// Data routes
			webhelpers.NewEchoRoute(http.MethodGet, "/htmx/trouble-reports/data", h.handleGetData),
			webhelpers.NewEchoRoute(http.MethodDelete, "/htmx/trouble-reports/data", h.handleDeleteData),

			// Attachments preview routes
			webhelpers.NewEchoRoute(http.MethodGet, "/htmx/trouble-reports/attachments-preview", h.handleGetAttachmentsPreview),

			// Modifications routes
			webhelpers.NewEchoRoute(http.MethodGet, "/htmx/trouble-reports/modifications/:id", func(c echo.Context) error {
				return h.handleGetModifications(c, nil)
			}),
			webhelpers.NewEchoRoute(http.MethodPost, "/htmx/trouble-reports/modifications/:id", h.handlePostModifications),
		},
	)
}
