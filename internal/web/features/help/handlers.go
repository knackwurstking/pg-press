package help

import (
	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/web/features/help/templates"
	"github.com/knackwurstking/pgpress/internal/web/shared/handlers"
	"github.com/knackwurstking/pgpress/pkg/logger"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	*handlers.BaseHandler
}

func NewHandler(db *database.DB) *Handler {
	return &Handler{
		BaseHandler: handlers.NewBaseHandler(db, logger.NewComponentLogger("Help")),
	}
}

func (h *Handler) GetMarkdown(c echo.Context) error {
	page := templates.MarkdownHelpPage()
	if err := page.Render(c.Request().Context(), c.Response().Writer); err != nil {
		return h.RenderInternalError(c, "render help page failed: "+err.Error())
	}
	return nil
}
