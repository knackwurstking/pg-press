package htmxhandler

import (
	"errors"

	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/labstack/echo/v4"
)

type Tools struct {
	DB *database.DB
}

func (h *Tools) RegisterRoutes(e *echo.Echo) {
	e.GET(serverPathPrefix+"/htmx/tools/list-all", h.handleListAll)
	e.GET(serverPathPrefix+"/htmx/tools/edit", h.handleEdit)
}

func (h *Tools) handleListAll(c echo.Context) error {
	// TODO: Implement list all handler

	return errors.New("under construction")
}

func (h *Tools) handleEdit(c echo.Context) error {
	// TODO: Implement edit handler

	return errors.New("under construction")
}
