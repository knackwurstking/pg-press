package editor

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
			// Editor page
			helpers.NewEchoRoute(http.MethodGet, "/editor",
				r.handler.GetEditorPage),

			// Save content
			helpers.NewEchoRoute(http.MethodPost, "/editor/save",
				r.handler.PostSaveContent),
		},
	)
}
