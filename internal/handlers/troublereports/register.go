package troublereports

import (
	"net/http"

	"github.com/knackwurstking/pg-press/internal/env"

	ui "github.com/knackwurstking/ui/ui-templ"

	"github.com/labstack/echo/v4"
)

func Register(e *echo.Echo, path string) {
	ui.RegisterEchoRoutes(
		e,
		env.ServerPathPrefix,
		[]*ui.EchoRoute{
			// Pages
			ui.NewEchoRoute(http.MethodGet, path, GetPage),
			ui.NewEchoRoute(http.MethodGet, path+"/share-pdf", GetSharePDF),
			ui.NewEchoRoute(http.MethodGet, path+"/attachment", GetAttachment),
			ui.NewEchoRoute(http.MethodGet, path+"/modifications/:id", GetModificationsForID),

			// HTMX
			ui.NewEchoRoute(http.MethodGet, path+"/data", HTMXGetData),
			ui.NewEchoRoute(http.MethodDelete, path+"/data", HTMXDeleteTroubleReport),
			ui.NewEchoRoute(http.MethodGet, path+"/attachments-preview", HTMXGetAttachmentsPreview),
			ui.NewEchoRoute(http.MethodPost, path+"/rollback", HTMXPostRollback),
		},
	)
}
