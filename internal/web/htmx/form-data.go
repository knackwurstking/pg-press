package htmx

import (
	"time"

	pressmodels "github.com/knackwurstking/pgpress/internal/database/models/press"
	toolmodels "github.com/knackwurstking/pgpress/internal/database/models/tool"
)

type ToolEditFormData struct {
	Position toolmodels.Position      // Position form field name "position"
	Format   toolmodels.Format        // Format form field names "width" and "height"
	Type     string                   // Type form field name "type"
	Code     string                   // Code form field name "code"
	Press    *pressmodels.PressNumber // Press form field name "press-selection"
}

type CycleEditFormData struct {
	TotalCycles  int64 // TotalCycles form field name "total_cycles"
	PressNumber  *pressmodels.PressNumber
	Date         time.Time // OriginalDate form field name "original_date"
	Regenerating bool
}
