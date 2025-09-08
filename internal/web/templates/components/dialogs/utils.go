package dialogs

import pressmodels "github.com/knackwurstking/pgpress/internal/database/models/press"

func isPress(p *pressmodels.PressNumber, v int) bool {
	if p == nil {
		return false
	}

	return int(*p) == v
}
