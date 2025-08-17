package htmxhandler

import (
	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/labstack/echo/v4"
)

type Tools struct {
	DB *database.DB
}

func (h *Tools) RegisterRoutes(e *echo.Echo) {
	// TODO ...
}
