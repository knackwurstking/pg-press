package htmx

import (
	"time"

	"github.com/knackwurstking/pgpress/pkg/models"
)

type CycleEditFormData struct {
	TotalCycles  int64 // TotalCycles form field name "total_cycles"
	PressNumber  *models.PressNumber
	Date         time.Time // OriginalDate form field name "original_date"
	Regenerating bool
}
