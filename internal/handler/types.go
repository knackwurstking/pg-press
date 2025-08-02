package handler

import (
	"embed"

	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/htmxhandler"
)

type Base struct {
	DB               *database.DB
	ServerPathPrefix string
	Templates        embed.FS
}

func (b *Base) NewHTMX(prefix string) *htmxhandler.Base {
	return &htmxhandler.Base{
		DB:               b.DB,
		ServerPathPrefix: b.ServerPathPrefix + prefix,
		Templates:        b.Templates,
	}
}
