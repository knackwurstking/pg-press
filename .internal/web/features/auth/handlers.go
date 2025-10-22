package auth

import (
	"github.com/knackwurstking/pgpress/internal/services"
	"github.com/knackwurstking/pgpress/internal/web/shared/base"
	"github.com/knackwurstking/pgpress/pkg/logger"
)

type Handler struct {
	*base.Handler
}

func NewHandler(db *services.Registry) *Handler {
	return &Handler{
		Handler: base.NewHandler(db, logger.NewComponentLogger("Auth")),
	}
}
