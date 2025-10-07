package troublereports

import (
	"net/http"

	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/web/shared/helpers"

	"github.com/labstack/echo/v4"
)

type Routes struct {
	handler *Handler
}

func NewRoutes(db *database.DB) *Routes {
	return &Routes{
		handler: NewHandler(db),
	}
}

func (r *Routes) RegisterRoutes(e *echo.Echo) {
	helpers.RegisterEchoRoutes(
		e,
		[]*helpers.EchoRoute{
			// Pages
			helpers.NewEchoRoute(http.MethodGet, "/trouble-reports",
				r.handler.GetPage),

			helpers.NewEchoRoute(http.MethodGet, "/trouble-reports/share-pdf",
				r.handler.GetSharePDF),

			helpers.NewEchoRoute(http.MethodGet, "/trouble-reports/attachment",
				r.handler.GetAttachment),

			helpers.NewEchoRoute(http.MethodGet, "/trouble-reports/modifications/:id",
				r.handler.GetModificationsForID),

			// HTMX
			// Data routes
			helpers.NewEchoRoute(http.MethodGet, "/htmx/trouble-reports/data",
				r.handler.HTMXGetData,
			),

			helpers.NewEchoRoute(http.MethodDelete, "/htmx/trouble-reports/data",
				r.handler.HTMXDeleteTroubleReport,
			),

			// Attachments preview routes
			helpers.NewEchoRoute(http.MethodGet, "/htmx/trouble-reports/attachments-preview",
				r.handler.HTMXGetAttachmentsPreview),

			// Rollback route
			helpers.NewEchoRoute(http.MethodPost, "/htmx/trouble-reports/rollback",
				r.handler.HTMXPostRollback),
		},
	)
}
