package htmxhandler

import (
	"embed"

	"github.com/knackwurstking/pg-vis/internal/database"
)

type Base struct {
	DB               *database.DB
	ServerPathPrefix string
	Templates        embed.FS
}
