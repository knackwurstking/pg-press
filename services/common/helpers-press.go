package common

import (
	"github.com/knackwurstking/pg-press/services/shared"
)

func GetPressNumberForTool(db DB, toolID shared.EntityID) (shared.PressNumber, shared.Slot, error) {
	var pressNumber shared.PressNumber
	var slotUp, slotDown shared.EntityID

	err := db.Press.Press.DB().QueryRow(
		"SELECT id, slot_up, slot_down FROM presses WHERE slot_up = ? OR slot_down = ?",
		toolID, toolID,
	).Scan(&pressNumber, &slotUp, &slotDown)
	if err != nil {
		return -1, shared.SlotUnknown, err
	}

	if slotUp == toolID {
		return pressNumber, shared.SlotUp, nil
	}
	if slotDown == toolID {
		return pressNumber, shared.SlotDown, nil
	}
	return pressNumber, shared.SlotUnknown, nil
}
