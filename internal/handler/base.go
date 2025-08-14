package handler

import (
	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/htmxhandler"
)

type Base struct {
	DB               *database.DB
	ServerPathPrefix string
}

func (b *Base) NewHTMXHandlerBase(prefix string) *htmxhandler.Base {
	return &htmxhandler.Base{
		DB:               b.DB,
		ServerPathPrefix: b.ServerPathPrefix + prefix,
	}
}
