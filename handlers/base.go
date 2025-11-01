package handlers

import (
	"github.com/knackwurstking/pg-press/logger"
	"github.com/knackwurstking/pg-press/services"
)

type Base struct {
	Registry *services.Registry
	Log      *logger.Logger
}

func NewBase(r *services.Registry, l *logger.Logger) *Base {
	return &Base{
		Registry: r,
		Log:      l,
	}
}
