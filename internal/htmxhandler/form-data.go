package htmxhandler

import "github.com/knackwurstking/pgpress/internal/database"

type ToolEditFormData struct {
	Position database.Position   // Position form field name "position"
	Format   database.ToolFormat // Format form field names "width" and "height"
	Type     string              // Type form field name "type"
	Code     string              // Code form field name "code"
}

type CycleEditFormData struct {
	TotalCycles int64                 // TotalCycles form field name "total_cycles"
	PressNumber *database.PressNumber // PressNumber form field name "press_number", numbers lower 0 means disabled
}
