package home

import (
	"github.com/knackwurstking/pgpress/internal/services"
	"github.com/knackwurstking/pgpress/internal/web/features/home/templates"
	"github.com/knackwurstking/pgpress/internal/web/shared/base"
	"github.com/knackwurstking/pgpress/pkg/logger"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	*base.Handler
}

func NewHandler(db *services.Registry) *Handler {
	return &Handler{
		Handler: base.NewHandler(db, logger.NewComponentLogger("Home")),
	}
}

func (h *Handler) GetHomePage(c echo.Context) error {
	page := templates.HomePage()
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c,
			"failed to render home page: "+err.Error())
	}
	return nil
}
