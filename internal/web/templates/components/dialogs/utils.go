package dialogs

import "github.com/knackwurstking/pgpress/internal/database/models"

func isPress(p *models.PressNumber, v int) bool {
	if p == nil {
		return false
	}

	return int(*p) == v
}
