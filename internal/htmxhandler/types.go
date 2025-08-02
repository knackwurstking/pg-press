package htmxhandler

import (
	"embed"

	"github.com/knackwurstking/pgpress/internal/database"
)

type Base struct {
	DB               *database.DB
	ServerPathPrefix string
	Templates        embed.FS
}
