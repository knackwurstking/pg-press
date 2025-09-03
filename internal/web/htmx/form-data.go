package htmx

import "github.com/knackwurstking/pgpress/internal/models"

type ToolEditFormData struct {
	Position models.Position     // Position form field name "position"
	Format   models.ToolFormat   // Format form field names "width" and "height"
	Type     string              // Type form field name "type"
	Code     string              // Code form field name "code"
	Press    *models.PressNumber // Press form field name "press-selection"
}

type CycleEditFormData struct {
	TotalCycles int64               // TotalCycles form field name "total_cycles"
	PressNumber *models.PressNumber // PressNumber form field name "press_number", numbers lower 0 means disabled
}
