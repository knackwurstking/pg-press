package htmx

import (
	"time"

	presscyclemodels "github.com/knackwurstking/pgpress/internal/database/models/presscycle"
	toolmodels "github.com/knackwurstking/pgpress/internal/database/models/tool"
)

type ToolEditFormData struct {
	Position toolmodels.Position           // Position form field name "position"
	Format   toolmodels.ToolFormat         // Format form field names "width" and "height"
	Type     string                        // Type form field name "type"
	Code     string                        // Code form field name "code"
	Press    *presscyclemodels.PressNumber // Press form field name "press-selection"
}

type CycleEditFormData struct {
	TotalCycles int64 // TotalCycles form field name "total_cycles"
	PressNumber *presscyclemodels.PressNumber
	Date        time.Time // OriginalDate form field name "original_date"
}
