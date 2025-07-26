// Package home provides HTTP route handlers for the home page.
package home

import (
	"embed"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/knackwurstking/pg-vis/routes/constants"
	"github.com/knackwurstking/pg-vis/routes/internal/utils"
)

// Handler handles home page HTTP requests.
type Handler struct {
	db               *pgvis.DB
	serverPathPrefix string
	templates        embed.FS
}

// NewHandler creates a new home page handler.
func NewHandler(db *pgvis.DB, serverPathPrefix string, templates embed.FS) *Handler {
	return &Handler{
		db:               db,
		serverPathPrefix: serverPathPrefix,
		templates:        templates,
	}
}

// RegisterRoutes registers all home page routes.
func (h *Handler) RegisterRoutes(e *echo.Echo) {
	e.GET(h.serverPathPrefix+"/", h.handleHome)
}

// handleHome handles the home page request.
func (h *Handler) handleHome(c echo.Context) error {
	return utils.HandleTemplate(
		c,
		nil,
		h.templates,
		constants.HomePageTemplates,
	)
}
