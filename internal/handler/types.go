package handler

import (
	"embed"

	"github.com/knackwurstking/pg-vis/internal/database"
	"github.com/knackwurstking/pg-vis/internal/htmxhandler"
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
