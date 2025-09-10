package dialogs

import "github.com/knackwurstking/pgpress/pkg/models"

func isPress(p *models.PressNumber, v int) bool {
	if p == nil {
		return false
	}

	return int(*p) == v
}
