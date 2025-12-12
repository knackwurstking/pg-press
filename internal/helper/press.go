package helper

import (
	"github.com/knackwurstking/pg-press/internal/common"
	"github.com/knackwurstking/pg-press/internal/shared"
)

const (
	SQLGetPressNumberForTool string = `
		SELECT id, slot_up, slot_down FROM presses WHERE slot_up = :slot_up OR slot_down = :slot_down;
	`
)

func GetPressNumberForTool(db *common.DB, toolID shared.EntityID) (shared.PressNumber, shared.Slot) {
	var pressNumber shared.PressNumber = -1
	var slotUp, slotDown shared.EntityID

	_ = db.Press.Press.DB().QueryRow(SQLGetPressNumberForTool, toolID, toolID).Scan(&pressNumber, &slotUp, &slotDown)

	switch toolID {
	case slotUp:
		return pressNumber, shared.SlotUpper
	case slotDown:
		return pressNumber, shared.SlotLower
	default:
		return pressNumber, shared.SlotUnknown
	}
}
