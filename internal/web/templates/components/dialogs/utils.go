package dialogs

import toolmodels "github.com/knackwurstking/pgpress/internal/database/models/tool"

func isPress(p *toolmodels.PressNumber, v int) bool {
	if p == nil {
		return false
	}

	return int(*p) == v
}
