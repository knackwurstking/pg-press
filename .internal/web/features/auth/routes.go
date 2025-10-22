package auth

import (
	"github.com/knackwurstking/pgpress/internal/services"
)

type Routes struct {
	handler *Handler
}

func NewRoutes(db *services.Registry) *Routes {
	return &Routes{
		handler: NewHandler(db),
	}
}
