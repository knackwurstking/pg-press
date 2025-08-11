package htmxhandler

import (
	"github.com/knackwurstking/pgpress/internal/database"
)

type Base struct {
	DB               *database.DB
	ServerPathPrefix string
}
