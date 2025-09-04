package htmx

import (
	"net/http"

	database "github.com/knackwurstking/pgpress/internal/database/core"
	"github.com/knackwurstking/pgpress/internal/utils"
	"github.com/labstack/echo/v4"
)

type TroubleReports struct {
	DB *database.DB
}

func (h *TroubleReports) RegisterRoutes(e *echo.Echo) {
	utils.RegisterEchoRoutes(
		e,
		[]*utils.EchoRoute{
			// Dialog edit routes
			utils.NewEchoRoute(http.MethodGet, "/htmx/trouble-reports/dialog-edit", func(c echo.Context) error {
				return h.handleGetDialogEdit(c, nil)
			}),
			utils.NewEchoRoute(http.MethodPost, "/htmx/trouble-reports/dialog-edit", h.handlePostDialogEdit),
			utils.NewEchoRoute(http.MethodPut, "/htmx/trouble-reports/dialog-edit", h.handlePutDialogEdit),

			// Data routes
			utils.NewEchoRoute(http.MethodGet, "/htmx/trouble-reports/data", h.handleGetData),
			utils.NewEchoRoute(http.MethodDelete, "/htmx/trouble-reports/data", h.handleDeleteData),

			// Attachments preview routes
			utils.NewEchoRoute(http.MethodGet, "/htmx/trouble-reports/attachments-preview", h.handleGetAttachmentsPreview),

			// Modifications routes
			utils.NewEchoRoute(http.MethodGet, "/htmx/trouble-reports/modifications/:id", func(c echo.Context) error {
				return h.handleGetModifications(c, nil)
			}),
			utils.NewEchoRoute(http.MethodPost, "/htmx/trouble-reports/modifications/:id", h.handlePostModifications),
		},
	)
}
