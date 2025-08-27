package htmxhandler

import "github.com/knackwurstking/pgpress/internal/database"

type ToolEditFormData struct {
	Position database.Position
	Format   database.ToolFormat
	Type     string
	Code     string
}

type CycleEditFormData struct {
	TotalCycles int64
}
