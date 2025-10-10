package notes

import (
	"net/http"

	"github.com/knackwurstking/pgpress/internal/services"
	"github.com/knackwurstking/pgpress/internal/web/shared/helpers"

	"github.com/labstack/echo/v4"
)

type Routes struct {
	handler *Handler
}

func NewRoutes(db *services.Registry) *Routes {
	return &Routes{
		handler: NewHandler(db),
	}
}

func (r *Routes) RegisterRoutes(e *echo.Echo) {
	helpers.RegisterEchoRoutes(
		e,
		[]*helpers.EchoRoute{
			// Notes page
			helpers.NewEchoRoute(http.MethodGet, "/notes",
				r.handler.GetNotesPage),

			// HTMX routes for notes dialog editing
			helpers.NewEchoRoute(http.MethodGet, "/htmx/notes/edit",
				r.handler.HTMXGetEditNoteDialog),

			helpers.NewEchoRoute(http.MethodPost, "/htmx/notes/edit",
				r.handler.HTMXPostEditNoteDialog),

			helpers.NewEchoRoute(http.MethodPut, "/htmx/notes/edit",
				r.handler.HTMXPutEditNoteDialog),

			helpers.NewEchoRoute(http.MethodDelete, "/htmx/notes/delete",
				r.handler.HTMXDeleteNote),
		},
	)
}
