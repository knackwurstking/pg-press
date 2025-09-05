package htmx

import (
	"time"

	"github.com/knackwurstking/pgpress/internal/database/models"
)

type ToolEditFormData struct {
	Position models.Position     // Position form field name "position"
	Format   models.ToolFormat   // Format form field names "width" and "height"
	Type     string              // Type form field name "type"
	Code     string              // Code form field name "code"
	Press    *models.PressNumber // Press form field name "press-selection"
}

type CycleEditFormData struct {
	TotalCycles int64 // TotalCycles form field name "total_cycles"
	PressNumber *models.PressNumber
	Date        time.Time // OriginalDate form field name "original_date"
}
